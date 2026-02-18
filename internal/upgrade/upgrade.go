package upgrade

import (
	"math"

	"github.com/clicker-org/clicker/internal/config"
)

// BuyOn is the interface all passive income generators must implement.
type BuyOn interface {
	ID() string
	Name() string
	Description() string
	BaseCost() float64
	CostScaling() float64
	BaseCPS() float64
	LevelRequirement() int
}

// BuyOnState tracks runtime ownership state for a single buy-on type.
type BuyOnState struct {
	BuyOnID string
	Count   int
}

// CostForNext returns the coin cost to purchase the next unit of a buy-on
// given the current ownership count.
// Formula: BaseCost * CostScaling^count
func CostForNext(b BuyOn, count int) float64 {
	return b.BaseCost() * math.Pow(b.CostScaling(), float64(count))
}

// ConfigBuyOn is a BuyOn backed by a config.BuyOnConfig.
type ConfigBuyOn struct {
	cfg config.BuyOnConfig
}

// NewConfigBuyOn wraps a BuyOnConfig as a BuyOn interface.
func NewConfigBuyOn(cfg config.BuyOnConfig) *ConfigBuyOn {
	return &ConfigBuyOn{cfg: cfg}
}

func (c *ConfigBuyOn) ID() string               { return c.cfg.ID }
func (c *ConfigBuyOn) Name() string              { return c.cfg.Name }
func (c *ConfigBuyOn) Description() string       { return c.cfg.Description }
func (c *ConfigBuyOn) BaseCost() float64         { return c.cfg.BaseCost }
func (c *ConfigBuyOn) CostScaling() float64      { return c.cfg.CostScaling }
func (c *ConfigBuyOn) BaseCPS() float64          { return c.cfg.BaseCPS }
func (c *ConfigBuyOn) LevelRequirement() int     { return c.cfg.LevelRequirement }
