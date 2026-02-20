package economy

// ExchangeBoostResult holds the result of an exchange boost calculation.
type ExchangeBoostResult struct {
	GeneralCoinsEarned float64
	WorldCoinsCost     float64
	NewExchangeRate    float64
}

// ExchangeBoostCostPercent is the fraction of current balance sacrificed per boost.
const ExchangeBoostCostPercent = 0.20

// ExchangeBoostRateGain is the multiplicative factor applied to the exchange rate per boost.
const ExchangeBoostRateGain = 1.01

// CalculateExchangeBoost computes the result of an exchange boost.
// currentBalance: current world coin balance.
// exchangeRate: current world coin → GC rate.
func CalculateExchangeBoost(currentBalance, exchangeRate float64) ExchangeBoostResult {
	cost := currentBalance * ExchangeBoostCostPercent
	gc := cost * exchangeRate
	newRate := exchangeRate * ExchangeBoostRateGain
	return ExchangeBoostResult{
		GeneralCoinsEarned: gc,
		WorldCoinsCost:     cost,
		NewExchangeRate:    newRate,
	}
}
