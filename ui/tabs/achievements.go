package tabs

import tea "github.com/charmbracelet/bubbletea"

// AchievementsTabModel is the [A]chievements tab (stub for Phase 0).
type AchievementsTabModel struct{}

func (m AchievementsTabModel) Init() tea.Cmd                            { return nil }
func (m AchievementsTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m AchievementsTabModel) View() string {
	return "\n\n  ACHIEVEMENTS — coming in Phase 3\n"
}
