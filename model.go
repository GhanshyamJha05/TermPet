package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type tickMsg time.Time

// Particle represents animated sparkles, hearts, or fire sparks
type Particle struct {
	x        int
	y        int
	char     string
	color    string
	lifetime int
}

// PoopItem represents a mess left on the terminal floor
type PoopItem struct {
	x int
	y int
}

// Model represents the bubbletea application model
type Model struct {
	width     int
	height    int
	species   string
	compact   bool
	statePath string
	petState  *PetState

	// Position and movement
	x          int
	y          int
	targetX    int
	targetY    int
	dir        Direction
	state      StateType
	stateTimer int // ticks left in current state
	tickCount  int // animation tick

	// Shell command execution & terminal history
	terminalInput textinput.Model
	history       []string
	scrollOffset  int // 0 = scrolled to bottom, >0 = scrolled up by N lines
	cwd           string

	// Command history navigation
	cmdHistory []string
	cmdIndex   int // -1 when not browsing or at new prompt
	savedInput string

	// Feedback message ("Fed Luna! +30%", "Pet Luna <3")
	notice     string
	noticeTime int

	// Visual particles & poops
	particles []Particle
	poops     []PoopItem

	// Sound / Terminal bell toggle
	soundEnabled bool

	// Random source
	rng *rand.Rand
}

func NewModel(species string, compact bool, statePath string, petState *PetState) Model {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	ti := textinput.New()
	ti.Placeholder = "Type shell command or pet command (:feed, :pet, :clean, :bell, :name <txt>)..."
	ti.Focus()
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))

	initialPoops := make([]PoopItem, 0)
	if petState.HasPoop && petState.PoopCount > 0 {
		for i := 0; i < petState.PoopCount; i++ {
			initialPoops = append(initialPoops, PoopItem{
				x: 5 + (i * 8),
				y: 4 + (i % 3),
			})
		}
	}

	return Model{
		species:       species,
		compact:       compact,
		statePath:     statePath,
		petState:      petState,
		x:             15,
		y:             3,
		targetX:       15,
		targetY:       3,
		dir:           DirRight,
		state:         StateIdle,
		stateTimer:    30, // 3 seconds initially
		tickCount:     0,
		rng:           r,
		terminalInput: ti,
		cwd:           currentDir,
		cmdHistory:    make([]string, 0),
		cmdIndex:      -1,
		particles:     make([]Particle, 0),
		poops:         initialPoops,
		soundEnabled:  true,
		history: []string{
			"✨ TermPet companion initialized! Watch your pet react to your commands.",
			"💡 Use standard terminal commands (cd, dir, ls, git, echo, etc.)",
			"💡 Pet commands: ':f' (feed), ':p' (pet), ':clean' (clean mess), ':bell' (toggle sound), ':name <txt>'",
			"💡 Navigation: [↑/↓] command history, [Tab] completion, [PgUp/PgDn] scrollback",
		},
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, textinput.Blink, tickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.terminalInput.Width = msg.Width - 6
		m.clampPosition()
		return m, nil

	case tea.KeyMsg:
		// If user is actively typing, pet gets curious and watches them!
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace || msg.Type == tea.KeyBackspace {
			m.handleTypingReaction()
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			_ = SaveState(m.statePath, m.petState)
			return m, tea.Quit

		case tea.KeyPgUp:
			m.scrollUp(5)
			return m, nil

		case tea.KeyPgDown:
			m.scrollDown(5)
			return m, nil

		case tea.KeyUp:
			if msg.Alt || (msg.Runes != nil && len(msg.Runes) > 0) {
				m.scrollUp(1)
				return m, nil
			}
			m.navigateHistory(-1)
			return m, nil

		case tea.KeyDown:
			if msg.Alt {
				m.scrollDown(1)
				return m, nil
			}
			m.navigateHistory(1)
			return m, nil

		case tea.KeyTab:
			m.handleTabCompletion()
			return m, nil

		case tea.KeyEnter:
			inputStr := strings.TrimSpace(m.terminalInput.Value())
			if inputStr != "" {
				// Record to command history
				if len(m.cmdHistory) == 0 || m.cmdHistory[len(m.cmdHistory)-1] != inputStr {
					m.cmdHistory = append(m.cmdHistory, inputStr)
				}
				m.cmdIndex = -1
				m.savedInput = ""
				m.scrollOffset = 0

				m.handleCommand(inputStr)
				m.terminalInput.Reset()
			}
			return m, nil
		}

	case tickMsg:
		m.tickCount++

		// Stats decay slowly over time
		m.decayStats()

		// Update particles
		m.updateParticles()

		// Notice countdown
		if m.noticeTime > 0 {
			m.noticeTime--
			if m.noticeTime == 0 {
				m.notice = ""
			}
		}

		// Periodically save state (every 100 ticks = 10s)
		if m.tickCount%100 == 0 {
			_ = SaveState(m.statePath, m.petState)
		}

		// Update pet state machine
		m.updateBehavior()

		return m, tickCmd()
	}

	m.terminalInput, cmd = m.terminalInput.Update(msg)
	return m, cmd
}

