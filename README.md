# termpet 🐾

A cozy terminal pet that roams freely around your entire terminal window. Built with Go, [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

No CGO required. Runs cleanly across Linux, macOS, and Windows Terminal.

---

## Features

- **2D Free Movement**: Roams the full terminal window, not just a single line. Gracefully adapts to terminal resizing and bounds.
- **Dynamic State Machine**:
  - `idle`: Observes surroundings calmly.
  - `walking`: Picks a random target in the window, flips its facing direction (left/right), and paces smoothly.
  - `sleeping`: Rests peacefully with floating `z Z` animations.
  - `happy`: Celebrates with hearts and stars when petted or fed.
  - `hungry`: When hunger reaches \(\ge 75\%\), face expressions change and the pet begs with `feed me!`.
- **Species Options**: Built-in sprites for `cat`, `dog`, and `ghost`.
- **Compact / Tmux Mode**: Pass `--small` for a 1-line sprite ideal for narrow tmux panes or status bars.
- **Persistence**: Remembers your pet's name, hunger level, and last-seen time in `~/.termpet/state.json`. Accumulates offline hunger realistically (capped at 100%).
- **Smooth Animation**: Rendered at 10 FPS with `tea.Tick`.

---

## Installation

### Using `go install`

```bash
go install .
```
*(Or install globally from a git remote if hosted)*

### Building from Source

```bash
# Clone or navigate to the directory
cd "local pet"

# Build executable
go build -o termpet .
```

---

## Usage

```bash
# Run with default species (cat)
termpet

# Choose a different species: dog or ghost
termpet -species dog
termpet -species ghost

# Compact 1-line mode for tmux panes
termpet -small

# Launch with a specific name
termpet -name Mochi
```

---

## Key Bindings

| Key | Action |
|-----|--------|
| `f` | **Feed**: Lowers hunger by 25% and makes your pet happy |
| `p` | **Pet**: Show affection to your pet (triggers happy animation) |
| `n` | **Rename**: Enter renaming mode (`Enter` to save, `Esc` to cancel) |
| `q` / `Esc` / `Ctrl+C` | **Quit**: Cleanly saves pet state and exits alt-screen |

---

## How to Add a New Species

Sprites are decoupled from app logic in [sprites.go](file:///sprites.go). To add a new species (e.g. `bunny`):

1. Open `sprites.go`.
2. Add an entry to the `SpeciesMap` map:

```go
"bunny": {
    Name: "bunny",
    Normal: SpeciesSprites{
        Idle: SpriteFrames{
            Left:  []string{" (\\(\\ \n (-.-) \n o_(\")(\")"},
            Right: []string{" /)/) \n (-.-) \n (\")(\")_o"},
        },
        Walking: SpriteFrames{
            Left:  []string{" (\\(\\ \n ('.' ) \n o_(\")(\")", " (\\(\\ \n ('.' ) \n  /  \\ "},
            Right: []string{" /)/) \n ( '.' ) \n (\")(\")_o", " /)/) \n ( '.' ) \n  /  \\ "},
        },
        Sleeping: SpriteFrames{
            Left:  []string{" (\\(\\  z\n (-.-)  Z\n o_(\")(\")"},
            Right: []string{"z  /)/)\n Z(-.-) \n  (\")(\")_o"},
        },
        Happy: SpriteFrames{
            Left:  []string{" (\\(\\  ♥\n (^.^) \n o_(\")(\")"},
            Right: []string{"♥ /)/)\n  (^.^) \n  (\")(\")_o"},
        },
        Hungry: SpriteFrames{
            Left:  []string{" (\\(\\ feed me!\n (;.;) \n o_(\")(\")"},
            Right: []string{"feed me! /)/)\n        (;.;) \n        (\")(\")_o"},
        },
    },
    Compact: SpeciesSprites{
        Idle:     SpriteFrames{Left: []string{"(\\( -.-)"}, Right: []string{"(-.- /)/)"}},
        Walking:  SpriteFrames{Left: []string{"(\\( '.' )~"}, Right: []string{"~( '.' /)/)"}},
        Sleeping: SpriteFrames{Left: []string{"(-.-)zZ"}, Right: []string{"Zz(-.-)"}},
        Happy:    SpriteFrames{Left: []string{"(^.^)♥"}, Right: []string{"♥(^.^)"}},
        Hungry:   SpriteFrames{Left: []string{"(;.;)[feed!]"}, Right: []string{"[feed!](;.;)"}},
    },
},
```

Now you can run:
```bash
termpet -species bunny
```

---

## Running Tests

```bash
go test -v ./...
```

