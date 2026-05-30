package detector

import (
	"testing"

	"github.com/networkchaos/BRD_Currency/internal/fetcher"
)

func TestAnalyse_BuySignal(t *testing.T) {
	// IRR weakens: rate goes UP (more IRR needed per USD = bad for IRR holders)
	previous := fetcher.Rates{"IRR": 40000.0}
	current := fetcher.Rates{"IRR": 44000.0} // +10% → BUY signal

	signals := Analyse(previous, current)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != SignalBuy {
		t.Errorf("expected BUY signal, got %s", signals[0].Type)
	}
	if signals[0].Currency != "IRR" {
		t.Errorf("expected IRR, got %s", signals[0].Currency)
	}
}

func TestAnalyse_SellSignal(t *testing.T) {
	// IRR strengthens: rate goes DOWN (fewer IRR needed per USD = good for IRR holders)
	previous := fetcher.Rates{"IRR": 44000.0}
	current := fetcher.Rates{"IRR": 40000.0} // -9.09% → SELL signal

	signals := Analyse(previous, current)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != SignalSell {
		t.Errorf("expected SELL signal, got %s", signals[0].Type)
	}
}

func TestAnalyse_NoSignal_SmallChange(t *testing.T) {
	// Only 1% change — below threshold, no alert
	previous := fetcher.Rates{"IRR": 40000.0}
	current := fetcher.Rates{"IRR": 40400.0} // +1%

	signals := Analyse(previous, current)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for small change, got %d", len(signals))
	}
}

func TestAnalyse_NewCurrencySkipped(t *testing.T) {
	// ZWL appears in current but not in previous — no baseline, skip it
	previous := fetcher.Rates{"IRR": 40000.0}
	current := fetcher.Rates{"IRR": 40000.0, "ZWL": 361.9}

	signals := Analyse(previous, current)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals (no baseline for ZWL), got %d", len(signals))
	}
}

func TestAnalyse_MultipleSignals(t *testing.T) {
	previous := fetcher.Rates{
		"IRR": 40000.0,
		"ZWL": 300.0,
		"VES": 30.0,
	}
	current := fetcher.Rates{
		"IRR": 44000.0, // +10%  → BUY
		"ZWL": 270.0,   // -10%  → SELL
		"VES": 30.3,    // +1%   → nothing
	}

	signals := Analyse(previous, current)

	if len(signals) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(signals))
	}
}

func TestChangePercent(t *testing.T) {
	tests := []struct {
		name     string
		old, new float64
		expected float64
	}{
		{"10% increase", 100, 110, 10.0},
		{"50% decrease", 100, 50, -50.0},
		{"no change", 100, 100, 0.0},
		{"zero old rate", 0, 100, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChangePercent(tt.old, tt.new)
			if got != tt.expected {
				t.Errorf("ChangePercent(%v, %v) = %v, want %v", tt.old, tt.new, got, tt.expected)
			}
		})
	}
}
