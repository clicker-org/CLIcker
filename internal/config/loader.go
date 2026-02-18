package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

var validIDRe = regexp.MustCompile(`^[a-z_]+$`)

// BuyOnConfig holds configuration for a single buy-on (passive income generator).
type BuyOnConfig struct {
	ID               string  `toml:"id"`
	Name             string  `toml:"name"`
	Description      string  `toml:"description"`
	BaseCost         float64 `toml:"base_cost"`
	CostScaling      float64 `toml:"cost_scaling"`
	BaseCPS          float64 `toml:"base_cps"`
	LevelRequirement int     `toml:"level_requirement"`
}

// UpgradeConfig holds configuration for a one-time buy-on upgrade.
type UpgradeConfig struct {
	ID               string  `toml:"id"`
	Name             string  `toml:"name"`
	Description      string  `toml:"description"`
	TargetBuyOnID    string  `toml:"target_buy_on"`
	Multiplier       float64 `toml:"multiplier"`
	Cost             float64 `toml:"cost"`
	LevelRequirement int     `toml:"level_requirement"`
}

// PrestigeThresholdConfig defines when a world prestige becomes available.
type PrestigeThresholdConfig struct {
	Type  string  `toml:"type"`
	Value float64 `toml:"value"`
}

// CompletionMilestone is a single milestone that contributes to world completion %.
type CompletionMilestone struct {
	ID          string  `toml:"id"`
	Description string  `toml:"description"`
	Type        string  `toml:"type"`
	Value       float64 `toml:"value"`
	Weight      float64 `toml:"weight"`
}

// WorldConfig is the full configuration for a single world loaded from TOML.
type WorldConfig struct {
	ID                   string                  `toml:"id"`
	Name                 string                  `toml:"name"`
	CoinName             string                  `toml:"coin_name"`
	CoinSymbol           string                  `toml:"coin_symbol"`
	AccentColor          string                  `toml:"accent_color"`
	AmbientAnimation     string                  `toml:"ambient_animation"`
	BaseExchangeRate     float64                 `toml:"base_exchange_rate"`
	OfflinePercentage    float64                 `toml:"offline_percentage"`
	OfflineCapHours      float64                 `toml:"offline_cap_hours"`
	BuyOns               []BuyOnConfig           `toml:"buy_ons"`
	BuyOnUpgrades        []UpgradeConfig         `toml:"buy_on_upgrades"`
	PrestigeThreshold    PrestigeThresholdConfig `toml:"prestige_threshold"`
	CompletionMilestones []CompletionMilestone   `toml:"completion_milestones"`
}

// LoadWorld loads a single WorldConfig from the given TOML file path.
func LoadWorld(path string) (WorldConfig, error) {
	var cfg WorldConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return WorldConfig{}, fmt.Errorf("config: loading %q: %w", path, err)
	}
	return cfg, nil
}

// LoadWorldDir loads all *.toml files from dir and returns them as a slice.
func LoadWorldDir(dir string) ([]WorldConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("config: reading dir %q: %w", dir, err)
	}
	var cfgs []WorldConfig
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		cfg, err := LoadWorld(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, cfg)
	}
	return cfgs, nil
}

// Validate checks a WorldConfig for consistency errors and returns a list of
// human-readable error strings. An empty slice means the config is valid.
func Validate(cfg WorldConfig) []string {
	var errs []string

	if !validIDRe.MatchString(cfg.ID) {
		errs = append(errs, fmt.Sprintf("world ID %q must match [a-z_]+", cfg.ID))
	}

	buyOnIDs := map[string]bool{}
	for _, b := range cfg.BuyOns {
		if !validIDRe.MatchString(b.ID) {
			errs = append(errs, fmt.Sprintf("buy_on ID %q must match [a-z_]+", b.ID))
		}
		if b.CostScaling < 1.0 {
			errs = append(errs, fmt.Sprintf("buy_on %q cost_scaling %.2f must be >= 1.0", b.ID, b.CostScaling))
		}
		if b.BaseCPS <= 0 {
			errs = append(errs, fmt.Sprintf("buy_on %q base_cps %.4f must be > 0", b.ID, b.BaseCPS))
		}
		buyOnIDs[b.ID] = true
	}

	for _, u := range cfg.BuyOnUpgrades {
		if !buyOnIDs[u.TargetBuyOnID] {
			errs = append(errs, fmt.Sprintf("upgrade %q references unknown buy_on %q", u.ID, u.TargetBuyOnID))
		}
	}

	if len(cfg.CompletionMilestones) > 0 {
		total := 0.0
		for _, m := range cfg.CompletionMilestones {
			total += m.Weight
		}
		if math.Abs(total-1.0) > 0.001 {
			errs = append(errs, fmt.Sprintf("completion_milestones weights sum to %.4f, must be 1.0 ±0.001", total))
		}
	}

	return errs
}
