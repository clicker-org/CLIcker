package offline

import (
	"time"

	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
)

// Global offline income constants. These values are defined here (not in world
// TOML) per the architecture spec — they are engine-level defaults, not
// per-world balance values.
const (
	// WorldOfflinePct is the fraction of a world's CPS earned while offline.
	WorldOfflinePct = 0.10
	// WorldOfflineCapHours is the default maximum hours of offline accumulation.
	WorldOfflineCapHours = 8.0
	// OverviewOfflineBaseRate is the general-coin trickle rate (GC/sec) when
	// the player quits from the overview / galaxy map.
	OverviewOfflineBaseRate = 0.001
	// OverviewOfflineGCCap is the maximum general coins earned from an overview
	// offline session.
	OverviewOfflineGCCap = 10.0
	// MinReportDuration is the minimum time away before the offline report is
	// shown to the player on launch.
	MinReportDuration = 60 * time.Second
)

// Result holds the outcome of an offline income calculation.
type Result struct {
	Duration    time.Duration
	WorldCoins  float64
	GeneralCoins float64
	WorldID     string
}

// Apply computes offline income based on how long the game was closed and
// where the player was when they quit, then applies the earned amounts
// directly to gs. Returns the Result for display in the offline report.
//
// For world-screen sessions, world-specific offline settings are sourced from
// worldReg. If a world is missing from the registry, engine-level defaults are
// used as a fallback.
func Apply(lastScreen, lastWorldID string, savedAt time.Time, gs *gamestate.GameState, worldReg *world.WorldRegistry) Result {
	if savedAt.IsZero() {
		return Result{}
	}
	elapsed := time.Since(savedAt).Seconds()
	if elapsed <= 0 {
		return Result{}
	}

	result := Result{
		Duration: time.Duration(elapsed * float64(time.Second)),
		WorldID:  lastWorldID,
	}

	if lastScreen == "world" && lastWorldID != "" {
		if ws, ok := gs.Worlds[lastWorldID]; ok {
			offlinePct := WorldOfflinePct
			capHours := WorldOfflineCapHours
			if worldReg != nil {
				if w, ok := worldReg.Get(lastWorldID); ok {
					if p := w.OfflinePercentage(); p > 0 {
						offlinePct = p
					}
					if baseCap := w.OfflineCapHours(); baseCap > 0 {
						capHours = world.EffectiveOfflineCapHours(ws, baseCap)
					}
				}
			}
			coins := CalculateOfflineIncome(ws.CPS, offlinePct, elapsed, capHours)
			if coins > 0 {
				ws.Coins += coins
				ws.TotalCoinsEarned += coins
			}
			result.WorldCoins = coins
		}
	} else {
		gc := CalculateOverviewOfflineIncome(OverviewOfflineBaseRate, elapsed, OverviewOfflineGCCap)
		if gc > 0 {
			gs.Player.GeneralCoins += gc
		}
		result.GeneralCoins = gc
	}

	return result
}

// CalculateOfflineIncome computes coins earned while the game was closed
// for a world that was active when the player quit.
//
// Formula: min(cps * offlinePct * elapsedSecs, cps * capHours * 3600 * offlinePct)
func CalculateOfflineIncome(cps, offlinePct, elapsedSecs, capHours float64) float64 {
	if cps <= 0 || offlinePct <= 0 || elapsedSecs <= 0 || capHours <= 0 {
		return 0
	}
	earned := cps * offlinePct * elapsedSecs
	cap := cps * capHours * 3600 * offlinePct
	if earned > cap {
		return cap
	}
	return earned
}

// CalculateOverviewOfflineIncome computes general coins earned while the
// player was on the overview/galaxy map screen when they quit.
//
// Formula: min(baseRatePerSec * elapsedSecs, gcCap)
func CalculateOverviewOfflineIncome(baseRatePerSec, elapsedSecs, gcCap float64) float64 {
	if baseRatePerSec <= 0 || elapsedSecs <= 0 || gcCap <= 0 {
		return 0
	}
	earned := baseRatePerSec * elapsedSecs
	if earned > gcCap {
		return gcCap
	}
	return earned
}
