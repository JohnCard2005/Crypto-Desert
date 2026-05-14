package characters

// StatusEffect represents a temporary condition applied to a character
type StatusEffect string

const (
	StatusFrozen    StatusEffect = "frozen"    // skips turn
	StatusPoisoned  StatusEffect = "poisoned"  // loses HP each turn
	StatusBurning   StatusEffect = "burning"   // loses HP, reduced defense
	StatusParalyzed StatusEffect = "paralyzed" // 50% chance to skip turn
	StatusBuffed    StatusEffect = "buffed"    // increased attack
	StatusDefending StatusEffect = "defending" // increased CA until next turn
	StatusBullRun   StatusEffect = "bull_run"  // crypto bonus amplified (2x factor)
	StatusBearTrap  StatusEffect = "bear_trap" // crypto penalty amplified (2x negative)
)

type ActiveStatus struct {
	Effect   StatusEffect
	Duration int // turns remaining
	Power    int // effect intensity (damage per turn for DoT, bonus value for buffs)
}

// StatusModifiers returns the combat modifiers applied by an active status
type StatusModifiers struct {
	AttackBonus   int
	DefenseBonus  int
	SkipTurn      bool
	SkipChance    float64 // 0.0–1.0
	DamagePerTurn int
	CryptoMult    float64 // multiplier on top of crypto factor
}

func (s ActiveStatus) Modifiers() StatusModifiers {
	switch s.Effect {
	case StatusFrozen:
		return StatusModifiers{SkipTurn: true, DefenseBonus: -2}
	case StatusPoisoned:
		return StatusModifiers{DamagePerTurn: s.Power}
	case StatusBurning:
		return StatusModifiers{DamagePerTurn: s.Power, DefenseBonus: -1}
	case StatusParalyzed:
		return StatusModifiers{SkipChance: 0.5, AttackBonus: -2}
	case StatusBuffed:
		return StatusModifiers{AttackBonus: s.Power}
	case StatusDefending:
		return StatusModifiers{DefenseBonus: s.Power}
	case StatusBullRun:
		return StatusModifiers{CryptoMult: 2.0}
	case StatusBearTrap:
		return StatusModifiers{CryptoMult: 0.5}
	}
	return StatusModifiers{}
}
