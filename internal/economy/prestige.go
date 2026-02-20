package economy

import "math"

// PrestigeReward holds the result of a prestige calculation.
type PrestigeReward struct {
	GeneralCoinsEarned float64
	PrestigeMultiplier float64 // the NEW multiplier after this prestige
	XPGrant            int
}

// PrestigeMultiplierGain returns the factor by which the prestige multiplier grows
// when performing the n-th prestige (0-indexed: first prestige is n=0).
// Formula: 1 + 0.5/sqrt(n+1) — starts at ×1.5, diminishes with each prestige.
func PrestigeMultiplierGain(n int) float64 {
	return 1.0 + 0.5/math.Sqrt(float64(n+1))
}

// CalculatePrestigeReward computes the reward for prestiging a world.
// totalCoinsEarned: lifetime coins earned in this world (used as progress proxy).
// prestigeCount: number of times the world has already been prestiged (0 = first prestige).
// currentMultiplier: the world's current prestige multiplier (starts at 1.0).
func CalculatePrestigeReward(totalCoinsEarned float64, prestigeCount int, currentMultiplier float64) PrestigeReward {
	// General coins: proportional to sqrt of total coins earned.
	// 1 M TC → ~100 GC; 100 M TC → ~1 000 GC; 1 B TC → ~31 623 GC.
	gc := math.Sqrt(totalCoinsEarned) * 0.1

	// New multiplier stacks multiplicatively with diminishing returns.
	newMult := currentMultiplier * PrestigeMultiplierGain(prestigeCount)

	// XP scales with the prestige count: first prestige = 500 XP, second = 1 000 XP, etc.
	xp := 500 * (prestigeCount + 1)

	return PrestigeReward{
		GeneralCoinsEarned: gc,
		PrestigeMultiplier: newMult,
		XPGrant:            xp,
	}
}
