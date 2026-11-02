package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig represents global user configuration for TermPet.
type AppConfig struct {
	DefaultTheme   string `json:"default_theme"`
	DefaultSpecies string `json:"default_species"`
	TickRateMs     int    `json:"tick_rate_ms"`
	SoundEnabled   bool   `json:"sound_enabled"`
	MaxScrollback  int    `json:"max_scrollback"`
}

// DefaultConfig returns default sensible application configuration.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		DefaultTheme:   "dracula",
		DefaultSpecies: "dragon",
		TickRateMs:     1000,
		SoundEnabled:   true,
		MaxScrollback:  500,
	}
}

// DefaultConfigPath returns ~/.termpet/config.json
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".termpet", "config.json"), nil
}

// LoadConfig reads config from configPath. If missing or invalid, returns default config.
func LoadConfig(configPath string) (*AppConfig, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), err
	}

	// Validate & sanitize bounds
	if cfg.TickRateMs < 100 {
		cfg.TickRateMs = 1000
	}
	if cfg.MaxScrollback < 50 {
		cfg.MaxScrollback = 500
	}
	if cfg.DefaultTheme == "" {
		cfg.DefaultTheme = "dracula"
	}
	if cfg.DefaultSpecies == "" {
		cfg.DefaultSpecies = "dragon"
	}

	return cfg, nil
}

// SaveConfig persists the configuration to disk.
func SaveConfig(configPath string, cfg *AppConfig) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
