package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
	"github.com/clicker-org/clicker/ui/components"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// OverviewModel is the galaxy map / overview screen.
type OverviewModel struct {
	t        theme.Theme
	gs       *gamestate.GameState
	worldReg *world.WorldRegistry
	gmap     components.GalaxyMap
	width    int
	height   int
}

// NewOverviewModel creates an OverviewModel.
func NewOverviewModel(
	t theme.Theme,
	gs *gamestate.GameState,
	worldReg *world.WorldRegistry,
	width, height int,
) OverviewModel {
	mapH := height - 3
	if mapH < 3 {
		mapH = 3
	}
	return OverviewModel{
		t:        t,
		gs:       gs,
		worldReg: worldReg,
		gmap:     components.GalaxyMap{Width: width, Height: mapH},
		width:    width,
		height:   height,
	}
}

func (m OverviewModel) Init() tea.Cmd { return nil }

func (m OverviewModel) Update(msg tea.Msg) (OverviewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		worldIDs := m.worldReg.IDs()
		switch msg.String() {
		case "d", "D":
			return m, func() tea.Msg { return messages.NavigateToDashboardMsg{} }
		case "enter":
			id := m.gmap.FocusedWorldID(worldIDs)
			if id == "" && len(worldIDs) > 0 {
				id = worldIDs[0]
			}
			if id != "" {
				wid := id
				return m, func() tea.Msg { return messages.NavigateToWorldMsg{WorldID: wid} }
			}
		case "left", "h":
			m.gmap.MoveLeft()
		case "right", "l":
			m.gmap.MoveRight(len(worldIDs))
		case "up", "k":
			m.gmap.MoveUp()
		case "down", "j":
			m.gmap.MoveDown(len(worldIDs))
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.gmap.Width = msg.Width
		mapH := msg.Height - 3
		if mapH < 3 {
			mapH = 3
		}
		m.gmap.Height = mapH
	}
	return m, nil
}

func (m OverviewModel) View() string {
	bg := lipgloss.Color(m.t.Background())
	worldIDs := m.worldReg.IDs()

	// Galaxy map fills available space; gmap handles its own height via lipgloss.Place.
	mapArea := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Render(m.gmap.View(worldIDs, m.t))

	divider := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Foreground(lipgloss.Color(m.t.BorderColor())).
		Render(strings.Repeat("─", m.width))

	statsLine := ""
	if m.gs != nil {
		p := m.gs.Player
		statsLine = fmt.Sprintf("  General Coins: %.2f GC  |  LVL: %d  |  XP: %d", p.GeneralCoins, p.Level, p.XP)
	}
	styledStats := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Foreground(lipgloss.Color(m.t.CoinColor())).
		Render(statsLine)

	helpLine := "  [Enter] Enter World   [D] Dashboard   [Q] Quit   [?] Help"
	styledHelp := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Foreground(lipgloss.Color(m.t.DimText())).
		Render(helpLine)

	return mapArea + "\n" + divider + "\n" + styledStats + "\n" + styledHelp
}
