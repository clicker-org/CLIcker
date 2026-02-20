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

// ConfirmModal is a compact two-button dialog overlay built on BaseModal.
// It sizes itself to roughly half the outer width with a fixed 7-row height,
// keeping it visually distinct from the larger TabModal.
//
// Navigation:
//
//	NavUpMsg      → focus the [Esc] button (cancels on Enter)
//	NavDownMsg    → unfocus the [Esc] button
//	NavLeftMsg    → focus the Confirm button (when [Esc] not focused)
//	NavRightMsg   → focus the Cancel button (when [Esc] not focused)
//	NavConfirmMsg → emit ConfirmMsg{Confirmed: current selection}
type ConfirmModal struct {
	inner          BaseModal
	t              theme.Theme
	label          string // inner text of the confirm button, e.g. "Quit", "Prestige"
	confirmFocused bool   // true = Confirm focused (default), false = Cancel focused
}

// NewConfirmModal returns a ConfirmModal with Confirm pre-focused.
// label is the inner text shown on the confirm button (e.g. "Quit", "Prestige").
func NewConfirmModal(t theme.Theme, label string) ConfirmModal {
	return ConfirmModal{inner: NewBaseModal(t), t: t, label: label, confirmFocused: true}
}

// ConfirmFocused reports whether the confirm button (not cancel, not [Esc])
// will be triggered by the next NavConfirmMsg.
func (m ConfirmModal) ConfirmFocused() bool { return m.confirmFocused && !m.inner.EscFocused() }

// Update handles navigation and emits ConfirmMsg on selection.
func (m ConfirmModal) Update(msg tea.Msg) (ConfirmModal, tea.Cmd) {
	switch msg.(type) {
	case messages.NavUpMsg, messages.NavDownMsg:
		var cmd tea.Cmd
		m.inner, cmd = m.inner.Update(msg)
		return m, cmd

	case messages.NavLeftMsg:
		if !m.inner.EscFocused() {
			m.confirmFocused = true
		}

	case messages.NavRightMsg:
		if !m.inner.EscFocused() {
			m.confirmFocused = false
		}

	case messages.NavConfirmMsg:
		if m.inner.EscFocused() {
			var cmd tea.Cmd
			m.inner, cmd = m.inner.Update(msg)
			return m, cmd
		}
		confirmed := m.confirmFocused
		return m, func() tea.Msg { return ConfirmMsg{Confirmed: confirmed} }

	case ModalCloseMsg:
		return m, func() tea.Msg { return ConfirmMsg{Confirmed: false} }
	}
	return m, nil
}

// View renders the confirm dialog centered over bgContent.
// bgContent must be a pre-rendered string of exactly outerWidth×outerHeight cells.
func (m ConfirmModal) View(title, question, bgContent string, outerWidth, outerHeight int) string {
	boxWidth := min(max(outerWidth/2, 44), outerWidth-4)
	const boxHeight = 7
	innerWidth := boxWidth - 2
	content := m.renderContent(question, innerWidth)
	return m.inner.View(title, content, bgContent, outerWidth, outerHeight, boxWidth, boxHeight)
}

// renderContent builds the question + button rows for the interior of the box.
// innerWidth is the usable width between the side borders.
func (m ConfirmModal) renderContent(question string, innerWidth int) string {
	questionW := lipgloss.Width(question)
	qPad := max(innerWidth-questionW, 0)
	qLeft := qPad / 2
	qRight := qPad - qLeft
	questionLine := strings.Repeat(" ", qLeft) + question + strings.Repeat(" ", qRight)

	const cancelLabel = "[ Cancel ]"
	confirmLabel := "[ " + m.label + " ]"
	const buttonGap = "   "

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.t.DimText()))
	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.t.AccentColor())).
		Bold(true).
		Underline(true)

	var confirmStr, cancelStr string
	switch {
	case m.inner.EscFocused():
		confirmStr = dimStyle.Render(confirmLabel)
		cancelStr = dimStyle.Render(cancelLabel)
	case m.confirmFocused:
		confirmStr = activeStyle.Render(confirmLabel)
		cancelStr = dimStyle.Render(cancelLabel)
	default:
		confirmStr = dimStyle.Render(confirmLabel)
		cancelStr = activeStyle.Render(cancelLabel)
	}

	totalButtonW := lipgloss.Width(confirmLabel) + lipgloss.Width(buttonGap) + lipgloss.Width(cancelLabel)
	bPad := max(innerWidth-totalButtonW, 0)
	bLeft := bPad / 2
	bRight := bPad - bLeft
	buttonLine := strings.Repeat(" ", bLeft) + confirmStr + buttonGap + cancelStr + strings.Repeat(" ", bRight)

	blank := strings.Repeat(" ", innerWidth)

	return strings.Join([]string{
		blank,
		questionLine,
		blank,
		buttonLine,
		blank,
	}, "\n")
}
