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
	ti.Placeholder = "Type shell command or pet command (:feed, :pet, :name <newname>)..."
	ti.Focus()
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))

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
		history: []string{
			"✨ TermPet initialized! Your companion roams around as you type commands.",
			"💡 Use standard terminal commands (cd, dir, ls, git, echo, etc.)",
			"💡 Scroll: [PgUp] / [PgDn] / [Shift+Up] / [Shift+Down] | Command history: [Up] / [Down] | [Tab] for completion",
			"💡 Pet Shortcuts: ':f' to feed, ':p' to pet, ':name <new>' to rename, 'exit'/'quit' to quit.",
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

				// Automatically reset scroll offset on new command
				m.scrollOffset = 0

				m.handleCommand(inputStr)
				m.terminalInput.Reset()
			}
			return m, nil
		}

	case tickMsg:
		m.tickCount++

		// Accumulate hunger slowly while running (0.005% per tick = 1% every 20s)
		m.petState.Hunger += 0.005
		if m.petState.Hunger > 100 {
			m.petState.Hunger = 100
		}

		// Update notice countdown
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
		// Stay at oldest
		m.cmdIndex = 0
		m.terminalInput.SetValue(m.cmdHistory[0])
		m.terminalInput.SetCursor(len(m.terminalInput.Value()))
	} else if newIdx >= len(m.cmdHistory) {
		// Returned past newest, restore saved input
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

	// Pet commands auto-complete
	petCommands := []string{":feed", ":f", ":pet", ":p", ":name ", ":quit", ":q", "clear", "cls", "exit"}
	if len(parts) <= 1 && (strings.HasPrefix(prefix, ":") || strings.HasPrefix(prefix, "c") || strings.HasPrefix(prefix, "e")) {
		for _, cmd := range petCommands {
			if strings.HasPrefix(cmd, prefix) && cmd != prefix {
				m.terminalInput.SetValue(cmd + val[cursor:])
				m.terminalInput.SetCursor(len(cmd))
				return
			}
		}
	}

	// File / Directory path auto-completion
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
		// Replace matched segment
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

	// Built-in pet commands & exit
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
		m.state = StateHappy
		m.stateTimer = 25
		m.notice = "Yum! Fed " + m.petState.Name + "! (Hunger -25%)"
		m.noticeTime = 25
		m.addHistoryLine(fmt.Sprintf("🍗 Fed %s! Hunger is now %.0f%%.", m.petState.Name, m.petState.Hunger))
		_ = SaveState(m.statePath, m.petState)
		return
	}

	if lower == ":p" || lower == ":pet" || lower == "pet" {
		m.state = StateHappy
		m.stateTimer = 25
		m.notice = "Purr! You petted " + m.petState.Name + "!"
		m.noticeTime = 25
		m.addHistoryLine(fmt.Sprintf("💖 You gave %s warm pets! They are happy.", m.petState.Name))
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

	// Directory Persistence: cd support
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
		} else {
			m.cwd = newPath
			_ = os.Chdir(newPath)
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
		return
	}

	lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	limit := 100
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	for _, l := range lines {
		m.addHistoryLine(strings.TrimRight(l, "\r"))
	}
}

func (m *Model) addHistoryLine(line string) {
	m.history = append(m.history, line)
	// Keep history bounded to 500 lines
	if len(m.history) > 500 {
		m.history = m.history[len(m.history)-500:]
	}
}

func (m *Model) updateBehavior() {
	if m.petState.Hunger >= 75.0 && m.state != StateHappy {
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

	case StateHappy:
		if m.stateTimer <= 0 {
			m.pickNextState()
		}

	case StateSleeping:
		if m.stateTimer <= 0 {
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

	roll := m.rng.Float32()
	switch {
	case roll < 0.50: // 50% walk
		m.state = StateWalking
		m.pickRandomTarget()
		m.stateTimer = m.rng.Intn(60) + 20

	case roll < 0.80: // 30% stay idle
		m.state = StateIdle
		m.stateTimer = m.rng.Intn(40) + 20

	default: // 20% sleep
		m.state = StateSleeping
		m.stateTimer = m.rng.Intn(70) + 40
	}
}

func (m *Model) pickRandomTarget() {
	sprite := GetFrame(m.species, m.compact, m.state, m.dir, m.tickCount)
	spW, spH := GetSpriteDimensions(sprite)

	maxX := m.width - spW - 2
	// Leave bottom 3 rows for input prompt and status bar
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

	// Vertical movement every 2 ticks
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

	// Layout budget:
	// - Line height-2: Terminal command input prompt
	// - Line height-1: Status bar
	// - Lines 0 to height-3: Terminal history workspace where pet roams!
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

	// 3. Render pet over the terminal lines
	screenLines := make([]string, usableHeight)
	for r := 0; r < usableHeight; r++ {
		bgText := historyLines[r]
		bgWidth := runewidth.StringWidth(bgText)

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
		} else {
			screenLines[r] = bgText
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

	stateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1FA8C"))

	scrollIndicator := ""
	if m.scrollOffset > 0 {
		scrollIndicator = fmt.Sprintf(" [▲ %d lines]", m.scrollOffset)
	}

	noticeStr := ""
	if m.notice != "" {
		noticeStr = fmt.Sprintf(" | %s", m.notice)
	}

	info := fmt.Sprintf(" 🐉 %s (%s) | Hunger: %s | State: %s%s%s",
		nameStyle.Render(m.petState.Name),
		m.species,
		hungerStyle.Render(fmt.Sprintf("%d%%", hungerVal)),
		stateStyle.Render(string(m.state)),
		scrollIndicator,
		noticeStr,
	)

	hints := fmt.Sprintf("[%s]:feed [%s]:pet [%s]hist [%s]scroll [%s]quit ",
		keyStyle.Render(":f"),
		keyStyle.Render(":p"),
		keyStyle.Render("↑↓"),
		keyStyle.Render("PgUp/Dn"),
		keyStyle.Render("exit"),
	)

	infoLen := runewidth.StringWidth(info)
	hintsLen := runewidth.StringWidth(hints)
	gap := m.width - infoLen - hintsLen
	if gap < 1 {
		gap = 1
	}

	return barStyle.Render(info + strings.Repeat(" ", gap) + hints)
}
