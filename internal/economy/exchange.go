package economy

// ExchangeBoostResult holds the result of an exchange boost calculation.
type ExchangeBoostResult struct {
	GeneralCoinsEarned float64
	WorldCoinsCost     float64
	NewExchangeRate    float64
}

// CalculateExchangeBoost computes the result of an exchange boost.
// costPercent: fraction of current balance to sacrifice (e.g. 0.20).
// exchangeRate: current world coin -> GC rate.
// rateImprovement: additive improvement to exchange rate per boost.
// Stub: formula TBD in Phase 2.
func CalculateExchangeBoost(
	currentBalance, costPercent, exchangeRate, rateImprovement float64,
) ExchangeBoostResult {
	return ExchangeBoostResult{}
}
