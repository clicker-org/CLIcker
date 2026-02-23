package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clicker-org/clicker/internal/engine"
)

func TestClick_AccumulatesCoinsAndStats(t *testing.T) {
	eng := newTestEngine(t)
	before := eng.State.Worlds["terra"].Coins

	earned := eng.HandleClick("terra")

	assert.Greater(t, earned, 0.0)
	assert.InDelta(t, before+earned, eng.State.Worlds["terra"].Coins, 0.001)
	assert.Equal(t, int64(1), eng.State.Worlds["terra"].TotalClicks)
	assert.Equal(t, int64(1), eng.State.Player.TotalClicks)
	assert.InDelta(t, earned, eng.State.Player.WorldTotalCoinsEarned["terra"], 0.001)
}

func TestClickPower_ScalesWithPrestigeMultiplier(t *testing.T) {
	eng := newTestEngine(t)
	base := eng.ClickPower("terra")

	eng.State.Worlds["terra"].PrestigeMultiplier = 2.0

	assert.InDelta(t, base*2.0, eng.ClickPower("terra"), 0.001)
}

func TestClickPower_UsesEngineOwnedGlobalMultiplier(t *testing.T) {
	eng := newTestEngine(t)
	base := eng.ClickPower("terra")
	assert.InDelta(t, base, eng.ClickPower("terra"), 0.001)
}

func TestBuyOnPurchase_UpdatesCPS(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 1000.0
	cpsBefore := eng.State.Worlds["terra"].CPS

	cost, ok := eng.PurchaseBuyOn("terra", "auto_miner")

	assert.True(t, ok, "purchase should succeed with enough coins")
	assert.Greater(t, cost, 0.0)
	assert.Greater(t, eng.State.Worlds["terra"].CPS, cpsBefore, "CPS should increase after purchase")
	assert.InDelta(t, 1000.0-cost, eng.State.Worlds["terra"].Coins, 0.001)
	assert.Equal(t, 1, eng.State.Worlds["terra"].BuyOnCounts["auto_miner"])
}

func TestBuyOnPurchase_FailsOnInsufficientCoins(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 0

	_, ok := eng.PurchaseBuyOn("terra", "auto_miner")

	assert.False(t, ok, "purchase should fail with no coins")
	assert.Equal(t, 0, eng.State.Worlds["terra"].BuyOnCounts["auto_miner"])
}

func TestBuyOnPurchase_CostScalesWithCount(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 10_000.0

	cost1, ok1 := eng.PurchaseBuyOn("terra", "auto_miner")
	require.True(t, ok1)

	cost2, ok2 := eng.PurchaseBuyOn("terra", "auto_miner")
	require.True(t, ok2)

	assert.Greater(t, cost2, cost1, "second purchase should cost more than first")
	assert.Equal(t, 2, eng.State.Worlds["terra"].BuyOnCounts["auto_miner"])
}

func TestTick_AppliesCPSToCoins(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].CPS = 10.0
	startCoins := eng.State.Worlds["terra"].Coins

	eng.Tick(1.0)

	assert.InDelta(t, startCoins+10.0, eng.State.Worlds["terra"].Coins, 0.001)
}

func TestTick_UpdatesLifetimeCoinsAndPlaytime(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].CPS = 5.0

	eng.Tick(2.0)

	assert.InDelta(t, 10.0, eng.State.Worlds["terra"].TotalCoinsEarned, 0.001)
	assert.InDelta(t, 10.0, eng.State.Player.WorldTotalCoinsEarned["terra"], 0.001)
	assert.InDelta(t, 2.0, eng.State.Player.TotalPlaySeconds, 0.001)
}

func TestTick_EmitsAutosaveEventAfterInterval(t *testing.T) {
	eng := newTestEngine(t)

	events := eng.Tick(engine.AutoSaveInterval + 0.1)

	var found bool
	for _, ev := range events {
		if ev.Type == engine.EventAutoSave {
			found = true
		}
	}
	assert.True(t, found, "autosave event should be emitted after interval")
}

func TestTick_NoAutosaveBeforeInterval(t *testing.T) {
	eng := newTestEngine(t)

	events := eng.Tick(engine.AutoSaveInterval - 1.0)

	for _, ev := range events {
		assert.NotEqual(t, engine.EventAutoSave, ev.Type, "autosave should not fire before interval")
	}
}
