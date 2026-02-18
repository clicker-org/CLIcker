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
	t            theme.Theme
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
		t:       t,
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

// Resize returns a copy of the model with updated dimensions.
func (m ClickTabModel) Resize(w, h int) ClickTabModel {
	m.width = w
	m.height = h
	return m
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
	bg := lipgloss.Color(m.t.Background())
	coinFg := lipgloss.Color(m.t.CoinColor())
	dimFg := lipgloss.Color(m.t.DimText())

	// Click box width: half the terminal, minimum 28 chars.
	boxWidth := m.width / 2
	if boxWidth < 28 {
		boxWidth = 28
	}

	borderSt := m.borderStyle.Width(boxWidth)
	if m.clickFlash {
		borderSt = m.flashStyle.Width(boxWidth)
	}

	// Click zone.
	clickBox := borderSt.Align(lipgloss.Center).Render("PRESS SPACEBAR\nto mine coins")

	// Coin float (briefly visible after each click).
	var floatLine string
	if m.coinFloat && m.lastCoinGain > 0 {
		floatLine = lipgloss.NewStyle().Foreground(coinFg).Render(
			fmt.Sprintf("+%.1f TC", m.lastCoinGain))
	}

	// Stats.
	clickPower := m.eng.ClickPower(m.worldID, 1.0)
	cps := 0.0
	if ws := m.eng.State.Worlds[m.worldID]; ws != nil {
		cps = ws.CPS
	}
	statsLine := lipgloss.NewStyle().Foreground(dimFg).Render(
		fmt.Sprintf("Click Power: %.2f TC/click    CPS: %.2f", clickPower, cps))

	// Center block: click box + coin float placeholder + blank line + stats.
	centerBlock := strings.Join([]string{clickBox, floatLine, "", statsLine}, "\n")

	// Lay out: 3-line animation strips pinned to top and bottom edges,
	// with the center block placed in the remaining space.
	const animH = 3
	if m.anim != nil && m.height > animH*2+4 {
		innerH := m.height - animH*2
		topAnim := lipgloss.NewStyle().Width(m.width).Background(bg).
			Render(m.anim.View(m.width, animH))
		bottomAnim := lipgloss.NewStyle().Width(m.width).Background(bg).
			Render(m.anim.View(m.width, animH))
		inner := lipgloss.Place(m.width, innerH,
			lipgloss.Center, lipgloss.Center,
			centerBlock,
			lipgloss.WithWhitespaceBackground(bg))
		return topAnim + "\n" + inner + "\n" + bottomAnim
	}

	// No animation (or too short): center content in available space.
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		centerBlock,
		lipgloss.WithWhitespaceBackground(bg))
}