func (m *Model) decayStats() {
	// Hunger accumulates (0.005% per tick = ~1% per 20 seconds)
	m.petState.Hunger += 0.005
	if m.petState.Hunger > 100 {
		m.petState.Hunger = 100
	}

	// Energy drains while active
	if m.state != StateSleeping {
		m.petState.Energy -= 0.004
		if m.petState.Energy < 0 {
			m.petState.Energy = 0
		}
	} else {
		// Energy recovers while sleeping
		m.petState.Energy += 0.03
		if m.petState.Energy > 100 {
			m.petState.Energy = 100
		}
	}

	// Happiness drops if hungry or surrounded by poop
	if m.petState.Hunger > 75.0 || m.petState.HasPoop {
		m.petState.Happiness -= 0.008
		if m.petState.Happiness < 0 {
			m.petState.Happiness = 0
		}
	}

	// High hunger warning bell pulse
	if int(m.petState.Hunger) >= 90 && m.tickCount%150 == 0 {
		m.beep()
	}

	// Poop mechanic: after eating well and time passes, pet leaves a mess
	if m.petState.Hunger < 40 && m.tickCount > 0 && m.tickCount%1200 == 0 && len(m.poops) < 5 {
		m.petState.HasPoop = true
		m.poops = append(m.poops, PoopItem{x: m.x, y: m.y + 1})
		m.petState.PoopCount = len(m.poops)
		m.petState.Cleanliness = 100.0 - float64(len(m.poops)*20)
		m.notice = fmt.Sprintf("💩 %s made a mess! Type ':clean' to tidy up.", m.petState.Name)
		m.noticeTime = 40
	}
}

func (m *Model) beep() {
	if m.soundEnabled {
		// Terminal audible bell
		fmt.Print("\a")
	}
}

func (m *Model) spawnParticles(x, y int, chars []string, colors []string, count int) {
	for i := 0; i < count; i++ {
		offsetX := m.rng.Intn(7) - 3
		offsetY := m.rng.Intn(3) - 1
		ch := chars[m.rng.Intn(len(chars))]
		col := colors[m.rng.Intn(len(colors))]
		m.particles = append(m.particles, Particle{
			x:        x + offsetX,
			y:        y + offsetY,
			char:     ch,
			color:    col,
			lifetime: m.rng.Intn(8) + 6,
		})
	}
}

func (m *Model) updateParticles() {
	surviving := make([]Particle, 0, len(m.particles))
	for _, p := range m.particles {
		p.lifetime--
		if p.lifetime > 0 {
			// Float upward gently
			if p.lifetime%2 == 0 {
				p.y--
			}
			surviving = append(surviving, p)
		}
	}
	m.particles = surviving
}

func (m *Model) handleTypingReaction() {
	// If pet is idle or walking, 30% chance to become curious and step toward prompt
	if (m.state == StateIdle || m.state == StateWalking) && m.rng.Float32() < 0.35 {
		m.state = StateCurious
		m.stateTimer = 20
		// Walk towards lower-left area where prompt is
		m.targetX = m.rng.Intn(18) + 2
		m.targetY = m.height - 7
		if m.targetY < 1 {
			m.targetY = 1
		}
		if m.targetX > m.x {
			m.dir = DirRight
		} else {
			m.dir = DirLeft
		}
	}
}

