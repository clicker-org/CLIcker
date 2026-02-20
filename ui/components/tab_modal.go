package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clicker-org/clicker/ui/theme"
)

// TabModal is a large modal overlay for tab content (Shop, Prestige,
// Achievements). It sizes itself to 3/4 of the outer width and fills most of
// the available height, leaving a small margin on all sides.
type TabModal struct {
	inner BaseModal
}

// NewTabModal returns a fresh TabModal.
func NewTabModal(t theme.Theme) TabModal {
	return TabModal{inner: NewBaseModal(t)}
}

// Update handles [Esc] button navigation.
func (m TabModal) Update(msg tea.Msg) (TabModal, tea.Cmd) {
	var cmd tea.Cmd
	m.inner, cmd = m.inner.Update(msg)
	return m, cmd
}

// View renders the tab modal centered over bgContent.
// bgContent must be a pre-rendered string of exactly outerWidth×outerHeight cells.
func (m TabModal) View(title, content, bgContent string, outerWidth, outerHeight int) string {
	boxWidth := min(max(outerWidth*3/4, 40), outerWidth-4)
	boxHeight := max(outerHeight-4, 6)
	return m.inner.View(title, content, bgContent, outerWidth, outerHeight, boxWidth, boxHeight)
}

// InnerWidth returns the usable content width inside this modal given an outer
// terminal width. Prestige tab uses this to size its progress bar correctly.
func (m TabModal) InnerWidth(outerWidth int) int {
	boxWidth := min(max(outerWidth*3/4, 40), outerWidth-4)
	return boxWidth - 2
}
