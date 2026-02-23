package screens

import (
	"testing"

	"github.com/clicker-org/clicker/internal/achievement"
	"github.com/clicker-org/clicker/internal/engine"
	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
	"github.com/clicker-org/clicker/ui/messages"
	_ "github.com/clicker-org/clicker/internal/world/worlds"
	"github.com/clicker-org/clicker/ui/theme/themes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAchievementsView_ListsAllWithPassFailMarkers(t *testing.T) {
	reg := achievement.NewAchievementRegistry()
	reg.Register(achievement.Achievement{
		ID:          "first_click",
		Name:        "First Click",
		Description: "Click once.",
	})
	reg.Register(achievement.Achievement{
		ID:          "hidden_one",
		Name:        "Secret Finder",
		Description: "Find the hidden thing.",
		Hidden:      true,
	})

	gs := gamestate.NewGameState()
	for _, w := range world.DefaultRegistry.List() {
		gs.Worlds[w.ID()] = world.NewWorldState(w.ID(), w.BaseExchangeRate())
	}
	eng := engine.New(gs, world.DefaultRegistry, reg)
	eng.Earned["first_click"] = true

	m := NewAchievementsModel(themes.SpaceTheme{}, eng, 120, 40)
	view := m.View()

	assert.Contains(t, view, "[PASSED] First Click")
	assert.Contains(t, view, "[NOT PASSED] Secret Finder")
	assert.Contains(t, view, "Click once.")
	assert.Contains(t, view, "Find the hidden thing.")
}

func TestAchievementsNavigation_ArrowAndVimMessagesMoveCursor(t *testing.T) {
	reg := achievement.NewAchievementRegistry()
	reg.Register(achievement.Achievement{ID: "a1", Name: "A1", Description: "A1"})
	reg.Register(achievement.Achievement{ID: "a2", Name: "A2", Description: "A2"})
	reg.Register(achievement.Achievement{ID: "a3", Name: "A3", Description: "A3"})

	gs := gamestate.NewGameState()
	for _, w := range world.DefaultRegistry.List() {
		gs.Worlds[w.ID()] = world.NewWorldState(w.ID(), w.BaseExchangeRate())
	}
	eng := engine.New(gs, world.DefaultRegistry, reg)

	m := NewAchievementsModel(themes.SpaceTheme{}, eng, 120, 18)
	require.Equal(t, 0, m.cursor)

	m, _ = m.Update(messages.NavDownMsg{})
	assert.Equal(t, 1, m.cursor)

	m, _ = m.Update(messages.NavDownMsg{})
	assert.Equal(t, 2, m.cursor)

	m, _ = m.Update(messages.NavDownMsg{})
	assert.Equal(t, 2, m.cursor, "cursor should clamp at end")

	m, _ = m.Update(messages.NavUpMsg{})
	assert.Equal(t, 1, m.cursor)

	m, _ = m.Update(messages.NavUpMsg{})
	assert.Equal(t, 0, m.cursor)

	m, _ = m.Update(messages.NavUpMsg{})
	assert.Equal(t, 0, m.cursor, "cursor should clamp at start")
}