func (m *Model) scrollUp(delta int) {
	usableHeight := m.height - 2
	if usableHeight < 1 {
		usableHeight = 1
	}
	maxScroll := len(m.history) - usableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	m.scrollOffset += delta
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
}

func (m *Model) scrollDown(delta int) {
	m.scrollOffset -= delta
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *Model) navigateHistory(dir int) {
	if len(m.cmdHistory) == 0 {
		return
	}

	if m.cmdIndex == -1 {
		if dir == -1 {
			m.savedInput = m.terminalInput.Value()
			m.cmdIndex = len(m.cmdHistory) - 1
			m.terminalInput.SetValue(m.cmdHistory[m.cmdIndex])
			m.terminalInput.SetCursor(len(m.terminalInput.Value()))
		}
		return
	}

	newIdx := m.cmdIndex + dir
	if newIdx < 0 {
		m.cmdIndex = 0
		m.terminalInput.SetValue(m.cmdHistory[0])
		m.terminalInput.SetCursor(len(m.terminalInput.Value()))
	} else if newIdx >= len(m.cmdHistory) {
		m.cmdIndex = -1
		m.terminalInput.SetValue(m.savedInput)
		m.terminalInput.SetCursor(len(m.savedInput))
	} else {
		m.cmdIndex = newIdx
		m.terminalInput.SetValue(m.cmdHistory[newIdx])
		m.terminalInput.SetCursor(len(m.terminalInput.Value()))
	}
}

func (m *Model) handleTabCompletion() {
	val := m.terminalInput.Value()
	cursor := m.terminalInput.Position()
	if cursor > len(val) {
		cursor = len(val)
	}

	prefix := val[:cursor]
	parts := strings.Fields(prefix)

	petCommands := []string{":feed", ":f", ":pet", ":p", ":clean", ":bell", ":name ", ":quit", ":q", "clear", "cls", "exit"}
	if len(parts) <= 1 && (strings.HasPrefix(prefix, ":") || strings.HasPrefix(prefix, "c") || strings.HasPrefix(prefix, "e")) {
		for _, cmd := range petCommands {
			if strings.HasPrefix(cmd, prefix) && cmd != prefix {
				m.terminalInput.SetValue(cmd + val[cursor:])
				m.terminalInput.SetCursor(len(cmd))
				return
			}
		}
	}

	targetWord := ""
	startIdx := strings.LastIndexAny(prefix, " \t\"'")
	if startIdx >= 0 {
		targetWord = prefix[startIdx+1:]
	} else {
		targetWord = prefix
	}

	resolvedDir := m.cwd
	matchPrefix := targetWord

	if strings.ContainsAny(targetWord, "/\\") {
		dirPart := filepath.Dir(targetWord)
		basePart := filepath.Base(targetWord)
		if filepath.IsAbs(dirPart) {
			resolvedDir = dirPart
		} else {
			resolvedDir = filepath.Join(m.cwd, dirPart)
		}
		matchPrefix = basePart
	}

	entries, err := os.ReadDir(resolvedDir)
	if err != nil {
		return
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(matchPrefix)) {
			completed := name
			if entry.IsDir() {
				completed += string(filepath.Separator)
			}
			matches = append(matches, completed)
		}
	}

	if len(matches) == 1 {
		var fullCompleted string
		if strings.ContainsAny(targetWord, "/\\") {
			dirPart := filepath.Dir(targetWord)
			fullCompleted = filepath.Join(dirPart, matches[0])
		} else {
			fullCompleted = matches[0]
		}

		newVal := prefix[:len(prefix)-len(targetWord)] + fullCompleted + val[cursor:]
		m.terminalInput.SetValue(newVal)
		m.terminalInput.SetCursor(len(prefix) - len(targetWord) + len(fullCompleted))
	} else if len(matches) > 1 {
		m.addHistoryLine("Matches: " + strings.Join(matches, "  "))
	}
}

