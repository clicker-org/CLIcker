package engine

import (
	"math"

	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/player"
)

// EngineEventType identifies the kind of engine event.
type EngineEventType string

const (
	EventAchievementUnlocked EngineEventType = "achievement_unlocked"
	EventLevelUp             EngineEventType = "level_up"
	EventAutoSave            EngineEventType = "autosave"
)

// EngineEvent is emitted by Tick to communicate side-effects to the UI layer.
type EngineEvent struct {
	Type EngineEventType
	// For EventAchievementUnlocked: ID of the achievement.
	AchievementID string
	// For EventLevelUp: the new level.
	NewLevel int
}

// Timing constants.
const (
	AutoSaveInterval    = 30.0  // seconds
	AchievCheckInterval = 5.0   // seconds
	TickIntervalMs      = 100   // milliseconds per tick
)

// Tick advances the engine by dt seconds and returns any events that occurred.
func (e *Engine) Tick(dt float64) []EngineEvent {
	var events []EngineEvent

	// 1. Apply CPS to all active worlds.
	for _, ws := range e.State.Worlds {
		if ws.CPS > 0 {
			earned := ws.CPS * dt
			ws.Coins += earned
			ws.TotalCoinsEarned += earned
			if e.State.Player.WorldTotalCoinsEarned == nil {
				e.State.Player.WorldTotalCoinsEarned = make(map[string]float64)
			}
			e.State.Player.WorldTotalCoinsEarned[ws.WorldID] += earned
		}
	}

	// 2. Update total play seconds.
	e.State.Player.TotalPlaySeconds += dt

	// 3. Debounced achievement check.
	e.achievCheckTimer += dt
	if e.achievCheckTimer >= AchievCheckInterval {
		e.achievCheckTimer = 0
		newlyUnlocked := achievement.CheckAchievements(e.State, e.AchievReg, e.Earned)
		for _, id := range newlyUnlocked {
			e.Earned[id] = true
			prevLevel := e.State.Player.Level
			if a, ok := e.AchievReg.Get(id); ok {
				if a.XPGrant > 0 {
					player.AddXP(&e.State.Player, a.XPGrant)
				}
				if a.Reward != nil {
					switch a.Reward.Type {
					case achievement.RewardTypeXP:
						player.AddXP(&e.State.Player, int(math.Round(a.Reward.Value)))
					case achievement.RewardTypeGeneralCoins:
						e.State.Player.GeneralCoins += a.Reward.Value
						e.State.Player.LifetimeGeneralCoins += a.Reward.Value
					}
				}
			}
			if e.State.Player.Level > prevLevel {
				events = append(events, EngineEvent{
					Type:     EventLevelUp,
					NewLevel: e.State.Player.Level,
				})
			}
			events = append(events, EngineEvent{
				Type:          EventAchievementUnlocked,
				AchievementID: id,
			})
		}
	}

	// 4. Autosave timer.
	e.autosaveTimer += dt
	if e.autosaveTimer >= AutoSaveInterval {
		e.autosaveTimer = 0
		events = append(events, EngineEvent{Type: EventAutoSave})
	}

	return events
}
