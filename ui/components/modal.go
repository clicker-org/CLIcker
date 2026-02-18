package components

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// ModalCloseMsg is emitted by Modal when it requests to be closed
// (user pressed Enter on the focused [Esc] button).
type ModalCloseMsg struct{}

// Modal is a centered overlay box with a white border and a focusable [Esc]
// button embedded in the top-left corner of the border.
//
// Navigation:
//   - NavUpMsg    → move focus to the [Esc] button
//   - NavDownMsg  → move focus away from the [Esc] button
//   - NavConfirmMsg (when Esc focused) → emit ModalCloseMsg
type Modal struct {
	escFocused bool
	t          theme.Theme
}

// NewModal returns a fresh Modal with no focused state.
func NewModal(t theme.Theme) Modal {
	return Modal{t: t}
}

// Update handles the navigation messages that control the [Esc] button focus.
func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
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

// View renders the modal overlaid on top of bgContent.
// bgContent must be a pre-rendered string of exactly width×height cells.
// The modal is centered; rows above/below and columns to the left of the modal
// show the live background content.
func (m Modal) View(title, content, bgContent string, width, height int) string {
	box := m.renderBox(title, content, width, height)
	boxLines := strings.Split(box, "\n")
	boxH := len(boxLines)
	boxW := lipgloss.Width(boxLines[0])

	bgLines := strings.Split(bgContent, "\n")
	// Ensure bgLines has exactly height rows.
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
			// Above or below the modal: show the full background row.
			result[i] = bgLine
			continue
		}
		// This row intersects the modal.
		modalLine := boxLines[oi]

		// Left portion: first startX visual columns from the background.
		leftPart := ansi.Truncate(bgLine, startX, "")
		leftVisual := lipgloss.Width(leftPart)
		if leftVisual < startX {
			leftPart += strings.Repeat(" ", startX-leftVisual)
		}

		// Right portion: solid background fill.
		// (Extracting the right side of an ANSI string requires a full ANSI parser;
		// background-color fill is the practical alternative.)
		rightWidth := max(width-startX-boxW, 0)
		var rightPart string
		if rightWidth > 0 {
			rightPart = lipgloss.NewStyle().
				Background(lipgloss.Color(m.t.Background())).
				Render(strings.Repeat(" ", rightWidth))
		}

		result[i] = leftPart + modalLine + rightPart
	}

	return strings.Join(result, "\n")
}

// renderBox builds the raw modal box string (no placement/padding).
func (m Modal) renderBox(title, content string, width, height int) string {
	const borderColor = "#ffffff"
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))

	modalWidth := max(width*3/4, 40)
	modalWidth = min(modalWidth, width-4)
	modalHeight := max(height-4, 6)

	innerWidth := modalWidth - 2
	innerHeight := modalHeight - 2

	// [Esc] button: reversed (dark text on white bg) when focused.
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
	escVisualLen := lipgloss.Width(escLabel) // always 5, ANSI-stripped

	// Top border: ┌[Esc]──────── TITLE ─┐
	titleStr := " " + title + " "
	dashCount := max(innerWidth-escVisualLen-len(titleStr), 0)
	topBorder := borderStyle.Render("┌") + escLabel +
		borderStyle.Render(strings.Repeat("─", dashCount)+titleStr+"┐")

	// Bottom border: └──────────────────┘
	bottomBorder := borderStyle.Render("└" + strings.Repeat("─", innerWidth) + "┘")

	// Interior rows padded to innerWidth with side borders.
	contentLines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	side := borderStyle.Render("│")
	rows := make([]string, 0, modalHeight)
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
