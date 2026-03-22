package tui

import (
	"fmt"
	"strings"

	"github.com/andragon31/fenrir/internal/graph"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	fenrirGreen  = lipgloss.Color("#4ade80")
	fenrirYellow = lipgloss.Color("#fbbf24")
	fenrirRed    = lipgloss.Color("#f87171")
	fenrirBlue   = lipgloss.Color("#60a5fa")
	fenrirGray   = lipgloss.Color("#6b7280")
	fenrirDark   = lipgloss.Color("#1f2937")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(fenrirGreen)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(fenrirBlue)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5e7eb"))

	dimStyle = lipgloss.NewStyle().
			Foreground(fenrirGray)

	errorStyle = lipgloss.NewStyle().
			Foreground(fenrirRed)

	warningStyle = lipgloss.NewStyle().
			Foreground(fenrirYellow)
)

type Model struct {
	graph     *graph.Graph
	view      string
	cursor    int
	sessions  []graph.Session
	nodes     []graph.Node
	drifts    []graph.DriftScore
	stats     *graph.Stats
	searchQuery string
	err       error
	quitting  bool
}

func NewModel(g *graph.Graph) *Model {
	return &Model{
		graph: g,
		view:  "dashboard",
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		stats, err := m.graph.GetStats()
		if err != nil {
			return errMsg{err}
		}
		return statsMsg{stats}
	}
}

type errMsg struct {
	err error
}

type statsMsg struct {
	stats *graph.Stats
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			m.cursor++
			if m.cursor >= len(m.nodes) {
				m.cursor = 0
			}
		case "k", "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.nodes) - 1
			}
		case "enter":
			m.selectItem()
		case "1":
			m.view = "dashboard"
			m.loadDashboard()
		case "2":
			m.view = "sessions"
			m.loadSessions()
		case "3":
			m.view = "drift"
			m.loadDrift()
		case "4":
			m.view = "search"
		case "5":
			m.view = "stats"
			m.loadStats()
		case "0":
			m.view = "help"
		}
	case statsMsg:
		m.stats = msg.stats
	case errMsg:
		m.err = msg.err
	}
	return m, nil
}

func (m *Model) selectItem() {
	if m.view == "sessions" && len(m.sessions) > m.cursor {
		m.view = "session-detail"
	}
}

func (m *Model) loadDashboard() {
	stats, err := m.graph.GetStats()
	if err != nil {
		m.err = err
		return
	}
	m.stats = stats
}

func (m *Model) loadSessions() {
	sessions, err := m.graph.ListSessions(50)
	if err != nil {
		m.err = err
		return
	}
	m.sessions = sessions
	m.cursor = 0
}

func (m *Model) loadDrift() {
	drifts, err := m.graph.GetDriftScores("")
	if err != nil {
		m.err = err
		return
	}
	m.drifts = drifts
	m.cursor = 0
}

func (m *Model) loadStats() {
	m.loadDashboard()
}

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("╔══════════════════════════════════════════╗\n"))
	s.WriteString(titleStyle.Render("║           FENRIR - AI Governance         ║\n"))
	s.WriteString(titleStyle.Render("╚══════════════════════════════════════════╝\n\n"))

	s.WriteString(dimStyle.Render("[1] Dashboard  [2] Sessions  [3] Drift  [4] Search  [5] Stats  [0] Help  [q] Quit\n\n"))

	switch m.view {
	case "dashboard":
		s.WriteString(m.dashboardView())
	case "sessions":
		s.WriteString(m.sessionsView())
	case "drift":
		s.WriteString(m.driftView())
	case "search":
		s.WriteString(m.searchView())
	case "stats":
		s.WriteString(m.statsView())
	case "help":
		s.WriteString(m.helpView())
	}

	return s.String()
}

func (m *Model) dashboardView() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("┌─ Dashboard ─────────────────────────────────┐\n"))

	if m.stats != nil {
		s.WriteString(fmt.Sprintf("│  Nodes:       %-30d│\n", m.stats.TotalNodes))
		s.WriteString(fmt.Sprintf("│  Edges:       %-30d│\n", m.stats.TotalEdges))
		s.WriteString(fmt.Sprintf("│  Sessions:    %-30d│\n", m.stats.TotalSessions))
		s.WriteString(fmt.Sprintf("│  Active:      %-30d│\n", m.stats.ActiveSessions))
		s.WriteString(fmt.Sprintf("│  Decisions:   %-30d│\n", m.stats.TotalDecisions))
		s.WriteString(fmt.Sprintf("│  Audit Logs:  %-30d│\n", m.stats.AuditEntries))
	} else {
		s.WriteString("│  Loading...                               │\n")
	}

	s.WriteString("└──────────────────────────────────────────────┘\n")

	return s.String()
}

