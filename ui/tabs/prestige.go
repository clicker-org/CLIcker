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

// prestigeConfirmType tracks which action the inline confirm dialog is for.
type prestigeConfirmType int

const (
	confirmPrestigeNone        prestigeConfirmType = iota
	confirmPrestigeAction                          // awaiting prestige confirm
	confirmExchangeBoostAction                     // awaiting exchange boost confirm
)

// PrestigeTabModel is the [P]restige tab.
type PrestigeTabModel struct {
	eng     *engine.Engine
	worldID string
	t       theme.Theme
	width   int
	height  int

	progressBar components.ProgressBar

	// Inline confirm dialog state.
	confirmType prestigeConfirmType
	confirmLeft bool // true = confirm button focused, false = cancel button focused
}

// modalInnerWidth mirrors the width math used by components.Modal.renderBox so
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
		confirmType: confirmPrestigeNone,
		confirmLeft: true,
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "p", "P":
			if m.confirmType == confirmPrestigeNone && m.eng.CanPrestige(m.worldID) {
				m.confirmType = confirmPrestigeAction
				m.confirmLeft = true
			}
			return m, barCmd

		case "e", "E":
			if m.confirmType == confirmPrestigeNone && m.eng.CanExchangeBoost(m.worldID) {
				m.confirmType = confirmExchangeBoostAction
				m.confirmLeft = true
			}
			return m, barCmd

		case "esc":
			m.confirmType = confirmPrestigeNone
			return m, barCmd
		}

	case messages.NavLeftMsg:
		if m.confirmType != confirmPrestigeNone {
			m.confirmLeft = true
		}
		return m, barCmd

	case messages.NavRightMsg:
		if m.confirmType != confirmPrestigeNone {
			m.confirmLeft = false
		}
		return m, barCmd

	case messages.NavConfirmMsg:
		if m.confirmType == confirmPrestigeNone {
			return m, barCmd
		}
		if !m.confirmLeft {
			// Cancel selected.
			m.confirmType = confirmPrestigeNone
			return m, barCmd
		}
		// Confirm selected — execute the action.
		switch m.confirmType {
		case confirmPrestigeAction:
			m.eng.ExecutePrestige(m.worldID)
		case confirmExchangeBoostAction:
			m.eng.ExecuteExchangeBoost(m.worldID)
		}
		m.confirmType = confirmPrestigeNone
		return m, barCmd
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

	content := sb.String()

	// ── Inline confirm overlay ─────────────────────────────────────────
	if m.confirmType != confirmPrestigeNone {
		var title, question, actionLabel string
		switch m.confirmType {
		case confirmPrestigeAction:
			title = "CONFIRM PRESTIGE"
			question = fmt.Sprintf("Reset world. Earn: +%s GC  ×%.2f  +%d XP",
				economy.FormatCoinsBare(preview.GeneralCoinsEarned),
				preview.PrestigeMultiplier,
				preview.XPGrant,
			)
			actionLabel = "Prestige"
		case confirmExchangeBoostAction:
			boost := m.eng.ExchangeBoostPreview(m.worldID)
			title = "CONFIRM EXCHANGE BOOST"
			question = fmt.Sprintf("Sacrifice %s %s → earn %s GC",
				economy.FormatCoinsBare(boost.WorldCoinsCost),
				coinSymbol,
				economy.FormatCoinsBare(boost.GeneralCoinsEarned),
			)
			actionLabel = "Exchange"
		}

		box := renderInlineConfirm(title, question, actionLabel, m.confirmLeft, innerW, m.t)
		return content + "\n" + box
	}

	return content
}

// renderInlineConfirm builds a small bordered confirm dialog embedded inside
// the prestige tab's content area. The box width equals innerW so it aligns
// with the modal's own content column.
func renderInlineConfirm(title, question, actionLabel string, confirmFocused bool, innerW int, t theme.Theme) string {
	const borderHex = "#ffffff"
	borderSt := lipgloss.NewStyle().Foreground(lipgloss.Color(borderHex))
	dimSt := lipgloss.NewStyle().Foreground(lipgloss.Color(t.DimText()))
	activeSt := lipgloss.NewStyle().Foreground(lipgloss.Color(t.AccentColor())).Bold(true).Underline(true)

	// The box fits inside the modal's inner column with a 2-char left indent.
	const indent = "  "
	boxW := innerW - len(indent)
	if boxW < 30 {
		boxW = 30
	}

	// Top border with centered title.
	titleStr := " " + title + " "
	dashes := max(boxW-len(titleStr), 0)
	leftD := dashes / 2
	rightD := dashes - leftD
	top := borderSt.Render("┌" + strings.Repeat("─", leftD) + titleStr + strings.Repeat("─", rightD) + "┐")
	bot := borderSt.Render("└" + strings.Repeat("─", boxW) + "┘")
	side := borderSt.Render("│")

	// Question line (truncate if it exceeds the inner box width).
	qRunes := []rune(question)
	innerBoxW := boxW - 2 // space on each side inside the border
	if len(qRunes) > innerBoxW {
		if innerBoxW > 1 {
			qRunes = append(qRunes[:innerBoxW-1], '…')
		} else {
			qRunes = qRunes[:innerBoxW]
		}
	}
	questionStr := string(qRunes)
	qPad := max(innerBoxW-len([]rune(questionStr)), 0)
	qLeft := qPad / 2
	qRight := qPad - qLeft
	questionLine := side + strings.Repeat(" ", qLeft) + questionStr + strings.Repeat(" ", qRight) + side

	// Buttons.
	cancelLabel := "[ Cancel ]"
	confirmLabel := "[ " + actionLabel + " ]"
	const gap = "   "

	var cancelStr, confirmStr string
	if confirmFocused {
		confirmStr = activeSt.Render(confirmLabel)
		cancelStr = dimSt.Render(cancelLabel)
	} else {
		confirmStr = dimSt.Render(confirmLabel)
		cancelStr = activeSt.Render(cancelLabel)
	}

	totalBtnW := lipgloss.Width(confirmLabel) + len(gap) + lipgloss.Width(cancelLabel)
	bPad := max(innerBoxW-totalBtnW, 0)
	bLeft := bPad / 2
	bRight := bPad - bLeft
	buttonLine := side + strings.Repeat(" ", bLeft) + confirmStr + gap + cancelStr + strings.Repeat(" ", bRight) + side

	blank := side + strings.Repeat(" ", boxW) + side

	lines := []string{
		indent + top,
		indent + blank,
		indent + questionLine,
		indent + blank,
		indent + buttonLine,
		indent + blank,
		indent + bot,
	}
	return strings.Join(lines, "\n")
}
