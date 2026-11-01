package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	speciesFlag := flag.String("species", "dragon", "Species of the pet: dragon, cat, dog, ghost, or custom")
	themeFlag := flag.String("theme", "dracula", "Color theme: dracula, fire, emerald, cyber, sakura")
	smallFlag := flag.Bool("small", false, "Compact one-line sprite mode for tmux / small panes")
	nameFlag := flag.String("name", "", "Custom pet name (defaults to saved name or Draco)")
	flag.Parse()

	// Load any custom sprites from ~/.termpet/sprites/*.json
	LoadCustomSprites()

	chosenSpecies := strings.ToLower(strings.TrimSpace(*speciesFlag))
	if _, ok := SpeciesMap[chosenSpecies]; !ok {
		fmt.Fprintf(os.Stderr, "Unknown species %q. Available built-ins: dragon, cat, dog, ghost (or add JSON in ~/.termpet/sprites/)\n", chosenSpecies)
		os.Exit(1)
	}

	statePath, err := DefaultStatePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve state directory: %v\n", err)
		os.Exit(1)
	}

	defaultName := "Draco"
	if *nameFlag != "" {
		defaultName = *nameFlag
	}

	state, err := LoadState(statePath, defaultName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load state (%v), starting fresh\n", err)
		state = &PetState{
			Name:   defaultName,
			Hunger: 20.0,
			Theme:  "dracula",
		}
	}

	if *nameFlag != "" {
		state.Name = *nameFlag
	}
	if *themeFlag != "" {
		if _, ok := Themes[strings.ToLower(*themeFlag)]; ok {
			state.Theme = strings.ToLower(*themeFlag)
		}
	}

	m := NewModel(chosenSpecies, *smallFlag, statePath, state)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running termpet: %v\n", err)
		os.Exit(1)
	}
}
