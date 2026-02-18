package upgrade

import "github.com/clicker-org/clicker/internal/config"

// EffectiveMultiplier returns the combined CPS multiplier for a specific buy-on,
// considering all purchased upgrades that target it.
func EffectiveMultiplier(buyOnID string, upgrades []config.UpgradeConfig, purchased map[string]bool) float64 {
	mult := 1.0
	for _, u := range upgrades {
		if u.TargetBuyOnID == buyOnID && purchased[u.ID] {
			mult *= u.Multiplier
		}
	}
	return mult
}

// CalculateWorldCPS aggregates the total coins-per-second for a world.
//
//   total_CPS = sum(buyOn.BaseCPS * count * upgradeMult) * worldPrestigeMult * globalCPSMult
func CalculateWorldCPS(
	registry *WorldUpgradeRegistry,
	buyOnCounts map[string]int,
	purchasedUpgrades map[string]bool,
	worldPrestigeMult float64,
	globalCPSMult float64,
) float64 {
	upgrades := registry.ListUpgrades()
	total := 0.0
	for _, b := range registry.ListBuyOns() {
		count := buyOnCounts[b.ID()]
		if count == 0 {
			continue
		}
		mult := EffectiveMultiplier(b.ID(), upgrades, purchasedUpgrades)
		total += b.BaseCPS() * float64(count) * mult
	}
	return total * worldPrestigeMult * globalCPSMult
}
