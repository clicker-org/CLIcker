package player

// XPForLevel returns the cumulative XP required to reach level n from level 1.
// Formula per level step: cost(n→n+1) = int(100 * 1.5^(n-1))
func XPForLevel(n int) int {
	if n <= 1 {
		return 0
	}
	total := 0
	xp := 100.0
	for i := 1; i < n; i++ {
		total += int(xp)
		xp *= 1.5
	}
	return total
}

// XPNeededForNextLevel returns XP needed to advance from level n to n+1.
func XPNeededForNextLevel(n int) int {
	return XPForLevel(n+1) - XPForLevel(n)
}

// AddXP adds xp to the player, leveling up as needed.
// Returns true if at least one level-up occurred.
func AddXP(p *Player, xp int) bool {
	p.XP += xp
	leveled := false
	for p.XP >= XPForLevel(p.Level+1) {
		p.Level++
		leveled = true
	}
	return leveled
}

// LevelGateCheck returns true if the player meets the required level.
func LevelGateCheck(p Player, required int) bool {
	if required <= 0 {
		return true
	}
	return p.Level >= required
}
