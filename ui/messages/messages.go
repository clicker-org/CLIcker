package messages

import "github.com/clicker-org/clicker/internal/engine"

// NavigateToOverviewMsg navigates to the overview/galaxy map screen.
type NavigateToOverviewMsg struct{}

// NavigateToDashboardMsg navigates to the dashboard screen.
type NavigateToDashboardMsg struct{}

// NavigateToAchievementsMsg navigates to the global achievements screen.
type NavigateToAchievementsMsg struct{}

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

// PrestigeConfirmRequestedMsg is emitted by the prestige tab when the user
// triggers a prestige action that requires confirmation.
type PrestigeConfirmRequestedMsg struct{}

// ExchangeBoostConfirmRequestedMsg is emitted by the prestige tab when the
// user triggers an exchange boost that requires confirmation.
type ExchangeBoostConfirmRequestedMsg struct{}

// Directional navigation messages — emitted by App for arrow/vim keys.
// Screens respond to these instead of raw key strings so that new screens get
// keyboard navigation without repeating key-string switch cases.
//
// The pattern:  Nav*Msg moves focus (cursor),  NavConfirmMsg activates it.
// First-letter shortcuts bypass this flow and activate instantly.
type NavUpMsg struct{}
type NavDownMsg struct{}
type NavLeftMsg struct{}
type NavRightMsg struct{}
type NavConfirmMsg struct{}
