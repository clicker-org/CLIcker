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

const (
	overviewTitleHeight = 6
	overviewFooterLines = 3
	overviewTitleGap    = 1
)

// NewOverviewModel creates an OverviewModel.
func NewOverviewModel(
	t theme.Theme,
	gs *gamestate.GameState,
	worldReg *world.WorldRegistry,
	width, height int,
) OverviewModel {
	mapH := height - (overviewTitleHeight + overviewTitleGap + overviewFooterLines)
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
	worlds := m.worldVisuals()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d", "D":
			return m, func() tea.Msg { return messages.NavigateToDashboardMsg{} }
		case "a", "A":
			return m, func() tea.Msg { return messages.NavigateToAchievementsMsg{} }
		}
	case messages.NavConfirmMsg:
		id := m.gmap.FocusedWorldID(worlds)
		if id == "" && len(worlds) > 0 {
			id = worlds[0].ID
		}
		if id != "" {
			wid := id
			return m, func() tea.Msg { return messages.NavigateToWorldMsg{WorldID: wid} }
		}
	case messages.NavLeftMsg:
		m.gmap.MoveLeft(len(worlds))
	case messages.NavRightMsg:
		m.gmap.MoveRight(len(worlds))
	case messages.NavUpMsg:
		m.gmap.MoveUp(len(worlds))
	case messages.NavDownMsg:
		m.gmap.MoveDown(len(worlds))
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.gmap.Width = msg.Width
		mapH := msg.Height - (overviewTitleHeight + overviewTitleGap + overviewFooterLines)
		if mapH < 3 {
			mapH = 3
		}
		m.gmap.Height = mapH
	}
	return m, nil
}

func (m OverviewModel) View() string {
	bg := lipgloss.Color(m.t.Background())
	accent := lipgloss.Color(m.t.AccentColor())
	worlds := m.worldVisuals()

	title := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Foreground(accent).
		Bold(true).
		Render(strings.Join([]string{
			"   _____ _      _____      _            ",
			"  / ____| |    |_   _|    | |           ",
			" | |    | |      | |   ___| | _____ _ __",
			" | |    | |      | |  / __| |/ / _ \\ '__|",
			" | |____| |____ _| |_| (__|   <  __/ |   ",
			"  \\_____|______|_____\\___|_|\\_\\___|_|   ",
		}, "\n"))

	// Galaxy map fills available space; gmap handles its own height via lipgloss.Place.
	mapArea := lipgloss.NewStyle().
		Width(m.width).
		Height(m.gmap.Height).
		Background(bg).
		Render(m.gmap.View(worlds, m.t))

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

	helpLine := "  [Enter] Enter World   [D] Dashboard   [A] Achievements   [Q] Quit   [?] Help"
	styledHelp := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Foreground(lipgloss.Color(m.t.DimText())).
		Render(helpLine)

	return title + "\n" + mapArea + "\n" + divider + "\n" + styledStats + "\n" + styledHelp
}

func (m OverviewModel) worldVisuals() []components.WorldVisual {
	list := m.worldReg.List()
	visuals := make([]components.WorldVisual, 0, len(list))
	for _, w := range list {
		var ws *world.WorldState
		if m.gs != nil {
			ws = m.gs.Worlds[w.ID()]
		}
		v := components.WorldVisual{
			ID:          w.ID(),
			Name:        w.Name(),
			AccentColor: w.AccentColor(),
		}
		if ws != nil {
			v.Completion = ws.CompletionPercent
			v.Coins = ws.Coins
			v.CPS = ws.CPS
			v.Prestige = ws.PrestigeCount
		}
		visuals = append(visuals, v)
	}
	return visuals
}
