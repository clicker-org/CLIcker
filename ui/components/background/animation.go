package background

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// AnimTickMsg is sent on each animation frame tick.
type AnimTickMsg struct {
	AnimationName string
}

// BackgroundAnimation is the interface for ambient background animations.
type BackgroundAnimation interface {
	Init() tea.Cmd
	Update(tea.Msg) (BackgroundAnimation, tea.Cmd)
	View(width, height int) string
	Name() string
}

// AnimTickCmd returns a tea.Cmd that fires an AnimTickMsg after the given interval.
func AnimTickCmd(name string, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return AnimTickMsg{AnimationName: name}
	})
}
