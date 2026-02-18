package economy

// PrestigeReward holds the result of a prestige calculation.
type PrestigeReward struct {
	GeneralCoinsEarned float64
	PrestigeMultiplier float64
	XPGrant            int
}

// CalculatePrestigeReward computes the reward for prestiging a world.
// totalCoinsEarned: lifetime coins earned in this world.
// prestigeCount: number of times the world has already been prestiged.
// Stub: formula TBD in Phase 2.
func CalculatePrestigeReward(totalCoinsEarned float64, prestigeCount int) PrestigeReward {
	return PrestigeReward{}
}
