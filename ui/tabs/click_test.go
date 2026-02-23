package tabs

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
	_ "github.com/clicker-org/clicker/internal/world/worlds"
	"github.com/clicker-org/clicker/ui/theme/themes"
	"github.com/stretchr/testify/assert"
)

func TestClickTab_IgnoresSpaceAutoRepeat(t *testing.T) {
	tab := newTestClickTab(t)
	base := time.Unix(1_700_000_000, 0)
	now := base
	tab.now = func() time.Time { return now }

	tab = updateClickTab(t, tab, spaceKeyMsg())
	for i := 1; i <= 10; i++ {
		now = base.Add(time.Duration(i*30) * time.Millisecond)
		tab = updateClickTab(t, tab, spaceKeyMsg())
	}

	assert.Equal(t, int64(1), tab.eng.State.Player.TotalClicks)
	assert.Equal(t, int64(1), tab.eng.State.Worlds["terra"].TotalClicks)
}

func TestClickTab_AllowsNextClickAfterPause(t *testing.T) {
	tab := newTestClickTab(t)
	base := time.Unix(1_700_000_000, 0)
	now := base
	tab.now = func() time.Time { return now }

	tab = updateClickTab(t, tab, spaceKeyMsg())
	now = now.Add(40 * time.Millisecond)
	tab = updateClickTab(t, tab, spaceKeyMsg())
	now = now.Add(300 * time.Millisecond)
	tab = updateClickTab(t, tab, spaceKeyMsg())

	assert.Equal(t, int64(2), tab.eng.State.Player.TotalClicks)
	assert.Equal(t, int64(2), tab.eng.State.Worlds["terra"].TotalClicks)
}

func newTestClickTab(t *testing.T) ClickTabModel {
	t.Helper()
	gs := gamestate.NewGameState()
	for _, w := range world.DefaultRegistry.List() {
		gs.Worlds[w.ID()] = world.NewWorldState(w.ID(), w.BaseExchangeRate())
	}
	eng := engine.New(gs, world.DefaultRegistry, achievement.NewAchievementRegistry())
	return NewClickTab(eng, "terra", themes.SpaceTheme{}, nil, "", 100, 30)
}

func updateClickTab(t *testing.T, m ClickTabModel, msg tea.Msg) ClickTabModel {
	t.Helper()
	updated, _ := m.Update(msg)
	tab, ok := updated.(ClickTabModel)
	if !ok {
		t.Fatalf("expected ClickTabModel, got %T", updated)
	}
	return tab
}

func spaceKeyMsg() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
}
