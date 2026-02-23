package engine

import (
	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/economy"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/player"
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
func (e *Engine) ClickPower(worldID string) float64 {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return 0
	}
	base := 1.0 * ws.PrestigeMultiplier
	if m := e.globalClickMultiplier(); m > 0 {
		base *= m
	}
	return base
}

// globalClickMultiplier returns the effective global click multiplier sourced
// from engine-owned state. Phase 0 has no general shop effects yet.
func (e *Engine) globalClickMultiplier() float64 {
	return 1.0
}

// HandleClick records a manual click for the given world, adds coins, and
// returns the number of coins earned by this click.
func (e *Engine) HandleClick(worldID string) float64 {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return 0
	}
	earned := e.ClickPower(worldID)
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
func (e *Engine) PurchaseBuyOn(worldID, buyOnID string) (float64, bool) {
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
	if b.LevelRequirement() > e.State.Player.Level {
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

// CanPrestige reports whether the player has met the prestige threshold for the
// given world. Returns false for unknown worlds.
func (e *Engine) CanPrestige(worldID string) bool {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return false
	}
	w, ok := e.WorldReg.Get(worldID)
	if !ok {
		return false
	}
	threshold := w.Config().PrestigeThreshold
	switch threshold.Type {
	case "coins_earned":
		return ws.TotalCoinsEarned >= threshold.Value
	case "buy_ons_owned":
		total := 0
		for _, count := range ws.BuyOnCounts {
			total += count
		}
		return float64(total) >= threshold.Value
	case "completion_percent":
		return ws.CompletionPercent*100 >= threshold.Value
	default:
		return false
	}
}

// PrestigeProgress returns (current, threshold) for the active prestige metric
// in the given world. Used by the UI to render a progress bar.
func (e *Engine) PrestigeProgress(worldID string) (current, threshold float64) {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return 0, 1
	}
	w, ok := e.WorldReg.Get(worldID)
	if !ok {
		return 0, 1
	}
	cfg := w.Config().PrestigeThreshold
	switch cfg.Type {
	case "coins_earned":
		return ws.TotalCoinsEarned, cfg.Value
	case "buy_ons_owned":
		total := 0
		for _, count := range ws.BuyOnCounts {
			total += count
		}
		return float64(total), cfg.Value
	case "completion_percent":
		return ws.CompletionPercent * 100, cfg.Value
	default:
		return 0, cfg.Value
	}
}

// ExecutePrestige performs a world prestige: computes rewards, applies them to
// the player, and resets the world state. Returns (reward, true) on success or
// (zero, false) if the prestige threshold has not been met.
func (e *Engine) ExecutePrestige(worldID string) (economy.PrestigeReward, bool) {
	if !e.CanPrestige(worldID) {
		return economy.PrestigeReward{}, false
	}
	ws := e.State.Worlds[worldID]

	reward := economy.CalculatePrestigeReward(ws.TotalCoinsEarned, ws.PrestigeCount, ws.PrestigeMultiplier)

	// Apply rewards to the player.
	e.State.Player.GeneralCoins += reward.GeneralCoinsEarned
	e.State.Player.LifetimeGeneralCoins += reward.GeneralCoinsEarned
	player.AddXP(&e.State.Player, reward.XPGrant)

	// Update prestige state.
	ws.PrestigeCount++
	ws.PrestigeMultiplier = reward.PrestigeMultiplier

	// Reset world — coins, buy-ons, CPS. Keep: prestige count/multiplier,
	// exchange rate, offline cap upgrade level, lifetime stats (TotalCoinsEarned,
	// TotalClicks) and completion progress.
	ws.Coins = 0
	ws.BuyOnCounts = make(map[string]int)
	ws.PurchasedUpgrades = make(map[string]bool)
	ws.CPS = 0

	return reward, true
}

// CanExchangeBoost reports whether the player has a non-zero world coin balance
// to sacrifice for an exchange boost.
func (e *Engine) CanExchangeBoost(worldID string) bool {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return false
	}
	return ws.Coins > 0
}

// ExchangeBoostPreview returns the projected result of an exchange boost
// without actually executing it. Safe to call at any time.
func (e *Engine) ExchangeBoostPreview(worldID string) economy.ExchangeBoostResult {
	ws, ok := e.State.Worlds[worldID]
	if !ok {
		return economy.ExchangeBoostResult{}
	}
	return economy.CalculateExchangeBoost(ws.Coins, ws.ExchangeRate)
}

// ExecuteExchangeBoost performs an exchange boost: sacrifices a portion of the
// world coin balance and converts it to general coins at the current exchange
// rate. The exchange rate is permanently improved. Returns (result, true) on
// success or (zero, false) if the balance is zero.
func (e *Engine) ExecuteExchangeBoost(worldID string) (economy.ExchangeBoostResult, bool) {
	if !e.CanExchangeBoost(worldID) {
		return economy.ExchangeBoostResult{}, false
	}
	ws := e.State.Worlds[worldID]

	result := economy.CalculateExchangeBoost(ws.Coins, ws.ExchangeRate)

	ws.Coins -= result.WorldCoinsCost
	ws.ExchangeRate = result.NewExchangeRate

	e.State.Player.GeneralCoins += result.GeneralCoinsEarned
	e.State.Player.LifetimeGeneralCoins += result.GeneralCoinsEarned

	return result, true
}
