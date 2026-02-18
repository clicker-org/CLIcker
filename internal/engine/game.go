package engine

import (
	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
)

// Engine is the core game engine. It owns mutable state and exposes
// deterministic operations that the UI layer calls in response to user input
// and timer ticks.
type Engine struct {
	State    gamestate.GameState
	WorldReg *world.WorldRegistry
	AchievReg *achievement.AchievementRegistry

	// Earned achievements map (worldID -> true if earned).
	Earned map[string]bool

	autosaveTimer  float64
	achievCheckTimer float64
}

// New creates and returns a new Engine.
func New(
	gs gamestate.GameState,
	worldReg *world.WorldRegistry,
	achievReg *achievement.AchievementRegistry,
) *Engine {
	return &Engine{
		State:     gs,
		WorldReg:  worldReg,
		AchievReg: achievReg,
		Earned:    make(map[string]bool),
	}
}

// ClickPower returns the coins generated per manual click in the given world.
// globalClickMult is the combined global click multiplier from general shop purchases.
func (e *Engine) ClickPower(worldID string, globalClickMult float64) float64 {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return 0
	}
	base := 1.0 * ws.PrestigeMultiplier
	if globalClickMult > 0 {
		base *= globalClickMult
	}
	return base
}

// HandleClick records a manual click for the given world, adds coins, and
// returns the number of coins earned by this click.
func (e *Engine) HandleClick(worldID string, globalClickMult float64) float64 {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return 0
	}
	earned := e.ClickPower(worldID, globalClickMult)
	ws.Coins += earned
	ws.TotalCoinsEarned += earned
	ws.TotalClicks++
	e.State.Player.TotalClicks++
	if e.State.Player.WorldTotalCoinsEarned == nil {
		e.State.Player.WorldTotalCoinsEarned = make(map[string]float64)
	}
	e.State.Player.WorldTotalCoinsEarned[worldID] += earned
	return earned
}
