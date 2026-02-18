package offline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateOfflineIncome(t *testing.T) {
	tests := []struct {
		name        string
		cps         float64
		offlinePct  float64
		elapsedSecs float64
		capHours    float64
		expected    float64
	}{
		{"zero_cps", 0, 0.10, 3600, 8, 0},
		{"zero_elapsed", 10, 0.10, 0, 8, 0},
		{"under_cap", 10, 0.10, 1800, 8, 1800},    // 10*0.1*1800=1800, cap=28800
		{"at_cap", 10, 0.10, 28800, 8, 28800},     // exactly at cap
		{"over_cap", 10, 0.10, 100000, 8, 28800},  // capped
		{"low_pct", 10, 0.05, 3600, 8, 1800},      // 10*0.05*3600=1800
		{"negative_elapsed", 10, 0.10, -100, 8, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateOfflineIncome(tc.cps, tc.offlinePct, tc.elapsedSecs, tc.capHours)
			assert.InDelta(t, tc.expected, got, 0.001)
		})
	}
}

func TestCalculateOverviewOfflineIncome(t *testing.T) {
	tests := []struct {
		name           string
		baseRatePerSec float64
		elapsedSecs    float64
		gcCap          float64
		expected       float64
	}{
		{"under_cap", 0.01, 1000, 100, 10},   // 0.01*1000=10
		{"over_cap", 0.01, 20000, 100, 100},  // capped at 100
		{"zero_rate", 0, 3600, 100, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateOverviewOfflineIncome(tc.baseRatePerSec, tc.elapsedSecs, tc.gcCap)
			assert.InDelta(t, tc.expected, got, 0.001)
		})
	}
}