func (m *Model) handleCommand(rawCmd string) {
	lower := strings.ToLower(rawCmd)

	if lower == "quit" || lower == "exit" || lower == ":q" {
		_ = SaveState(m.statePath, m.petState)
		m.history = append(m.history, "Goodbye! Exiting...")
		return
	}

	if lower == ":f" || lower == ":feed" || lower == "feed" {
		m.petState.Hunger -= 25.0
		if m.petState.Hunger < 0 {
			m.petState.Hunger = 0
		}
		m.petState.Happiness += 15.0
		if m.petState.Happiness > 100 {
			m.petState.Happiness = 100
		}
		m.state = StateHappy
		m.stateTimer = 25
		m.notice = "Yum! Fed " + m.petState.Name + "! (Hunger -25%, Happiness +15%)"
		m.noticeTime = 25
		m.addHistoryLine(fmt.Sprintf("🍗 Fed %s! Hunger: %.0f%% | Happiness: %.0f%%.", m.petState.Name, m.petState.Hunger, m.petState.Happiness))
		m.spawnParticles(m.x+4, m.y+1, []string{"♥", "★", "✨"}, []string{"#FF79C6", "#F1FA8C", "#50FA7B"}, 6)
		_ = SaveState(m.statePath, m.petState)
		return
	}

	if lower == ":p" || lower == ":pet" || lower == "pet" {
		m.petState.Happiness += 20.0
		if m.petState.Happiness > 100 {
			m.petState.Happiness = 100
		}
		m.state = StateHappy
		m.stateTimer = 25
		m.notice = "Purr! You petted " + m.petState.Name + "! (Happiness +20%)"
		m.noticeTime = 25
		m.addHistoryLine(fmt.Sprintf("💖 You gave %s warm pets! Happiness is now %.0f%%.", m.petState.Name, m.petState.Happiness))
		m.spawnParticles(m.x+4, m.y, []string{"♥", "💖", "✨"}, []string{"#FF79C6", "#FF92D0", "#BD93F9"}, 7)
		_ = SaveState(m.statePath, m.petState)
		return
	}

	if lower == ":clean" || lower == "clean" {
		if len(m.poops) > 0 {
			count := len(m.poops)
			m.poops = nil
			m.petState.HasPoop = false
			m.petState.PoopCount = 0
			m.petState.Cleanliness = 100.0
			m.petState.Happiness += 15.0
			if m.petState.Happiness > 100 {
				m.petState.Happiness = 100
			}
			m.state = StateHappy
			m.stateTimer = 20
			m.notice = fmt.Sprintf("✨ Cleaned up %d messes! Cleanliness is 100%%.", count)
			m.noticeTime = 25
			m.addHistoryLine(fmt.Sprintf("🧹 Cleaned up after %s! Fresh and tidy.", m.petState.Name))
			m.spawnParticles(m.x+4, m.y+1, []string{"✨", "⭐", "🫧"}, []string{"#8BE9FD", "#F1FA8C", "#50FA7B"}, 8)
			_ = SaveState(m.statePath, m.petState)
		} else {
			m.addHistoryLine("🧹 All clean! No messes found.")
		}
		return
	}

	if lower == ":bell" || lower == ":sound" {
		m.soundEnabled = !m.soundEnabled
		status := "enabled"
		if !m.soundEnabled {
			status = "muted"
		}
		m.notice = fmt.Sprintf("🔔 Sound alerts %s.", status)
		m.noticeTime = 20
		m.addHistoryLine(fmt.Sprintf("🔔 Sound alerts are now %s.", status))
		return
	}

	if strings.HasPrefix(lower, ":name ") {
		newName := strings.TrimSpace(rawCmd[6:])
		if newName != "" {
			m.petState.Name = newName
			m.notice = fmt.Sprintf("Renamed to %s!", newName)
			m.noticeTime = 20
			m.addHistoryLine(fmt.Sprintf("🏷️ Pet renamed to %s!", newName))
			_ = SaveState(m.statePath, m.petState)
		}
		return
	}

	if lower == "cls" || lower == "clear" {
		m.history = nil
		m.scrollOffset = 0
		return
	}

	if strings.HasPrefix(lower, "cd ") || lower == "cd" {
		m.addHistoryLine("❯ " + rawCmd)
		target := strings.TrimSpace(rawCmd[2:])
		target = strings.Trim(target, "\"'")

		if target == "" || target == "~" {
			home, err := os.UserHomeDir()
			if err == nil {
				target = home
			}
		}

		var newPath string
		if filepath.IsAbs(target) {
			newPath = filepath.Clean(target)
		} else {
			newPath = filepath.Clean(filepath.Join(m.cwd, target))
		}

		info, err := os.Stat(newPath)
		if err != nil || !info.IsDir() {
			m.addHistoryLine(fmt.Sprintf("cd: no such directory: %s", target))
			// Command failed reaction
			m.onCommandResult(false)
		} else {
			m.cwd = newPath
			_ = os.Chdir(newPath)
			// Command succeeded reaction
			m.onCommandResult(true)
		}
		return
	}

	// Execute as real terminal command in OS shell
	m.addHistoryLine("❯ " + rawCmd)

	var execCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		execCmd = exec.Command("powershell", "-NoProfile", "-Command", rawCmd)
	} else {
		execCmd = exec.Command("sh", "-c", rawCmd)
	}
	execCmd.Dir = m.cwd

	out, err := execCmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		m.addHistoryLine("Error: " + err.Error())
		m.onCommandResult(false)
		return
	}

	// Trigger pet reaction based on execution outcome!
	m.onCommandResult(err == nil)

	lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	limit := 100
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	for _, l := range lines {
		m.addHistoryLine(strings.TrimRight(l, "\r"))
	}
}

