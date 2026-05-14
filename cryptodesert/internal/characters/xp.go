package characters

// XPToNextLevel returns the XP required to reach the next level.
// Uses a quadratic curve: XP(n) = 100 * n^2
func XPToNextLevel(level int) int {
	return 100 * level * level
}

// TotalXPForLevel returns cumulative XP needed to reach a given level from 0
func TotalXPForLevel(level int) int {
	total := 0
	for i := 1; i < level; i++ {
		total += XPToNextLevel(i)
	}
	return total
}

// LevelFromXP returns the level corresponding to a total XP amount
func LevelFromXP(totalXP int) int {
	level := 1
	for totalXP >= TotalXPForLevel(level+1) {
		level++
		if level >= MaxLevel {
			return MaxLevel
		}
	}
	return level
}

const MaxLevel = 30

// LevelUpStats holds how much each stat grows per level per class
type LevelUpStats struct {
	HPPerLevel          int
	AttackModPerLevel   int
	StrengthModPerLevel int
	CAPerLevel          int
	ManaPerLevel        int
}

var ClassLevelScaling = map[string]LevelUpStats{
	"warrior": {HPPerLevel: 12, AttackModPerLevel: 1, StrengthModPerLevel: 1, CAPerLevel: 1, ManaPerLevel: 0},
	"mage":    {HPPerLevel: 6, AttackModPerLevel: 1, StrengthModPerLevel: 0, CAPerLevel: 0, ManaPerLevel: 15},
	"archer":  {HPPerLevel: 8, AttackModPerLevel: 1, StrengthModPerLevel: 1, CAPerLevel: 1, ManaPerLevel: 5},
	"rogue":   {HPPerLevel: 7, AttackModPerLevel: 1, StrengthModPerLevel: 1, CAPerLevel: 0, ManaPerLevel: 8},
	"shaman":  {HPPerLevel: 9, AttackModPerLevel: 0, StrengthModPerLevel: 1, CAPerLevel: 0, ManaPerLevel: 12},
}
