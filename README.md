# TermPet 🐾🐉

An interactive terminal pet that roams across your **real live terminal workspace** while allowing you to run shell commands in place!

You **never have to quit or switch spaces** to run terminal commands (`dir`, `ls`, `git`, `cd`, `echo`, etc.). Your pet roams above and across your terminal output in real time.

---

## Features

1. **Interactive Shell Integration**:
   - Type any terminal command directly at the `❯ ` prompt with persistent directory support (`cd`).
   - Output renders naturally in the terminal history window while your pet explores above it.
2. **Directory Persistence (`cd` support)**:
   - Built-in `cd <dir>`, `cd ..`, `cd ~` support that persists across commands.
   - Current directory folder name is displayed next to the command prompt.
3. **Scrollback Navigation**:
   - Press **`PageUp`** / **`PageDown`** to scroll through output history.
   - Status bar clearly shows how many lines you are scrolled up (`[▲ N lines]`), and automatically snaps back to bottom on new commands.
4. **Command History Navigation**:
   - Press **`Up`** / **`Down`** arrows to browse previously entered shell and pet commands.
5. **Tab Auto-Completion**:
   - Press **`Tab`** to auto-complete pet commands (`:feed`, `:pet`, `:name`, etc.) as well as local file and folder paths.
6. **Baby Dragon Species (`dragon`) by default**:
   - Detailed ASCII wings, horns, tail, animated paws, resting sleep pose (`(___m_m)`), happy heart wings (`~(<3|W|<3)~`), and hungry fire breath.
   - Also includes `cat`, `dog`, and `ghost` sprites.
7. **Compact Mode**:
   - Pass `--small` for a compact one-line sprite ideal for narrow tmux panes or status splits.
8. **Persistence**:
   - Pet name, hunger, and last-seen timestamp are stored in `~/.termpet/state.json`.

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

## Navigation Shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Cycle through previous command history |
| `Tab` | Auto-complete commands and file/directory paths |
| `PageUp` | Scroll up terminal output history by 5 lines |
| `PageDown` | Scroll down terminal output history by 5 lines |
| `Ctrl+C` / `Esc` | Cleanly save state and quit |

---

## Building & Running

### 1. Build
```powershell
go build -o termpet.exe .
```

### 2. Run
```powershell
# Run with default Baby Dragon ("Draco")
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
