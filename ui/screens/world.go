package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/economy"
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

// worldConfirmType identifies which action a confirm dialog is asking about.
type worldConfirmType int

const (
	worldConfirmNone     worldConfirmType = iota
	worldConfirmPrestige                  // confirm before prestige reset
	worldConfirmExchange                  // confirm before exchange boost
)

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
	modal         components.TabModal

	// confirmModal is the confirm dialog that overlays the prestige modal.
	// It is only shown when confirmOpen is true.
	confirmType  worldConfirmType
	confirmModal components.ConfirmModal
	confirmOpen  bool

	statusBar    components.StatusBar
	width        int
	height       int
	activeStyle  lipgloss.Style
	focusedStyle lipgloss.Style
	dimStyle     lipgloss.Style
}

// toggleModalHotkey applies consistent open/close behavior for tab hotkeys.
// Pressing the hotkey of an open modal closes it; otherwise it opens/switches to it.
func (m WorldModel) toggleModalHotkey(target ModalType, headerIdx int) WorldModel {
	if m.activeModal == target {
		m.activeModal = ModalNone
		m.focusedHeader = 0
		return m
	}
	m.activeModal = target
	m.focusedHeader = headerIdx
	m.modal = components.NewTabModal(m.t)
	return m
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
		statusBar:    components.NewStatusBar(t, width, eng.WorldReg),
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

// handleConfirmInput is the gate that intercepts all input when a confirm
// dialog is open. Returns (handled, newModel, cmd). Non-input messages (ticks,
// window size, etc.) are not consumed so they continue to drive animations.
func (m WorldModel) handleConfirmInput(msg tea.Msg) (bool, WorldModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.confirmOpen = false
		}
		return true, m, nil

	case messages.NavLeftMsg, messages.NavRightMsg,
		messages.NavUpMsg, messages.NavDownMsg:
		m.confirmModal, _ = m.confirmModal.Update(msg)
		return true, m, nil

	case messages.NavConfirmMsg:
		if m.confirmModal.ConfirmFocused() {
			switch m.confirmType {
			case worldConfirmPrestige:
				m.eng.ExecutePrestige(m.worldID)
			case worldConfirmExchange:
				m.eng.ExecuteExchangeBoost(m.worldID)
			}
		}
		m.confirmOpen = false
		return true, m, nil
	}
	return false, m, nil
}

// confirmContent returns the title and question string for the active confirm.
func (m WorldModel) confirmContent() (title, question string) {
	ws := m.eng.State.Worlds[m.worldID]
	switch m.confirmType {
	case worldConfirmPrestige:
		preview := economy.CalculatePrestigeReward(ws.TotalCoinsEarned, ws.PrestigeCount, ws.PrestigeMultiplier)
		return "CONFIRM PRESTIGE", fmt.Sprintf(
			"Reset world. Earn: +%s GC  ×%.2f  +%d XP",
			economy.FormatCoinsBare(preview.GeneralCoinsEarned),
			preview.PrestigeMultiplier,
			preview.XPGrant,
		)
	case worldConfirmExchange:
		boost := m.eng.ExchangeBoostPreview(m.worldID)
		coinSymbol := m.worldID
		if w, ok := m.eng.WorldReg.Get(m.worldID); ok {
			coinSymbol = w.CoinSymbol()
		}
		return "CONFIRM EXCHANGE BOOST", fmt.Sprintf(
			"Sacrifice %s %s → earn %s GC",
			economy.FormatCoinsBare(boost.WorldCoinsCost),
			coinSymbol,
			economy.FormatCoinsBare(boost.GeneralCoinsEarned),
		)
	}
	return "", ""
}

func (m WorldModel) Update(msg tea.Msg) (WorldModel, tea.Cmd) {
	// Confirm dialog acts as a gate: it captures all input when open.
	// Non-input messages (ticks, window resize) pass through unchanged so
	// background animations keep running.
	if m.confirmOpen {
		if handled, newM, cmd := m.handleConfirmInput(msg); handled {
			return newM, cmd
		}
	}

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
			// Esc closes layers from the top down.
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
			m = m.toggleModalHotkey(ModalShop, 1)
			return m, nil

		case "p", "P":
			m = m.toggleModalHotkey(ModalPrestige, 2)
			return m, nil

		case "e", "E":
			if m.activeModal == ModalPrestige {
				// Forward E to the prestige tab (may emit ExchangeBoostConfirmRequestedMsg).
				newModel, c := m.prestigeTab.Update(msg)
				if pt, ok := newModel.(tabs.PrestigeTabModel); ok {
					m.prestigeTab = pt
				}
				return m, c
			}
			return m, nil

		case "a", "A":
			m = m.toggleModalHotkey(ModalAchievements, 3)
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

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.activeModal == ModalShop {
				newModel, c := m.shopTab.Update(msg)
				if st, ok := newModel.(tabs.ShopTabModel); ok {
					m.shopTab = st
				}
				return m, c
			}
		}
		// Block all other key input from reaching tabs when a modal is open.
		if m.activeModal != ModalNone {
			return m, nil
		}

	case components.ModalCloseMsg:
		m.activeModal = ModalNone
		return m, nil

	// Prestige tab emits these when P/E are pressed inside the prestige modal.
	case messages.PrestigeConfirmRequestedMsg:
		m.confirmType = worldConfirmPrestige
		m.confirmModal = components.NewConfirmModal(m.t, "Prestige")
		m.confirmOpen = true
		return m, nil

	case messages.ExchangeBoostConfirmRequestedMsg:
		m.confirmType = worldConfirmExchange
		m.confirmModal = components.NewConfirmModal(m.t, "Exchange")
		m.confirmOpen = true
		return m, nil

	// Arrow-key cursor: moves the header focus when no modal is open.
	case messages.NavLeftMsg:
		if m.activeModal == ModalNone {
			m.focusedHeader = (m.focusedHeader + headerCount - 1) % headerCount
		}
		return m, nil

	case messages.NavRightMsg:
		if m.activeModal == ModalNone {
			m.focusedHeader = (m.focusedHeader + 1) % headerCount
		}
		return m, nil
	}

	var cmds []tea.Cmd

	// Forward nav messages to the active modal's content tab.
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
				// Enter confirms prestige from the modal content; arrows still control [Esc] focus.
				if _, isConfirm := msg.(messages.NavConfirmMsg); isConfirm {
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
			} else {
				// For Achievements (and any other non-interactive modal content),
				// route nav input to the outer modal's [Esc] button.
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
				m.modal = components.NewTabModal(m.t)
			case 2:
				m.activeModal = ModalPrestige
				m.modal = components.NewTabModal(m.t)
			case 3:
				m.activeModal = ModalAchievements
				m.modal = components.NewTabModal(m.t)
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
			style = m.activeStyle
		case m.activeModal == ModalNone && i == m.focusedHeader:
			style = m.focusedStyle
		case m.activeModal == ModalNone && item.modal == ModalNone:
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
		modalView := m.modal.View(title, content, bgContent, m.width, contentHeight)

		if m.confirmOpen {
			// Overlay the confirm dialog on top of the prestige modal.
			confirmTitle, confirmQuestion := m.confirmContent()
			contentArea = m.confirmModal.View(confirmTitle, confirmQuestion, modalView, m.width, contentHeight)
		} else {
			contentArea = modalView
		}
	} else {
		contentArea = bgContent
	}

	ws := m.eng.State.Worlds[m.worldID]
	statusBar := m.statusBar.View(*m.gs, m.worldID, ws)

	return header + "\n" + divider + "\n" + contentArea + "\n" + divider + "\n" + statusBar
}