// React to command success or failure
func (m *Model) onCommandResult(success bool) {
	if success {
		// Victory reaction: spin & celebratory stars
		m.state = StateVictory
		m.stateTimer = 25
		m.petState.Happiness += 5.0
		if m.petState.Happiness > 100 {
			m.petState.Happiness = 100
		}
		m.spawnParticles(m.x+4, m.y, []string{"★", "✨", "🎉"}, []string{"#F1FA8C", "#50FA7B", "#BD93F9"}, 6)
	} else {
		// Angry reaction: breathe tiny fire puffs
		m.state = StateAngry
		m.stateTimer = 25
		m.spawnParticles(m.x+6, m.y+1, []string{"🔥", "💥", "💨"}, []string{"#FF5555", "#FFB86C", "#F1FA8C"}, 5)
		m.beep()
	}
}

func (m *Model) addHistoryLine(line string) {
	m.history = append(m.history, line)
	if len(m.history) > 500 {
		m.history = m.history[len(m.history)-500:]
	}
}

func (m *Model) updateBehavior() {
	if m.petState.Hunger >= 75.0 && m.state != StateHappy && m.state != StateAngry && m.state != StateVictory {
		m.state = StateHungry
	}

	m.stateTimer--

	switch m.state {
	case StateWalking:
		m.stepTowardTarget()
		if (m.x == m.targetX && m.y == m.targetY) || m.stateTimer <= 0 {
			m.pickNextState()
		}

	case StateHungry:
		if m.stateTimer <= 0 {
			if m.rng.Float32() < 0.6 {
				m.pickRandomTarget()
				m.state = StateWalking
				m.stateTimer = m.rng.Intn(40) + 20
			} else {
				m.stateTimer = m.rng.Intn(30) + 15
			}
		}

	case StateHappy, StateVictory, StateAngry, StateCurious:
		if m.stateTimer <= 0 {
			m.pickNextState()
		}

	case StateSleeping:
		// Wake up if energy is recovered
		if m.petState.Energy >= 95.0 || m.stateTimer <= 0 {
			m.pickNextState()
		}

	case StateIdle:
		if m.stateTimer <= 0 {
			m.pickNextState()
		}
	}

	m.clampPosition()
}

func (m *Model) pickNextState() {
	if m.petState.Hunger >= 75.0 {
		m.state = StateHungry
		m.stateTimer = 40
		return
	}

	// Low energy leads to sleep
	if m.petState.Energy <= 20.0 {
		m.state = StateSleeping
		m.stateTimer = m.rng.Intn(100) + 60
		return
	}

	roll := m.rng.Float32()
	switch {
	case roll < 0.45: // Walk
		m.state = StateWalking
		m.pickRandomTarget()
		m.stateTimer = m.rng.Intn(60) + 20

	case roll < 0.75: // Idle
		m.state = StateIdle
		m.stateTimer = m.rng.Intn(40) + 20

	default: // Rest/Sleep if tired
		if m.petState.Energy < 50.0 {
			m.state = StateSleeping
			m.stateTimer = m.rng.Intn(80) + 40
		} else {
			m.state = StateIdle
			m.stateTimer = 30
		}
	}
}

