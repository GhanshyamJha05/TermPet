# TermPet 🐾🐉

An interactive terminal pet that roams across your **real live terminal workspace** while allowing you to run shell commands in place!

You **never have to quit or switch spaces** to run terminal commands (`dir`, `ls`, `git`, `cd`, `echo`, etc.). Your pet roams above and across your terminal output in real time with active reactions to your work, dynamic leveling, evolution, and customizable color themes.

---

## What's New: Visuals & Customization 🎨✨

1. **TrueColor Gradient Sprites**:
   - Multi-line sprites are dynamically rendered with smooth vertical gradients matching your chosen theme (horns, wings, belly, and claws highlighted in distinct hues).
2. **Built-in Color Themes**:
   - Change your terminal palette on the fly with `:theme <name>`:
     - **`dracula`**: Iconic vampire violet, glowing amber, and hot pink accent.
     - **`fire`**: Blazing crimson, fiery orange, and bright spark yellow.
     - **`emerald`**: Lush radioactive lime, mint green, and electric cyan.
     - **`cyber`**: High-contrast cyberpunk neon cyan, magenta pink, and laser yellow.
     - **`sakura`**: Delicate cherry blossom pink, pastel magenta, and crisp white.
3. **Dynamic Dragon Evolution & Growth**:
   - Your pet now gains **XP** and levels up as you work and run shell commands (+20 XP per successful command, +10 XP per pet/feed, +25 XP per mess cleaned).
   - **Level 1–4 (Baby Stage)**: Small baby dragon hatchling.
   - **Level 5–9 (Juvenile Stage)**: Dragon sprouts large spread wings (`\\_//\\ /|WW|\\ /\\_//`) and bigger horns.
   - **Level 10+ (Ancient Stage)**: Massive ancient dragon with armored horns (`/\\_★_★_/\\`), broad wings (`<<\\\\>>`), and majestic presence.
   - Evolutionary milestones trigger celebratory particle showers (`👑`, `⚡`, `🌟`).
4. **Custom Sprites via JSON (`~/.termpet/sprites/`)**:
   - Drop any custom `.json` sprite definition in `~/.termpet/sprites/` without recompiling the source code!
   - Example format:
   ```json
   {
     "name": "phoenix",
     "normal": {
       "idle": { "left": ["~<v>~"], "right": ["~<v>~"] },
       "walking": { "left": ["~<v>~-"], "right": ["-~<v>~"] }
     },
     "compact": {
       "idle": { "left": ["<v>"], "right": ["<v>"] }
     }
   }
   ```
   - Launch with `termpet -species phoenix`.

---

## Pet Behavior & Interactivity 🎮

- **Typing Curiosity**: When you type in the prompt, your pet gets curious (`StateCurious`), tilts its head, and trots down to watch.
- **Outcome Reactions**:
  - **Success (exit 0)**: Dragon celebrates with a victory spin (`StateVictory`), raising its claws with celebratory stars (`★`, `✨`).
  - **Failure (non-zero exit)**: Dragon enters `StateAngry`, stomping and puffing fire sparks (`🔥`, `💥`).
- **Multi-Stat Simulation**:
  - **Hunger**: Rises over time; triggers audible alert at \(\ge 90\%\).
  - **Mood / Happiness**: Boosted by care and successful commands.
  - **Energy**: Drains while active; pet sleeps when tired.
  - **Cleanliness & Poop Mechanic**: Messes (`💩`) appear after long sessions; type `:clean` to tidy up with bubble particles.
- **Audible Alerts (`:bell`)**: Terminal bell alerts on critical hunger or command failures.

---

## Command Reference

| Command | Action |
|---|---|
| `:theme <name>` | Switch color theme (`dracula`, `fire`, `emerald`, `cyber`, `sakura`) |
| `:f` or `:feed` | Feed your pet (hunger -25%, happiness +15%, +10 XP) |
| `:p` or `:pet` | Pet your dragon (happiness +20%, +10 XP) |
| `:clean` | Clean up messes (`💩`) on the terminal floor (+25 XP per mess) |
| `:bell` | Toggle sound alerts on/off |
| `:name <txt>` | Rename your pet on the fly |
| `clear` or `cls` | Clears terminal scrollback |
| `exit` or `quit` | Saves pet status to `~/.termpet/state.json` and exits cleanly |

---

## Navigation Shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Cycle command history |
| `Tab` | Auto-complete commands, themes, & file paths |
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

# Start with a specific theme
.\termpet.exe -theme fire
.\termpet.exe -theme cyber

# Custom name
.\termpet.exe -name Sparky

# Compact mode for small tmux panes
.\termpet.exe -small
```

### 3. Run Unit Tests
```powershell
go test -v ./...
```
