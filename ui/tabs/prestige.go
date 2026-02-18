package tabs

import tea "github.com/charmbracelet/bubbletea"

// PrestigeTabModel is the [P]restige tab (stub for Phase 0).
type PrestigeTabModel struct{}

func (m PrestigeTabModel) Init() tea.Cmd                            { return nil }
func (m PrestigeTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m PrestigeTabModel) View() string {
	return "\n\n  PRESTIGE — coming in Phase 2\n"
}
