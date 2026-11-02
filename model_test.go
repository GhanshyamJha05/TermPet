package main

import (
	"os"
	"path/filepath"
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
		Theme:     "dracula",
		LastSeen:  time.Now(),
	}

	m := NewModel("dragon", false, statePath, petState, DefaultConfig())

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

	// Test terminal command execution (echo hello) -> should trigger StateVictory and award XP
	m.handleCommand("echo hello")
	if m.state != StateVictory {
		t.Errorf("Successful command should trigger StateVictory, got %v", m.state)
	}
	if m.petState.XP <= 0 {
		t.Errorf("Expected XP to increase from command execution, got %d", m.petState.XP)
	}

	// Test View rendering
	rendered := m.View()
	if !strings.Contains(rendered, "Sparky") {
		t.Errorf("View output should contain pet name 'Sparky'")
	}
}

func TestThemeCustomization(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, Theme: "dracula", LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newM.(Model)

	// Switch theme to cyber
	m.handleCommand(":theme cyber")
	if m.petState.Theme != "cyber" {
		t.Errorf("Expected theme 'cyber', got %s", m.petState.Theme)
	}

	// Render view with cyber theme
	rendered := m.View()
	if !strings.Contains(rendered, "cyber") {
		t.Errorf("Expected theme name in status bar view")
	}
}

func TestDragonEvolutionAndStages(t *testing.T) {
	lvl1, stage1 := CalculateLevelAndStage(0)
	if lvl1 != 1 || stage1 != StageBaby {
		t.Errorf("0 XP should be Level 1 Baby, got Lv.%d %s", lvl1, stage1)
	}

	lvl5, stage5 := CalculateLevelAndStage(450)
	if lvl5 != 5 || stage5 != StageYoung {
		t.Errorf("450 XP should be Level 5 Juvenile, got Lv.%d %s", lvl5, stage5)
	}

	lvl10, stage10 := CalculateLevelAndStage(950)
	if lvl10 != 10 || stage10 != StageAdult {
		t.Errorf("950 XP should be Level 10 Ancient, got Lv.%d %s", lvl10, stage10)
	}

	// Test frame rendering changes across stages
	babyFrame := GetFrame("dragon", false, StateIdle, StageBaby, DirRight, 0)
	juvFrame := GetFrame("dragon", false, StateIdle, StageYoung, DirRight, 0)
	ancientFrame := GetFrame("dragon", false, StateIdle, StageAdult, DirRight, 0)

	if babyFrame == juvFrame || juvFrame == ancientFrame {
		t.Errorf("Frames should differ between Baby, Juvenile, and Ancient evolutionary stages")
	}
}

func TestCustomSpriteLoading(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("No home dir found")
	}
	customDir := filepath.Join(home, ".termpet", "sprites")
	_ = os.MkdirAll(customDir, 0755)

	customJSON := `{
		"name": "phoenix",
		"normal": {
			"idle": {
				"left": ["~<v>~"],
				"right": ["~<v>~"]
			}
		},
		"compact": {
			"idle": {
				"left": ["<v>"],
				"right": ["<v>"]
			}
		}
	}`
	testFile := filepath.Join(customDir, "phoenix.json")
	_ = os.WriteFile(testFile, []byte(customJSON), 0644)
	defer os.Remove(testFile)

	LoadCustomSprites()
	if _, ok := SpeciesMap["phoenix"]; !ok {
		t.Errorf("Expected custom species 'phoenix' to be loaded into SpeciesMap")
	}
}

func TestTypingReactionAndCuriosity(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, Happiness: 80.0, Energy: 80.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())
	m.state = StateIdle

	m.handleTypingReaction()
	if m.state == StateCurious {
		if m.targetY == 0 {
			t.Errorf("Expected targetY to be updated near prompt")
		}
	}
}

func TestPoopMechanicAndCleaning(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, Happiness: 80.0, Energy: 80.0, HasPoop: true, PoopCount: 2, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())

	if len(m.poops) != 2 {
		t.Fatalf("Expected 2 poops initialized, got %d", len(m.poops))
	}

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
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())

	m.onCommandResult(false)
	if m.state != StateAngry {
		t.Errorf("Expected StateAngry on command failure, got %v", m.state)
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
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())

	m.handleCommand("cd " + tempDir)
	if m.cwd != tempDir {
		t.Errorf("Expected cwd to be %s, got %s", tempDir, m.cwd)
	}
}

func TestCommandHistoryNavigation(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())

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
}

func TestScrollbackNavigation(t *testing.T) {
	tempDir := t.TempDir()
	petState := &PetState{Name: "Draco", Hunger: 20.0, LastSeen: time.Now()}
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())
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
	m := NewModel("dragon", false, tempDir+"/state.json", petState, DefaultConfig())

	m.terminalInput.SetValue(":the")
	m.terminalInput.SetCursor(4)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newM.(Model)

	if !strings.HasPrefix(m.terminalInput.Value(), ":theme") {
		t.Errorf("Expected completion ':theme', got %s", m.terminalInput.Value())
	}
}
