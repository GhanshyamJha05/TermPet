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

	// Test cd to tempDir
	m.handleCommand("cd " + tempDir)
	if m.cwd != tempDir {
		t.Errorf("Expected cwd to be %s, got %s", tempDir, m.cwd)
	}
}

func TestCommandHistoryNavigation(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)

	// Simulate typing and entering commands
	m.terminalInput.SetValue("cmd1")
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)

	m.terminalInput.SetValue("cmd2")
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)

	if len(m.cmdHistory) != 2 {
		t.Fatalf("Expected 2 commands in history, got %d", len(m.cmdHistory))
	}

	// Navigate Up
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.terminalInput.Value() != "cmd2" {
		t.Errorf("Expected 'cmd2' on first Up, got %s", m.terminalInput.Value())
	}

	// Navigate Up again
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newM.(Model)
	if m.terminalInput.Value() != "cmd1" {
		t.Errorf("Expected 'cmd1' on second Up, got %s", m.terminalInput.Value())
	}

	// Navigate Down
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(Model)
	if m.terminalInput.Value() != "cmd2" {
		t.Errorf("Expected 'cmd2' on Down, got %s", m.terminalInput.Value())
	}
}

func TestScrollbackNavigation(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState)
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newM.(Model)

	// Add 50 history lines
	for i := 0; i < 50; i++ {
		m.addHistoryLine(string(rune('A' + (i % 26))))
	}

	// Test PgUp
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = newM.(Model)
	if m.scrollOffset != 5 {
		t.Errorf("Expected scrollOffset 5 after PgUp, got %d", m.scrollOffset)
	}

	// Test PgDn
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

	// Type prefix ":fe" and press Tab
	m.terminalInput.SetValue(":fe")
	m.terminalInput.SetCursor(3)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)

	if m.terminalInput.Value() != ":feed" {
		t.Errorf("Expected completion ':feed', got %s", m.terminalInput.Value())
	}
}
