package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/economy"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
	"github.com/clicker-org/clicker/ui/theme"
)

// StatusBar is the always-visible bottom status bar.
type StatusBar struct {
	style    lipgloss.Style
	width    int
	worldReg *world.WorldRegistry
}

// NewStatusBar creates a StatusBar with the given theme and width.
func NewStatusBar(t theme.Theme, width int, worldReg *world.WorldRegistry) StatusBar {
	return StatusBar{
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.PrimaryText())).
			Background(lipgloss.Color(t.Background())).
			Width(width).
			Padding(0, 1),
		width:    width,
		worldReg: worldReg,
	}
}

// SetWidth updates the bar width.
func (s *StatusBar) SetWidth(w int) {
	s.width = w
	s.style = s.style.Width(w)
}

// View renders the status bar. ws may be nil for overview/dashboard screens.
func (s StatusBar) View(gs gamestate.GameState, activeWorldID string, ws *world.WorldState) string {
	if ws == nil {
		return s.style.Render(fmt.Sprintf(
			"GC: %s  |  LVL: %d  |  XP: %d",
			economy.FormatCoinsBare(gs.Player.GeneralCoins),
			gs.Player.Level,
			gs.Player.XP,
		))
	}
	coinName := activeWorldID
	if s.worldReg != nil {
		if w, ok := s.worldReg.Get(activeWorldID); ok {
			coinName = w.CoinName()
		}
	}
	return s.style.Render(fmt.Sprintf(
		"%s | %s: %s | CPS: %.1f | Prestige: %d | LVL: %d XP: %d",
		activeWorldID,
		coinName,
		economy.FormatCoinsBare(ws.Coins),
		ws.CPS,
		ws.PrestigeCount,
		gs.Player.Level,
		gs.Player.XP,
	))
}
