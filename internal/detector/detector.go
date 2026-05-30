// Package detector analyses exchange rate changes and produces buy/sell signals.
//
// How it works:
//   - A currency DROPS vs USD  → more units needed to buy 1 USD → currency weakened
//     e.g. IRR goes 40000 → 80000 means IRR lost 50% of its value → BUY signal
//   - A currency RISES vs USD  → fewer units needed to buy 1 USD → currency strengthened
//     e.g. IRR goes 80000 → 40000 means IRR doubled in value → SELL signal

package detector

import (
	"fmt"
	"log"

	"github.com/networkchaos/BRD_Currency/internal/fetcher"
)

//triggers threshold

const (
	BuyThreshold  = 5.0 //dorp -5% vs usd trigger for buy signal
	SellThreshold = 5.0 //rise +5  trigger for sell signal
)

// signaltype tells me what action to take
type SignalType string

const (
	SignalBuy  SignalType = "BUY"
	SignalSell SignalType = "SELL"
	SignalNone SignalType = "NONE"
)

//signal detects currency change and produces a message to send to telegram

type Signal struct {
	Currency  string
	Type      SignalType
	OldRate   float64
	NewRate   float64
	Changepct float64 //+ve for sell -ve for buy
	Message   string  //telegram message to send
}

//Analyse compares previous rates to current rate and returns a list of signals to send to telegram
//both maps are currency -> units-per-USD (higher = weaker currency)

func Analyse(previous, current fetcher.Rates) []Signal {
	//returns a slice of signals to send to telegram
	var signals []Signal

	for currency, newRate := range current {
		oldRate, exists := previous[currency]
		if !exists {
			log.Printf("currency %s not found in previous rates, skipping analysis", currency)
			continue
		}
		if oldRate == 0 {
			log.Printf("old rate for currency %s is zero, skipping analysis to avoid division by zero", currency)
			continue
		}
		//changepct > 0 meanns the rate went UP (more units per USD = currency weakened) → SELL signal
		//changepct < 0 means the rate went DOWN (fewer units per USD = currency strengthened) → BUY signal
		changepct := ((newRate - oldRate) / oldRate) * 100

		sig := Signal{
			Currency:  currency,
			OldRate:   oldRate,
			NewRate:   newRate,
			Changepct: changepct,
		}

		switch {
		case changepct >= BuyThreshold:
			sig.Type = SignalBuy
			sig.Message = formatBuyMessage(currency, oldRate, newRate, changepct)
			signals = append(signals, sig)

		case changepct <= -SellThreshold:
			sig.Type = SignalSell
			sig.Message = formatSellMessage(currency, oldRate, newRate, changepct)
			signals = append(signals, sig)

		}
	}
	return signals

}

//changepct calculates pct change between two rate
//exported for testing

func ChangePct(oldRate, newRate float64) float64 {
	if oldRate == 0 {
		return 0 //avoid division by zero
	}
	return ((newRate - oldRate) / oldRate) * 100
}
func formatBuyMessage(curreny string, oldRate, newRate, changepct float64) string {
	return fmt.Sprint(
		" BUY SIGNAL : %s\n"+
			"currency dropped %.2f%%\n"+
			"old rate: 1 USD = %.4f%s\n"+
			"New rate: 1 USD = %.4f%s\n"+
			"----------------------\n"+
			" hey cool aid %s has weakened. Consider buying if you believe it will recover.",
		currency, changepct, oldRate, currency, newRate, currency, currency,
	)
}
func formatSellMessage(currency string, oldRate, newRate, changePct float64) string {
	return fmt.Sprintf(
		"🔴 SELL SIGNAL: %s\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"📈 Currency rose %.2f%%\n"+
			"Old rate: 1 USD = %.4f %s\n"+
			"New rate: 1 USD = %.4f %s\n"+
			"━━━━━━━━━━━━━━━━━━\n"+
			"💡 %s has strengthened. If you hold it, now may be a good time to sell.",
		currency, -changePct, oldRate, currency, newRate, currency, currency,
	)
}
