package player

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXPForLevel(t *testing.T) {
	assert.Equal(t, 0, XPForLevel(1))
	assert.Equal(t, 100, XPForLevel(2))
	assert.Equal(t, 250, XPForLevel(3))   // 100 + 150
	assert.Equal(t, 1318, XPForLevel(6))  // 100+150+225+337+506
}

func TestXPNeededForNextLevel(t *testing.T) {
	assert.Equal(t, 100, XPNeededForNextLevel(1))
	assert.Equal(t, 150, XPNeededForNextLevel(2))
}

func TestAddXP_NoLevelUp(t *testing.T) {
	p := NewPlayer()
	leveled := AddXP(&p, 50)
	assert.False(t, leveled)
	assert.Equal(t, 1, p.Level)
	assert.Equal(t, 50, p.XP)
}

func TestAddXP_ExactLevelUp(t *testing.T) {
	p := NewPlayer()
	leveled := AddXP(&p, 100)
	assert.True(t, leveled)
	assert.Equal(t, 2, p.Level)
}

func TestAddXP_MultiLevelUp(t *testing.T) {
	p := NewPlayer()
	leveled := AddXP(&p, 10000)
	assert.True(t, leveled)
	assert.True(t, p.Level > 3)
}

func TestLevelGateCheck(t *testing.T) {
	p := NewPlayer() // level 1
	assert.True(t, LevelGateCheck(p, 0))
	assert.True(t, LevelGateCheck(p, 1))
	assert.False(t, LevelGateCheck(p, 2))
}
