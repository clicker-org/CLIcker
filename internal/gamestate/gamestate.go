// Package gamestate is a pure data package holding the top-level game state.
// It has no Bubble Tea imports and no game logic.
// Both internal/engine and internal/achievement import this package.
package gamestate

import (
	"github.com/clicker-org/clicker/internal/player"
	"github.com/clicker-org/clicker/internal/world"
)

// GameState is the top-level container for all mutable game state.
type GameState struct {
	Player        player.Player
	Worlds        map[string]*world.WorldState
	LastScreen    string
	LastWorldID   string
	ActiveWorldID string
}

// NewGameState returns a freshly initialized GameState with no worlds.
func NewGameState() GameState {
	return GameState{
		Player:        player.NewPlayer(),
		Worlds:        make(map[string]*world.WorldState),
		LastScreen:    "overview",
		LastWorldID:   "",
		ActiveWorldID: "",
	}
}
