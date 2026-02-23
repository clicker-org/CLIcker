package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/player"
	"github.com/clicker-org/clicker/internal/save"
	"github.com/clicker-org/clicker/internal/world"
)

// TestFullSaveLoadCycle drives a real Engine into a non-trivial state, writes it
// to disk, reads it back, and asserts that every field survived the roundtrip.
// This test crosses: engine → gamestate → save.Save → save.Load.
func TestFullSaveLoadCycle(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "save.json")

	eng := newTestEngine(t)

	// Put the engine in a known, non-trivial state.
	eng.State.Player.XP = 500
	eng.State.Player.Level = 3
	eng.State.Player.GeneralCoins = 99.9
	eng.State.LastScreen = "world"
	eng.State.LastWorldID = "terra"

	ws := eng.State.Worlds["terra"]
	ws.Coins = 1234.5
	ws.TotalCoinsEarned = 5000.0
	ws.BuyOnCounts["auto_miner"] = 3
	ws.PrestigeCount = 1
	ws.PrestigeMultiplier = 1.5

	earned := map[string]bool{"first_click": true}
	settings := save.Settings{AnimationsEnabled: false, ActiveTheme: "space"}

	require.NoError(t, save.Save(eng.State, earned, settings, savePath))

	sf, err := save.Load(savePath)
	require.NoError(t, err)

	assert.Equal(t, save.CurrentVersion, sf.Version)
	assert.Equal(t, "world", sf.LastScreen)
	assert.Equal(t, "terra", sf.LastWorldID)

	assert.Equal(t, 500, sf.Player.XP)
	assert.Equal(t, 3, sf.Player.Level)
	assert.InDelta(t, 99.9, sf.Player.GeneralCoins, 0.001)

	terraData, ok := sf.Worlds["terra"]
	require.True(t, ok, "terra world missing from save file")
	assert.InDelta(t, 1234.5, terraData.Coins, 0.001)
	assert.InDelta(t, 5000.0, terraData.TotalCoinsEarned, 0.001)
	assert.Equal(t, 3, terraData.BuyOnCounts["auto_miner"])
	assert.Equal(t, 1, terraData.PrestigeCount)
	assert.InDelta(t, 1.5, terraData.PrestigeMultiplier, 0.001)

	assert.True(t, sf.Achievements["first_click"])
	assert.False(t, sf.Settings.AnimationsEnabled)
	assert.Equal(t, "space", sf.Settings.ActiveTheme)
}

// TestGameStateFromSave_ReconstitutesAllWorlds verifies that GameStateFromSave
// restores saved world data and creates fresh WorldState entries for any world
// IDs not present in the save file. This tests the new-world-added scenario.
func TestGameStateFromSave_ReconstitutesAllWorlds(t *testing.T) {
	sf := save.DefaultSaveFile()
	sf.Player = player.Player{
		XP:                    200,
		Level:                 2,
		GeneralCoins:          10.0,
		WorldTotalCoinsEarned: make(map[string]float64),
	}
	sf.Worlds["terra"] = save.WorldSaveData{
		WorldID:            "terra",
		Coins:              42.0,
		TotalCoinsEarned:   100.0,
		BuyOnCounts:        map[string]int{"auto_miner": 2},
		PurchasedUpgrades:  map[string]bool{},
		PrestigeMultiplier: 1.0,
	}

	worldIDs := world.DefaultRegistry.IDs()
	gs := save.GameStateFromSave(sf, world.DefaultRegistry)

	assert.Equal(t, 200, gs.Player.XP)
	assert.Equal(t, 2, gs.Player.Level)

	require.NotNil(t, gs.Worlds["terra"])
	assert.InDelta(t, 42.0, gs.Worlds["terra"].Coins, 0.001)
	assert.Equal(t, 2, gs.Worlds["terra"].BuyOnCounts["auto_miner"])

	// Every registered world ID must have a WorldState entry.
	for _, id := range worldIDs {
		assert.NotNil(t, gs.Worlds[id], "world %q missing from reconstructed state", id)
	}
}

