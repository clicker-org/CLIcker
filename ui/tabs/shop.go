package tabs

import tea "github.com/charmbracelet/bubbletea"

// ShopTabModel is the [S]hop tab (stub for Phase 0).
type ShopTabModel struct{}

func (m ShopTabModel) Init() tea.Cmd                            { return nil }
func (m ShopTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m ShopTabModel) View() string {
	return "\n\n  SHOP — coming in Phase 1\n"
}
