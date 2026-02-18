package economy

// GeneralShopItemType enumerates the kinds of items in the general coin shop.
type GeneralShopItemType string

const (
	ItemTypeGlobalCPSMultiplier   GeneralShopItemType = "global_cps_multiplier"
	ItemTypeGlobalClickMultiplier GeneralShopItemType = "global_click_multiplier"
	ItemTypeGlobalXPMultiplier    GeneralShopItemType = "global_xp_multiplier"
	ItemTypePerWorldCPSMultiplier GeneralShopItemType = "per_world_cps_multiplier"
	ItemTypeOfflineCapUpgrade     GeneralShopItemType = "offline_cap_upgrade"
	ItemTypeCosmetic              GeneralShopItemType = "cosmetic"
)

// GeneralShopItem represents a purchasable item in the general coin shop.
type GeneralShopItem struct {
	ID          string
	Name        string
	Description string
	Type        GeneralShopItemType
	Cost        float64
	// TargetWorldID is set for per-world items; empty for global items.
	TargetWorldID string
	// Value is the numeric effect (e.g. 0.05 for +5%).
	Value float64
}
