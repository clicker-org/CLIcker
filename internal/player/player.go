package player

// Player holds global state that persists across all prestiges.
type Player struct {
	XP                   int                `json:"xp"`
	Level                int                `json:"level"`
	GeneralCoins         float64            `json:"general_coins"`
	TotalClicks          int64              `json:"total_clicks"`
	TotalPlaySeconds     float64            `json:"total_play_seconds"`
	LifetimeGeneralCoins float64            `json:"lifetime_general_coins"`
	WorldTotalCoinsEarned map[string]float64 `json:"world_total_coins_earned"`
}

// NewPlayer returns a freshly initialized player.
func NewPlayer() Player {
	return Player{
		XP:                    0,
		Level:                 1,
		GeneralCoins:          0,
		TotalClicks:           0,
		TotalPlaySeconds:      0,
		LifetimeGeneralCoins:  0,
		WorldTotalCoinsEarned: make(map[string]float64),
	}
}
