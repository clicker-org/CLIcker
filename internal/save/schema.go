package save

import (
	"fmt"
	"time"

	"github.com/clicker-org/clicker/internal/player"
)

// CurrentVersion is the current save file schema version.
const CurrentVersion = 1

// signedEnvelope is the on-disk format for save files.
// Data holds the base64-encoded JSON of a SaveFile; Sig is its HMAC-SHA256.
type signedEnvelope struct {
	Data string `json:"data"`
	Sig  string `json:"sig"`
}

// WorldSaveData holds all persisted data for a single world.
type WorldSaveData struct {
	WorldID                string             `json:"world_id"`
	Coins                  float64            `json:"coins"`
	TotalCoinsEarned       float64            `json:"total_coins_earned"`
	CPS                    float64            `json:"cps"`
	BuyOnCounts            map[string]int     `json:"buy_on_counts"`
	PurchasedUpgrades      map[string]bool    `json:"purchased_upgrades"`
	PrestigeCount          int                `json:"prestige_count"`
	PrestigeMultiplier     float64            `json:"prestige_multiplier"`
	ExchangeRate           float64            `json:"exchange_rate"`
	OfflineCapUpgradeLevel int                `json:"offline_cap_upgrade_level"`
	CompletionPercent      float64            `json:"completion_percent"`
	TotalClicks            int64              `json:"total_clicks"`
}

// Settings holds user-configurable preferences.
type Settings struct {
	AnimationsEnabled bool   `json:"animations_enabled"`
	ActiveTheme       string `json:"active_theme"`
}

// SaveFile is the top-level save file structure.
type SaveFile struct {
	Version      int                       `json:"version"`
	SavedAt      time.Time                 `json:"saved_at"`
	LastScreen   string                    `json:"last_screen"`
	LastWorldID  string                    `json:"last_world_id"`
	Player       player.Player             `json:"player"`
	Worlds       map[string]WorldSaveData  `json:"worlds"`
	Achievements map[string]bool           `json:"achievements"`
	Settings     Settings                  `json:"settings"`
}

// DefaultSaveFile returns a fresh SaveFile with sensible defaults.
func DefaultSaveFile() SaveFile {
	return SaveFile{
		Version:      CurrentVersion,
		SavedAt:      time.Now(),
		LastScreen:   "overview",
		LastWorldID:  "",
		Player:       player.NewPlayer(),
		Worlds:       make(map[string]WorldSaveData),
		Achievements: make(map[string]bool),
		Settings: Settings{
			AnimationsEnabled: true,
			ActiveTheme:       "space",
		},
	}
}

// Migrate upgrades a SaveFile to the current schema version, applying any
// required migration functions in sequence.
func Migrate(sf *SaveFile) error {
	if sf.Version > CurrentVersion {
		return fmt.Errorf("save: file version %d is newer than current version %d", sf.Version, CurrentVersion)
	}
	// Each migration block below handles one version step.
	// Example for future migrations:
	// if sf.Version < 2 { migrateV1toV2(sf); sf.Version = 2 }
	sf.Version = CurrentVersion
	return nil
}
