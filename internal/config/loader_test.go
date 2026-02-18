package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWorld_ValidFile(t *testing.T) {
	path := findTerraToml(t)
	cfg, err := LoadWorld(path)
	require.NoError(t, err)
	assert.Equal(t, "terra", cfg.ID)
	assert.NotEmpty(t, cfg.BuyOns)
	assert.NotEmpty(t, cfg.CompletionMilestones)
}

func TestLoadWorld_MissingFile(t *testing.T) {
	_, err := LoadWorld("/nonexistent/path/world.toml")
	assert.Error(t, err)
}

func TestValidate_Valid(t *testing.T) {
	path := findTerraToml(t)
	cfg, err := LoadWorld(path)
	require.NoError(t, err)
	errs := Validate(cfg)
	assert.Empty(t, errs, "terra.toml should be valid: %v", errs)
}

func TestValidate_WeightsDontSumToOne(t *testing.T) {
	cfg := WorldConfig{
		ID: "test_world",
		BuyOns: []BuyOnConfig{
			{ID: "a", CostScaling: 1.15, BaseCPS: 0.1},
		},
		CompletionMilestones: []CompletionMilestone{
			{ID: "m1", Weight: 0.5},
			{ID: "m2", Weight: 0.3},
			// total = 0.8, not 1.0
		},
	}
	errs := Validate(cfg)
	assert.NotEmpty(t, errs)
}

func TestValidate_InvalidTargetBuyOn(t *testing.T) {
	cfg := WorldConfig{
		ID: "test_world",
		BuyOns: []BuyOnConfig{
			{ID: "miner", CostScaling: 1.15, BaseCPS: 0.1},
		},
		BuyOnUpgrades: []UpgradeConfig{
			{ID: "turbo", TargetBuyOnID: "nonexistent"},
		},
		CompletionMilestones: []CompletionMilestone{
			{ID: "m1", Weight: 1.0},
		},
	}
	errs := Validate(cfg)
	assert.NotEmpty(t, errs)
}

func TestValidate_CostScalingBelowOne(t *testing.T) {
	cfg := WorldConfig{
		ID: "test_world",
		BuyOns: []BuyOnConfig{
			{ID: "bad_miner", CostScaling: 0.5, BaseCPS: 0.1},
		},
		CompletionMilestones: []CompletionMilestone{
			{ID: "m1", Weight: 1.0},
		},
	}
	errs := Validate(cfg)
	assert.NotEmpty(t, errs)
}

// findTerraToml walks up from the test directory to find configs/worlds/terra.toml.
func findTerraToml(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		candidate := filepath.Join(dir, "configs", "worlds", "terra.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find configs/worlds/terra.toml")
		}
		dir = parent
	}
}
