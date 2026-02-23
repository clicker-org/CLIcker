package achievement

import "github.com/clicker-org/clicker/internal/gamestate"

// RegisterDefaults registers the baseline achievement set.
func RegisterDefaults(reg *AchievementRegistry) {
	reg.Register(Achievement{
		ID:          "first_click",
		Name:        "First Click",
		Description: "Click once in any world.",
		XPGrant:     25,
		Condition: func(gs gamestate.GameState) bool {
			return gs.Player.TotalClicks >= 1
		},
	})

	reg.Register(Achievement{
		ID:          "click_apprentice",
		Name:        "Click Apprentice",
		Description: "Reach 100 total clicks.",
		XPGrant:     50,
		Condition: func(gs gamestate.GameState) bool {
			return gs.Player.TotalClicks >= 100
		},
	})

	reg.Register(Achievement{
		ID:          "first_buyon",
		Name:        "Automation Begins",
		Description: "Buy your first buy-on.",
		XPGrant:     40,
		Condition: func(gs gamestate.GameState) bool {
			return totalOwnedBuyOns(gs) >= 1
		},
	})

	reg.Register(Achievement{
		ID:          "collector_10",
		Name:        "Collector",
		Description: "Own 10 total buy-ons across all worlds.",
		XPGrant:     100,
		Condition: func(gs gamestate.GameState) bool {
			return totalOwnedBuyOns(gs) >= 10
		},
	})

	reg.Register(Achievement{
		ID:          "terra_million",
		Name:        "Terra Millionaire",
		Description: "Earn 1,000,000 Terra-Coins.",
		XPGrant:     150,
		Condition: func(gs gamestate.GameState) bool {
			ws, ok := gs.Worlds["terra"]
			return ok && ws.TotalCoinsEarned >= 1_000_000
		},
	})

	reg.Register(Achievement{
		ID:          "aqua_million",
		Name:        "Aqua Millionaire",
		Description: "Earn 1,000,000 Aqua-Coins.",
		XPGrant:     150,
		Condition: func(gs gamestate.GameState) bool {
			ws, ok := gs.Worlds["aqua"]
			return ok && ws.TotalCoinsEarned >= 1_000_000
		},
	})

	reg.Register(Achievement{
		ID:          "first_prestige",
		Name:        "Ascension I",
		Description: "Perform your first prestige.",
		XPGrant:     180,
		Reward: &Reward{
			Type:  RewardTypeGeneralCoins,
			Value: 25,
		},
		Condition: func(gs gamestate.GameState) bool {
			return totalPrestiges(gs) >= 1
		},
	})

	reg.Register(Achievement{
		ID:          "prestige_10",
		Name:        "Prestige Veteran",
		Description: "Reach 10 total prestiges across worlds.",
		XPGrant:     300,
		Reward: &Reward{
			Type:  RewardTypeGeneralCoins,
			Value: 100,
		},
		Condition: func(gs gamestate.GameState) bool {
			return totalPrestiges(gs) >= 10
		},
	})

	reg.Register(Achievement{
		ID:          "worldhopper",
		Name:        "Worldhopper",
		Description: "Earn coins in two different worlds.",
		XPGrant:     90,
		Condition: func(gs gamestate.GameState) bool {
			activeWorlds := 0
			for _, earned := range gs.Player.WorldTotalCoinsEarned {
				if earned > 0 {
					activeWorlds++
				}
			}
			return activeWorlds >= 2
		},
	})

	reg.Register(Achievement{
		ID:          "level_5",
		Name:        "Rising Star",
		Description: "Reach account level 5.",
		XPGrant:     120,
		Condition: func(gs gamestate.GameState) bool {
			return gs.Player.Level >= 5
		},
	})
}

func totalOwnedBuyOns(gs gamestate.GameState) int {
	total := 0
	for _, ws := range gs.Worlds {
		for _, count := range ws.BuyOnCounts {
			total += count
		}
	}
	return total
}

func totalPrestiges(gs gamestate.GameState) int {
	total := 0
	for _, ws := range gs.Worlds {
		total += ws.PrestigeCount
	}
	return total
}
