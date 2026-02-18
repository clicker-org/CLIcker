package engine

// ScreenID identifies the active UI screen.
type ScreenID string

const (
	ScreenOverview      ScreenID = "overview"
	ScreenWorld         ScreenID = "world"
	ScreenDashboard     ScreenID = "dashboard"
	ScreenOfflineReport ScreenID = "offline_report"
)

// NavigateTo returns the target screen ID.
// Phase 0: simple passthrough. Future phases may add transition logic.
func NavigateTo(current, target ScreenID) ScreenID {
	return target
}
