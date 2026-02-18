package economy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCoinsBare(t *testing.T) {
	tests := []struct {
		amount   float64
		expected string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.00K"},
		{1500, "1.50K"},
		{1_230_000, "1.23M"},
		{1_000_000_000, "1.00B"},
		{1_000_000_000_000, "1.00T"},
		{1_000_000_000_000_000, "1.00Q"},
		{-500, "-500"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, FormatCoinsBare(tc.amount))
		})
	}
}
