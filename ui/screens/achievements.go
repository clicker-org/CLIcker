package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/achievement"
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
	cursor int
	scroll int
}

// NewAchievementsModel creates an AchievementsModel.
func NewAchievementsModel(t theme.Theme, eng *engine.Engine, width, height int) AchievementsModel {
	return AchievementsModel{t: t, eng: eng, width: width, height: height}
}

func (m AchievementsModel) Init() tea.Cmd { return nil }

func (m AchievementsModel) Update(msg tea.Msg) (AchievementsModel, tea.Cmd) {
	total := 0
	if m.eng != nil && m.eng.AchievReg != nil {
		total = m.eng.AchievReg.Total()
	}
	if total > 0 && m.cursor >= total {
		m.cursor = total - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return messages.NavigateToOverviewMsg{} }
		case "d", "D":
			return m, func() tea.Msg { return messages.NavigateToDashboardMsg{} }
		}
	case messages.NavUpMsg:
		if total == 0 {
			return m, nil
		}
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scroll {
				m.scroll = m.cursor
			}
		}
	case messages.NavDownMsg:
		if total == 0 {
			return m, nil
		}
		if m.cursor < total-1 {
			m.cursor++
			vis := m.visibleCount()
			if m.cursor >= m.scroll+vis {
				m.scroll = m.cursor - vis + 1
			}
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
	all := []achievement.Achievement{}
	if m.eng != nil && m.eng.AchievReg != nil {
		total = m.eng.AchievReg.Total()
		all = m.eng.AchievReg.GetAll()
		for _, a := range all {
			if m.eng.Earned[a.ID] {
				unlocked++
			}
		}
	}

	progress := 0.0
	if total > 0 {
		progress = float64(unlocked) / float64(total)
	}

	contentW := min(max(m.width-8, 60), 110)
	if contentW > m.width {
		contentW = m.width
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Width(contentW).
		Align(lipgloss.Center).
		Foreground(accent).
		Bold(true).
		Render("ACHIEVEMENTS"))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Width(contentW).
		Align(lipgloss.Center).
		Foreground(dimFg).
		Render(fmt.Sprintf("Unlocked %d / %d", unlocked, total)))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Width(contentW).
		Align(lipgloss.Center).
		Foreground(dimFg).
		Render(renderAchievementProgress(progress)))
	sb.WriteString("\n\n")

	if total == 0 {
		sb.WriteString(lipgloss.NewStyle().
			Width(contentW).
			Align(lipgloss.Center).
			Foreground(dimFg).
			Render("No achievements registered yet.\n"))
	} else {
		vis := m.visibleCount()
		end := min(m.scroll+vis, len(all))
		if m.scroll > 0 {
			sb.WriteString(lipgloss.NewStyle().Width(contentW).Foreground(dimFg).Render("↑ more above"))
			sb.WriteString("\n")
		}
		for i := m.scroll; i < end; i++ {
			a := all[i]
			sb.WriteString(m.renderAchievementCard(a, m.eng.Earned[a.ID], i == m.cursor, contentW))
			if i < end-1 {
				sb.WriteString("\n")
			}
		}
		if end < len(all) {
			sb.WriteString("\n")
			sb.WriteString(lipgloss.NewStyle().Width(contentW).Foreground(dimFg).Render("↓ more below"))
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
		Render("  [↑/↓] or [j/k] Navigate   [Esc] Back to Overview   [D] Dashboard")

	return body + "\n" + divider + "\n" + helpLine
}

func (m AchievementsModel) renderAchievementCard(a achievement.Achievement, passed, selected bool, contentW int) string {
	dim := lipgloss.Color(m.t.DimText())
	primary := lipgloss.Color(m.t.PrimaryText())
	success := lipgloss.Color(m.t.SuccessColor())
	errorC := lipgloss.Color(m.t.ErrorColor())
	accent := lipgloss.Color(m.t.AccentColor())

	statusText := "[NOT PASSED]"
	statusColor := errorC
	borderColor := m.t.BorderColor()
	if passed {
		statusText = "[PASSED]"
		statusColor = success
		borderColor = m.t.SuccessColor()
	}
	if selected {
		borderColor = m.t.AccentColor()
	}

	borderSt := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))
	top := borderSt.Render("┌" + strings.Repeat("─", contentW) + "┐")
	bot := borderSt.Render("└" + strings.Repeat("─", contentW) + "┘")
	side := borderSt.Render("│")

	statusRender := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText)
	nameStyle := lipgloss.NewStyle().Foreground(primary).Bold(true)
	if selected {
		nameStyle = nameStyle.Foreground(accent)
	}
	nameRender := nameStyle.Render(a.Name)
	leftW := max(contentW-14, 20)
	row1 := achPadVisual(statusRender+" "+nameRender, leftW) + achPadVisual(lipgloss.NewStyle().Foreground(accent).Render(fmt.Sprintf("XP %+d", a.XPGrant)), contentW-leftW)
	row2 := lipgloss.NewStyle().Foreground(dim).Render(achTruncStr(a.Description, contentW))
	row3 := lipgloss.NewStyle().Foreground(dim).Render("ID: " + a.ID)

	makeRow := func(content string) string {
		visW := lipgloss.Width(content)
		pad := max(contentW-visW, 0)
		return side + content + strings.Repeat(" ", pad) + side
	}

	return strings.Join([]string{
		top,
		makeRow(row1),
		makeRow(row2),
		makeRow(row3),
		bot,
	}, "\n")
}

// visibleCount returns how many achievement cards fit in the available body.
func (m AchievementsModel) visibleCount() int {
	// Reserve rows for header + footer labels in the body.
	available := m.height - 14
	if available < 5 {
		return 1
	}
	// Card is 5 lines, with 1 separator line between cards.
	count := available / 6
	if count < 1 {
		count = 1
	}
	return count
}

func achTruncStr(s string, maxW int) string {
	r := []rune(s)
	if len(r) <= maxW {
		return s
	}
	if maxW <= 1 {
		return "…"
	}
	return string(r[:maxW-1]) + "…"
}

func achPadVisual(s string, w int) string {
	v := lipgloss.Width(s)
	if v >= w {
		return s
	}
	return s + strings.Repeat(" ", w-v)
}

func renderAchievementProgress(p float64) string {
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	const width = 24
	filled := int(p * width)
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}
