package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/ui/theme"
)

// NotificationDismissMsg is sent when a notification's timer expires.
type NotificationDismissMsg struct{}

// Notification is a transient toast notification component.
type Notification struct {
	text    string
	visible bool
	style   lipgloss.Style
}

// NewNotification creates a Notification with the given theme.
func NewNotification(t theme.Theme) Notification {
	return Notification{
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.PrimaryText())).
			Background(lipgloss.Color(t.SecondaryAccent())).
			Padding(0, 1).
			Bold(true),
	}
}

// Show displays a notification and returns a Cmd that dismisses it after duration.
func (n *Notification) Show(text string, duration time.Duration) tea.Cmd {
	n.text = text
	n.visible = true
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return NotificationDismissMsg{}
	})
}

// Update handles notification lifecycle messages.
func (n Notification) Update(msg tea.Msg) (Notification, tea.Cmd) {
	if _, ok := msg.(NotificationDismissMsg); ok {
		n.visible = false
		n.text = ""
	}
	return n, nil
}

// View renders the notification, or empty string if not visible.
func (n Notification) View() string {
	if !n.visible {
		return ""
	}
	return n.style.Render("* " + n.text)
}
