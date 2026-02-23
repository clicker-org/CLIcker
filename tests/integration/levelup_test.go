package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clicker-org/clicker/internal/player"
)

// TestLevelUp_UnlocksGatedBuyOn is the core level-gate integration test.
// It flows through: player.AddXP mutates eng.State.Player.Level → engine reads
// that level in PurchaseBuyOn → purchase is allowed. Bugs in how the engine
// reads player level from state would surface here.
func TestLevelUp_UnlocksGatedBuyOn(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 1_000_000.0

	// Smelter requires level 2. Player starts at level 1.
	_, ok := eng.PurchaseBuyOn("terra", "smelter")
	assert.False(t, ok, "smelter should be blocked at level 1")

	// Grant XP directly to the engine's player state — simulating what the
	// engine will do automatically in Phase 4 (prestige XP, achievement XP, etc.)
	player.AddXP(&eng.State.Player, player.XPForLevel(2))
	assert.Equal(t, 2, eng.State.Player.Level)

	// The engine reads eng.State.Player.Level in PurchaseBuyOn — it should now
	// see level 2 and allow the purchase.
	_, ok = eng.PurchaseBuyOn("terra", "smelter")
	assert.True(t, ok, "smelter should be unlocked after level-up")
	assert.Equal(t, 1, eng.State.Worlds["terra"].BuyOnCounts["smelter"])
}

// TestLevelUp_XPAccumulates_ThenUnlocks verifies that partial XP grants
// accumulate on the engine's player state before crossing a threshold. This
// catches bugs where the engine discards or resets XP between calls.
func TestLevelUp_XPAccumulates_ThenUnlocks(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 1_000_000.0

	// 60 XP — not enough for level 2 (100 needed). Gate should still hold.
	player.AddXP(&eng.State.Player, 60)
	assert.Equal(t, 1, eng.State.Player.Level)
	_, ok := eng.PurchaseBuyOn("terra", "smelter")
	assert.False(t, ok, "still blocked: only 60 of 100 XP accumulated")

	// Another 60 XP — total 120, crosses the level 2 threshold.
	player.AddXP(&eng.State.Player, 60)
	assert.Equal(t, 2, eng.State.Player.Level)
	_, ok = eng.PurchaseBuyOn("terra", "smelter")
	assert.True(t, ok, "unlocked after XP crossed level 2 threshold")
}

func TestLevelGate_BlocksPurchaseBelowRequiredLevel(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 1_000_000_000.0

	// quantum_extractor requires level 10; player starts at level 1.
	eng.State.Player.Level = 1
	_, ok := eng.PurchaseBuyOn("terra", "quantum_extractor")

	assert.False(t, ok, "purchase should be blocked by level gate")
	assert.Equal(t, 0, eng.State.Worlds["terra"].BuyOnCounts["quantum_extractor"])
}

func TestLevelGate_AllowsPurchaseAtRequiredLevel(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 1_000_000_000.0

	eng.State.Player.Level = 10
	_, ok := eng.PurchaseBuyOn("terra", "quantum_extractor")

	assert.True(t, ok, "purchase should succeed at required level")
	assert.Equal(t, 1, eng.State.Worlds["terra"].BuyOnCounts["quantum_extractor"])
}

func TestLevelGate_DeepExcavator_RequiresLevel5(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 1_000_000_000.0

	eng.State.Player.Level = 4
	_, ok := eng.PurchaseBuyOn("terra", "deep_excavator")
	assert.False(t, ok, "deep_excavator should be blocked below level 5")

	eng.State.Player.Level = 5
	_, ok = eng.PurchaseBuyOn("terra", "deep_excavator")
	assert.True(t, ok, "deep_excavator should be allowed at level 5")
}
