package worlds

import (
	"bytes"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/clicker-org/clicker/configs"
	"github.com/clicker-org/clicker/internal/config"
	"github.com/clicker-org/clicker/internal/world"
)

type terraWorld struct {
	cfg config.WorldConfig
}

func init() {
	var cfg config.WorldConfig
	if _, err := toml.NewDecoder(bytes.NewReader(configs.TerraToml)).Decode(&cfg); err != nil {
		log.Panicf("worlds: failed to load terra config: %v", err)
	}
	world.DefaultRegistry.Register(&terraWorld{cfg: cfg})
}

func (t *terraWorld) ID() string                  { return t.cfg.ID }
func (t *terraWorld) Name() string                { return t.cfg.Name }
func (t *terraWorld) CoinName() string            { return t.cfg.CoinName }
func (t *terraWorld) CoinSymbol() string          { return t.cfg.CoinSymbol }
func (t *terraWorld) AccentColor() string         { return t.cfg.AccentColor }
func (t *terraWorld) AmbientAnimation() string    { return t.cfg.AmbientAnimation }
func (t *terraWorld) BaseExchangeRate() float64   { return t.cfg.BaseExchangeRate }
func (t *terraWorld) OfflinePercentage() float64  { return t.cfg.OfflinePercentage }
func (t *terraWorld) OfflineCapHours() float64    { return t.cfg.OfflineCapHours }
func (t *terraWorld) Config() config.WorldConfig  { return t.cfg }
