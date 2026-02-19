package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/offline"
	"github.com/clicker-org/clicker/internal/world"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// OfflineReportModel shows the offline income popup on game launch.
type OfflineReportModel struct {
	t        theme.Theme
	result   offline.Result
	visible  bool
	boxStyle lipgloss.Style
}

// NewOfflineReportModel creates an OfflineReportModel. The report is shown
// only if the player was away for at least offline.MinReportDuration.
func NewOfflineReportModel(t theme.Theme, result offline.Result) OfflineReportModel {
	return OfflineReportModel{
		t:       t,
		result:  result,
		visible: result.Duration >= offline.MinReportDuration,
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(t.AccentColor())).
			Padding(1, 2).
			Width(44),
	}
}

// IsVisible returns true if the report should be shown.
func (m OfflineReportModel) IsVisible() bool { return m.visible }

func (m OfflineReportModel) Init() tea.Cmd { return nil }

func (m OfflineReportModel) Update(msg tea.Msg) (OfflineReportModel, tea.Cmd) {
	dismiss := func() tea.Cmd {
		m.visible = false
		return func() tea.Msg { return messages.OfflineReportDismissedMsg{} }
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return m, dismiss()
		}
	case messages.NavConfirmMsg:
		return m, dismiss()
	}
	return m, nil
}

func (m OfflineReportModel) View() string {
	if !m.visible {
		return ""
	}

	hours := int(m.result.Duration.Hours())
	mins := int(m.result.Duration.Minutes()) % 60

	var sb strings.Builder
	sb.WriteString("         WELCOME BACK!\n\n")
	sb.WriteString(fmt.Sprintf("  You were away for: %dh %dm\n\n", hours, mins))

	if m.result.WorldID != "" {
		sb.WriteString(fmt.Sprintf("  While offline in %s:\n", m.result.WorldID))
		if m.result.WorldCoins > 0 {
			coinName := m.result.WorldID
			if w, ok := world.DefaultRegistry.Get(m.result.WorldID); ok {
				coinName = w.CoinName()
			}
			sb.WriteString(fmt.Sprintf("  + %.0f %s\n\n", m.result.WorldCoins, coinName))
		} else {
			sb.WriteString("  No passive income yet. Buy upgrades!\n\n")
		}
	}
	if m.result.GeneralCoins > 0 {
		sb.WriteString(fmt.Sprintf("  + %.2f GC (overview trickle)\n\n", m.result.GeneralCoins))
	}
	sb.WriteString("  [Enter] Continue")

	return m.boxStyle.Render(sb.String())
}
