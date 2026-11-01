package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"time"
)

// PetState stores persistent state on disk.
type PetState struct {
	Name     string    `json:"name"`
	Hunger   float64   `json:"hunger"` // 0.0 to 100.0
	LastSeen time.Time `json:"last_seen"`
	Species  string    `json:"species,omitempty"`
}

// HungerRatePerHour determines how much hunger accumulates per hour offline.
// E.g., 10% per hour means full hunger takes 10 hours.
const HungerRatePerHour = 10.0

// DefaultStatePath returns ~/.termpet/state.json
func DefaultStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".termpet", "state.json"), nil
}

// LoadState reads state from path, calculating hunger elapsed since LastSeen.
func LoadState(path string, defaultName string) (*PetState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &PetState{
				Name:     defaultName,
				Hunger:   20.0,
				LastSeen: time.Now(),
			}, nil
		}
		return nil, err
	}

	var state PetState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	if state.Name == "" {
		state.Name = defaultName
	}

	// Calculate accumulated hunger while offline
	now := time.Now()
	if !state.LastSeen.IsZero() && now.After(state.LastSeen) {
		elapsedHours := now.Sub(state.LastSeen).Hours()
		state.Hunger = CalculateOfflineHunger(state.Hunger, elapsedHours)
	}
	state.LastSeen = now

	return &state, nil
}

// CalculateOfflineHunger adds accumulated hunger capped between 0 and 100.
func CalculateOfflineHunger(currentHunger float64, elapsedHours float64) float64 {
	if elapsedHours < 0 {
		elapsedHours = 0
	}
	newHunger := currentHunger + (elapsedHours * HungerRatePerHour)
	if newHunger > 100.0 {
		return 100.0
	}
	if newHunger < 0.0 {
		return 0.0
	}
	return math.Round(newHunger*10) / 10
}

// SaveState saves the current state to disk.
func SaveState(path string, state *PetState) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	state.LastSeen = time.Now()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
