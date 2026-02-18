package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/ui/theme"
)

// GalaxyMap renders a stub ASCII galaxy map with navigation.
type GalaxyMap struct {
	Width        int
	Height       int
	FocusedIndex int
}

// MoveLeft moves the focused world index left/up.
func (g *GalaxyMap) MoveLeft() {
	if g.FocusedIndex > 0 {
		g.FocusedIndex--
	}
}

// MoveRight moves the focused world index right/down.
func (g *GalaxyMap) MoveRight(max int) {
	if g.FocusedIndex < max-1 {
		g.FocusedIndex++
	}
}

// MoveUp moves the focused world index up.
func (g *GalaxyMap) MoveUp() {
	if g.FocusedIndex > 0 {
		g.FocusedIndex--
	}
}

// MoveDown moves the focused world index down.
func (g *GalaxyMap) MoveDown(max int) {
	if g.FocusedIndex < max-1 {
		g.FocusedIndex++
	}
}

// FocusedWorldID returns the ID of the currently focused world, or empty string.
func (g *GalaxyMap) FocusedWorldID(worldIDs []string) string {
	if len(worldIDs) == 0 {
		return ""
	}
	if g.FocusedIndex >= len(worldIDs) {
		g.FocusedIndex = len(worldIDs) - 1
	}
	if g.FocusedIndex < 0 {
		g.FocusedIndex = 0
	}
	return worldIDs[g.FocusedIndex]
}

// View renders a placeholder galaxy map with world list centered in the available space.
func (g GalaxyMap) View(worldIDs []string, t theme.Theme) string {
	bg := lipgloss.Color(t.Background())
	accentFg := lipgloss.Color(t.AccentColor())
	dimFg := lipgloss.Color(t.DimText())
	primaryFg := lipgloss.Color(t.PrimaryText())

	focusedStyle := lipgloss.NewStyle().Foreground(accentFg).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(dimFg)
	headerStyle := lipgloss.NewStyle().Foreground(primaryFg)

	var lines []string
	lines = append(lines, headerStyle.Render("·  ✦        · *   ✦  ·"))
	lines = append(lines, headerStyle.Render("      GALAXY MAP"))
	lines = append(lines, "")

	for i, id := range worldIDs {
		if i == g.FocusedIndex {
			lines = append(lines, focusedStyle.Render("▶  "+id))
		} else {
			lines = append(lines, normalStyle.Render("   "+id))
		}
	}
	if len(worldIDs) == 0 {
		lines = append(lines, normalStyle.Render("(no worlds registered)"))
	}

	h := g.Height
	if h <= 0 {
		h = 20
	}
	return lipgloss.Place(g.Width, h,
		lipgloss.Center, lipgloss.Center,
		strings.Join(lines, "\n"),
		lipgloss.WithWhitespaceBackground(bg))
}
