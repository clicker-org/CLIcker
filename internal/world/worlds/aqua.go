package worlds

import (
	"bytes"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/clicker-org/clicker/configs"
	"github.com/clicker-org/clicker/internal/config"
	"github.com/clicker-org/clicker/internal/world"
)

type aquaWorld struct {
	cfg config.WorldConfig
}

func init() {
	var cfg config.WorldConfig
	if _, err := toml.NewDecoder(bytes.NewReader(configs.AquaToml)).Decode(&cfg); err != nil {
		log.Panicf("worlds: failed to load aqua config: %v", err)
	}
	world.DefaultRegistry.Register(&aquaWorld{cfg: cfg})
}

func (a *aquaWorld) ID() string                 { return a.cfg.ID }
func (a *aquaWorld) Name() string               { return a.cfg.Name }
func (a *aquaWorld) CoinName() string           { return a.cfg.CoinName }
func (a *aquaWorld) CoinSymbol() string         { return a.cfg.CoinSymbol }
func (a *aquaWorld) AccentColor() string        { return a.cfg.AccentColor }
func (a *aquaWorld) AmbientAnimation() string   { return a.cfg.AmbientAnimation }
func (a *aquaWorld) BaseExchangeRate() float64  { return a.cfg.BaseExchangeRate }
func (a *aquaWorld) OfflinePercentage() float64 { return a.cfg.OfflinePercentage }
func (a *aquaWorld) OfflineCapHours() float64   { return a.cfg.OfflineCapHours }
func (a *aquaWorld) Config() config.WorldConfig { return a.cfg }
