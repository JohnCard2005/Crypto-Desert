package game

import (
	"math/rand"

	"crypto-desert/internal/characters"
	"crypto-desert/internal/combat"
	"crypto-desert/internal/enemies"
)

// AIDecision represents the action the AI has chosen
type AIDecision struct {
	Action ActionType
	Target *characters.Character // nil for self-targeting or non-targeted actions
	Reason string                // human-readable explanation (for battle log)
}

// ── Decision Engine ───────────────────────────────────────────────────────────

// DecideAction selects the best action for an enemy given the current battle state.
// Each AIBehavior implements a distinct heuristic.
func DecideAction(enemy *enemies.Enemy, targets []*characters.Character) AIDecision {
	// Filter to only living targets
	live := livingTargets(targets)
	if len(live) == 0 {
		return AIDecision{Action: ActionDefend, Reason: "sem alvos vivos"}
	}

	switch enemy.Behavior {
	case enemies.BehaviorAggressive:
		return aggressiveAI(enemy, live)
	case enemies.BehaviorDefensive:
		return defensiveAI(enemy, live)
	case enemies.BehaviorBerserker:
		return berserkerAI(enemy, live)
	case enemies.BehaviorSupport:
		return supportAI(enemy, live)
	case enemies.BehaviorRandom:
		return randomAI(enemy, live)
	default:
		return aggressiveAI(enemy, live)
	}
}

// ── Behavior Implementations ──────────────────────────────────────────────────

// aggressiveAI: always attacks, targets the weakest (lowest HP) player
func aggressiveAI(enemy *enemies.Enemy, targets []*characters.Character) AIDecision {
	target := weakestTarget(targets)

	// Use ability if available — aggressive enemies love burst damage
	if enemy.CanUseAbility() {
		return AIDecision{
			Action: ActionAbility,
			Target: target,
			Reason: "usando habilidade especial no alvo mais fraco",
		}
	}

	return AIDecision{
		Action: ActionAttack,
		Target: target,
		Reason: "atacando o alvo mais fraco",
	}
}

// defensiveAI: defends when below HP threshold, attacks the strongest enemy otherwise
func defensiveAI(enemy *enemies.Enemy, targets []*characters.Character) AIDecision {
	// Prioritize defending when low HP
	if enemy.IsLowHP() && !enemy.HasStatus(characters.StatusDefending) {
		return AIDecision{
			Action: ActionDefend,
			Reason: "HP baixo — assumindo postura defensiva",
		}
	}

	// When healthy, target the most dangerous enemy (highest attack)
	target := mostDangerousTarget(targets)

	if enemy.CanUseAbility() && !enemy.IsLowHP() {
		return AIDecision{
			Action: ActionAbility,
			Target: target,
			Reason: "usando habilidade enquanto tem HP confortável",
		}
	}

	return AIDecision{
		Action: ActionAttack,
		Target: target,
		Reason: "atacando o alvo mais perigoso",
	}
}

// berserkerAI: uses ability as soon as available, otherwise attacks weakest
// Goes into a frenzy when low HP — never defends
func berserkerAI(enemy *enemies.Enemy, targets []*characters.Character) AIDecision {
	target := weakestTarget(targets)

	// Berserkers prioritize ability above everything
	if enemy.CanUseAbility() {
		return AIDecision{
			Action: ActionAbility,
			Target: target,
			Reason: "FRENESI — descarregando habilidade especial",
		}
	}

	// When low HP, berserker attacks with extra aggression (no defending)
	if enemy.IsLowHP() {
		target = weakestTarget(targets) // stays on weakest to finish kills
		return AIDecision{
			Action: ActionAttack,
			Target: target,
			Reason: "FRENESI — HP crítico, ignorando defesa",
		}
	}

	return AIDecision{
		Action: ActionAttack,
		Target: target,
		Reason: "atacando o alvo mais fraco",
	}
}