// TestBuyOnPurchase_CPSPreservedAcrossSaveLoad is the key integration test for
// CPS serialization. CPS is earned through real purchases (not injected), saved
// to disk, and loaded back. A field rename or serialization bug in WorldSaveData
// would produce a zero or wrong CPS after load and this test would catch it.
func TestBuyOnPurchase_CPSPreservedAcrossSaveLoad(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "save.json")

	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 10_000.0

	_, ok1 := eng.PurchaseBuyOn("terra", "auto_miner", 0)
	require.True(t, ok1)
	_, ok2 := eng.PurchaseBuyOn("terra", "auto_miner", 0)
	require.True(t, ok2)

	cpsAfterPurchase := eng.State.Worlds["terra"].CPS
	require.Greater(t, cpsAfterPurchase, 0.0, "real purchases should produce non-zero CPS")

	require.NoError(t, save.Save(eng.State, map[string]bool{}, save.Settings{AnimationsEnabled: true, ActiveTheme: "space"}, savePath))

	sf, err := save.Load(savePath)
	require.NoError(t, err)

	assert.InDelta(t, cpsAfterPurchase, sf.Worlds["terra"].CPS, 0.001,
		"CPS should be identical after save/load")
	assert.Equal(t, 2, sf.Worlds["terra"].BuyOnCounts["auto_miner"],
		"buy-on count should be preserved")
}

// TestBuyOnPurchase_SaveLoad_TickEarnsCoins is the fully end-to-end test for
// CPS persistence: buy-ons are purchased (earning real CPS), the game is saved
// and loaded into a brand-new engine, and a tick must earn coins at the correct
// rate. This is the chain that no other test covers:
//
//	PurchaseBuyOn → CPS calculation → Save → Load → GameStateFromSave → new Engine → Tick → coins
func TestBuyOnPurchase_SaveLoad_TickEarnsCoins(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "save.json")

	// Phase 1: earn CPS through a real purchase.
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 10_000.0
	_, ok := eng.PurchaseBuyOn("terra", "auto_miner", 0)
	require.True(t, ok)
	cps := eng.State.Worlds["terra"].CPS
	require.Greater(t, cps, 0.0)

	// Phase 2: save.
	require.NoError(t, save.Save(eng.State, map[string]bool{}, save.Settings{AnimationsEnabled: true, ActiveTheme: "space"}, savePath))

	// Phase 3: load into a completely fresh engine — no shared state with Phase 1.
	sf, err := save.Load(savePath)
	require.NoError(t, err)
	gs := save.GameStateFromSave(sf, world.DefaultRegistry)
	eng2 := engine.New(gs, world.DefaultRegistry, achievement.NewAchievementRegistry())

	// Phase 4: tick 1 second and verify the loaded CPS drives coin accumulation.
	coinsBefore := eng2.State.Worlds["terra"].Coins
	eng2.Tick(1.0)
	assert.InDelta(t, coinsBefore+cps, eng2.State.Worlds["terra"].Coins, 0.001,
		"tick should earn coins at the CPS rate restored from save")
}

// TestSaveLoadCycle_BuyOnCountsPreserved ensures buy-on ownership counts are
// preserved exactly across a save/load cycle (regression: nil map handling).
func TestSaveLoadCycle_BuyOnCountsPreserved(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "save.json")

	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 9999.0
	eng.State.Worlds["terra"].BuyOnCounts["auto_miner"] = 7
	eng.State.Worlds["terra"].BuyOnCounts["drill_bot"] = 3

	require.NoError(t, save.Save(eng.State, map[string]bool{}, save.Settings{AnimationsEnabled: true, ActiveTheme: "space"}, savePath))

	sf, err := save.Load(savePath)
	require.NoError(t, err)

	terraData := sf.Worlds["terra"]
	assert.Equal(t, 7, terraData.BuyOnCounts["auto_miner"])
	assert.Equal(t, 3, terraData.BuyOnCounts["drill_bot"])
	assert.InDelta(t, 9999.0, terraData.Coins, 0.001)
}
