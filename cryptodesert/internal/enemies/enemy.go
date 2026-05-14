package enemies

import "crypto-desert/internal/characters"

// AIBehavior defines how an enemy picks its action each turn
type AIBehavior string

const (
	BehaviorAggressive AIBehavior = "aggressive" // always attacks, prefers weakest target
	BehaviorDefensive  AIBehavior = "defensive"  // defends when low HP, attacks otherwise
	BehaviorBerserker  AIBehavior = "berserker"  // uses ability as soon as available, then attacks
	BehaviorSupport    AIBehavior = "support"    // applies debuffs before attacking
	BehaviorRandom     AIBehavior = "random"     // chaotic, unpredictable
)

// EnemyTier controls XP/loot reward scaling
type EnemyTier string

const (
	TierCommon  EnemyTier = "common"
	TierElite   EnemyTier = "elite"
	TierBoss    EnemyTier = "boss"
)

// Enemy wraps a Character with AI metadata.
// The Character holds all combat state; Enemy adds behavioral identity.
type Enemy struct {
	*characters.Character

	// AI
	Behavior AIBehavior

	// Identity
	Tier        EnemyTier
	Description string

	// Rewards on death
	XPReward   int
	GoldReward int

	// Internal AI state
	turnsSinceAbility int  // tracks turns to help decide ability usage
	lowHPThreshold    int  // HP value below which defensive AI switches to heal/defend
}

// IsEnemy marks this as an enemy combatant (used in battle to distinguish sides)
func (e *Enemy) IsEnemy() bool { return true }

// NewEnemy wraps an existing Character with enemy metadata.
func NewEnemy(c *characters.Character, behavior AIBehavior, tier EnemyTier, desc string, xpReward, goldReward int) *Enemy {
	lowHP := c.MaxHP / 3 // defensive threshold: 33% HP
	return &Enemy{
		Character:      c,
		Behavior:       behavior,
		Tier:           tier,
		Description:    desc,
		XPReward:       xpReward,
		GoldReward:     goldReward,
		lowHPThreshold: lowHP,
	}
}

// IncrementTurnCounter is called each time the enemy completes a turn
func (e *Enemy) IncrementTurnCounter() {
	e.turnsSinceAbility++
}

// ResetTurnCounter resets after ability use
func (e *Enemy) ResetTurnCounter() {
	e.turnsSinceAbility = 0
}

// TurnsSinceAbility returns how many turns have passed since last ability use
func (e *Enemy) TurnsSinceAbility() int {
	return e.turnsSinceAbility
}

// IsLowHP returns true if the enemy is below the defensive HP threshold
func (e *Enemy) IsLowHP() bool {
	return e.HP <= e.lowHPThreshold
}
