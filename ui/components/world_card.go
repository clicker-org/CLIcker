package components

import (
	"fmt"

	"github.com/clicker-org/clicker/internal/world"
)

// WorldCard renders a compact summary card for a world.
type WorldCard struct {
	Width int
}

// View renders a world card stub.
func (c WorldCard) View(w world.World, ws *world.WorldState) string {
	if w == nil {
		return "[unknown world]"
	}
	completion := 0.0
	if ws != nil {
		completion = ws.CompletionPercent
	}
	return fmt.Sprintf("[ %s | %.1f%% ]", w.Name(), completion)
}
