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
		if msg.String() == "esc" {
			return m, func() tea.Msg { return messages.NavigateToOverviewMsg{} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m DashboardModel) View() string {
	divider := strings.Repeat("─", m.width)
	var sb strings.Builder
	sb.WriteString("\n  DASHBOARD (Phase 5)\n")
	sb.WriteString("  " + divider[:min(18, len(divider))] + "\n\n")
	if m.gs != nil {
		p := m.gs.Player
		sb.WriteString(fmt.Sprintf("  Level:          %d\n", p.Level))
		sb.WriteString(fmt.Sprintf("  XP:             %d\n", p.XP))
		sb.WriteString(fmt.Sprintf("  General Coins:  %.2f GC\n", p.GeneralCoins))
		sb.WriteString(fmt.Sprintf("  Total Clicks:   %d\n", p.TotalClicks))
		sb.WriteString(fmt.Sprintf("  Time Played:    %.0fs\n", p.TotalPlaySeconds))
	}
	content := sb.String()
	// Fill available height; footer line pinned at bottom via Height.
	body := lipgloss.NewStyle().Width(m.width).Height(m.height - 2).Render(content)
	return body + "\n" + divider + "\n  [Esc] Back to Overview"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
