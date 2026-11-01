package main

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelLifecycleAndActions(t *testing.T) {
	tempDir := t.TempDir()
	statePath := tempDir + "/test_state.json"

	petState := &PetState{
		Name:      "Draco",
		Hunger:    80.0, // Hungry initially
		Happiness: 50.0,
		Energy:    60.0,
		LastSeen:  time.Now(),
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
	if len(m.particles) == 0 {
		t.Errorf("Feeding should spawn heart/sparkle particles")
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

	// Test terminal command execution (echo hello) -> should trigger StateVictory
	m.handleCommand("echo hello")
	if m.state != StateVictory {
		t.Errorf("Successful command should trigger StateVictory, got %v", m.state)
	}

	// Test View rendering
	rendered := m.View()
	if !strings.Contains(rendered, "Sparky") {
		t.Errorf("View output should contain pet name 'Sparky'")
	}
}

func TestTypingReactionAndCuriosity(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, Happiness: 80.0, Energy: 80.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)
	m.state = StateIdle

	// Force seed typing reaction directly
	m.handleTypingReaction()
	// Either stayed or turned curious based on roll; test calling directly works
	if m.state == StateCurious {
		if m.targetY == 0 {
			t.Errorf("Expected targetY to be updated near prompt")
		}
	}
}

func TestPoopMechanicAndCleaning(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, Happiness: 80.0, Energy: 80.0, HasPoop: true, PoopCount: 2, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)

	if len(m.poops) != 2 {
		t.Fatalf("Expected 2 poops initialized, got %d", len(m.poops))
	}

	// Test :clean command
	m.handleCommand(":clean")
	if len(m.poops) != 0 {
		t.Errorf("Expected 0 poops after cleaning, got %d", len(m.poops))
	}
	if m.petState.Cleanliness != 100.0 {
		t.Errorf("Expected cleanliness 100 after :clean, got %f", m.petState.Cleanliness)
	}
}

func TestCommandFailureReaction(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, Happiness: 80.0, Energy: 80.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)

	// Simulate command failure
	m.onCommandResult(false)
	if m.state != StateAngry {
		t.Errorf("Expected StateAngry on command failure, got %v", m.state)
	}
}

func TestSoundBellToggle(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, Happiness: 80.0, Energy: 80.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)

	orig := m.soundEnabled
	m.handleCommand(":bell")
	if m.soundEnabled == orig {
		t.Errorf("Expected :bell to toggle soundEnabled")
	}
}

func TestDirectoryPersistence(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	tempDir := t.TempDir()
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)

	m.handleCommand("cd " + tempDir)
	if m.cwd != tempDir {
		t.Errorf("Expected cwd to be %s, got %s", tempDir, m.cwd)
	}
}

func TestCommandHistoryNavigation(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)

	m.terminalInput.SetValue("cmd1")
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)

	m.terminalInput.SetValue("cmd2")
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)

	if len(m.cmdHistory) != 2 {
		t.Fatalf("Expected 2 commands in history, got %d", len(m.cmdHistory))
	}

	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.terminalInput.Value() != "cmd2" {
		t.Errorf("Expected 'cmd2' on first Up, got %s", m.terminalInput.Value())
	}

	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.terminalInput.Value() != "cmd1" {
		t.Errorf("Expected 'cmd1' on second Up, got %s", m.terminalInput.Value())
	}
}

func TestScrollbackNavigation(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newM.(Model)

	for i := 0; i < 50; i++ {
		m.addHistoryLine(string(rune('A' + (i % 26))))
	}

	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = newM.(Model)
	if m.scrollOffset != 5 {
		t.Errorf("Expected scrollOffset 5 after PgUp, got %d", m.scrollOffset)
	}

	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = newM.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset 0 after PgDn, got %d", m.scrollOffset)
	}
}

func TestTabCompletion(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)

	m.terminalInput.SetValue(":fe")
	m.terminalInput.SetCursor(3)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)

	if m.terminalInput.Value() != ":feed" {
		t.Errorf("Expected completion ':feed', got %s", m.terminalInput.Value())
	}
}
