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

// ModalType identifies which overlay modal is currently open.
type ModalType int

const (
	ModalNone         ModalType = iota
	ModalShop                   // focusedHeader index 1
	ModalPrestige               // focusedHeader index 2
	ModalAchievements           // focusedHeader index 3
)

// headerCount is the number of items in the top header bar.
const headerCount = 4

// WorldModel hosts the world screen. The click view is always the background;
// Shop, Prestige, and Achievements open as modal overlays on top.
type WorldModel struct {
	t       theme.Theme
	eng     *engine.Engine
	gs      *gamestate.GameState
	worldID string

	clickTab    tabs.ClickTabModel
	shopTab     tabs.ShopTabModel
	prestigeTab tabs.PrestigeTabModel
	achTab      tabs.AchievementsTabModel

	// focusedHeader is the header item the arrow-key cursor sits on (0–3).
	// activeModal is the modal currently displayed; ModalNone means no overlay.
	focusedHeader int
	activeModal   ModalType
	modal         components.Modal

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
	contentH := max(height-4, 3)
	return WorldModel{
		t:            t,
		eng:          eng,
		gs:           gs,
		worldID:      worldID,
		clickTab:     tabs.NewClickTab(eng, worldID, t, animReg, animKey, width, contentH),
		shopTab:      tabs.NewShopTab(eng, worldID, t, width, contentH),
		prestigeTab:  tabs.NewPrestigeTab(eng, worldID, t, width, contentH),
		statusBar:    components.NewStatusBar(t, width),
		width:        width,
		height:       height,
		activeModal:  ModalNone,
		activeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.AccentColor())).
			Bold(true).
			Underline(true),
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
		contentH := max(msg.Height-4, 3)
		m.clickTab = m.clickTab.Resize(msg.Width, contentH)
		m.shopTab = m.shopTab.Resize(msg.Width, contentH)
		m.prestigeTab = m.prestigeTab.Resize(msg.Width, contentH)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.activeModal != ModalNone {
				m.activeModal = ModalNone
				return m, nil
			}
			return m, func() tea.Msg { return messages.NavigateToOverviewMsg{} }

		case "c", "C":
			// C always returns to the plain click view (closes any open modal).
			m.activeModal = ModalNone
			m.focusedHeader = 0
			return m, nil

		case "s", "S":
			if m.activeModal == ModalNone {
				m.activeModal = ModalShop
				m.focusedHeader = 1
				m.modal = components.NewModal(m.t)
			} else if m.activeModal == ModalShop {
				m.activeModal = ModalNone
				m.focusedHeader = 0
			}
			return m, nil

		case "p", "P":
			if m.activeModal == ModalNone {
				// Open the prestige modal.
				m.activeModal = ModalPrestige
				m.focusedHeader = 2
				m.modal = components.NewModal(m.t)
				return m, nil
			} else if m.activeModal == ModalPrestige {
				// Forward P to the prestige tab (triggers prestige confirm).
				newModel, c := m.prestigeTab.Update(msg)
				if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
					m.prestigeTab = pt
				}
				return m, c
			}
			return m, nil

		case "e", "E":
			if m.activeModal == ModalPrestige {
				// Forward E to the prestige tab (triggers exchange boost confirm).
				newModel, c := m.prestigeTab.Update(msg)
				if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
					m.prestigeTab = pt
				}
				return m, c
			}
			return m, nil

		case "a", "A":
			if m.activeModal == ModalNone {
				m.activeModal = ModalAchievements
				m.focusedHeader = 3
				m.modal = components.NewModal(m.t)
			} else if m.activeModal == ModalAchievements {
				m.activeModal = ModalNone
				m.focusedHeader = 0
			}
			return m, nil

		case "tab":
			if m.activeModal == ModalNone {
				m.focusedHeader = (m.focusedHeader + 1) % headerCount
			}
			return m, nil

		case "shift+tab":
			if m.activeModal == ModalNone {
				m.focusedHeader = (m.focusedHeader + headerCount - 1) % headerCount
			}
			return m, nil
		}
		// Block all other key input from reaching tabs when a modal is open.
		if m.activeModal != ModalNone {
			return m, nil
		}

	case components.ModalCloseMsg:
		m.activeModal = ModalNone
		return m, nil

	// Arrow-key cursor: moves the header focus when no modal is open.
	// When the prestige modal is open, Left/Right are forwarded to the tab
	// (for confirm button navigation).
	case messages.NavLeftMsg:
		if m.activeModal == ModalNone {
			m.focusedHeader = (m.focusedHeader + headerCount - 1) % headerCount
			return m, nil
		}
		if m.activeModal == ModalPrestige {
			newModel, c := m.prestigeTab.Update(msg)
			if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
				m.prestigeTab = pt
			}
			return m, c
		}
		return m, nil

	case messages.NavRightMsg:
		if m.activeModal == ModalNone {
			m.focusedHeader = (m.focusedHeader + 1) % headerCount
			return m, nil
		}
		if m.activeModal == ModalPrestige {
			newModel, c := m.prestigeTab.Update(msg)
			if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
				m.prestigeTab = pt
			}
			return m, c
		}
		return m, nil
	}

	var cmds []tea.Cmd

	// Forward nav messages to the active modal's content tab.
	// The shop tab uses up/down to navigate its list and confirm to purchase.
	// The prestige tab uses up/down/confirm for its inline confirm dialog.
	// Achievements and other modals use the outer modal's Esc-button behaviour.
	if m.activeModal != ModalNone {
		switch msg.(type) {
		case messages.NavUpMsg, messages.NavDownMsg, messages.NavConfirmMsg:
			if m.activeModal == ModalShop {
				newModel, c := m.shopTab.Update(msg)
				if st, ok := newModel.(tabs.ShopTabModel); ok {
					m.shopTab = st
				}
				if c != nil {
					cmds = append(cmds, c)
				}
			} else if m.activeModal == ModalPrestige {
				newModel, c := m.prestigeTab.Update(msg)
				if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
					m.prestigeTab = pt
				}
				if c != nil {
					cmds = append(cmds, c)
				}
			} else {
				newModal, c := m.modal.Update(msg)
				m.modal = newModal
				if c != nil {
					cmds = append(cmds, c)
				}
			}
			return m, tea.Batch(cmds...)
		}
	} else {
		// When no modal: Enter (NavConfirmMsg) opens the focused header's modal.
		if _, ok := msg.(messages.NavConfirmMsg); ok {
			switch m.focusedHeader {
			case 1:
				m.activeModal = ModalShop
				m.modal = components.NewModal(m.t)
			case 2:
				m.activeModal = ModalPrestige
				m.modal = components.NewModal(m.t)
			case 3:
				m.activeModal = ModalAchievements
				m.modal = components.NewModal(m.t)
			}
			return m, nil
		}
	}

	// Click tab always receives non-key messages (background animations).
	// It also receives everything when no modal is open.
	_, isKey := msg.(tea.KeyMsg)
	if !isKey || m.activeModal == ModalNone {
		newModel, c := m.clickTab.Update(msg)
		if ct, ok := newModel.(tabs.ClickTabModel); ok {
			m.clickTab = ct
		}
		if c != nil {
			cmds = append(cmds, c)
		}
	}

	// Active modal tab receives messages when open.
	switch m.activeModal {
	case ModalShop:
		newModel, c := m.shopTab.Update(msg)
		if st, ok := newModel.(tabs.ShopTabModel); ok {
			m.shopTab = st
		}
		if c != nil {
			cmds = append(cmds, c)
		}
	case ModalPrestige:
		newModel, c := m.prestigeTab.Update(msg)
		if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
			m.prestigeTab = pt
		}
		if c != nil {
			cmds = append(cmds, c)
		}
	case ModalAchievements:
		newModel, c := m.achTab.Update(msg)
		if at, ok := newModel.(tabs.AchievementsTabModel); ok {
			m.achTab = at
		}
		if c != nil {
			cmds = append(cmds, c)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m WorldModel) View() string {
	bg := lipgloss.Color(m.t.Background())
	borderFg := lipgloss.Color(m.t.BorderColor())

	// Header: [C]lick is always the active background view.
	// Arrow keys move focusedHeader; Enter opens the focused modal.
	type headerItem struct {
		label string
		modal ModalType
	}
	items := []headerItem{
		{"[C]lick", ModalNone},
		{"[S]hop", ModalShop},
		{"[P]restige", ModalPrestige},
		{"[A]chievements", ModalAchievements},
	}

	headerParts := make([]string, len(items))
	for i, item := range items {
		var style lipgloss.Style
		switch {
		case m.activeModal != ModalNone && item.modal == m.activeModal:
			// The currently-open modal's header item.
			style = m.activeStyle
		case m.activeModal == ModalNone && i == m.focusedHeader:
			// Arrow-key cursor is here.
			style = m.focusedStyle
		case m.activeModal == ModalNone && item.modal == ModalNone:
			// Click is the active background view (when cursor is elsewhere).
			style = m.activeStyle
		default:
			style = m.dimStyle
		}
		headerParts[i] = style.Render(item.label)
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

	contentHeight := max(m.height-4, 3)

	// Always render the click tab as the background content.
	bgContent := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Background(bg).
		Render(m.clickTab.View())

	var contentArea string
	if m.activeModal != ModalNone {
		var title, content string
		switch m.activeModal {
		case ModalShop:
			title, content = "SHOP", m.shopTab.View()
		case ModalPrestige:
			title, content = "PRESTIGE", m.prestigeTab.View()
		case ModalAchievements:
			title, content = "ACHIEVEMENTS", m.achTab.View()
		}
		// Overlay the modal on top of the live click-tab background.
		contentArea = m.modal.View(title, content, bgContent, m.width, contentHeight)
	} else {
		contentArea = bgContent
	}

	ws := m.eng.State.Worlds[m.worldID]
	statusBar := m.statusBar.View(*m.gs, m.worldID, ws)

	return header + "\n" + divider + "\n" + contentArea + "\n" + divider + "\n" + statusBar
}
