package save

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/player"
)

func TestSavePath(t *testing.T) {
	path := SavePath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "clicker")
	assert.Contains(t, path, "save.json")
}

func TestDefaultSaveFile(t *testing.T) {
	sf := DefaultSaveFile()
	assert.Equal(t, CurrentVersion, sf.Version)
	assert.False(t, sf.SavedAt.IsZero())
	assert.Equal(t, "overview", sf.LastScreen)
	assert.NotNil(t, sf.Worlds)
	assert.NotNil(t, sf.Achievements)
	assert.True(t, sf.Settings.AnimationsEnabled)
	assert.Equal(t, "space", sf.Settings.ActiveTheme)
}

func TestRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")

	gs := gamestate.NewGameState()
	gs.Player = player.NewPlayer()
	gs.Player.XP = 150
	gs.Player.Level = 2
	gs.Player.GeneralCoins = 42.5
	gs.LastScreen = "world"
	gs.LastWorldID = "terra"

	earned := map[string]bool{"first_click": true}
	settings := Settings{AnimationsEnabled: false, ActiveTheme: "space"}

	err := Save(gs, earned, settings, path)
	require.NoError(t, err)

	sf, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, CurrentVersion, sf.Version)
	assert.Equal(t, 150, sf.Player.XP)
	assert.Equal(t, 2, sf.Player.Level)
	assert.InDelta(t, 42.5, sf.Player.GeneralCoins, 0.001)
	assert.Equal(t, "world", sf.LastScreen)
	assert.Equal(t, "terra", sf.LastWorldID)
	assert.True(t, sf.Achievements["first_click"])
	assert.False(t, sf.Settings.AnimationsEnabled)
}

func TestLoad_MissingFile(t *testing.T) {
	sf, err := Load("/nonexistent/path/save.json")
	// Missing file should return DefaultSaveFile with no error.
	assert.NoError(t, err)
	assert.Equal(t, CurrentVersion, sf.Version)
}

func TestLoad_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	require.NoError(t, os.WriteFile(path, []byte("not valid json {{{"), 0o600))

	sf, err := Load(path)
	// Corrupt file should log a warning and return DefaultSaveFile with no error.
	assert.NoError(t, err)
	assert.Equal(t, CurrentVersion, sf.Version)
}

func TestMigrate_AlreadyCurrent(t *testing.T) {
	sf := DefaultSaveFile()
	sf.Version = CurrentVersion
	sf.SavedAt = time.Now()
	err := Migrate(&sf)
	require.NoError(t, err)
	assert.Equal(t, CurrentVersion, sf.Version)
}
