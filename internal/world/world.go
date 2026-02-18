package world

import "github.com/clicker-org/clicker/internal/config"

const HoursPerOfflineCapUpgrade = 2.0

type World interface {
	ID() string
	Name() string
	CoinName() string
	CoinSymbol() string
	AccentColor() string
	AmbientAnimation() string
	BaseExchangeRate() float64
	OfflinePercentage() float64
	OfflineCapHours() float64
	Config() config.WorldConfig
}

type WorldState struct {
	WorldID string `json:"world_id"`

	Coins            float64 `json:"coins"`
	TotalCoinsEarned float64 `json:"total_coins_earned"`
	CPS              float64 `json:"cps"`

	BuyOnCounts       map[string]int  `json:"buy_on_counts"`
	PurchasedUpgrades map[string]bool `json:"purchased_upgrades"`

	PrestigeCount      int     `json:"prestige_count"`
	PrestigeMultiplier float64 `json:"prestige_multiplier"`

	ExchangeRate float64 `json:"exchange_rate"`

	OfflineCapUpgradeLevel int `json:"offline_cap_upgrade_level"`

	CompletionPercent float64 `json:"completion_percent"`

	TotalClicks int64 `json:"total_clicks"`
}

func NewWorldState(worldID string, baseExchangeRate float64) *WorldState {
	return &WorldState{
		WorldID:           worldID,
		Coins:             0,
		TotalCoinsEarned:  0,
		CPS:               0,
		BuyOnCounts:       make(map[string]int),
		PurchasedUpgrades: make(map[string]bool),
		PrestigeCount:     0,
		PrestigeMultiplier: 1.0,
		ExchangeRate:      baseExchangeRate,
		OfflineCapUpgradeLevel: 0,
		CompletionPercent: 0,
		TotalClicks:       0,
	}
}

// EffectiveOfflineCapHours returns the total offline cap hours for a world,
// accounting for purchased upgrades. This follows a linear function
func EffectiveOfflineCapHours(ws *WorldState, baseCapHours float64) float64 {
	return baseCapHours + float64(ws.OfflineCapUpgradeLevel)*HoursPerOfflineCapUpgrade
}
