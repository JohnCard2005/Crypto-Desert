package characters

import (
	"fmt"
	"math/rand"
)

// ── Vitals ───────────────────────────────────────────────────────────────────

func (c *Character) TakeDamage(damage int) {
	c.HP -= damage
	if c.HP <= 0 {
		c.HP = 0
		c.Alive = false
	}
}

func (c *Character) Heal(amount int) {
	c.HP += amount
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
}

func (c *Character) SpendMana(amount int) bool {
	if c.Mana < amount {
		return false
	}
	c.Mana -= amount
	return true
}

func (c *Character) RestoreMana(amount int) {
	c.Mana += amount
	if c.Mana > c.MaxMana {
		c.Mana = c.MaxMana
	}
}

func (c *Character) IsAlive() bool {
	return c.Alive
}

// ── Status Effects ────────────────────────────────────────────────────────────

// AddStatus applies a status effect. If already present, refreshes duration.
func (c *Character) AddStatus(effect StatusEffect, duration, power int) {
	for i, s := range c.Statuses {
		if s.Effect == effect {
			c.Statuses[i].Duration = duration
			c.Statuses[i].Power = power
			return
		}
	}
	c.Statuses = append(c.Statuses, ActiveStatus{
		Effect:   effect,
		Duration: duration,
		Power:    power,
	})
}

// HasStatus returns true if the character has the given status active
func (c *Character) HasStatus(effect StatusEffect) bool {
	for _, s := range c.Statuses {
		if s.Effect == effect {
			return true
		}
	}
	return false
}

// RemoveStatus removes a specific status effect
func (c *Character) RemoveStatus(effect StatusEffect) {
	filtered := c.Statuses[:0]
	for _, s := range c.Statuses {
		if s.Effect != effect {
			filtered = append(filtered, s)
		}
	}
	c.Statuses = filtered
}

// TickStatuses processes end-of-turn status effects.
// Returns a log of what happened (damage taken, statuses expired, etc.)
func (c *Character) TickStatuses() []string {
	var log []string
	remaining := c.Statuses[:0]

	for _, s := range c.Statuses {
		mods := s.Modifiers()

		// Apply DoT damage
		if mods.DamagePerTurn > 0 {
			c.TakeDamage(mods.DamagePerTurn)
			log = append(log, fmt.Sprintf("%s sofreu %d de dano por %s", c.Name, mods.DamagePerTurn, s.Effect))
		}

		s.Duration--
		if s.Duration > 0 {
			remaining = append(remaining, s)
		} else {
			log = append(log, fmt.Sprintf("%s: status %q expirou", c.Name, s.Effect))
		}
	}

	c.Statuses = remaining
	return log
}

// GetCombatModifiers aggregates all active status modifiers
func (c *Character) GetCombatModifiers() StatusModifiers {
	combined := StatusModifiers{CryptoMult: 1.0}
	for _, s := range c.Statuses {
		m := s.Modifiers()
		combined.AttackBonus += m.AttackBonus
		combined.DefenseBonus += m.DefenseBonus
		combined.DamagePerTurn += m.DamagePerTurn
		if m.SkipTurn {
			combined.SkipTurn = true
		}
		if m.SkipChance > combined.SkipChance {
			combined.SkipChance = m.SkipChance
		}
		if m.CryptoMult != 0 {
			combined.CryptoMult *= m.CryptoMult
		}
	}
	if combined.CryptoMult == 0 {
		combined.CryptoMult = 1.0
	}
	return combined
}

// ShouldSkipTurn returns true if the character must skip (frozen or paralyzed check)
func (c *Character) ShouldSkipTurn() bool {
	mods := c.GetCombatModifiers()
	if mods.SkipTurn {
		return true
	}
	if mods.SkipChance > 0 {
		return rand.Float64() < mods.SkipChance
	}
	return false
}

// EffectiveCA returns the character's current CA including status bonuses and Defend action
func (c *Character) EffectiveCA() int {
	mods := c.GetCombatModifiers()
	ca := c.CA + mods.DefenseBonus
	if c.Defending {
		ca += 4 // Defend action bonus
	}
	return ca
}

// EffectiveAttackMod returns attack modifier including active status effects
func (c *Character) EffectiveAttackMod() int {
	mods := c.GetCombatModifiers()
	return c.AttackMod + mods.AttackBonus
}

// ── Ability ───────────────────────────────────────────────────────────────────

// CanUseAbility checks if the special ability is available this turn
func (c *Character) CanUseAbility() bool {
	if c.Ability.Cooldown == 0 {
		return !c.AbilityUsed
	}
	return c.AbilityCooldownLeft == 0
}

// UseAbility marks the ability as used and sets cooldown
func (c *Character) UseAbility() bool {
	if !c.CanUseAbility() {
		return false
	}
	if c.Ability.ManaCost > 0 && !c.SpendMana(c.Ability.ManaCost) {
		return false
	}
	if c.Ability.Cooldown == 0 {
		c.AbilityUsed = true
	} else {
		c.AbilityCooldownLeft = c.Ability.Cooldown
	}
	return true
}

// TickAbilityCooldown decrements the cooldown counter at end of turn
func (c *Character) TickAbilityCooldown() {
	if c.AbilityCooldownLeft > 0 {
		c.AbilityCooldownLeft--
	}
}

// ResetForNewBattle resets per-battle state (cooldowns, one-time abilities, defend stance)
func (c *Character) ResetForNewBattle() {
	c.AbilityUsed = false
	c.AbilityCooldownLeft = 0
	c.Defending = false
	c.Statuses = nil
	c.Alive = c.HP > 0
}

// ── XP and Leveling ───────────────────────────────────────────────────────────

// GainXP adds XP and triggers level-ups if the threshold is crossed.
// Returns the number of levels gained.
func (c *Character) GainXP(amount int) int {
	if c.Level >= MaxLevel {
		return 0
	}

	c.XP += amount
	levelsGained := 0

	for c.Level < MaxLevel && c.XP >= TotalXPForLevel(c.Level+1) {
		c.applyLevelUp()
		levelsGained++
	}

	c.XPToNext = TotalXPForLevel(c.Level+1) - c.XP
	if c.Level >= MaxLevel {
		c.XPToNext = 0
	}

	return levelsGained
}

// applyLevelUp increments level and scales stats according to class scaling table
func (c *Character) applyLevelUp() {
	c.Level++

	scaling, ok := ClassLevelScaling[c.Class]
	if !ok {
		scaling = LevelUpStats{HPPerLevel: 8, AttackModPerLevel: 1, ManaPerLevel: 5}
	}

	hpGain := scaling.HPPerLevel
	c.MaxHP += hpGain
	c.HP += hpGain // heal the gained HP on level up

	c.MaxMana += scaling.ManaPerLevel
	c.Mana = c.MaxMana // restore mana on level up

	c.AttackMod += scaling.AttackModPerLevel
	c.StrengthMod += scaling.StrengthModPerLevel
	c.CA += scaling.CAPerLevel
}
