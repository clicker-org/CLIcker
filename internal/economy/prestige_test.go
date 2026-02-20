package economy

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrestigeMultiplierGain(t *testing.T) {
	tests := []struct {
		n        int
		wantGain float64
	}{
		{0, 1.5},           // first prestige: +50%
		{1, 1.0 + 0.5/math.Sqrt(2)}, // second: ≈1.354
		{3, 1.25},          // fourth: 1 + 0.5/sqrt(4) = 1.25
	}
	for _, tc := range tests {
		got := PrestigeMultiplierGain(tc.n)
		assert.InDelta(t, tc.wantGain, got, 0.001, "n=%d", tc.n)
	}
}

func TestCalculatePrestigeReward(t *testing.T) {
	tests := []struct {
		name              string
		totalCoins        float64
		prestigeCount     int
		currentMultiplier float64
		wantGC            float64
		wantXP            int
		wantMultGT        float64 // new multiplier must be > this
	}{
		{
			name:              "first prestige at 1M coins",
			totalCoins:        1_000_000,
			prestigeCount:     0,
			currentMultiplier: 1.0,
			wantGC:            100.0,
			wantXP:            500,
			wantMultGT:        1.0,
		},
		{
			name:              "second prestige at 10M coins",
			totalCoins:        10_000_000,
			prestigeCount:     1,
			currentMultiplier: 1.5,
			wantGC:            math.Sqrt(10_000_000) * 0.1,
			wantXP:            1000,
			wantMultGT:        1.5,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := CalculatePrestigeReward(tc.totalCoins, tc.prestigeCount, tc.currentMultiplier)
			assert.InDelta(t, tc.wantGC, r.GeneralCoinsEarned, 0.01)
			assert.Equal(t, tc.wantXP, r.XPGrant)
			assert.Greater(t, r.PrestigeMultiplier, tc.wantMultGT)
		})
	}
}

func TestCalculateExchangeBoost(t *testing.T) {
	t.Run("standard boost", func(t *testing.T) {
		r := CalculateExchangeBoost(1000, 0.001)
		assert.InDelta(t, 200.0, r.WorldCoinsCost, 0.001)  // 20% of 1000
		assert.InDelta(t, 0.2, r.GeneralCoinsEarned, 0.001) // 200 * 0.001
		assert.InDelta(t, 0.00101, r.NewExchangeRate, 0.000001) // 0.001 * 1.01
	})

	t.Run("zero balance", func(t *testing.T) {
		r := CalculateExchangeBoost(0, 0.001)
		assert.Equal(t, 0.0, r.WorldCoinsCost)
		assert.Equal(t, 0.0, r.GeneralCoinsEarned)
	})
}
