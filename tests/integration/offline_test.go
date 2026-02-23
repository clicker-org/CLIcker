package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clicker-org/clicker/internal/offline"
	"github.com/clicker-org/clicker/internal/save"
	"github.com/clicker-org/clicker/internal/world"
)

// TestOfflineApply_WorldScreen_EarnsCoins simulates a player quitting from the
// world screen 4 hours ago with active CPS, verifying coins are earned and
// applied to the world state.
func TestOfflineApply_WorldScreen_EarnsCoins(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].CPS = 100.0
	eng.State.Worlds["terra"].Coins = 0

	savedAt := time.Now().Add(-4 * time.Hour)
	result := offline.Apply("world", "terra", savedAt, &eng.State, world.DefaultRegistry)

	w, ok := world.DefaultRegistry.Get("terra")
	require.True(t, ok)
	expected := 100.0 * w.OfflinePercentage() * (4 * 3600)
	// Tolerance of 50 coins accounts for test execution time (~5 seconds at 10 coins/sec offline).
	assert.InDelta(t, expected, result.WorldCoins, 50.0)
	assert.InDelta(t, expected, eng.State.Worlds["terra"].Coins, 50.0, "coins should be applied to world state")
	assert.Equal(t, "terra", result.WorldID)
	assert.Greater(t, result.Duration, 3*time.Hour)
}

// TestOfflineApply_WorldScreen_CappedAtMaxHours verifies that offline income is
// capped when the player has been away longer than the configured cap.
func TestOfflineApply_WorldScreen_CappedAtMaxHours(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].CPS = 100.0

	// 24 hours away — well over the 8h cap.
	savedAt := time.Now().Add(-24 * time.Hour)
	result := offline.Apply("world", "terra", savedAt, &eng.State, world.DefaultRegistry)

	// Since the cap is always hit, the result is deterministic.
	w, ok := world.DefaultRegistry.Get("terra")
	require.True(t, ok)
	expectedMax := 100.0 * w.OfflinePercentage() * w.OfflineCapHours() * 3600
	assert.InDelta(t, expectedMax, result.WorldCoins, 0.001)
}

// TestOfflineApply_OverviewScreen_EarnsGeneralCoins verifies that quitting from
// the overview/galaxy map screen generates general coins (not world coins).
func TestOfflineApply_OverviewScreen_EarnsGeneralCoins(t *testing.T) {
	eng := newTestEngine(t)
	startGC := eng.State.Player.GeneralCoins

	// 4 hours away — long enough to exceed the GC cap (cap hits at ~2.78h).
	savedAt := time.Now().Add(-4 * time.Hour)
	result := offline.Apply("overview", "", savedAt, &eng.State, world.DefaultRegistry)

	assert.InDelta(t, offline.OverviewOfflineGCCap, result.GeneralCoins, 0.001)
	assert.InDelta(t, startGC+offline.OverviewOfflineGCCap, eng.State.Player.GeneralCoins, 0.001)
	assert.Equal(t, float64(0), result.WorldCoins, "no world coins should be earned from overview session")
}

// TestOfflineApply_ZeroTimeAway_EarnsNothing verifies that a savedAt in the
// future (or equal to now) yields zero income.
func TestOfflineApply_ZeroTimeAway_EarnsNothing(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].CPS = 100.0

	savedAt := time.Now().Add(1 * time.Minute) // In the future — elapsed will be negative.
	result := offline.Apply("world", "terra", savedAt, &eng.State, world.DefaultRegistry)

	assert.Equal(t, float64(0), result.WorldCoins)
	assert.Equal(t, float64(0), result.GeneralCoins)
}

// TestOfflineApply_ZeroCPS_EarnsNothing verifies that a world with no buy-ons
// (CPS = 0) earns nothing during offline time.
func TestOfflineApply_ZeroCPS_EarnsNothing(t *testing.T) {
	eng := newTestEngine(t)
	// Default WorldState has CPS = 0 (no buy-ons purchased).
	assert.InDelta(t, 0.0, eng.State.Worlds["terra"].CPS, 0.001)

	savedAt := time.Now().Add(-1 * time.Hour)
	result := offline.Apply("world", "terra", savedAt, &eng.State, world.DefaultRegistry)

	assert.Equal(t, float64(0), result.WorldCoins)
}

