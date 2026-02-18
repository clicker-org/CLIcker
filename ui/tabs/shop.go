package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/economy"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/upgrade"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// linesPerCard is the number of terminal rows a single buy-on card occupies:
// top border + 3 content rows + bottom border.
const linesPerCard = 5

// ShopTabModel is the [S]hop tab: a scrollable list of buy-on cards.
type ShopTabModel struct {
	eng     *engine.Engine
	worldID string
	t       theme.Theme
	width   int // terminal (content-area) width
	height  int // content-area height (same value passed to the modal)
	cursor  int // index of the selected buy-on
	scroll  int // index of the first visible buy-on
}

// NewShopTab constructs a ShopTabModel for the given world.
func NewShopTab(eng *engine.Engine, worldID string, t theme.Theme, width, height int) ShopTabModel {
	return ShopTabModel{
		eng:     eng,
		worldID: worldID,
		t:       t,
		width:   width,
		height:  height,
	}
}

// Resize returns a copy with updated dimensions.
func (m ShopTabModel) Resize(w, h int) ShopTabModel {
	m.width = w
	m.height = h
	return m
}

func (m ShopTabModel) Init() tea.Cmd { return nil }

func (m ShopTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	reg, hasReg := m.eng.UpgradeReg[m.worldID]
	if !hasReg {
		return m, nil
	}
	items := reg.ListBuyOns()
	if len(items) == 0 {
		return m, nil
	}

	playerLevel := m.eng.State.Player.Level
	lastUnlocked := -1
	for i, b := range items {
		if b.LevelRequirement() <= playerLevel {
			lastUnlocked = i
		}
	}
	if lastUnlocked < 0 {
		return m, nil
	}

	visCount := m.visibleCount()

	switch msg.(type) {
	case messages.NavUpMsg:
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scroll {
				m.scroll = m.cursor
			}
		}
	case messages.NavDownMsg:
		if m.cursor < lastUnlocked {
			m.cursor++
			if m.cursor >= m.scroll+visCount {
				m.scroll = m.cursor - visCount + 1
			}
		}
	case messages.NavConfirmMsg:
		if m.cursor <= lastUnlocked {
			b := items[m.cursor]
			m.eng.PurchaseBuyOn(m.worldID, b.ID(), playerLevel)
		}
	}

	return m, nil
}

// visibleCount returns how many cards fit in the modal's inner content area.
func (m ShopTabModel) visibleCount() int {
	// The modal computes: modalHeight = max(height-4, 6); innerHeight = modalHeight-2.
	modalInnerH := max(m.height-4, 6) - 2
	count := modalInnerH / linesPerCard
	if count < 1 {
		count = 1
	}
	return count
}

func (m ShopTabModel) View() string {
	reg, hasReg := m.eng.UpgradeReg[m.worldID]
	ws := m.eng.State.Worlds[m.worldID]
	if !hasReg || ws == nil {
		return "  No items available."
	}

	items := reg.ListBuyOns()
	if len(items) == 0 {
		return "  No items available."
	}

	playerLevel := m.eng.State.Player.Level

	coinSymbol := ""
	if w, ok := m.eng.WorldReg.Get(m.worldID); ok {
		coinSymbol = w.CoinSymbol()
	}

	// Modal inner width: mirrors modal.renderBox math.
	modalW := max(m.width*3/4, 40)
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	innerW := modalW - 2     // content width between modal's │ borders
	cardContentW := innerW - 2 // content width inside each card's own borders
	if cardContentW < 30 {
		cardContentW = 30
	}

	visCount := m.visibleCount()
	end := m.scroll + visCount
	if end > len(items) {
		end = len(items)
	}

	var sb strings.Builder

	if m.scroll > 0 {
		sb.WriteString(
			lipgloss.NewStyle().Foreground(lipgloss.Color(m.t.DimText())).
				Render("  ↑ more above") + "\n",
		)
	}

	for i := m.scroll; i < end; i++ {
		b := items[i]
		count := ws.BuyOnCounts[b.ID()]
		cost := upgrade.CostForNext(b, count)
		locked := b.LevelRequirement() > playerLevel
		selected := i == m.cursor
		canAfford := ws.Coins >= cost

		sb.WriteString(m.renderCard(b, cost, count, coinSymbol, locked, selected, canAfford, cardContentW))
		if i < end-1 {
			sb.WriteString("\n")
		}
	}

	if end < len(items) {
		sb.WriteString("\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color(m.t.DimText())).
				Render("  ↓ more below"),
		)
	}

	return sb.String()
}

