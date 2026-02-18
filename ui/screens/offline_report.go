package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// OfflineReportModel shows the offline income popup on game launch.
type OfflineReportModel struct {
	t           theme.Theme
	worldID     string
	timeOffline time.Duration
	coinsEarned float64
	gcEarned    float64
	visible     bool
	boxStyle    lipgloss.Style
}

// NewOfflineReportModel creates an OfflineReportModel.
func NewOfflineReportModel(
	t theme.Theme,
	worldID string,
	timeOffline time.Duration,
	coinsEarned, gcEarned float64,
	visible bool,
) OfflineReportModel {
	return OfflineReportModel{
		t:           t,
		worldID:     worldID,
		timeOffline: timeOffline,
		coinsEarned: coinsEarned,
		gcEarned:    gcEarned,
		visible:     visible,
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc":
			m.visible = false
			return m, func() tea.Msg { return messages.OfflineReportDismissedMsg{} }
		}
	}
	return m, nil
}

func (m OfflineReportModel) View() string {
	if !m.visible {
		return ""
	}

	hours := int(m.timeOffline.Hours())
	mins := int(m.timeOffline.Minutes()) % 60

	var sb strings.Builder
	sb.WriteString("         WELCOME BACK!\n\n")
	sb.WriteString(fmt.Sprintf("  You were away for: %dh %dm\n\n", hours, mins))

	if m.worldID != "" && m.coinsEarned > 0 {
		sb.WriteString(fmt.Sprintf("  While offline in [%s]:\n", m.worldID))
		sb.WriteString(fmt.Sprintf("  + %.0f TC\n\n", m.coinsEarned))
	}
	if m.gcEarned > 0 {
		sb.WriteString(fmt.Sprintf("  + %.2f GC (overview trickle)\n\n", m.gcEarned))
	}
	sb.WriteString("  [Enter] Continue")

	return m.boxStyle.Render(sb.String())
}
