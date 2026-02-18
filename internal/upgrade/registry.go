package upgrade

import (
	"sync"

	"github.com/clicker-org/clicker/internal/config"
)

// WorldUpgradeRegistry holds all buy-ons and upgrades for a single world.
type WorldUpgradeRegistry struct {
	mu       sync.RWMutex
	buyOns   []BuyOn
	buyOnMap map[string]BuyOn
	upgrades []config.UpgradeConfig
	upgradeMap map[string]config.UpgradeConfig
}

// NewWorldUpgradeRegistry creates an empty registry.
func NewWorldUpgradeRegistry() *WorldUpgradeRegistry {
	return &WorldUpgradeRegistry{
		buyOnMap:   make(map[string]BuyOn),
		upgradeMap: make(map[string]config.UpgradeConfig),
	}
}

// RegisterBuyOn adds a BuyOn to the registry. Panics if the ID is already registered.
func (r *WorldUpgradeRegistry) RegisterBuyOn(b BuyOn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.buyOnMap[b.ID()]; exists {
		panic("upgrade: duplicate buy-on ID: " + b.ID())
	}
	r.buyOns = append(r.buyOns, b)
	r.buyOnMap[b.ID()] = b
}

// RegisterUpgrade adds an upgrade config to the registry.
func (r *WorldUpgradeRegistry) RegisterUpgrade(u config.UpgradeConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.upgradeMap[u.ID]; exists {
		panic("upgrade: duplicate upgrade ID: " + u.ID)
	}
	r.upgrades = append(r.upgrades, u)
	r.upgradeMap[u.ID] = u
}

// GetBuyOn returns the BuyOn with the given ID and a found flag.
func (r *WorldUpgradeRegistry) GetBuyOn(id string) (BuyOn, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.buyOnMap[id]
	return b, ok
}

// GetUpgrade returns the UpgradeConfig with the given ID and a found flag.
func (r *WorldUpgradeRegistry) GetUpgrade(id string) (config.UpgradeConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.upgradeMap[id]
	return u, ok
}

// ListBuyOns returns all registered buy-ons in insertion order.
func (r *WorldUpgradeRegistry) ListBuyOns() []BuyOn {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]BuyOn, len(r.buyOns))
	copy(out, r.buyOns)
	return out
}

// ListUpgrades returns all registered upgrades in insertion order.
func (r *WorldUpgradeRegistry) ListUpgrades() []config.UpgradeConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]config.UpgradeConfig, len(r.upgrades))
	copy(out, r.upgrades)
	return out
}
