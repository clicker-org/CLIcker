package components

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// ModalCloseMsg is emitted by BaseModal when the user confirms the [Esc] button.
type ModalCloseMsg struct{}

// BaseModal is the shared foundation for all modal overlays.
// It renders a bordered box with a focusable [Esc] button in the top-left
// corner and composites it centered over a pre-rendered background.
// Callers supply the box dimensions explicitly; the concrete modal types
// (TabModal, ConfirmModal) choose appropriate sizes for their use case.
//
// Navigation:
//   - NavUpMsg                       → focus the [Esc] button
//   - NavDownMsg                     → unfocus the [Esc] button
//   - NavConfirmMsg (when [Esc] focused) → emit ModalCloseMsg
type BaseModal struct {
	escFocused bool
	t          theme.Theme
}

// NewBaseModal returns a fresh BaseModal with no focused state.
func NewBaseModal(t theme.Theme) BaseModal {
	return BaseModal{t: t}
}

// EscFocused reports whether the [Esc] border button currently has focus.
func (m BaseModal) EscFocused() bool { return m.escFocused }

// Update handles the navigation messages that control the [Esc] button focus.
func (m BaseModal) Update(msg tea.Msg) (BaseModal, tea.Cmd) {
	switch msg.(type) {
	case messages.NavUpMsg:
		m.escFocused = true
	case messages.NavDownMsg:
		m.escFocused = false
	case messages.NavConfirmMsg:
		if m.escFocused {
			m.escFocused = false
			return m, func() tea.Msg { return ModalCloseMsg{} }
		}
	}
	return m, nil
}

// View renders a box of boxWidth×boxHeight centered over bgContent.
// bgContent must be a pre-rendered string of exactly outerWidth×outerHeight cells.
func (m BaseModal) View(title, content, bgContent string, outerWidth, outerHeight, boxWidth, boxHeight int) string {
	box := m.renderBox(title, content, boxWidth, boxHeight)
	return overlayOnBackground(box, bgContent, m.t.Background(), outerWidth, outerHeight)
}

// overlayOnBackground centers box (a multi-line pre-rendered string) over
// bgContent (a pre-rendered width×height terminal string) and returns the
// composite result. bgColor is used to fill any gap in the right margin.
func overlayOnBackground(box, bgContent, bgColor string, width, height int) string {
	boxLines := strings.Split(box, "\n")
	boxH := len(boxLines)
	boxW := lipgloss.Width(boxLines[0])

	bgLines := strings.Split(bgContent, "\n")
	for len(bgLines) < height {
		bgLines = append(bgLines, strings.Repeat(" ", width))
	}
	bgLines = bgLines[:height]

	startY := max((height-boxH)/2, 0)
	startX := max((width-boxW)/2, 0)

	result := make([]string, height)
	for i, bgLine := range bgLines {
		oi := i - startY
		if oi < 0 || oi >= boxH {
			result[i] = bgLine
			continue
		}
		modalLine := boxLines[oi]

		leftPart := ansi.Truncate(bgLine, startX, "")
		leftVisual := lipgloss.Width(leftPart)
		if leftVisual < startX {
			leftPart += strings.Repeat(" ", startX-leftVisual)
		}

		rightStart := startX + boxW
		rightPart := ansi.Cut(bgLine, rightStart, width)
		rightVisual := lipgloss.Width(rightPart)
		rightWidth := max(width-rightStart, 0)
		if rightVisual < rightWidth {
			rightPart += lipgloss.NewStyle().
				Background(lipgloss.Color(bgColor)).
				Render(strings.Repeat(" ", rightWidth-rightVisual))
		}

		result[i] = leftPart + modalLine + rightPart
	}

	return strings.Join(result, "\n")
}

// renderBox builds the raw bordered box string at the given dimensions.
func (m BaseModal) renderBox(title, content string, boxWidth, boxHeight int) string {
	borderColor := m.t.BorderColor()
	if borderColor == "" {
		borderColor = "#ffffff"
	}
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))

	innerWidth := boxWidth - 2
	innerHeight := boxHeight - 2

	// [Esc] button: reversed when focused.
	var escLabel string
	if m.escFocused {
		escLabel = borderStyle.Render("[") +
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.t.Background())).
				Background(lipgloss.Color(borderColor)).
				Bold(true).
				Render("Esc") +
			borderStyle.Render("]")
	} else {
		escLabel = borderStyle.Render("[Esc]")
	}
	escW := lipgloss.Width(escLabel) // always 5

	titleStr := " " + title + " "
	dashCount := max(innerWidth-escW-lipgloss.Width(titleStr), 0)
	topBorder := borderStyle.Render("┌") + escLabel +
		borderStyle.Render(strings.Repeat("─", dashCount)+titleStr+"┐")

	bottomBorder := borderStyle.Render("└" + strings.Repeat("─", innerWidth) + "┘")

	contentLines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	side := borderStyle.Render("│")
	rows := make([]string, 0, boxHeight)
	rows = append(rows, topBorder)
	for i := range innerHeight {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}
		pad := max(innerWidth-lipgloss.Width(line), 0)
		rows = append(rows, side+line+strings.Repeat(" ", pad)+side)
	}
	rows = append(rows, bottomBorder)

	return strings.Join(rows, "\n")
}
