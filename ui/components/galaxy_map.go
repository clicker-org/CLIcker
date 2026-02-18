package components

import (
	"strings"
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

// View renders a placeholder galaxy map with world list.
func (g GalaxyMap) View(worldIDs []string) string {
	var sb strings.Builder
	sb.WriteString("  ·  ✦        · *   ✦  ·\n")
	sb.WriteString("         GALAXY MAP\n\n")
	for i, id := range worldIDs {
		cursor := "  "
		if i == g.FocusedIndex {
			cursor = "> "
		}
		sb.WriteString(cursor + "[ " + id + " ]\n")
	}
	if len(worldIDs) == 0 {
		sb.WriteString("  (no worlds registered)\n")
	}
	return sb.String()
}
