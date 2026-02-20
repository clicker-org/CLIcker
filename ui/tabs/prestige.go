package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/economy"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/ui/components"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme"
)

// progressBarPrefix is the text that precedes the progress bar on its line.
// The bar width must leave exactly this much room so the line does not overflow.
const progressBarPrefix = "  Progress:  "

// PrestigeTabModel is the [P]restige tab.
type PrestigeTabModel struct {
	eng     *engine.Engine
	worldID string
	t       theme.Theme
	width   int
	height  int

	progressBar components.ProgressBar
}

// modalInnerWidth mirrors the width math used by components.TabModal so
// that content rendered by the prestige tab fits inside the modal exactly.
func modalInnerWidth(tabWidth int) int {
	modalW := max(tabWidth*3/4, 40)
	if modalW > tabWidth-4 {
		modalW = tabWidth - 4
	}
	return modalW - 2
}

// barWidth returns the correct progress-bar pixel width given the tab's outer
// width, subtracting the prefix label that sits to its left on the same line.
func barWidth(tabWidth int) int {
	w := modalInnerWidth(tabWidth) - len(progressBarPrefix)
	if w < 10 {
		w = 10
	}
	return w
}

// NewPrestigeTab constructs a PrestigeTabModel.
func NewPrestigeTab(eng *engine.Engine, worldID string, t theme.Theme, width, height int) PrestigeTabModel {
	return PrestigeTabModel{
		eng:         eng,
		worldID:     worldID,
		t:           t,
		width:       width,
		height:      height,
		progressBar: components.NewProgressBar(t, barWidth(width), ""),
	}
}

// Resize returns a copy with updated dimensions and a resized progress bar.
func (m PrestigeTabModel) Resize(w, h int) PrestigeTabModel {
	m.width = w
	m.height = h
	m.progressBar.SetWidth(barWidth(w))
	return m
}

func (m PrestigeTabModel) Init() tea.Cmd { return nil }

func (m PrestigeTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always forward messages to the progress bar so its animations stay live.
	var barCmd tea.Cmd
	m.progressBar, barCmd = m.progressBar.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "p", "P":
			if m.eng.CanPrestige(m.worldID) {
				return m, func() tea.Msg { return messages.PrestigeConfirmRequestedMsg{} }
			}
		case "e", "E":
			if m.eng.CanExchangeBoost(m.worldID) {
				return m, func() tea.Msg { return messages.ExchangeBoostConfirmRequestedMsg{} }
			}
		}
	}

	return m, barCmd
}

