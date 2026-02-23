package save

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/clicker-org/clicker/internal/gamestate"
	"github.com/clicker-org/clicker/internal/world"
)

// configDir returns the OS-appropriate config directory for clicker.
// Respects XDG_CONFIG_HOME on Linux; falls back to ~/.config/clicker.
func configDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "."
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "clicker")
}

// SavePath returns the OS-appropriate path for the save file.
func SavePath() string {
	return filepath.Join(configDir(), "save.json")
}

// LogPath returns the OS-appropriate path for the log file.
func LogPath() string {
	return filepath.Join(configDir(), "clicker.log")
}

// Save writes the current game state to path as a signed envelope.
// The save file contains a base64-encoded JSON payload and its HMAC-SHA256
// signature. Load will reject files whose signature does not match.
func Save(gs gamestate.GameState, earned map[string]bool, settings Settings, path string) error {
	sf := SaveFileFromGameState(gs, earned, settings)
	payload, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("save: marshal: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(payload)
	envelope := signedEnvelope{
		Data: encoded,
		Sig:  sign([]byte(encoded)),
	}
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("save: marshal envelope: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("save: mkdir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("save: write: %w", err)
	}
	return nil
}

// Load reads a SaveFile from path. If the file does not exist, returns
// DefaultSaveFile with no error. If the file is corrupt or its HMAC signature
// does not match, logs a warning and returns DefaultSaveFile. Applies
// migrations after a successful load.
func Load(path string) (SaveFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultSaveFile(), nil
	}
	if err != nil {
		return DefaultSaveFile(), fmt.Errorf("save: read %q: %w", path, err)
	}

	var envelope signedEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.Data == "" {
		log.Printf("save: unrecognised save format at %q, starting fresh", path)
		return DefaultSaveFile(), nil
	}

	if !verify([]byte(envelope.Data), envelope.Sig) {
		log.Printf("save: signature mismatch at %q — file may have been tampered with, starting fresh", path)
		return DefaultSaveFile(), nil
	}

	payload, err := base64.StdEncoding.DecodeString(envelope.Data)
	if err != nil {
		log.Printf("save: base64 decode failed at %q, starting fresh: %v", path, err)
		return DefaultSaveFile(), nil
	}

	var sf SaveFile
	if err := json.Unmarshal(payload, &sf); err != nil {
		log.Printf("save: corrupt save payload at %q, starting fresh: %v", path, err)
		return DefaultSaveFile(), nil
	}

	if err := Migrate(&sf); err != nil {
		return DefaultSaveFile(), err
	}
	return sf, nil
}

// GameStateFromSave reconstructs a GameState from a SaveFile.
//
// worldReg is used to ensure all registered worlds have a WorldState entry,
// even when missing from the save (e.g. after adding a new world). Missing
// worlds are initialized with their configured base exchange rate.
func GameStateFromSave(sf SaveFile, worldReg *world.WorldRegistry) gamestate.GameState {
	gs := gamestate.NewGameState()
	gs.Player = sf.Player
	if gs.Player.WorldTotalCoinsEarned == nil {
		gs.Player.WorldTotalCoinsEarned = make(map[string]float64)
	}
	gs.LastScreen = sf.LastScreen
	gs.LastWorldID = sf.LastWorldID
	gs.ActiveWorldID = sf.LastWorldID

	// Reconstruct worlds — use saved data where available, otherwise fresh state.
	for _, id := range worldReg.IDs() {
		if data, ok := sf.Worlds[id]; ok {
			ws := &world.WorldState{
				WorldID:                data.WorldID,
				Coins:                  data.Coins,
				TotalCoinsEarned:       data.TotalCoinsEarned,
				CPS:                    data.CPS,
				BuyOnCounts:            data.BuyOnCounts,
				PurchasedUpgrades:      data.PurchasedUpgrades,
				PrestigeCount:          data.PrestigeCount,
				PrestigeMultiplier:     data.PrestigeMultiplier,
				ExchangeRate:           data.ExchangeRate,
				OfflineCapUpgradeLevel: data.OfflineCapUpgradeLevel,
				CompletionPercent:      data.CompletionPercent,
				TotalClicks:            data.TotalClicks,
			}
			if ws.BuyOnCounts == nil {
				ws.BuyOnCounts = make(map[string]int)
			}
			if ws.PurchasedUpgrades == nil {
				ws.PurchasedUpgrades = make(map[string]bool)
			}
			gs.Worlds[id] = ws
		} else {
			baseRate := 0.0
			if w, ok := worldReg.Get(id); ok {
				baseRate = w.BaseExchangeRate()
			}
			gs.Worlds[id] = world.NewWorldState(id, baseRate)
		}
	}

	return gs
}

// SaveFileFromGameState creates a SaveFile snapshot from the current game state.
func SaveFileFromGameState(gs gamestate.GameState, earned map[string]bool, settings Settings) SaveFile {
	sf := DefaultSaveFile()
	sf.Player = gs.Player
	sf.LastScreen = gs.LastScreen
	sf.LastWorldID = gs.LastWorldID

	achCopy := make(map[string]bool, len(earned))
	for k, v := range earned {
		achCopy[k] = v
	}
	sf.Achievements = achCopy
	sf.Settings = settings

	for id, ws := range gs.Worlds {
		buyOnCopy := make(map[string]int, len(ws.BuyOnCounts))
		for k, v := range ws.BuyOnCounts {
			buyOnCopy[k] = v
		}
		upgCopy := make(map[string]bool, len(ws.PurchasedUpgrades))
		for k, v := range ws.PurchasedUpgrades {
			upgCopy[k] = v
		}
		sf.Worlds[id] = WorldSaveData{
			WorldID:                ws.WorldID,
			Coins:                  ws.Coins,
			TotalCoinsEarned:       ws.TotalCoinsEarned,
			CPS:                    ws.CPS,
			BuyOnCounts:            buyOnCopy,
			PurchasedUpgrades:      upgCopy,
			PrestigeCount:          ws.PrestigeCount,
			PrestigeMultiplier:     ws.PrestigeMultiplier,
			ExchangeRate:           ws.ExchangeRate,
			OfflineCapUpgradeLevel: ws.OfflineCapUpgradeLevel,
			CompletionPercent:      ws.CompletionPercent,
			TotalClicks:            ws.TotalClicks,
		}
	}

	return sf
}