func (m *Model) pickRandomTarget() {
	sprite := GetFrame(m.species, m.compact, m.state, m.dir, m.tickCount)
	spW, spH := GetSpriteDimensions(sprite)

	maxX := m.width - spW - 2
	maxY := m.height - spH - 4

	if maxX <= 0 {
		maxX = 1
	}
	if maxY <= 0 {
		maxY = 1
	}

	m.targetX = m.rng.Intn(maxX)
	m.targetY = m.rng.Intn(maxY)

	if m.targetX > m.x {
		m.dir = DirRight
	} else if m.targetX < m.x {
		m.dir = DirLeft
	}
}

func (m *Model) stepTowardTarget() {
	if m.x < m.targetX {
		m.x++
		m.dir = DirRight
	} else if m.x > m.targetX {
		m.x--
		m.dir = DirLeft
	}

	if m.tickCount%2 == 0 {
		if m.y < m.targetY {
			m.y++
		} else if m.y > m.targetY {
			m.y--
		}
	}
}

func (m *Model) clampPosition() {
	sprite := GetFrame(m.species, m.compact, m.state, m.dir, m.tickCount)
	spW, spH := GetSpriteDimensions(sprite)

	maxX := m.width - spW - 1
	maxY := m.height - spH - 4

	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}

	if m.x < 0 {
		m.x = 0
	}
	if m.x > maxX {
		m.x = maxX
	}
	if m.y < 0 {
		m.y = 0
	}
	if m.y > maxY {
		m.y = maxY
	}
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Starting TermPet..."
	}

	usableHeight := m.height - 2
	if usableHeight < 1 {
		usableHeight = 1
	}

	// 1. Get terminal background lines with scrollOffset support
	historyLines := make([]string, usableHeight)
	histCount := len(m.history)

	endIdx := histCount - m.scrollOffset
	if endIdx > histCount {
		endIdx = histCount
	}
	if endIdx < 0 {
		endIdx = 0
	}

	startIdx := endIdx - usableHeight
	if startIdx < 0 {
		startIdx = 0
	}

	for r := 0; r < usableHeight; r++ {
		hIdx := startIdx + r
		if hIdx >= 0 && hIdx < endIdx {
			historyLines[r] = m.history[hIdx]
		} else {
			historyLines[r] = ""
		}
	}

	// 2. Fetch current sprite frame
	sprite := GetFrame(m.species, m.compact, m.state, m.dir, m.tickCount)

	spriteStyle := lipgloss.NewStyle()
	switch m.state {
	case StateHappy:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#FF79C6")).Bold(true)
	case StateVictory:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#F1FA8C")).Bold(true)
	case StateAngry:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#FF5555")).Bold(true)
	case StateCurious:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#BD93F9")).Bold(true)
	case StateHungry:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#FF5555")).Bold(true)
	case StateSleeping:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#8BE9FD"))
	case StateWalking:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#50FA7B"))
	default:
		spriteStyle = spriteStyle.Foreground(lipgloss.Color("#FFB86C")).Bold(true)
	}

	styledSprite := spriteStyle.Render(sprite)
	spriteLines := strings.Split(styledSprite, "\n")

	// 3. Render base screen lines (history + poops + pet + particles)
	screenLines := make([]string, usableHeight)
	for r := 0; r < usableHeight; r++ {
		bgText := historyLines[r]
		bgWidth := runewidth.StringWidth(bgText)

		// Render poops on this line if present
		poopStr := ""
		for _, p := range m.poops {
			if p.y == r {
				poopStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
				poopStr = poopStyle.Render("💩")
				break
			}
		}

		if r >= m.y && r < m.y+len(spriteLines) {
			spriteLineIdx := r - m.y
			sprLine := spriteLines[spriteLineIdx]
			sprWidth := runewidth.StringWidth(sprLine)

			if bgWidth <= m.x {
				pad := strings.Repeat(" ", m.x-bgWidth)
				screenLines[r] = bgText + pad + sprLine
			} else {
				prefix := truncateOrPad(bgText, m.x)
				suffix := sliceFromWidth(bgText, m.x+sprWidth)
				screenLines[r] = prefix + sprLine + suffix
			}
		} else if poopStr != "" {
			if bgWidth <= 10 {
				pad := strings.Repeat(" ", 10-bgWidth)
				screenLines[r] = bgText + pad + poopStr
			} else {
				screenLines[r] = bgText + " " + poopStr
			}
		} else {
			screenLines[r] = bgText
		}

		// Overlay floating particles on this line
		for _, pt := range m.particles {
			if pt.y == r && pt.x >= 0 && pt.x < m.width {
				ptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(pt.color))
				pRender := ptStyle.Render(pt.char)
				lineW := runewidth.StringWidth(screenLines[r])
				if lineW <= pt.x {
					pad := strings.Repeat(" ", pt.x-lineW)
					screenLines[r] = screenLines[r] + pad + pRender
				}
			}
		}
	}

	// 4. Command Input Bar with current working directory prompt
	dirBase := filepath.Base(m.cwd)
	if dirBase == "." || dirBase == "/" || dirBase == "\\" {
		dirBase = m.cwd
	}
	cwdPrompt := lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Render(dirBase) + " "
	inputBar := cwdPrompt + m.terminalInput.View()

	// 5. Status Bar
	statusBar := m.renderStatusBar()

	return strings.Join(screenLines, "\n") + "\n" + inputBar + "\n" + statusBar
}