func (m PrestigeTabModel) View() string {
	ws := m.eng.State.Worlds[m.worldID]
	if ws == nil {
		return "  No world data."
	}

	w, hasWorld := m.eng.WorldReg.Get(m.worldID)

	coinSymbol := m.worldID
	if hasWorld {
		coinSymbol = w.CoinSymbol()
	}

	accent := lipgloss.Color(m.t.AccentColor())
	dim := lipgloss.Color(m.t.DimText())
	primary := lipgloss.Color(m.t.PrimaryText())
	coin := lipgloss.Color(m.t.CoinColor())
	success := lipgloss.Color(m.t.SuccessColor())
	warn := lipgloss.Color(m.t.WarningColor())

	accentSt := lipgloss.NewStyle().Foreground(accent)
	dimSt := lipgloss.NewStyle().Foreground(dim)
	primarySt := lipgloss.NewStyle().Foreground(primary)
	coinSt := lipgloss.NewStyle().Foreground(coin)
	successSt := lipgloss.NewStyle().Foreground(success)
	warnSt := lipgloss.NewStyle().Foreground(warn)

	innerW := modalInnerWidth(m.width)
	if innerW < 30 {
		innerW = 30
	}

	divider := dimSt.Render(strings.Repeat("─", innerW))

	var sb strings.Builder

	// ── Section 1: Prestige stats ──────────────────────────────────────
	sb.WriteString(fmt.Sprintf("  %s   %s\n",
		primarySt.Render(fmt.Sprintf("Prestige Count: %s",
			accentSt.Bold(true).Render(fmt.Sprintf("%d", ws.PrestigeCount)))),
		primarySt.Render(fmt.Sprintf("Multiplier: %s",
			accentSt.Bold(true).Render(fmt.Sprintf("%.2f×", ws.PrestigeMultiplier)))),
	))
	sb.WriteString("\n")

	// ── Progress bar ───────────────────────────────────────────────────
	current, threshold := m.eng.PrestigeProgress(m.worldID)
	pct := 0.0
	if threshold > 0 {
		pct = min(current/threshold, 1.0)
	}

	canPrestige := m.eng.CanPrestige(m.worldID)
	var progressLabel string
	if canPrestige {
		progressLabel = successSt.Bold(true).Render("READY TO PRESTIGE!")
	} else {
		progressLabel = dimSt.Render(fmt.Sprintf("%s / %s %s",
			economy.FormatCoinsBare(current),
			economy.FormatCoinsBare(threshold),
			coinSymbol,
		))
	}
	// progressBarPrefix + m.progressBar.View(pct) fits exactly within innerW.
	sb.WriteString(progressBarPrefix + m.progressBar.View(pct) + "\n")
	sb.WriteString(strings.Repeat(" ", len(progressBarPrefix)) + progressLabel + "\n")
	sb.WriteString("\n")

	// ── Prestige reward preview ────────────────────────────────────────
	preview := economy.CalculatePrestigeReward(ws.TotalCoinsEarned, ws.PrestigeCount, ws.PrestigeMultiplier)
	sb.WriteString(dimSt.Render("  Next prestige reward:") + "\n")
	sb.WriteString(fmt.Sprintf("    %s   %s   %s\n",
		coinSt.Render(fmt.Sprintf("+%s GC", economy.FormatCoinsBare(preview.GeneralCoinsEarned))),
		accentSt.Render(fmt.Sprintf("×%.2f multiplier", preview.PrestigeMultiplier)),
		successSt.Render(fmt.Sprintf("+%d XP", preview.XPGrant)),
	))
	sb.WriteString("\n")

	// Prestige action hint.
	if canPrestige {
		sb.WriteString("  " + warnSt.Bold(true).Render("[P]") + primarySt.Render(" Prestige (resets world)") + "\n")
	} else {
		sb.WriteString("  " + dimSt.Render("[P] Prestige (not yet available)") + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(divider + "\n")
	sb.WriteString("\n")

	// ── Section 2: Exchange Boost ──────────────────────────────────────
	sb.WriteString(primarySt.Bold(true).Render("  Exchange Boost") + "\n")
	sb.WriteString(dimSt.Render(fmt.Sprintf("  Current rate: %.6f GC per %s", ws.ExchangeRate, coinSymbol)) + "\n")
	sb.WriteString("\n")

	canExchange := m.eng.CanExchangeBoost(m.worldID)
	if canExchange {
		boost := m.eng.ExchangeBoostPreview(m.worldID)
		sb.WriteString(fmt.Sprintf("  Sacrifice %s (%s %s) → %s\n",
			warnSt.Render("20%"),
			coinSt.Render(economy.FormatCoinsBare(boost.WorldCoinsCost)),
			coinSymbol,
			coinSt.Bold(true).Render(fmt.Sprintf("+%s GC", economy.FormatCoinsBare(boost.GeneralCoinsEarned))),
		))
		sb.WriteString(dimSt.Render(fmt.Sprintf("  Rate after boost: %.6f GC/%s", boost.NewExchangeRate, coinSymbol)) + "\n")
		sb.WriteString("\n")
		sb.WriteString("  " + successSt.Bold(true).Render("[E]") + primarySt.Render(" Exchange Boost (no reset)") + "\n")
	} else {
		sb.WriteString(dimSt.Render("  No coins to exchange.") + "\n")
		sb.WriteString("\n")
		sb.WriteString("  " + dimSt.Render("[E] Exchange Boost (no reset)") + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString("  " + dimSt.Render(fmt.Sprintf("General Coins: %s",
		lipgloss.NewStyle().Foreground(coin).Bold(true).Render(
			economy.FormatCoinsBare(m.eng.State.Player.GeneralCoins)+" GC",
		),
	)) + "\n")

	return sb.String()
}
