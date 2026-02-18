package offline

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