func truncateOrPad(s string, width int) string {
	var sb strings.Builder
	cur := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if cur+rw > width {
			break
		}
		sb.WriteRune(r)
		cur += rw
	}
	if cur < width {
		sb.WriteString(strings.Repeat(" ", width-cur))
	}
	return sb.String()
}

func sliceFromWidth(s string, startWidth int) string {
	cur := 0
	runes := []rune(s)
	for idx, r := range runes {
		rw := runewidth.RuneWidth(r)
		if cur >= startWidth {
			return string(runes[idx:])
		}
		cur += rw
	}
	return ""
}

func (m Model) renderStatusBar() string {
	hungerVal := int(m.petState.Hunger)
	if hungerVal > 100 {
		hungerVal = 100
	}
	if hungerVal < 0 {
		hungerVal = 0
	}

	happyVal := int(m.petState.Happiness)
	energyVal := int(m.petState.Energy)

	hungerColor := "#50FA7B"
	if hungerVal >= 75 {
		hungerColor = "#FF5555"
	} else if hungerVal >= 50 {
		hungerColor = "#FFB86C"
	}

	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#282A36")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Width(m.width)

	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true)

	hungerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(hungerColor)).
		Bold(true)

	happyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF79C6")).
		Bold(true)

	energyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD")).
		Bold(true)

	stateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1FA8C"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B"))

	scrollIndicator := ""
	if m.scrollOffset > 0 {
		scrollIndicator = fmt.Sprintf(" [▲ %d lines]", m.scrollOffset)
	}

	poopIndicator := ""
	if len(m.poops) > 0 {
		poopIndicator = fmt.Sprintf(" [💩 x%d]", len(m.poops))
	}

	bellIcon := "🔔"
	if !m.soundEnabled {
		bellIcon = "🔕"
	}

	noticeStr := ""
	if m.notice != "" {
		noticeStr = fmt.Sprintf(" | %s", m.notice)
	}

	info := fmt.Sprintf(" 🐉 %s | Hunger: %s | Mood: %s | Energy: %s | %s%s%s%s",
		nameStyle.Render(m.petState.Name),
		hungerStyle.Render(fmt.Sprintf("%d%%", hungerVal)),
		happyStyle.Render(fmt.Sprintf("%d%%", happyVal)),
		energyStyle.Render(fmt.Sprintf("%d%%", energyVal)),
		stateStyle.Render(string(m.state)),
		poopIndicator,
		scrollIndicator,
		noticeStr,
	)

	hints := fmt.Sprintf("[%s]:f [%s]:p [%s]:clean [%s]%s [%s]exit ",
		keyStyle.Render(":feed"),
		keyStyle.Render(":pet"),
		keyStyle.Render(":clean"),
		keyStyle.Render(":bell"),
		bellIcon,
		keyStyle.Render("quit"),
	)

	infoLen := runewidth.StringWidth(info)
	hintsLen := runewidth.StringWidth(hints)
	gap := m.width - infoLen - hintsLen
	if gap < 1 {
		gap = 1
	}

	return barStyle.Render(info + strings.Repeat(" ", gap) + hints)
}
