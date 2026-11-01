# TermPet 🐾🐉

An interactive terminal pet that roams across your **real live terminal workspace** while allowing you to run shell commands in place!

You **never have to quit or switch spaces** to run terminal commands (`dir`, `ls`, `git`, `cd`, `echo`, etc.). Your pet roams above and across your terminal output in real time with active reactions to your work.

---

## What's New: Advanced Pet Behavior & Interactivity 🎮

1. **Typing Curiosity**:
   - As you type commands into the prompt, your pet perks up, tilts its head with curious eyes (`StateCurious`), and trots down toward your prompt to watch you type.
2. **Command Outcome Reactions**:
   - **On Success (exit 0)**: Your dragon celebrates with a victory spin (`StateVictory`), raising its claws and showering celebratory sparkles (`★`, `✨`, `🎉`). Mood/Happiness goes up!
   - **On Error / Failure (non-zero exit)**: Dragon enters `StateAngry`, stomping and puffing warning flame sparks (`🔥`, `💥`).
3. **Multi-Stat Simulation (Hunger, Mood/Happiness, Energy, Cleanliness)**:
   - **Hunger**: Increases slowly over time.
   - **Mood / Happiness**: Drops if neglected, but soars when you feed (`:feed`), pet (`:pet`), clean (`:clean`), or successfully run commands.
   - **Energy**: Drains while active. When depleted, your pet naturally lies down to sleep and recover energy.
   - **Cleanliness & Poop Mechanic**: If left well-fed without cleaning, messes (`💩`) appear on the terminal floor. Type `:clean` to tidy up with fresh bubbly particle effects.
4. **Visual Particles**:
   - Floating hearts (`♥`), stars (`★`), and celebratory sparkles appear when interacting with your pet.
5. **Audible Terminal Bell (`:bell`)**:
   - Emits a gentle terminal alert pulse when hunger is critical (\(\ge 90\%\)) or on command failure. Toggle on/off anytime with `:bell`.

---

## Features & Navigation

- **Directory Persistence**: Full `cd <dir>`, `cd ..`, and `cd ~` support that tracks your working directory across shell commands.
- **Scrollback Navigation**: Use **`PageUp`** / **`PageDown`** to scroll long terminal outputs; status bar shows `[▲ N lines]`.
- **Command History**: **`↑`** / **`↓`** arrows to browse through command history.
- **Tab Auto-Completion**: **`Tab`** auto-completes pet commands as well as filesystem paths.
- **Compact Mode**: `--small` provides a compact one-line sprite for tmux or narrow splits.

---

## Pet Commands

| Command | Action |
|---|---|
| `:f` or `:feed` | Feed your pet (hunger -25%, happiness +15%, spawns hearts) |
| `:p` or `:pet` | Pet your companion (happiness +20%, spawns heart sparkles) |
| `:clean` | Clean up messes (`💩`) on the terminal floor (cleanliness 100%) |
| `:bell` | Toggle audible terminal alerts on/off |
| `:name <txt>` | Rename your pet on the fly |
| `clear` or `cls` | Clears terminal scrollback |
| `exit` or `quit` | Saves pet status to `~/.termpet/state.json` and exits cleanly |

---

## Key Shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Cycle command history |
| `Tab` | Auto-complete commands & paths |
| `PageUp` / `PageDown` | Scroll terminal history up/down |
| `Ctrl+C` / `Esc` | Save state and quit |

---

## Building & Running

### 1. Build
```powershell
go build -o termpet.exe .
```

### 2. Run
```powershell
# Run with default Baby Dragon
.\termpet.exe

# Custom pet name
.\termpet.exe -name Sparky

# Compact mode for small panes
.\termpet.exe -small
```

### 3. Run Unit Tests
```powershell
go test -v ./...
```