// renderCard builds the 5-line bordered card string for a single buy-on.
func (m ShopTabModel) renderCard(
	b upgrade.BuyOn,
	cost float64,
	count int,
	coinSymbol string,
	locked, selected, canAfford bool,
	contentW int,
) string {
	dim := lipgloss.Color(m.t.DimText())
	primary := lipgloss.Color(m.t.PrimaryText())
	accent := lipgloss.Color(m.t.AccentColor())
	coinC := lipgloss.Color(m.t.CoinColor())
	errC := lipgloss.Color(m.t.ErrorColor())

	// Border color: accent when selected, white for normal, dim for locked.
	var borderHex string
	switch {
	case locked:
		borderHex = m.t.DimText()
	case selected:
		borderHex = m.t.AccentColor()
	default:
		borderHex = "#ffffff"
	}
	borderC := lipgloss.Color(borderHex)

	// Layout: fixed right column (level + own), variable left column.
	const rightW = 12
	leftW := contentW - rightW
	if leftW < 10 {
		leftW = 10
	}

	// ── Row 1: Name (left) │ LVL: N (right) ──────────────────────────
	nameText := shopTruncStr(b.Name(), leftW-1) // reserve 1 for the leading space
	var nameStyle lipgloss.Style
	switch {
	case locked:
		nameStyle = lipgloss.NewStyle().Foreground(dim).Bold(true)
	case selected:
		nameStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	default:
		nameStyle = lipgloss.NewStyle().Foreground(primary).Bold(true)
	}
	nameRender := nameStyle.Render(nameText)

	lvlText := fmt.Sprintf("LVL: %d", b.LevelRequirement())
	var lvlRender string
	if locked {
		lvlRender = lipgloss.NewStyle().Foreground(errC).Render(lvlText)
	} else {
		lvlRender = lipgloss.NewStyle().Foreground(dim).Render(lvlText)
	}

	row1 := shopPadVisual(" "+nameRender, leftW) + shopPadVisual(lvlRender, rightW)

	// ── Row 2: Description (left) │ Own: N (right) ────────────────────
	descText := shopTruncStr(b.Description(), leftW-1)
	descRender := lipgloss.NewStyle().Foreground(dim).Render(descText)

	ownText := fmt.Sprintf("Own: %d", count)
	ownRender := lipgloss.NewStyle().Foreground(dim).Render(ownText)

	row2 := shopPadVisual(" "+descRender, leftW) + shopPadVisual(ownRender, rightW)

	// ── Row 3: Cost (left) │ hint/lock (right) ─────────────────────────
	costText := "Cost: " + economy.FormatCoinsBare(cost) + " " + coinSymbol
	var costC lipgloss.Color
	switch {
	case locked:
		costC = dim
	case canAfford:
		costC = coinC
	default:
		costC = errC
	}
	costRender := lipgloss.NewStyle().Foreground(costC).Render(costText)

	var row3 string
	switch {
	case locked:
		lockRender := lipgloss.NewStyle().Foreground(errC).Bold(true).Render("[LOCKED]")
		row3 = shopPadVisual(" "+costRender, leftW) + shopPadVisual(lockRender, rightW)
	case selected:
		hintRender := lipgloss.NewStyle().Foreground(dim).Render("[ENTER]")
		row3 = shopPadVisual(" "+costRender, leftW) + shopPadVisual(hintRender, rightW)
	default:
		row3 = shopPadVisual(" "+costRender, contentW)
	}

	// ── Assemble bordered card ─────────────────────────────────────────
	borderSt := lipgloss.NewStyle().Foreground(borderC)
	top := borderSt.Render("┌" + strings.Repeat("─", contentW) + "┐")
	bot := borderSt.Render("└" + strings.Repeat("─", contentW) + "┘")
	side := borderSt.Render("│")

	makeRow := func(content string) string {
		visW := lipgloss.Width(content)
		pad := max(contentW-visW, 0)
		return side + content + strings.Repeat(" ", pad) + side
	}

	return strings.Join([]string{
		top,
		makeRow(row1),
		makeRow(row2),
		makeRow(row3),
		bot,
	}, "\n")
}

// shopTruncStr truncates s to at most maxW runes, appending "…" if cut.
func shopTruncStr(s string, maxW int) string {
	r := []rune(s)
	if len(r) <= maxW {
		return s
	}
	if maxW <= 1 {
		return "…"
	}
	return string(r[:maxW-1]) + "…"
}

// shopPadVisual right-pads s with spaces until its visual width equals w.
// If s is already >= w visually, it is returned unchanged.
func shopPadVisual(s string, w int) string {
	v := lipgloss.Width(s)
	if v >= w {
		return s
	}
	return s + strings.Repeat(" ", w-v)
}
