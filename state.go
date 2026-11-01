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
	Name       string    `json:"name"`
	Hunger     float64   `json:"hunger"`     // 0.0 to 100.0 (higher = hungrier)
	Happiness  float64   `json:"happiness"`  // 0.0 to 100.0 (higher = happier)
	Energy     float64   `json:"energy"`     // 0.0 to 100.0 (lower = sleepy)
	Cleanliness float64  `json:"cleanliness"`// 0.0 to 100.0 (lower = dirty/poop)
	HasPoop    bool      `json:"has_poop"`
	PoopCount  int       `json:"poop_count"`
	LastSeen   time.Time `json:"last_seen"`
	Species    string    `json:"species,omitempty"`
}

// HungerRatePerHour determines how much hunger accumulates per hour offline.
const HungerRatePerHour = 10.0

// DefaultStatePath returns ~/.termpet/state.json
func DefaultStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".termpet", "state.json"), nil
}

// LoadState reads state from path, calculating hunger & decay elapsed since LastSeen.
func LoadState(path string, defaultName string) (*PetState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &PetState{
				Name:        defaultName,
				Hunger:      20.0,
				Happiness:   80.0,
				Energy:      90.0,
				Cleanliness: 100.0,
				HasPoop:     false,
				PoopCount:   0,
				LastSeen:    time.Now(),
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
	if state.Happiness == 0 && state.Energy == 0 {
		state.Happiness = 80.0
		state.Energy = 90.0
		state.Cleanliness = 100.0
	}

	// Calculate accumulated changes while offline
	now := time.Now()
	if !state.LastSeen.IsZero() && now.After(state.LastSeen) {
		elapsedHours := now.Sub(state.LastSeen).Hours()
		state.Hunger = CalculateOfflineHunger(state.Hunger, elapsedHours)
		
		// Offline happiness drops slowly if neglected
		state.Happiness -= elapsedHours * 5.0
		if state.Happiness < 10.0 {
			state.Happiness = 10.0
		}
		// Offline energy recovers if pet was resting
		state.Energy += elapsedHours * 15.0
		if state.Energy > 100.0 {
			state.Energy = 100.0
		}
		// Poop accumulation if away for over 4 hours
		if elapsedHours >= 4.0 && !state.HasPoop {
			state.HasPoop = true
			state.PoopCount = int(math.Min(float64(int(elapsedHours/4.0)), 5.0))
			state.Cleanliness = math.Max(10.0, 100.0-float64(state.PoopCount*20))
		}
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
