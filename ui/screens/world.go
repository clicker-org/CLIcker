package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/ui/components"
	"github.com/clicker-org/clicker/ui/components/background"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/tabs"
	"github.com/clicker-org/clicker/ui/theme"
)

// TabID identifies which world tab is active.
type TabID int

const (
	TabClick TabID = iota
	TabShop
	TabPrestige
	TabAchievements
)

// WorldModel hosts the tabbed world screen.
type WorldModel struct {
	t       theme.Theme
	eng     *engine.Engine
	gs      *gamestate.GameState
	worldID string

	// focusedTab is where the arrow-key cursor sits; activeTab is what's displayed.
	// They differ only while the player is navigating with arrows before pressing Enter.
	// First-letter shortcuts and Tab/Shift+Tab keep them in sync (instant switch).
	focusedTab  TabID
	activeTab   TabID
	clickTab    tabs.ClickTabModel
	shopTab     tabs.ShopTabModel
	prestigeTab tabs.PrestigeTabModel
	achTab      tabs.AchievementsTabModel

	statusBar    components.StatusBar
	width        int
	height       int
	activeStyle  lipgloss.Style
	focusedStyle lipgloss.Style
	dimStyle     lipgloss.Style
}

// NewWorldModel creates a WorldModel for the given world.
func NewWorldModel(
	t theme.Theme,
	eng *engine.Engine,
	gs *gamestate.GameState,
	worldID string,
	animReg *background.AnimationRegistry,
	animKey string,
	width, height int,
) WorldModel {
	contentH := height - 4
	if contentH < 3 {
		contentH = 3
	}
	return WorldModel{
		t:          t,
		eng:        eng,
		gs:         gs,
		worldID:    worldID,
		focusedTab: TabClick,
		activeTab:  TabClick,
		clickTab:   tabs.NewClickTab(eng, worldID, t, animReg, animKey, width, contentH),
		statusBar:  components.NewStatusBar(t, width),
		width:      width,
		height:     height,
		activeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.AccentColor())).
			Bold(true).
			Underline(true),
		// Focused-but-not-active: accent colour without bold so the cursor is
		// visible but clearly distinct from the currently displayed tab.
		focusedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.AccentColor())),
		dimStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.DimText())),
	}
}

func (m WorldModel) Init() tea.Cmd {
	return m.clickTab.Init()
}

func (m WorldModel) Update(msg tea.Msg) (WorldModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetWidth(msg.Width)
		contentH := msg.Height - 4
		if contentH < 3 {
			contentH = 3
		}
		m.clickTab = m.clickTab.Resize(msg.Width, contentH)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return messages.NavigateToOverviewMsg{} }
		// First-letter shortcuts: instant switch, keeps focused in sync.
		case "c", "C":
			m.focusedTab, m.activeTab = TabClick, TabClick
			return m, nil
		case "s", "S":
			m.focusedTab, m.activeTab = TabShop, TabShop
			return m, nil
		case "p", "P":
			m.focusedTab, m.activeTab = TabPrestige, TabPrestige
			return m, nil
		case "a", "A":
			m.focusedTab, m.activeTab = TabAchievements, TabAchievements
			return m, nil
		// Tab / Shift+Tab: cycle instantly, keeps focused in sync.
		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
			m.focusedTab = m.activeTab
			return m, nil
		case "shift+tab":
			m.activeTab = (m.activeTab + 3) % 4
			m.focusedTab = m.activeTab
			return m, nil
		}
	// Arrow keys move the focused cursor; Enter (NavConfirmMsg) activates it.
	case messages.NavLeftMsg:
		m.focusedTab = (m.focusedTab + 3) % 4
		return m, nil
	case messages.NavRightMsg:
		m.focusedTab = (m.focusedTab + 1) % 4
		return m, nil
	case messages.NavConfirmMsg:
		m.activeTab = m.focusedTab
		return m, nil
	// NavUp/NavDown fall through to the active tab for list scrolling.
	}

	// Route message to the active tab.
	var cmd tea.Cmd
	switch m.activeTab {
	case TabClick:
		newModel, c := m.clickTab.Update(msg)
		if ct, ok := newModel.(tabs.ClickTabModel); ok {
			m.clickTab = ct
		}
		cmd = c
	case TabShop:
		newModel, c := m.shopTab.Update(msg)
		if st, ok := newModel.(tabs.ShopTabModel); ok {
			m.shopTab = st
		}
		cmd = c
	case TabPrestige:
		newModel, c := m.prestigeTab.Update(msg)
		if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
			m.prestigeTab = pt
		}
		cmd = c
	case TabAchievements:
		newModel, c := m.achTab.Update(msg)
		if at, ok := newModel.(tabs.AchievementsTabModel); ok {
			m.achTab = at
		}
		cmd = c
	}
	return m, cmd
}

func (m WorldModel) View() string {
	bg := lipgloss.Color(m.t.Background())
	borderFg := lipgloss.Color(m.t.BorderColor())

	type tabDef struct {
		id    TabID
		label string
	}
	tabDefs := []tabDef{
		{TabClick, "[C]lick"},
		{TabShop, "[S]hop"},
		{TabPrestige, "[P]restige"},
		{TabAchievements, "[A]chievements"},
	}

	var headerParts []string
	for _, tab := range tabDefs {
		switch {
		case tab.id == m.activeTab:
			// Active: bold + underline (content is showing).
			headerParts = append(headerParts, m.activeStyle.Render(tab.label))
		case tab.id == m.focusedTab:
			// Focused but not yet confirmed: accent colour, no bold — cursor is here.
			headerParts = append(headerParts, m.focusedStyle.Render(tab.label))
		default:
			headerParts = append(headerParts, m.dimStyle.Render(tab.label))
		}
	}
	header := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Padding(0, 1).
		Render(strings.Join(headerParts, "  "))

	divider := lipgloss.NewStyle().
		Width(m.width).
		Background(bg).
		Foreground(borderFg).
		Render(strings.Repeat("─", m.width))

	var content string
	switch m.activeTab {
	case TabClick:
		content = m.clickTab.View()
	case TabShop:
		content = m.shopTab.View()
	case TabPrestige:
		content = m.prestigeTab.View()
	case TabAchievements:
		content = m.achTab.View()
	}

	// Content area fills the space between the two dividers.
	// Layout: header(1) + divider(1) + content(n) + divider(1) + statusbar(1) = height
	contentHeight := m.height - 4
	if contentHeight < 3 {
		contentHeight = 3
	}
	contentArea := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Background(bg).
		Render(content)

	ws := m.eng.State.Worlds[m.worldID]
	statusBar := m.statusBar.View(*m.gs, m.worldID, ws)

	return header + "\n" + divider + "\n" + contentArea + "\n" + divider + "\n" + statusBar
}
