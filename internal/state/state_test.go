package state

import (
	"os"
	"testing"

	"github.com/networkchaos/BRD_Currency/internal/fetcher"
)

func TestSaveAndLoad(t *testing.T) {
	// Use a temp file so tests don't pollute the repo
	tmp, err := os.CreateTemp("", "state-*.json")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	rates := fetcher.Rates{
		"IRR": 42000.0,
		"ZWL": 361.9,
	}

	// Save
	if err := Save(tmp.Name(), rates); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load
	loaded, err := Load(tmp.Name())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify values round-trip correctly
	if loaded["IRR"] != 42000.0 {
		t.Errorf("expected IRR=42000, got %v", loaded["IRR"])
	}
	if loaded["ZWL"] != 361.9 {
		t.Errorf("expected ZWL=361.9, got %v", loaded["ZWL"])
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/state.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSave_EmptyRates(t *testing.T) {
	err := Save("/tmp/test-empty.json", fetcher.Rates{})
	if err == nil {
		t.Fatal("expected error for empty rates, got nil")
	}
}

func TestLoad_CorruptFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "corrupt-*.json")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	// Write garbage JSON
	tmp.WriteString("not valid json {{{")
	tmp.Close()

	_, err = Load(tmp.Name())
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}
