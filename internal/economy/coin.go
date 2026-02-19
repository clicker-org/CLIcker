package economy

import "fmt"

type siTier struct {
	threshold float64
	suffix    string
}

var siTiers = []siTier{
	{1e15, "Q"},
	{1e12, "T"},
	{1e9, "B"},
	{1e6, "M"},
	{1e3, "K"},
}

// FormatCoinsBare formats a coin amount with SI suffixes, no symbol.
func FormatCoinsBare(amount float64) string {
	neg := ""
	if amount < 0 {
		neg = "-"
		amount = -amount
	}
	for _, t := range siTiers {
		if amount >= t.threshold {
			return fmt.Sprintf("%s%.2f%s", neg, amount/t.threshold, t.suffix)
		}
	}
	return fmt.Sprintf("%s%.0f", neg, amount)
}

// FormatCoins formats a coin amount with SI suffixes and a symbol prefix.
func FormatCoins(amount float64, symbol string) string {
	return symbol + ": " + FormatCoinsBare(amount)
}

// FormatCPS formats a CPS value with SI suffixes and always two decimal places.
func FormatCPS(amount float64) string {
	for _, t := range siTiers {
		if amount >= t.threshold {
			return fmt.Sprintf("%.2f%s", amount/t.threshold, t.suffix)
		}
	}
	return fmt.Sprintf("%.2f", amount)
}
