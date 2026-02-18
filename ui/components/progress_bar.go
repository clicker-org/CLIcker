package components

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clicker-org/clicker/ui/theme"
)

// ProgressBar wraps the bubbles progress bar with theme support.
type ProgressBar struct {
	bar   progress.Model
	width int
}

// NewProgressBar creates a styled ProgressBar.
func NewProgressBar(t theme.Theme, width int, accentHex string) ProgressBar {
	color := accentHex
	if color == "" && t != nil {
		color = t.AccentColor()
	}
	bar := progress.New(
		progress.WithScaledGradient(color, color),
		progress.WithoutPercentage(),
	)
	bar.Width = width
	return ProgressBar{bar: bar, width: width}
}

// SetWidth updates the bar width.
func (p *ProgressBar) SetWidth(w int) {
	p.width = w
	p.bar.Width = w
}

// View renders the progress bar at the given percentage (0.0–1.0).
func (p ProgressBar) View(pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	return p.bar.ViewAs(pct)
}

// Update handles progress bar animations.
func (p ProgressBar) Update(msg tea.Msg) (ProgressBar, tea.Cmd) {
	barModel, cmd := p.bar.Update(msg)
	if m, ok := barModel.(progress.Model); ok {
		p.bar = m
	}
	return p, cmd
}
