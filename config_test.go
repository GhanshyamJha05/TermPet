package main

import (
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "config.json")

	cfg, err := LoadConfig(nonExistent)
	if err == nil {
		t.Errorf("Expected error loading non-existent config, got nil")
	}
	if cfg.DefaultTheme != "dracula" {
		t.Errorf("Expected default theme dracula, got %q", cfg.DefaultTheme)
	}
	if cfg.DefaultSpecies != "dragon" {
		t.Errorf("Expected default species dragon, got %q", cfg.DefaultSpecies)
	}
	if !cfg.SoundEnabled {
		t.Errorf("Expected SoundEnabled to be true by default")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	customCfg := &AppConfig{
		DefaultTheme:   "fire",
		DefaultSpecies: "cat",
		TickRateMs:     800,
		SoundEnabled:   false,
		MaxScrollback:  1000,
	}

	if err := SaveConfig(configPath, customCfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.DefaultTheme != "fire" {
		t.Errorf("Expected fire theme, got %s", loaded.DefaultTheme)
	}
	if loaded.DefaultSpecies != "cat" {
		t.Errorf("Expected cat species, got %s", loaded.DefaultSpecies)
	}
	if loaded.TickRateMs != 800 {
		t.Errorf("Expected 800ms tick rate, got %d", loaded.TickRateMs)
	}
	if loaded.SoundEnabled != false {
		t.Errorf("Expected sound to be disabled")
	}
	if loaded.MaxScrollback != 1000 {
		t.Errorf("Expected 1000 scrollback, got %d", loaded.MaxScrollback)
	}
}
