package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCalculateOfflineHunger(t *testing.T) {
	tests := []struct {
		name          string
		currentHunger float64
		elapsedHours  float64
		wantMin       float64
		wantMax       float64
	}{
		{
			name:          "Zero elapsed hours",
			currentHunger: 20.0,
			elapsedHours:  0,
			wantMin:       20.0,
			wantMax:       20.0,
		},
		{
			name:          "One hour elapsed adds 10%",
			currentHunger: 20.0,
			elapsedHours:  1.0,
			wantMin:       30.0,
			wantMax:       30.0,
		},
		{
			name:          "Long offline time capped at 100%",
			currentHunger: 50.0,
			elapsedHours:  24.0,
			wantMin:       100.0,
			wantMax:       100.0,
		},
		{
			name:          "Negative elapsed time (clock skew) does not decrease hunger",
			currentHunger: 30.0,
			elapsedHours:  -2.0,
			wantMin:       30.0,
			wantMax:       30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateOfflineHunger(tt.currentHunger, tt.elapsedHours)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateOfflineHunger(%v, %v) = %v, want between %v and %v",
					tt.currentHunger, tt.elapsedHours, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestPersistenceSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "state.json")

	initialState := &PetState{
		Name:     "Mochi",
		Hunger:   45.5,
		LastSeen: time.Now(),
		Species:  "cat",
	}

	// Test Save
	err := SaveState(statePath, initialState)
	if err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist", statePath)
	}

	// Test Load immediately (elapsed ~ 0)
	loaded, err := LoadState(statePath, "DefaultName")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.Name != "Mochi" {
		t.Errorf("Expected pet name Mochi, got %q", loaded.Name)
	}

	if loaded.Hunger < 45.0 || loaded.Hunger > 46.0 {
		t.Errorf("Expected hunger near 45.5, got %v", loaded.Hunger)
	}
}

func TestSpriteDimensions(t *testing.T) {
	sprite := " /\\_/\\\n( o.o )\n > ^ < "
	w, h := GetSpriteDimensions(sprite)
	if h != 3 {
		t.Errorf("Expected height 3, got %d", h)
	}
	if w < 7 {
		t.Errorf("Expected width >= 7, got %d", w)
	}
}
