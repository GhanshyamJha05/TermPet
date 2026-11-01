# termpet 🐾🐉

An interactive terminal pet that roams across your **real live terminal workspace** while allowing you to run shell commands in place!

You **never have to quit or switch spaces** to run terminal commands (`dir`, `ls`, `git`, `cd`, `echo`, etc.). Your pet roams above and across your terminal output in real time.

---

## What's New
1. **Interactive Shell Integration**: Type any terminal or PowerShell/Bash command right at the `❯ ` prompt. Output scrolls naturally in the terminal history window while your pet explores above it.
2. **Baby Dragon Species (`dragon`) by default**:
   - Detailed ASCII wings, horns, tail, animated paws, resting sleep pose (`(___m_m)`), happy heart wings (`~(<3|W|<3)~`), and hungry fire breath (`🔥`).
   - Also supports `cat`, `dog`, and `ghost`.
3. **Compact Mode**: `--small` provides a compact one-line sprite for narrow tmux or terminal split panes.

---

## Built-In Pet Commands & Shortcuts

You can type these directly into the terminal prompt:

| Command | Action |
|---|---|
| `:f` or `:feed` or `feed` | Feed your pet (hunger -25%, triggers happy celebration) |
| `:p` or `:pet` or `pet` | Pet your dragon (triggers happy affection) |
| `:name <newname>` | Rename your pet (e.g., `:name Sparky`) |
| `clear` or `cls` | Clears terminal history scrollback |
| `exit` or `quit` or `Ctrl+C` | Saves pet status to `~/.termpet/state.json` and exits cleanly |

---

## State Machine & Behavior

- **Idle**: Stretches wings, blinks, looks around.
- **Walking**: Flaps baby wings, paces left/right towards random points on screen.
- **Sleeping**: Curled up comfortably with floating `z Z` snore effects.
- **Happy**: Glowing eyes, fluttering wings, and hearts/sparkles.
- **Hungry**: When hunger hits \(\ge 75\%\), breathes tiny warning puffs with `feed me!`.
- **Persistence**: Hunger slowly accumulates while away (capped at 100%).

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

# Choose a custom name
.\termpet.exe -name Sparky

# Compact 1-line mode for tmux panes
.\termpet.exe -small
```

### 3. Run Unit Tests
```powershell
go test -v ./...
```
