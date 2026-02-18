package achievement

import "github.com/clicker-org/clicker/internal/gamestate"

// CheckAchievements evaluates all unearned achievements against the current game
// state and returns the IDs of any newly unlocked achievements.
// It does NOT mutate the earned map — the caller is responsible for recording unlocks.
func CheckAchievements(
	gs gamestate.GameState,
	registry *AchievementRegistry,
	earned map[string]bool,
) []string {
	var unlocked []string
	for _, a := range registry.GetAll() {
		if earned[a.ID] {
			continue
		}
		if a.Condition != nil && a.Condition(gs) {
			unlocked = append(unlocked, a.ID)
		}
	}
	return unlocked
}
