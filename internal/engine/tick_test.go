package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
	_ "github.com/clicker-org/clicker/internal/world/worlds"
)

func newTestEngineWithAchievement(t *testing.T, a achievement.Achievement) *Engine {
	t.Helper()
	achReg := achievement.NewAchievementRegistry()
	achReg.Register(a)

	gs := gamestate.NewGameState()
	for _, w := range world.DefaultRegistry.List() {
		gs.Worlds[w.ID()] = world.NewWorldState(w.ID(), w.BaseExchangeRate())
	}
	return New(gs, world.DefaultRegistry, achReg)
}

func countEvents(events []EngineEvent, typ EngineEventType) int {
	n := 0
	for _, ev := range events {
		if ev.Type == typ {
			n++
		}
	}
	return n
}

func TestTick_AchievementAppliesXPAndGeneralCoinsReward(t *testing.T) {
	eng := newTestEngineWithAchievement(t, achievement.Achievement{
		ID:      "test_reward_gc",
		Name:    "Test Reward GC",
		XPGrant: 120,
		Reward: &achievement.Reward{
			Type:  achievement.RewardTypeGeneralCoins,
			Value: 17.5,
		},
		Condition: func(gs gamestate.GameState) bool { return true },
	})

	events := eng.Tick(AchievCheckInterval)

	require.True(t, eng.Earned["test_reward_gc"])
	assert.Equal(t, 120, eng.State.Player.XP)
	assert.Equal(t, 2, eng.State.Player.Level, "120 XP should level from 1 to 2")
	assert.Equal(t, 17.5, eng.State.Player.GeneralCoins)
	assert.Equal(t, 17.5, eng.State.Player.LifetimeGeneralCoins)
	assert.Equal(t, 1, countEvents(events, EventAchievementUnlocked))
	assert.Equal(t, 1, countEvents(events, EventLevelUp))
}

func TestTick_AchievementRewardTypeXPAddsXP(t *testing.T) {
	eng := newTestEngineWithAchievement(t, achievement.Achievement{
		ID:      "test_reward_xp",
		Name:    "Test Reward XP",
		XPGrant: 0,
		Reward: &achievement.Reward{
			Type:  achievement.RewardTypeXP,
			Value: 60,
		},
		Condition: func(gs gamestate.GameState) bool { return true },
	})

	eng.Tick(AchievCheckInterval)

	require.True(t, eng.Earned["test_reward_xp"])
	assert.Equal(t, 60, eng.State.Player.XP)
	assert.Equal(t, 1, eng.State.Player.Level)
}
