// Package state persists exchange rates between runs using a local JSON file.
// Since GitHub Actions gives us a fresh environment each time, we commit
// state.json to the repo and let the workflow pull/push it.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/networkchaos/BRD_Currency/internal/fetcher"
)

// StateFile is what we write to disk
type StateFile struct {
	UpdatedAt string        `json:"updated_at"`
	Rates     fetcher.Rates `json:"rates"`
}

// Load reads a previously saved state file and returns the rates.
// Returns an error if the file doesn't exist yet (first run).
func Load(path string) (fetcher.Rates, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read state file: %w", err)
	}

	var sf StateFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("could not parse state file: %w", err)
	}

	if len(sf.Rates) == 0 {
		return nil, fmt.Errorf("state file contains no rates")
	}

	return sf.Rates, nil
}

// Save writes the current rates to a JSON file with a timestamp.
func Save(path string, rates fetcher.Rates) error {
	if len(rates) == 0 {
		return fmt.Errorf("cannot save empty rates")
	}

	sf := StateFile{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Rates:     rates,
	}

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write state file: %w", err)
	}

	return nil
}