// supportAI: applies debuffs first, then attacks; conserves ability for key moments
func supportAI(enemy *enemies.Enemy, targets []*characters.Character) AIDecision {
	// Find a target that doesn't have debuffs yet
	unbuffed := targetWithoutDebuff(targets)
	mainTarget := weakestTarget(targets)

	// Use ability on a clean target when available
	if enemy.CanUseAbility() && unbuffed != nil {
		return AIDecision{
			Action: ActionAbility,
			Target: unbuffed,
			Reason: "aplicando debuff em alvo sem status",
		}
	}

	// If ability is on cooldown, attack the target most affected by debuffs (already weakened)
	debuffed := mostDebuffedTarget(targets)
	if debuffed != nil {
		return AIDecision{
			Action: ActionAttack,
			Target: debuffed,
			Reason: "atacando alvo já debuffado",
		}
	}

	return AIDecision{
		Action: ActionAttack,
		Target: mainTarget,
		Reason: "atacando alvo principal",
	}
}

// randomAI: completely unpredictable — weighted random between all actions
func randomAI(enemy *enemies.Enemy, targets []*characters.Character) AIDecision {
	target := targets[rand.Intn(len(targets))]

	roll := rand.Intn(10)
	switch {
	case roll <= 0: // 10% flee attempt
		return AIDecision{
			Action: ActionFlee,
			Reason: "comportamento caótico — tentando fugir!",
		}
	case roll <= 2: // 20% defend
		return AIDecision{
			Action: ActionDefend,
			Reason: "comportamento caótico — defendendo aleatoriamente",
		}
	case roll <= 4 && enemy.CanUseAbility(): // 20% ability if available
		return AIDecision{
			Action: ActionAbility,
			Target: target,
			Reason: "comportamento caótico — usando habilidade por impulso",
		}
	default: // 50%+ attack
		return AIDecision{
			Action: ActionAttack,
			Target: target,
			Reason: "comportamento caótico — atacando aleatoriamente",
		}
	}
}

// ── Target Selection Helpers ──────────────────────────────────────────────────

func livingTargets(targets []*characters.Character) []*characters.Character {
	live := make([]*characters.Character, 0, len(targets))
	for _, t := range targets {
		if t.IsAlive() {
			live = append(live, t)
		}
	}
	return live
}

// weakestTarget returns the player with the lowest current HP
func weakestTarget(targets []*characters.Character) *characters.Character {
	if len(targets) == 0 {
		return nil
	}
	weakest := targets[0]
	for _, t := range targets[1:] {
		if t.HP < weakest.HP {
			weakest = t
		}
	}
	return weakest
}

// mostDangerousTarget returns the player with the highest AttackMod
func mostDangerousTarget(targets []*characters.Character) *characters.Character {
	if len(targets) == 0 {
		return nil
	}
	strongest := targets[0]
	for _, t := range targets[1:] {
		if t.AttackMod > strongest.AttackMod {
			strongest = t
		}
	}
	return strongest
}

// targetWithoutDebuff finds a player that has no active negative status
func targetWithoutDebuff(targets []*characters.Character) *characters.Character {
	negativeStatuses := []characters.StatusEffect{
		characters.StatusFrozen,
		characters.StatusPoisoned,
		characters.StatusBurning,
		characters.StatusParalyzed,
		characters.StatusBearTrap,
	}
	for _, t := range targets {
		hasDebuff := false
		for _, s := range negativeStatuses {
			if t.HasStatus(s) {
				hasDebuff = true
				break
			}
		}
		if !hasDebuff {
			return t
		}
	}
	return nil
}

// mostDebuffedTarget returns the player with the most active statuses
func mostDebuffedTarget(targets []*characters.Character) *characters.Character {
	if len(targets) == 0 {
		return nil
	}
	most := targets[0]
	for _, t := range targets[1:] {
		if len(t.Statuses) > len(most.Statuses) {
			most = t
		}
	}
	if len(most.Statuses) == 0 {
		return nil
	}
	return most
}

// ── Flee Resolution ───────────────────────────────────────────────────────────

// ResolveEnemyFlee handles a flee attempt for an enemy.
// Returns true if the enemy escaped (roll >= 15).
func ResolveEnemyFlee(enemy *enemies.Enemy) (escaped bool, roll int) {
	roll = combat.RollD20()
	return roll >= 15, roll
}
