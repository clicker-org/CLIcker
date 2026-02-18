package tabs

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/ui/components/background"
	"github.com/clicker-org/clicker/ui/theme"
)

type clickFlashEndMsg struct{}
type coinFloatEndMsg struct{}

// ClickTabModel is the [C]lick tab content model.
type ClickTabModel struct {
	eng          *engine.Engine
	worldID      string
	width        int
	height       int
	clickFlash   bool
	coinFloat    bool
	lastCoinGain float64
	anim         background.BackgroundAnimation
	borderStyle  lipgloss.Style
	flashStyle   lipgloss.Style
}

// NewClickTab creates a ClickTabModel for the given world.
func NewClickTab(
	eng *engine.Engine,
	worldID string,
	t theme.Theme,
	animReg *background.AnimationRegistry,
	animKey string,
	width, height int,
) ClickTabModel {
	var anim background.BackgroundAnimation
	if animReg != nil && animKey != "" {
		anim = animReg.New(animKey)
	}
	return ClickTabModel{
		eng:     eng,
		worldID: worldID,
		width:   width,
		height:  height,
		anim:    anim,
		borderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.AccentColor())).
			Padding(1, 3),
		flashStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.CoinColor())).
			Padding(1, 3),
	}
}

func (m ClickTabModel) Init() tea.Cmd {
	if m.anim != nil {
		return m.anim.Init()
	}
	return nil
}

func (m ClickTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == " " {
			gained := m.eng.HandleClick(m.worldID, 1.0)
			m.lastCoinGain = gained
			m.clickFlash = true
			m.coinFloat = true
			return m, tea.Batch(
				tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return clickFlashEndMsg{} }),
				tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return coinFloatEndMsg{} }),
			)
		}
	case clickFlashEndMsg:
		m.clickFlash = false
	case coinFloatEndMsg:
		m.coinFloat = false
	}
	if m.anim != nil {
		newAnim, cmd := m.anim.Update(msg)
		m.anim = newAnim
		return m, cmd
	}
	return m, nil
}

func (m ClickTabModel) View() string {
	ws := m.eng.State.Worlds[m.worldID]

	var sb strings.Builder

	// Top animation strip.
	if m.anim != nil {
		sb.WriteString(m.anim.View(m.width, 2))
		sb.WriteByte('\n')
	}

	// Click zone.
	clickContent := "  PRESS SPACEBAR  \n     to mine TC   "
	borderSt := m.borderStyle
	if m.clickFlash {
		borderSt = m.flashStyle
	}
	sb.WriteString(borderSt.Render(clickContent))
	sb.WriteByte('\n')

	// Coin float.
	if m.coinFloat && m.lastCoinGain > 0 {
		sb.WriteString(fmt.Sprintf("  +%.1f TC\n", m.lastCoinGain))
	} else {
		sb.WriteString("\n")
	}

	// Stats.
	clickPower := m.eng.ClickPower(m.worldID, 1.0)
	cps := 0.0
	if ws != nil {
		cps = ws.CPS
	}
	sb.WriteString(fmt.Sprintf("  Click Power: %.2f TC/click\n", clickPower))
	sb.WriteString(fmt.Sprintf("  CPS: %.2f\n", cps))

	// Bottom animation strip.
	if m.anim != nil {
		sb.WriteString(m.anim.View(m.width, 2))
	}

	return sb.String()
}
