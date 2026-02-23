package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// AchievementsModel is the global achievements screen.
type AchievementsModel struct {
	t      theme.Theme
	eng    *engine.Engine
	width  int
	height int
}

// NewAchievementsModel creates an AchievementsModel.
func NewAchievementsModel(t theme.Theme, eng *engine.Engine, width, height int) AchievementsModel {
	return AchievementsModel{t: t, eng: eng, width: width, height: height}
}

func (m AchievementsModel) Init() tea.Cmd { return nil }

func (m AchievementsModel) Update(msg tea.Msg) (AchievementsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return messages.NavigateToOverviewMsg{} }
		case "d", "D":
			return m, func() tea.Msg { return messages.NavigateToDashboardMsg{} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m AchievementsModel) View() string {
	bg := lipgloss.Color(m.t.Background())
	fg := lipgloss.Color(m.t.PrimaryText())
	dimFg := lipgloss.Color(m.t.DimText())
	accent := lipgloss.Color(m.t.AccentColor())
	borderFg := lipgloss.Color(m.t.BorderColor())

	dividerStr := strings.Repeat("─", max(m.width, 1))
	divider := lipgloss.NewStyle().Width(m.width).Background(bg).Foreground(borderFg).Render(dividerStr)

	total := 0
	unlocked := 0
	if m.eng != nil && m.eng.AchievReg != nil {
		total = m.eng.AchievReg.Total()
		for _, a := range m.eng.AchievReg.GetAll() {
			if m.eng.Earned[a.ID] {
				unlocked++
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("\n  ACHIEVEMENTS\n")
	sb.WriteString("  " + dividerStr[:min(21, len(dividerStr))] + "\n\n")
	sb.WriteString(
		lipgloss.NewStyle().Foreground(accent).Render(
			fmt.Sprintf("  Unlocked: %d / %d\n\n", unlocked, total),
		),
	)

	if total == 0 {
		sb.WriteString("  No achievements registered yet.\n")
	} else {
		for _, a := range m.eng.AchievReg.GetAll() {
			state := "[ ]"
			name := a.Name
			desc := a.Description
			if m.eng.Earned[a.ID] {
				state = "[x]"
			} else if a.Hidden {
				name = "???"
				desc = "Hidden achievement"
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", state, name))
			sb.WriteString(lipgloss.NewStyle().Foreground(dimFg).Render("      " + desc + "\n"))
		}
	}

	body := lipgloss.NewStyle().
		Width(m.width).
		Height(max(m.height-2, 1)).
		Background(bg).
		Foreground(fg).
		Render(sb.String())

	helpLine := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Foreground(dimFg).
		Render("  [Esc] Back to Overview   [D] Dashboard")

	return body + "\n" + divider + "\n" + helpLine
}