func (m *Model) sessionsView() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("┌─ Sessions ──────────────────────────────────┐\n"))

	if len(m.sessions) == 0 {
		s.WriteString("│  No sessions found                          │\n")
	} else {
		for i, session := range m.sessions {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}

			status := normalStyle.Render("active  ")
			if session.Status == "closed" {
				status = warningStyle.Render("closed  ")
			}

			goal := session.Goal
			if len(goal) > 35 {
				goal = goal[:32] + "..."
			}

			s.WriteString(fmt.Sprintf("│ %s [%s] %-35s│\n", cursor, status, goal))
		}
	}

	s.WriteString("└──────────────────────────────────────────────┘\n")
	s.WriteString(dimStyle.Render("j/k: navigate  Enter: view details\n"))

	return s.String()
}

func (m *Model) driftView() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("┌─ Drift Scores ──────────────────────────────┐\n"))

	if len(m.drifts) == 0 {
		s.WriteString("│  No drift data available                     │\n")
	} else {
		for _, drift := range m.drifts {
			bar := m.driftBar(drift.Score)
			_ = fenrirGreen // Color placeholder for future styling
			if drift.Score > 0.6 {
				bar = "⚠️ " + bar
			}

			s.WriteString(fmt.Sprintf("│ %-20s %s %.2f    │\n", drift.Module, bar, drift.Score))
		}
	}

	s.WriteString("└──────────────────────────────────────────────┘\n")

	return s.String()
}

func (m *Model) driftBar(score float64) string {
	filled := int(score * 20)
	empty := 20 - filled
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	return bar
}

func (m *Model) searchView() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("┌─ Search ────────────────────────────────────┐\n"))
	s.WriteString(fmt.Sprintf("│ Query: %-35s│\n", m.searchQuery))
	s.WriteString("│                                               │\n")
	s.WriteString("│ Type your search query and press Enter        │\n")
	s.WriteString("└──────────────────────────────────────────────┘\n")

	return s.String()
}

func (m *Model) statsView() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("┌─ Statistics ────────────────────────────────┐\n"))

	if m.stats != nil {
		s.WriteString(fmt.Sprintf("│  Total Nodes:       %-25d│\n", m.stats.TotalNodes))
		s.WriteString(fmt.Sprintf("│  Total Edges:       %-25d│\n", m.stats.TotalEdges))
		s.WriteString(fmt.Sprintf("│  Total Sessions:    %-25d│\n", m.stats.TotalSessions))
		s.WriteString(fmt.Sprintf("│  Active Sessions:   %-25d│\n", m.stats.ActiveSessions))
		s.WriteString(fmt.Sprintf("│  Decisions:         %-25d│\n", m.stats.TotalDecisions))
		s.WriteString(fmt.Sprintf("│  Audit Entries:    %-25d│\n", m.stats.AuditEntries))
		s.WriteString(fmt.Sprintf("│  Cached Packages:  %-25d│\n", m.stats.CachedPackages))
	} else {
		s.WriteString("│  Loading...                               │\n")
	}

	s.WriteString("└──────────────────────────────────────────────┘\n")

	return s.String()
}

func (m *Model) helpView() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("┌─ Help ──────────────────────────────────────┐\n"))
	s.WriteString("│  1 - Dashboard                                 │\n")
	s.WriteString("│  2 - Sessions                                  │\n")
	s.WriteString("│  3 - Drift Scores                              │\n")
	s.WriteString("│  4 - Search                                    │\n")
	s.WriteString("│  5 - Statistics                                │\n")
	s.WriteString("│  0 - Help                                      │\n")
	s.WriteString("│                                               │\n")
	s.WriteString("│  j/k - Navigate                                │\n")
	s.WriteString("│  Enter - Select                                 │\n")
	s.WriteString("│  q - Quit                                      │\n")
	s.WriteString("└──────────────────────────────────────────────┘\n")

	return s.String()
}

func NewProgram(model *Model) *tea.Program {
	return tea.NewProgram(model, tea.WithAltScreen())
}

func Run(g *graph.Graph) error {
	model := NewModel(g)
	_, err := NewProgram(model).Run()
	return err
}
