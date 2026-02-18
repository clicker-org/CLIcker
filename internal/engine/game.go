package engine

import (
	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/upgrade"
	"github.com/clicker-org/clicker/internal/world"
)

// Engine is the core game engine. It owns mutable state and exposes
// deterministic operations that the UI layer calls in response to user input
// and timer ticks.
type Engine struct {
	State     gamestate.GameState
	WorldReg  *world.WorldRegistry
	AchievReg *achievement.AchievementRegistry
	// UpgradeReg holds per-world buy-on and upgrade registries, keyed by world ID.
	UpgradeReg map[string]*upgrade.WorldUpgradeRegistry

	// Earned achievements map (achievementID -> true if earned).
	Earned map[string]bool

	autosaveTimer    float64
	achievCheckTimer float64
}

// New creates and returns a new Engine. It builds per-world upgrade registries
// from the world configs registered in worldReg.
func New(
	gs gamestate.GameState,
	worldReg *world.WorldRegistry,
	achievReg *achievement.AchievementRegistry,
) *Engine {
	upReg := make(map[string]*upgrade.WorldUpgradeRegistry)
	for _, w := range worldReg.List() {
		reg := upgrade.NewWorldUpgradeRegistry()
		for _, b := range w.Config().BuyOns {
			reg.RegisterBuyOn(upgrade.NewConfigBuyOn(b))
		}
		for _, u := range w.Config().BuyOnUpgrades {
			reg.RegisterUpgrade(u)
		}
		upReg[w.ID()] = reg
	}

	return &Engine{
		State:      gs,
		WorldReg:   worldReg,
		AchievReg:  achievReg,
		UpgradeReg: upReg,
		Earned:     make(map[string]bool),
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

// PurchaseBuyOn attempts to buy one unit of the given buy-on in the given world.
// Returns (cost, true) on success, or (0, false) if the purchase cannot proceed
// (insufficient coins, level gate not met, or unknown world/buy-on).
func (e *Engine) PurchaseBuyOn(worldID, buyOnID string, playerLevel int) (float64, bool) {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return 0, false
	}
	reg, ok := e.UpgradeReg[worldID]
	if !ok {
		return 0, false
	}
	b, ok := reg.GetBuyOn(buyOnID)
	if !ok {
		return 0, false
	}
	if b.LevelRequirement() > playerLevel {
		return 0, false
	}
	count := ws.BuyOnCounts[buyOnID]
	cost := upgrade.CostForNext(b, count)
	if ws.Coins < cost {
		return 0, false
	}
	ws.Coins -= cost
	ws.BuyOnCounts[buyOnID] = count + 1
	// Recompute CPS after the purchase.
	ws.CPS = upgrade.CalculateWorldCPS(reg, ws.BuyOnCounts, ws.PurchasedUpgrades, ws.PrestigeMultiplier, 1.0)
	return cost, true
}
