package achievement

import "github.com/clicker-org/clicker/internal/gamestate"

// RewardType enumerates the kinds of rewards an achievement can grant.
type RewardType string

const (
	RewardTypeNone            RewardType = "none"
	RewardTypeXP              RewardType = "xp"
	RewardTypeGeneralCoins    RewardType = "general_coins"
	RewardTypeMultiplier      RewardType = "multiplier"
	RewardTypeCosmetic        RewardType = "cosmetic"
)

// Reward describes what an achievement grants on unlock.
type Reward struct {
	Type  RewardType
	Value float64
	// CosmeticID is set for cosmetic rewards.
	CosmeticID string
}

// Achievement is a single trackable achievement in the game.
type Achievement struct {
	ID          string
	Name        string
	Description string
	// Hidden achievements show as "???" until earned.
	Hidden    bool
	XPGrant   int
	Reward    *Reward
	Condition func(gs gamestate.GameState) bool
}
