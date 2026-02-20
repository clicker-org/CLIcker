package integration_test

import (
	"testing"

	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
	_ "github.com/clicker-org/clicker/internal/world/worlds" 
)

// newTestEngine builds an Engine backed by the DefaultRegistry (Terra world) and
// an empty achievement registry. It is the shared starting point for all
// integration tests that need a running engine.
func newTestEngine(t *testing.T) *engine.Engine {
	t.Helper()
	gs := gamestate.NewGameState()
	for _, w := range world.DefaultRegistry.List() {
		gs.Worlds[w.ID()] = world.NewWorldState(w.ID(), w.BaseExchangeRate())
	}
	achReg := achievement.NewAchievementRegistry()
	return engine.New(gs, world.DefaultRegistry, achReg)
}
