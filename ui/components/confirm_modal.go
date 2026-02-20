package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// ConfirmMsg is sent when the user selects an option in the confirm dialog.
type ConfirmMsg struct{ Confirmed bool }

// ConfirmModal is a centered two-button dialog overlay.
//
// Navigation matches the rest of the UI:
//
//	NavLeftMsg    → focus the Cancel button
//	NavRightMsg   → focus the Confirm button
//	NavConfirmMsg → emit ConfirmMsg with the current selection
//	Esc           → handled by the caller as a cancel shortcut
type ConfirmModal struct {
	t              theme.Theme
	label          string // inner text of the confirm button, e.g. "Quit", "Prestige"
	confirmFocused bool   // false = Cancel focused (default), true = Confirm focused
}

// NewConfirmModal returns a ConfirmModal with Cancel pre-focused.
// label is the inner text shown on the confirm button (e.g. "Quit", "Confirm").
func NewConfirmModal(t theme.Theme, label string) ConfirmModal {
	return ConfirmModal{t: t, label: label}
}

// Update handles navigation between buttons and emits ConfirmMsg on selection.
func (m ConfirmModal) Update(msg tea.Msg) (ConfirmModal, tea.Cmd) {
	switch msg.(type) {
	case messages.NavLeftMsg:
		m.confirmFocused = false
	case messages.NavRightMsg:
		m.confirmFocused = true
	case messages.NavConfirmMsg:
		confirmed := m.confirmFocused
		return m, func() tea.Msg { return ConfirmMsg{Confirmed: confirmed} }
	}
	return m, nil
}

// View renders the confirm dialog centered over bgContent.
// bgContent must be a pre-rendered string of exactly width×height cells.
func (m ConfirmModal) View(title, question, bgContent string, width, height int) string {
	box := m.renderBox(title, question, width)
	return overlayOnBackground(box, bgContent, m.t.Background(), width, height)
}

// renderBox builds the raw confirm dialog box string.
//
// Layout (7 rows):
//
//	┌───────── Title ────────┐
//	│                        │
//	│         question       │
//	│                        │
//	│  [ Cancel ] [Confirm]  │
//	│                        │
//	└────────────────────────┘
func (m ConfirmModal) renderBox(title, question string, width int) string {
	const borderColor = "#ffffff"
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))

	modalWidth := max(width/2, 44)
	modalWidth = min(modalWidth, width-4)
	innerWidth := modalWidth - 2

	// Top border with centered title.
	titleStr := " " + title + " "
	titleW := lipgloss.Width(titleStr)
	dashes := max(innerWidth-titleW, 0)
	leftDash := dashes / 2
	rightDash := dashes - leftDash
	topBorder := borderStyle.Render(
		"┌" + strings.Repeat("─", leftDash) + titleStr + strings.Repeat("─", rightDash) + "┐",
	)
	bottomBorder := borderStyle.Render("└" + strings.Repeat("─", innerWidth) + "┘")

	side := borderStyle.Render("│")

	// Question line, centered.
	questionW := lipgloss.Width(question)
	qPad := max(innerWidth-questionW, 0)
	qLeft := qPad / 2
	qRight := qPad - qLeft
	questionLine := side + strings.Repeat(" ", qLeft) + question + strings.Repeat(" ", qRight) + side

	// Button labels.
	const cancelLabel = "[ Cancel ]"
	confirmLabel := "[ " + m.label + " ]"
	const buttonGap = "   "

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.t.DimText()))
	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.t.AccentColor())).
		Bold(true).
		Underline(true)

	var cancelStr, confirmStr string
	if m.confirmFocused {
		cancelStr = dimStyle.Render(cancelLabel)
		confirmStr = activeStyle.Render(confirmLabel)
	} else {
		cancelStr = activeStyle.Render(cancelLabel)
		confirmStr = dimStyle.Render(confirmLabel)
	}

	totalButtonW := lipgloss.Width(cancelLabel) + lipgloss.Width(buttonGap) + lipgloss.Width(confirmLabel)
	bPad := max(innerWidth-totalButtonW, 0)
	bLeft := bPad / 2
	bRight := bPad - bLeft
	buttonLine := side +
		strings.Repeat(" ", bLeft) + cancelStr + buttonGap + confirmStr + strings.Repeat(" ", bRight) +
		side

	blank := side + strings.Repeat(" ", innerWidth) + side

	return strings.Join([]string{
		topBorder,
		blank,
		questionLine,
		blank,
		buttonLine,
		blank,
		bottomBorder,
	}, "\n")
}
