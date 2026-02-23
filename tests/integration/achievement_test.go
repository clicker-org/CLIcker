package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
)

// newEngineWithAchievement creates an Engine wired to a registry that contains
// a single achievement with the given condition. Useful for targeted tests.
func newEngineWithAchievement(t *testing.T, a achievement.Achievement) *engine.Engine {
	t.Helper()
	achReg := achievement.NewAchievementRegistry()
	achReg.Register(a)

	gs := gamestate.NewGameState()
	for _, w := range world.DefaultRegistry.List() {
		gs.Worlds[w.ID()] = world.NewWorldState(w.ID(), w.BaseExchangeRate())
	}
	return engine.New(gs, world.DefaultRegistry, achReg)
}

func countUnlockEvents(events []engine.EngineEvent, id string) int {
	n := 0
	for _, ev := range events {
		if ev.Type == engine.EventAchievementUnlocked && ev.AchievementID == id {
			n++
		}
	}
	return n
}

// TestAchievement_UnlocksViaTick verifies the full flow: click → condition met →
// tick crosses debounce interval → EventAchievementUnlocked emitted → Earned updated.
func TestAchievement_UnlocksViaTick(t *testing.T) {
	eng := newEngineWithAchievement(t, achievement.Achievement{
		ID:   "test_first_click",
		Name: "First Click",
		Condition: func(gs gamestate.GameState) bool {
			return gs.Player.TotalClicks > 0
		},
	})

	eng.HandleClick("terra")

	// Tick exactly past the achievement check debounce interval.
	events := eng.Tick(engine.AchievCheckInterval)

	assert.Equal(t, 1, countUnlockEvents(events, "test_first_click"), "achievement unlock event should be emitted once")
	assert.True(t, eng.Earned["test_first_click"], "achievement should be marked as earned")
}

// TestAchievement_DoesNotFireBeforeConditionMet ensures the tracker does not
// emit unlock events when the condition is not satisfied.
func TestAchievement_DoesNotFireBeforeConditionMet(t *testing.T) {
	eng := newEngineWithAchievement(t, achievement.Achievement{
		ID:   "test_big_earner",
		Name: "Big Earner",
		Condition: func(gs gamestate.GameState) bool {
			ws, ok := gs.Worlds["terra"]
			return ok && ws.TotalCoinsEarned >= 1_000_000
		},
	})

	// Far below the threshold.
	eng.State.Worlds["terra"].TotalCoinsEarned = 100.0

	events := eng.Tick(engine.AchievCheckInterval)

	assert.Equal(t, 0, countUnlockEvents(events, "test_big_earner"))
	assert.False(t, eng.Earned["test_big_earner"])
}

// TestAchievement_DoesNotFireTwice verifies that once an achievement is earned,
// subsequent ticks do not re-emit the unlock event.
func TestAchievement_DoesNotFireTwice(t *testing.T) {
	eng := newEngineWithAchievement(t, achievement.Achievement{
		ID:        "test_always_true",
		Name:      "Always True",
		Condition: func(gs gamestate.GameState) bool { return true },
	})

	// First debounce window: should fire exactly once.
	events1 := eng.Tick(engine.AchievCheckInterval)
	assert.Equal(t, 1, countUnlockEvents(events1, "test_always_true"))

	// Second debounce window: already earned, must not fire again.
	events2 := eng.Tick(engine.AchievCheckInterval)
	assert.Equal(t, 0, countUnlockEvents(events2, "test_always_true"))
}

// TestAchievement_NoCheckBeforeDebounce verifies achievements are not evaluated
// before the debounce interval elapses (performance guard).
func TestAchievement_NoCheckBeforeDebounce(t *testing.T) {
	eng := newEngineWithAchievement(t, achievement.Achievement{
		ID:        "test_always_true_2",
		Name:      "Always True 2",
		Condition: func(gs gamestate.GameState) bool { return true },
	})

	// Tick for less than the debounce interval — no check should run.
	events := eng.Tick(engine.AchievCheckInterval - 0.1)

	assert.Equal(t, 0, countUnlockEvents(events, "test_always_true_2"))
	assert.False(t, eng.Earned["test_always_true_2"])
}
