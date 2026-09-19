package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen is one dashboard view.
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenSystem
	ScreenProcesses
	ScreenDocker
	ScreenGit
	ScreenNetwork
	ScreenDoctor
	ScreenCount
)

// Title names the screen in the menu and header.
func (s Screen) Title() string {
	switch s {
	case ScreenSystem:
		return "System"
	case ScreenProcesses:
		return "Processes"
	case ScreenDocker:
		return "Docker"
	case ScreenGit:
		return "Git"
	case ScreenNetwork:
		return "Network"
	case ScreenDoctor:
		return "Doctor"
	default:
		return "Dashboard"
	}
}

// dataMsg carries a finished background load back to the UI thread.
type dataMsg struct {
	screen Screen
	body   string
}

// loadCmd collects screen data off the UI thread so refresh
// never freezes the interface.
func loadCmd(s Screen) tea.Cmd {
	return func() tea.Msg {
		return dataMsg{screen: s, body: LoadScreen(s)}
	}
}

// Model is the dashboard state.
type Model struct {
	cursor  Screen
	active  Screen
	width   int
	height  int
	bodies  map[Screen]string
	loading map[Screen]bool
	theme   theme
	ready   bool
}

// New builds the dashboard model. noColor disables color output.
func New(noColor bool) Model {
	return Model{
		bodies:  make(map[Screen]string),
		loading: make(map[Screen]bool),
		theme:   newTheme(noColor),
	}
}

// Init loads the dashboard overview first.
func (m Model) Init() tea.Cmd {
	m.loading[ScreenDashboard] = true
	return loadCmd(ScreenDashboard)
}

// ensureLoad starts a background load unless the screen already
// has content or a load in flight.
func (m Model) ensureLoad(s Screen) tea.Cmd {
	if _, ok := m.bodies[s]; ok {
		return nil
	}
	if m.loading[s] {
		return nil
	}
	m.loading[s] = true
	return loadCmd(s)
}

// Update handles window resizes, loaded data, and keys:
// ↑/↓ navigate, Enter opens, r refreshes, Esc returns home, q quits.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		return m, nil
	case dataMsg:
		m.bodies[msg.screen] = msg.body
		m.loading[msg.screen] = false
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down":
			if m.cursor < ScreenCount-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			m.active = m.cursor
			return m, m.ensureLoad(m.active)
		case "esc":
			m.active, m.cursor = ScreenDashboard, ScreenDashboard
			return m, m.ensureLoad(ScreenDashboard)
		case "r":
			delete(m.bodies, m.active)
			m.loading[m.active] = true
			return m, loadCmd(m.active)
		}
	}
	return m, nil
}

// View renders header, menu + content, and footer. Narrow terminals
// stack the menu above the content instead of side-by-side.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.theme.title.Render("NEXUS"))
	b.WriteString("\n")
	b.WriteString(m.theme.subtitle.Render("Developer Command Center"))
	b.WriteString("\n\n")

	body, ok := m.bodies[m.active]
	if !ok {
		body = "Loading…"
	}
	content := m.theme.box.Render(
		m.theme.header.Render(m.active.Title()) + "\n\n" + body,
	)

	if m.width > 0 && m.width < 75 {
		b.WriteString(m.menuView() + "\n\n" + content)
	} else {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.menuView(), "  "+content))
	}

	b.WriteString("\n\n")
	b.WriteString(m.theme.footer.Render("↑/↓ navigate · Enter open · r refresh · Esc home · q quit"))
	if !m.ready {
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) menuView() string {
	var items []string
	for s := ScreenDashboard; s < ScreenCount; s++ {
		label := s.Title()
		if m.loading[s] {
			label += " …"
		}
		if s == m.cursor {
			items = append(items, m.theme.selected.Render("> "+label))
		} else {
			items = append(items, m.theme.normal.Render("  "+label))
		}
	}
	return strings.Join(items, "\n")
}
