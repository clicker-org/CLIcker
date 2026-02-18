package messages

import "github.com/clicker-org/clicker/internal/engine"

// NavigateToOverviewMsg navigates to the overview/galaxy map screen.
type NavigateToOverviewMsg struct{}

// NavigateToDashboardMsg navigates to the dashboard screen.
type NavigateToDashboardMsg struct{}

// NavigateToWorldMsg navigates to a specific world screen.
type NavigateToWorldMsg struct{ WorldID string }

// OfflineReportDismissedMsg is sent when the offline report is closed.
type OfflineReportDismissedMsg struct{}

// AchievementUnlockedMsg is sent when an achievement is newly unlocked.
type AchievementUnlockedMsg struct{ ID string }

// LevelUpMsg is sent when the player gains a level.
type LevelUpMsg struct{ NewLevel int }

// GameTickMsg carries engine events from a tick.
type GameTickMsg struct{ Events []engine.EngineEvent }
