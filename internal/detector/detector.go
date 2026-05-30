// Package detector analyses exchange rate changes and produces buy/sell signals.
//added to analyse across many windows and detect more complex patterns in the future

//the windows and their threshold :
// 5min -> 0.3% change (fast spike - very sensitive)
//1 hrs -> 1.0% change (sustained momentum)
// 6hrs -> 3.0% (string trnd)
//24 hrs -> 5.0% (major shift)

//
// How it works:
// rate = units of currency per 1 Usd
//rate goes up -> currency weakens (more units needed to buy 1 USD) → SELL signal
//rate goes down -> currency strengthens (fewer units needed to buy 1 USD) → BUY signal

package detector

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/networkchaos/BRD_Currency/internal/fetcher"
	"github.com/networkchaos/BRD_Currency/internal/state"
)

// WindoConfig defines a time window and how much change triggers an alert
type WindowConfig struct {
	Name          string
	Duration      time.Duration
	BuyThreshold  float64 //% rate increase that trigger an alert
	SellThreshold float64 //% rate decrease that trigger an alert
}

// Windows is the ordered lists of time windows we  check most senstive first
var Windows = []WindowConfig{
	{Name: "5min", Duration: 5 * time.Minute, BuyThreshold: 0.3, SellThreshold: 0.3},
	{Name: "1hour", Duration: 1 * time.Hour, BuyThreshold: 1.0, SellThreshold: 1.0},
	{Name: "6hour", Duration: 6 * time.Hour, BuyThreshold: 3.0, SellThreshold: 3.0},
	{Name: "24hour", Duration: 24 * time.Hour, BuyThreshold: 5.0, SellThreshold: 5.0},
}

// signaltype tells you what action to consider
type SignalType string

const (
	SignalBuy  SignalType = "BUY"
	SignalSell SignalType = "SELL"
)

// signal represent a detected opportunity for one currency on one time window
type Signal struct {
	Currency  string
	Type      SignalType
	Window    string
	OldRate   float64
	NewRate   float64
	ChangePct float64 // % change +ve for buy (weakening), -ve for sell (strengthening)
	Message   string
}

// AnalyseAll checks all time windows against the stored history and
// returns the most actionable signal per currency (highest urgency wins)
func AnalyseAll(sf *state.StateFile, current fetcher.Rates) []Signal {
	// dedupKey → best signal for that currency (we only send one alert per currency)
	best := map[string]Signal{}

	for _, window := range Windows {
		snap := sf.ClosestBefore(window.Duration)
		if snap == nil {
			// Not enough history yet for this window — skip
			continue
		}

		signals := analyseWindow(snap.Rates, current, window)
		for _, sig := range signals {
			existing, seen := best[sig.Currency]
			if !seen {
				best[sig.Currency] = sig
				continue
			}
			// Prefer the signal with the larger absolute change (more urgent)
			if abs(sig.ChangePct) > abs(existing.ChangePct) {
				best[sig.Currency] = sig
			}
		}
	}

	// Return as a sorted slice (deterministic order for tests)
	result := make([]Signal, 0, len(best))
	for _, sig := range best {
		result = append(result, sig)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Currency < result[j].Currency
	})
	return result
}

// analyseWindow compares one snapshot against current rates for one window.
func analyseWindow(previous, current fetcher.Rates, window WindowConfig) []Signal {
	var signals []Signal

	for currency, newRate := range current {
		oldRate, exists := previous[currency]
		if !exists || oldRate == 0 {
			continue
		}

		// changePct > 0 → rate went UP → currency dropped in value → BUY
		// changePct < 0 → rate went DOWN → currency gained value → SELL
		changePct := ChangePercent(oldRate, newRate)

		sig := Signal{
			Currency:  currency,
			Window:    window.Name,
			OldRate:   oldRate,
			NewRate:   newRate,
			ChangePct: changePct,
		}

		switch {
		case changePct >= window.BuyThreshold:
			sig.Type = SignalBuy
			sig.Message = formatBuyMessage(sig)
			signals = append(signals, sig)

		case changePct <= -window.SellThreshold:
			sig.Type = SignalSell
			sig.Message = formatSellMessage(sig)
			signals = append(signals, sig)
		}
	}

	return signals
}

// ChangePercent calculates the percentage change from oldRate to newRate.
func ChangePercent(oldRate, newRate float64) float64 {
	if oldRate == 0 {
		return 0
	}
	return ((newRate - oldRate) / oldRate) * 100
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func urgencyLabel(changePct float64, window string) string {
	a := abs(changePct)
	switch {
	case window == "5min" && a >= 0.3:
		return "⚡ FAST SPIKE"
	case window == "1hour" && a >= 1.0:
		return "🔥 SUSTAINED MOVE"
	case window == "6hour" && a >= 3.0:
		return "📊 STRONG TREND"
	default:
		return "🌊 MAJOR MOVE"
	}
}

func formatBuyMessage(s Signal) string {
	lines := []string{
		fmt.Sprintf("🟢 BUY SIGNAL — %s", s.Currency),
		strings.Repeat("━", 22),
		fmt.Sprintf("%s over %s", urgencyLabel(s.ChangePct, s.Window), s.Window),
		fmt.Sprintf("📉 Dropped %.4f%%", s.ChangePct),
		fmt.Sprintf("Was:  1 USD = %.6f %s", s.OldRate, s.Currency),
		fmt.Sprintf("Now:  1 USD = %.6f %s", s.NewRate, s.Currency),
		strings.Repeat("━", 22),
		fmt.Sprintf("💡 %s is weakening. If you believe it recovers, this may be a buying opportunity.", s.Currency),
		"⚠️  Always do your own research before trading.",
	}
	return strings.Join(lines, "\n")
}

func formatSellMessage(s Signal) string {
	lines := []string{
		fmt.Sprintf("🔴 SELL SIGNAL — %s", s.Currency),
		strings.Repeat("━", 22),
		fmt.Sprintf("%s over %s", urgencyLabel(s.ChangePct, s.Window), s.Window),
		fmt.Sprintf("📈 Gained %.4f%%", -s.ChangePct),
		fmt.Sprintf("Was:  1 USD = %.6f %s", s.OldRate, s.Currency),
		fmt.Sprintf("Now:  1 USD = %.6f %s", s.NewRate, s.Currency),
		strings.Repeat("━", 22),
		fmt.Sprintf("💡 %s is strengthening. If you're holding it, consider whether now is the time to sell.", s.Currency),
		"⚠️  Always do your own research before trading.",
	}
	return strings.Join(lines, "\n")
}
