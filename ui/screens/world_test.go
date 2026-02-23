package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
	_ "github.com/clicker-org/clicker/internal/world/worlds"
	"github.com/clicker-org/clicker/ui/messages"
	"github.com/clicker-org/clicker/ui/theme/themes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorldHotkeys_ToggleModalOpenAndClose(t *testing.T) {
	m := newTestWorldModel(t)

	m, _ = m.Update(runeKeyMsg('s'))
	assert.Equal(t, ModalShop, m.activeModal)

	m, _ = m.Update(runeKeyMsg('s'))
	assert.Equal(t, ModalNone, m.activeModal)

	m, _ = m.Update(runeKeyMsg('p'))
	assert.Equal(t, ModalPrestige, m.activeModal)

	m, _ = m.Update(runeKeyMsg('p'))
	assert.Equal(t, ModalNone, m.activeModal)

	m, _ = m.Update(runeKeyMsg('a'))
	assert.Equal(t, ModalAchievements, m.activeModal)

	m, _ = m.Update(runeKeyMsg('a'))
	assert.Equal(t, ModalNone, m.activeModal)
}

func TestWorldHotkeys_SwitchBetweenModals(t *testing.T) {
	m := newTestWorldModel(t)

	m, _ = m.Update(runeKeyMsg('s'))
	assert.Equal(t, ModalShop, m.activeModal)

	m, _ = m.Update(runeKeyMsg('p'))
	assert.Equal(t, ModalPrestige, m.activeModal)

	m, _ = m.Update(runeKeyMsg('a'))
	assert.Equal(t, ModalAchievements, m.activeModal)
}

func TestWorldPrestige_EnterRequestsConfirm(t *testing.T) {
	m := newTestWorldModel(t)
	ws := m.eng.State.Worlds["terra"]
	ws.TotalCoinsEarned = 1_000_000_000

	m, _ = m.Update(runeKeyMsg('p'))
	require.Equal(t, ModalPrestige, m.activeModal)

	_, cmd := m.Update(messages.NavConfirmMsg{})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(messages.PrestigeConfirmRequestedMsg)
	assert.True(t, ok, "expected PrestigeConfirmRequestedMsg, got %T", msg)
}

func TestWorldShop_NumberHotkeySelectsIndexedItem(t *testing.T) {
	m := newTestWorldModel(t)
	ws := m.eng.State.Worlds["terra"]
	ws.Coins = 1_000_000_000
	ws.TotalCoinsEarned = 1_000_000_000
	m.eng.State.Player.Level = 100

	items := m.eng.UpgradeReg["terra"].ListBuyOns()
	require.GreaterOrEqual(t, len(items), 2, "terra needs at least 2 buy-ons for index-hotkey test")

	firstID := items[0].ID()
	secondID := items[1].ID()

	m, _ = m.Update(runeKeyMsg('s'))
	require.Equal(t, ModalShop, m.activeModal)

	m, _ = m.Update(runeKeyMsg('2'))
	m, _ = m.Update(messages.NavConfirmMsg{})

	assert.Equal(t, 0, ws.BuyOnCounts[firstID], "first item should remain unpurchased")
	assert.Equal(t, 1, ws.BuyOnCounts[secondID], "second item should be purchased via [2] then Enter")
}

func newTestWorldModel(t *testing.T) WorldModel {
	t.Helper()
	gs := gamestate.NewGameState()
	for _, w := range world.DefaultRegistry.List() {
		gs.Worlds[w.ID()] = world.NewWorldState(w.ID(), w.BaseExchangeRate())
	}
	eng := engine.New(gs, world.DefaultRegistry, achievement.NewAchievementRegistry())
	return NewWorldModel(themes.SpaceTheme{}, eng, &eng.State, "terra", nil, "", 120, 40)
}

func runeKeyMsg(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}
