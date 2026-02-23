package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// DashboardModel is the statistics dashboard screen.
type DashboardModel struct {
	t      theme.Theme
	gs     *gamestate.GameState
	width  int
	height int
}

// NewDashboardModel creates a DashboardModel.
func NewDashboardModel(t theme.Theme, gs *gamestate.GameState, width, height int) DashboardModel {
	return DashboardModel{t: t, gs: gs, width: width, height: height}
}

func (m DashboardModel) Init() tea.Cmd { return nil }

func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return messages.NavigateToOverviewMsg{} }
		case "a", "A":
			return m, func() tea.Msg { return messages.NavigateToAchievementsMsg{} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m DashboardModel) View() string {
	bg := lipgloss.Color(m.t.Background())
	fg := lipgloss.Color(m.t.PrimaryText())
	dimFg := lipgloss.Color(m.t.DimText())
	borderFg := lipgloss.Color(m.t.BorderColor())

	dividerStr := strings.Repeat("─", m.width)
	divider := lipgloss.NewStyle().Width(m.width).Background(bg).Foreground(borderFg).Render(dividerStr)

	var sb strings.Builder
	sb.WriteString("\n  DASHBOARD (Phase 4)\n")
	sb.WriteString("  " + dividerStr[:min(18, len(dividerStr))] + "\n\n")
	if m.gs != nil {
		p := m.gs.Player
		sb.WriteString(fmt.Sprintf("  Level:          %d\n", p.Level))
		sb.WriteString(fmt.Sprintf("  XP:             %d\n", p.XP))
		sb.WriteString(fmt.Sprintf("  General Coins:  %.2f GC\n", p.GeneralCoins))
		sb.WriteString(fmt.Sprintf("  Total Clicks:   %d\n", p.TotalClicks))
		sb.WriteString(fmt.Sprintf("  Time Played:    %.0fs\n", p.TotalPlaySeconds))
	}
	body := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 2).
		Background(bg).
		Foreground(fg).
		Render(sb.String())

	helpLine := lipgloss.NewStyle().Width(m.width).Background(bg).Foreground(dimFg).Render("  [Esc] Back to Overview   [A] Achievements")
	return body + "\n" + divider + "\n" + helpLine
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