// TestOfflineApply_WithEarnedCPS_CorrectlyAccumulates is the fully end-to-end
// offline income test. CPS is earned through a real buy-on purchase (not
// injected), the state is saved and loaded, and offline income is applied to the
// reconstructed state. This is the chain that all three previous tests miss:
//
//	PurchaseBuyOn → CPS → Save → Load → GameStateFromSave → offline.Apply → coins
//
// A bug in CPS serialization would make result.WorldCoins diverge from the
// expected value based on the pre-save CPS.
func TestOfflineApply_WithEarnedCPS_CorrectlyAccumulates(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "save.json")

	// Phase 1: earn real CPS through a purchase.
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].Coins = 10_000.0
	_, ok := eng.PurchaseBuyOn("terra", "auto_miner")
	require.True(t, ok)
	cps := eng.State.Worlds["terra"].CPS
	require.Greater(t, cps, 0.0, "purchase should produce non-zero CPS")

	// Phase 2: save as if the player quit from the world screen.
	eng.State.LastScreen = "world"
	eng.State.LastWorldID = "terra"
	require.NoError(t, save.Save(eng.State, map[string]bool{}, save.Settings{AnimationsEnabled: true, ActiveTheme: "space"}, savePath))

	// Phase 3: load and reconstruct — completely separate from Phase 1.
	sf, err := save.Load(savePath)
	require.NoError(t, err)
	gs := save.GameStateFromSave(sf, world.DefaultRegistry)

	// Phase 4: apply offline income using the reconstructed state.
	savedAt := time.Now().Add(-1 * time.Hour)
	coinsBeforeOffline := gs.Worlds["terra"].Coins
	result := offline.Apply(sf.LastScreen, sf.LastWorldID, savedAt, &gs, world.DefaultRegistry)

	// Expected offline income: cps * 10% * 3600s (under 8h cap for small CPS).
	w, ok := world.DefaultRegistry.Get("terra")
	require.True(t, ok)
	expected := cps * w.OfflinePercentage() * 3600
	// Tolerance of 1.0 coin covers ~100 seconds of test execution at this CPS rate.
	assert.InDelta(t, expected, result.WorldCoins, 1.0,
		"offline income should use CPS that survived the save/load cycle")
	assert.Greater(t, gs.Worlds["terra"].Coins, coinsBeforeOffline,
		"offline coins should be applied to the reconstructed world state")
}

func TestOfflineApply_WorldScreen_CapUpgradeAffectsCap(t *testing.T) {
	eng := newTestEngine(t)
	eng.State.Worlds["terra"].CPS = 100.0
	eng.State.Worlds["terra"].OfflineCapUpgradeLevel = 1

	// Long enough away to hit cap.
	savedAt := time.Now().Add(-24 * time.Hour)
	result := offline.Apply("world", "terra", savedAt, &eng.State, world.DefaultRegistry)

	w, ok := world.DefaultRegistry.Get("terra")
	require.True(t, ok)
	capHours := world.EffectiveOfflineCapHours(eng.State.Worlds["terra"], w.OfflineCapHours())
	expectedMax := 100.0 * w.OfflinePercentage() * capHours * 3600
	assert.InDelta(t, expectedMax, result.WorldCoins, 0.001)
}

// TestEffectiveOfflineCapHours_ScalesWithUpgradeLevel verifies that purchasing
// offline cap upgrades correctly extends the effective cap.
func TestEffectiveOfflineCapHours_ScalesWithUpgradeLevel(t *testing.T) {
	cases := []struct {
		upgradeLevel int
		wantCap      float64
	}{
		{0, 8.0},
		{1, 10.0},
		{2, 12.0},
		{5, 18.0},
	}

	for _, tc := range cases {
		ws := world.NewWorldState("terra", 0.001)
		ws.OfflineCapUpgradeLevel = tc.upgradeLevel
		got := world.EffectiveOfflineCapHours(ws, 8.0)
		assert.InDelta(t, tc.wantCap, got, 0.001, "upgrade level %d", tc.upgradeLevel)
	}
}
