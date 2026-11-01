package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelLifecycleAndActions(t *testing.T) {
	tempDir := t.TempDir()
	statePath := tempDir + "/test_state.json"

	petState := &PetState{
		Name:     "Draco",
		Hunger:   80.0, // Hungry initially
		LastSeen: time.Now(),
	}

	m := NewModel("dragon", false, statePath, petState)

	// Simulate window size
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newM.(Model)

	if m.width != 80 || m.height != 24 {
		t.Fatalf("Expected dimensions 80x24, got %dx%d", m.width, m.height)
	}

	// Hunger >= 75 should trigger hungry state
	m.updateBehavior()
	if m.state != StateHungry {
		t.Errorf("Expected StateHungry, got %v", m.state)
	}

	// Test feeding command ":feed"
	m.handleCommand(":feed")
	if m.petState.Hunger > 56.0 {
		t.Errorf("Feeding should decrease hunger by 25%%, current hunger: %f", m.petState.Hunger)
	}
	if m.state != StateHappy {
		t.Errorf("Feeding should trigger StateHappy, got %v", m.state)
	}

	// Test petting command ":pet"
	m.handleCommand(":pet")
	if m.state != StateHappy {
		t.Errorf("Petting should trigger StateHappy, got %v", m.state)
	}

	// Test renaming command ":name Sparky"
	m.handleCommand(":name Sparky")
	if m.petState.Name != "Sparky" {
		t.Errorf("Expected pet name Sparky, got %q", m.petState.Name)
	}

	// Test terminal command execution (e.g. echo hello)
	m.handleCommand("echo hello")
	found := false
	for _, line := range m.history {
		if strings.Contains(line, "hello") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected command output 'hello' in history, got %v", m.history)
	}

	// Test View rendering
	rendered := m.View()
	if !strings.Contains(rendered, "Sparky") {
		t.Errorf("View output should contain pet name 'Sparky'")
	}
	if !strings.Contains(rendered, ":feed") {
		t.Errorf("View output should contain command hints")
	}
}
