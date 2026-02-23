package achievement

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDefaults_RegistersExpectedSet(t *testing.T) {
	reg := NewAchievementRegistry()
	RegisterDefaults(reg)

	all := reg.GetAll()
	require.Len(t, all, 10)

	seen := make(map[string]bool, len(all))
	for _, a := range all {
		assert.NotEmpty(t, a.ID)
		assert.NotEmpty(t, a.Name)
		assert.NotNil(t, a.Condition)
		assert.False(t, seen[a.ID], "duplicate achievement id %q", a.ID)
		seen[a.ID] = true
	}
}
