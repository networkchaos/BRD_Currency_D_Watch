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

/ Snapshot is one recorded reading at a point in time
type Snapshot struct {
	RecordedAt time.Time    `json:"recorded_at"`
	Rates      fetcher.Rates `json:"rates"`
}

// StateFile holds the rolling history we commit to the repo
type StateFile struct {
	UpdatedAt string     `json:"updated_at"`
	History   []Snapshot `json:"history"`
}

// Windows defines how far back we look for each comparison window.
// GitHub Actions runs every 5 min, so we keep enough history to cover 24h.
var Windows = map[string]time.Duration{
	"5min":  5 * time.Minute,
	"1hour": 1 * time.Hour,
	"6hour": 6 * time.Hour,
	"24hour": 24 * time.Hour,
}

// maxHistory caps how many snapshots we store (24h at 5-min intervals = 288)
const maxHistory = 300

// Load reads state.json and returns the full history.
// Returns an empty StateFile (not an error) if the file doesn't exist yet.
func Load(path string) (*StateFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First run — return empty state, not an error
			return &StateFile{}, nil
		}
		return nil, fmt.Errorf("could not read state file: %w", err)
	}

	// Empty file = treat as first run (temp files start empty)
	if len(data) == 0 {
		return &StateFile{}, nil
	}

	var sf StateFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("could not parse state file: %w", err)
	}

	return &sf, nil
}

// Save appends the current rates as a new snapshot and writes state.json.
// It prunes old snapshots beyond maxHistory to keep the file small.
func Save(path string, rates fetcher.Rates) error {
	if len(rates) == 0 {
		return fmt.Errorf("cannot save empty rates")
	}

	// Load existing history first so we can append to it
	sf, err := Load(path)
	if err != nil {
		return err
	}

	// Append new snapshot
	sf.History = append(sf.History, Snapshot{
		RecordedAt: time.Now().UTC(),
		Rates:      rates,
	})

	// Prune old snapshots
	if len(sf.History) > maxHistory {
		sf.History = sf.History[len(sf.History)-maxHistory:]
	}

	sf.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write state file: %w", err)
	}

	return nil
}

// SnapshotsBefore returns all snapshots recorded before the given duration ago.
// Used by the detector to find "what was the rate N minutes/hours ago?"
func (sf *StateFile) SnapshotsBefore(window time.Duration) []Snapshot {
	cutoff := time.Now().UTC().Add(-window)
	var result []Snapshot
	for _, s := range sf.History {
		if s.RecordedAt.Before(cutoff) {
			result = append(result, s)
		}
	}
	return result
}

// ClosestBefore returns the snapshot closest to (but before) the given duration ago.
// Returns nil if no snapshot exists for that window yet.
func (sf *StateFile) ClosestBefore(window time.Duration) *Snapshot {
	snapshots := sf.SnapshotsBefore(window)
	if len(snapshots) == 0 {
		return nil
	}
	// The last one in the list is the closest to our window cutoff
	s := snapshots[len(snapshots)-1]
	return &s
}
