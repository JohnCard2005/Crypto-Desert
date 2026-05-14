package combat

import (
	"math/rand"

	"crypto-desert/internal/characters"
)

// ── Dice ─────────────────────────────────────────────────────────────────────

func RollD20() int {
	return rand.Intn(20) + 1
}

func RollDice(sides int) int {
	if sides <= 0 {
		return 1
	}
	return rand.Intn(sides) + 1
}

// ── Crypto Factor ─────────────────────────────────────────────────────────────

// CryptoFactor converts a percentage variation into a damage multiplier.
// Clamped to [0.5, 2.0] to prevent absurd results.
//
//	+100% variation → factor 2.0 (double damage)
//	   0% variation → factor 1.0 (normal)
//	 -50% variation → factor 0.5 (minimum)
func CryptoFactor(variation float64) float64 {
	factor := 1.0 + (variation / 100.0)
	if factor < 0.5 {
		return 0.5
	}
	if factor > 2.0 {
		return 2.0
	}
	return factor
}

// ── Attack Resolution ─────────────────────────────────────────────────────────

// AttackOutcome describes the full result of a single attack roll
type AttackOutcome string

const (
	OutcomeCriticalMiss AttackOutcome = "critical_miss" // roll = 1
	OutcomeMiss         AttackOutcome = "miss"          // hit < target CA
	OutcomeHit          AttackOutcome = "hit"           // normal hit
	OutcomeCriticalHit  AttackOutcome = "critical_hit"  // roll = 20
)

// AttackResult holds everything that happened in one attack action
type AttackResult struct {
	Outcome      AttackOutcome
	Roll         int     // raw d20 result
	HitValue     int     // roll + attack modifier
	Damage       int     // final damage dealt (0 on miss)
	CryptoFactor float64 // the factor applied
	IsCrit       bool
}

// ResolveAttack performs a full attack from attacker → defender.
// It reads DamageDice, StrengthMod, EffectiveCA, and CryptoVariation directly
// from the Character structs. critMultiplier is 1.0 for normal, 2.0 for crit,
// 3.0 for special abilities like Warrior's Fúria.
func ResolveAttack(attacker, defender *characters.Character, critMultiplier float64) AttackResult {
	roll := RollD20()

	// Critical miss — automatic failure
	if roll == 1 {
		return AttackResult{
			Outcome: OutcomeCriticalMiss,
			Roll:    roll,
		}
	}

	// Critical hit — bypass CA, double (or more) damage
	if roll == 20 {
		if critMultiplier < 2.0 {
			critMultiplier = 2.0
		}
		dmg := calculateFinalDamage(attacker, defender, critMultiplier)
		defender.TakeDamage(dmg)
		return AttackResult{
			Outcome:      OutcomeCriticalHit,
			Roll:         roll,
			HitValue:     roll + attacker.EffectiveAttackMod(),
			Damage:       dmg,
			CryptoFactor: CryptoFactor(attacker.CryptoVariation),
			IsCrit:       true,
		}
	}

	// Normal attack — check against defender's CA
	hitValue := roll + attacker.EffectiveAttackMod()
	if hitValue < defender.EffectiveCA() {
		return AttackResult{
			Outcome:  OutcomeMiss,
			Roll:     roll,
			HitValue: hitValue,
		}
	}

	dmg := calculateFinalDamage(attacker, defender, critMultiplier)
	defender.TakeDamage(dmg)
	return AttackResult{
		Outcome:      OutcomeHit,
		Roll:         roll,
		HitValue:     hitValue,
		Damage:       dmg,
		CryptoFactor: CryptoFactor(attacker.CryptoVariation),
	}
}

// calculateFinalDamage computes the actual damage value following the spec formula:
//
//	dado_dano  = rolar(dado_da_classe)
//	dano_bruto = (dado_dano + mod_forca) * fator_cripto * critMultiplier * statusMult
//	dano_final = max(1, dano_bruto - defesa_alvo)
func calculateFinalDamage(attacker, defender *characters.Character, critMultiplier float64) int {
	// Roll the class damage die
	diceRoll := RollDice(attacker.DamageDice)

	// Get crypto factor, then apply any status multipliers on top
	mods := attacker.GetCombatModifiers()
	cryptoFact := CryptoFactor(attacker.CryptoVariation) * mods.CryptoMult

	// Clamp after status multiplication
	if cryptoFact < 0.5 {
		cryptoFact = 0.5
	}
	if cryptoFact > 3.0 {
		cryptoFact = 3.0 // bosses can push this high
	}

	rawDamage := float64(diceRoll+attacker.StrengthMod) * cryptoFact * critMultiplier

	// Subtract defender's flat defense
	final := int(rawDamage) - defender.Defense
	if final < 1 {
		return 1
	}
	return final
}
