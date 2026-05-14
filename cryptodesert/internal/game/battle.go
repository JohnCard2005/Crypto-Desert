package game

import (
	"fmt"
	"math/rand"

	"crypto-desert/internal/characters"
	"crypto-desert/internal/combat"
	"crypto-desert/internal/enemies"
)

// ── Action Types ──────────────────────────────────────────────────────────────

type ActionType string

const (
	ActionAttack  ActionType = "attack"
	ActionDefend  ActionType = "defend"
	ActionAbility ActionType = "ability"
	ActionFlee    ActionType = "flee"
)

// PlayerAction is sent by the UI/handler to represent a player's chosen action
type PlayerAction struct {
	Type   ActionType
	Target *characters.Character // required for Attack and Ability
}

// ── Battle State ──────────────────────────────────────────────────────────────

// BattleStatus represents the current phase of a battle
type BattleStatus string

const (
	BattleOngoing    BattleStatus = "ongoing"
	BattlePlayerWon  BattleStatus = "player_won"
	BattleEnemyWon   BattleStatus = "enemy_won"
	BattlePlayerFled BattleStatus = "player_fled"
)

// TurnEvent is a single narrated event within a turn (for logging and display)
type TurnEvent struct {
	Actor   string
	Message string
	Damage  int
	IsError bool // critical miss, failed flee, etc.
}

// TurnResult holds everything that happened in one full turn cycle
type TurnResult struct {
	TurnNumber  int
	ActorName   string
	Action      ActionType
	Events      []TurnEvent
	BattleState BattleStatus
}

// BattleResult is returned when a battle ends
type BattleResult struct {
	Status      BattleStatus
	TurnCount   int
	Log         []TurnResult
	XPGained    int
	GoldGained  int
	Survivors   []*characters.Character
}

// ── Battle ────────────────────────────────────────────────────────────────────

// Battle holds the full mutable state of an in-progress fight
type Battle struct {
	Players      []*characters.Character
	Enemies      []*enemies.Enemy
	Queue        *combat.InitiativeQueue
	TurnNumber   int
	Log          []TurnResult
	Status       BattleStatus

	// Channel for receiving player actions (populated by handler/UI)
	playerActions chan PlayerAction

	// Maps player name → pending action (for synchronous mode)
	pendingActions map[string]PlayerAction
}

// NewBattle creates and initializes a battle, rolls initiative, and builds the queue.
func NewBattle(players []*characters.Character, enemyGroup []*enemies.Enemy) *Battle {
	b := &Battle{
		Players:        players,
		Enemies:        enemyGroup,
		Queue:          combat.NewInitiativeQueue(),
		Status:         BattleOngoing,
		pendingActions: make(map[string]PlayerAction),
	}

	// Reset all combatants for a new battle
	for _, p := range players {
		p.ResetForNewBattle()
	}
	for _, e := range enemyGroup {
		e.ResetForNewBattle()
	}

	b.rollInitiative()
	return b
}

// rollInitiative rolls d20 + Speed for every combatant and fills the queue
func (b *Battle) rollInitiative() {
	for _, p := range b.Players {
		roll := combat.RollD20() + p.Speed
		b.Queue.Enqueue(&PlayerCombatant{Char: p}, roll, p.Speed)
	}
	for _, e := range b.Enemies {
		roll := combat.RollD20() + e.Speed
		b.Queue.Enqueue(&EnemyCombatant{Enemy: e}, roll, e.Speed)
	}
}

// InitiativeOrder returns a snapshot of the turn order for display
func (b *Battle) InitiativeOrder() []combat.InitiativeEntry {
	return b.Queue.TurnOrder()
}

// ── Turn Processing ───────────────────────────────────────────────────────────

// ProcessPlayerTurn executes a player's chosen action and advances the state.
// Returns the TurnResult and the updated BattleStatus.
func (b *Battle) ProcessPlayerTurn(actor *characters.Character, action PlayerAction) TurnResult {
	b.TurnNumber++
	result := TurnResult{
		TurnNumber:  b.TurnNumber,
		ActorName:   actor.Name,
		Action:      action.Type,
		BattleState: BattleOngoing,
	}

	// Check if this character must skip
	if actor.ShouldSkipTurn() {
		result.Events = append(result.Events, TurnEvent{
			Actor:   actor.Name,
			Message: fmt.Sprintf("%s está incapacitado e perde o turno!", actor.Name),
			IsError: true,
		})
		b.endOfTurn(actor, &result)
		return result
	}

	switch action.Type {

	case ActionAttack:
		if action.Target == nil || !action.Target.IsAlive() {
			action.Target = b.randomLivingEnemy()
		}
		if action.Target == nil {
			break
		}
		ar := combat.ResolveAttack(actor, action.Target, 1.0)
		result.Events = append(result.Events, narrateAttack(actor.Name, action.Target.Name, ar))

	case ActionDefend:
		actor.Defending = true
		actor.AddStatus(characters.StatusDefending, 1, 4)
		result.Events = append(result.Events, TurnEvent{
			Actor:   actor.Name,
			Message: fmt.Sprintf("%s assume postura defensiva (+4 CA até o próximo turno).", actor.Name),
		})

	case ActionAbility:
		if !actor.CanUseAbility() {
			result.Events = append(result.Events, TurnEvent{
				Actor:   actor.Name,
				Message: fmt.Sprintf("%s tentou usar %s mas a habilidade não está disponível!", actor.Name, actor.Ability.Name),
				IsError: true,
			})
			break
		}
		if !actor.UseAbility() {
			result.Events = append(result.Events, TurnEvent{
				Actor:   actor.Name,
				Message: fmt.Sprintf("%s não tem mana suficiente para %s!", actor.Name, actor.Ability.Name),
				IsError: true,
			})
			break
		}
		events := b.resolveAbility(actor, action.Target)
		result.Events = append(result.Events, events...)

	case ActionFlee:
		roll := combat.RollD20()
		if roll >= 15 {
			b.Status = BattlePlayerFled
			result.Events = append(result.Events, TurnEvent{
				Actor:   actor.Name,
				Message: fmt.Sprintf("%s rolou %d e conseguiu fugir da batalha!", actor.Name, roll),
			})
			result.BattleState = BattlePlayerFled
			b.Log = append(b.Log, result)
			return result
		}
		result.Events = append(result.Events, TurnEvent{
			Actor:   actor.Name,
			Message: fmt.Sprintf("%s tentou fugir (rolou %d) mas falhou! Precisa de 15+.", actor.Name, roll),
			IsError: true,
		})
	}

	b.endOfTurn(actor, &result)
	b.checkBattleEnd(&result)
	b.Log = append(b.Log, result)
	return result
}

// ProcessEnemyTurn runs the AI decision and resolves it automatically.
func (b *Battle) ProcessEnemyTurn(enemy *enemies.Enemy) TurnResult {
	b.TurnNumber++
	result := TurnResult{
		TurnNumber:  b.TurnNumber,
		ActorName:   enemy.Name,
		Action:      ActionAttack,
		BattleState: BattleOngoing,
	}

	// Skip check
	if enemy.ShouldSkipTurn() {
		result.Events = append(result.Events, TurnEvent{
			Actor:   enemy.Name,
			Message: fmt.Sprintf("%s está incapacitado e perde o turno!", enemy.Name),
			IsError: true,
		})
		b.endOfTurnEnemy(enemy, &result)
		return result
	}

	// AI decides
	decision := DecideAction(enemy, b.livingPlayers())
	result.Action = decision.Action

	result.Events = append(result.Events, TurnEvent{
		Actor:   enemy.Name,
		Message: fmt.Sprintf("[IA] %s: %s", enemy.Name, decision.Reason),
	})

	switch decision.Action {

	case ActionAttack:
		if decision.Target == nil {
			decision.Target = b.randomLivingPlayer()
		}
		if decision.Target == nil {
			break
		}
		ar := combat.ResolveAttack(enemy.Character, decision.Target, 1.0)
		result.Events = append(result.Events, narrateAttack(enemy.Name, decision.Target.Name, ar))

	case ActionDefend:
		enemy.Defending = true
		enemy.AddStatus(characters.StatusDefending, 1, 4)
		result.Events = append(result.Events, TurnEvent{
			Actor:   enemy.Name,
			Message: fmt.Sprintf("%s assume postura defensiva.", enemy.Name),
		})

	case ActionAbility:
		target := decision.Target
		if target == nil {
			target = b.randomLivingPlayer()
		}
		if !enemy.UseAbility() {
			// Fallback to attack if ability fails
			if target != nil {
				ar := combat.ResolveAttack(enemy.Character, target, 1.0)
				result.Events = append(result.Events, narrateAttack(enemy.Name, target.Name, ar))
			}
			break
		}
		events := b.resolveAbilityFromChar(enemy.Character, target)
		result.Events = append(result.Events, events...)

	case ActionFlee:
		escaped, roll := ResolveEnemyFlee(enemy)
		if escaped {
			enemy.Alive = false
			result.Events = append(result.Events, TurnEvent{
				Actor:   enemy.Name,
				Message: fmt.Sprintf("%s rolou %d e fugiu da batalha!", enemy.Name, roll),
			})
		} else {
			result.Events = append(result.Events, TurnEvent{
				Actor:   enemy.Name,
				Message: fmt.Sprintf("%s tentou fugir (rolou %d) mas falhou.", enemy.Name, roll),
				IsError: true,
			})
			// Fallback: attack
			if target := b.randomLivingPlayer(); target != nil {
				ar := combat.ResolveAttack(enemy.Character, target, 1.0)
				result.Events = append(result.Events, narrateAttack(enemy.Name, target.Name, ar))
			}
		}
	}

	b.endOfTurnEnemy(enemy, &result)
	b.checkBattleEnd(&result)
	b.Log = append(b.Log, result)
	return result
}

// ── Ability Resolution ────────────────────────────────────────────────────────

// resolveAbility executes a player character's special ability
func (b *Battle) resolveAbility(actor *characters.Character, target *characters.Character) []TurnEvent {
	return b.resolveAbilityFromChar(actor, target)
}

func (b *Battle) resolveAbilityFromChar(actor *characters.Character, target *characters.Character) []TurnEvent {
	ability := actor.Ability
	var events []TurnEvent

	events = append(events, TurnEvent{
		Actor:   actor.Name,
		Message: fmt.Sprintf("✦ %s usa %s!", actor.Name, ability.Name),
	})

	if ability.Targeting == characters.TargetSelf {
		// Self-buff abilities
		actor.AddStatus(characters.StatusBuffed, 2, 3)
		events = append(events, TurnEvent{
			Actor:   actor.Name,
			Message: fmt.Sprintf("%s recebe um bônus de ataque por 2 turnos.", actor.Name),
		})
		return events
	}

	if target == nil || !target.IsAlive() {
		target = b.randomLivingPlayer()
		if target == nil {
			target = b.randomLivingEnemy()
		}
	}
	if target == nil {
		return events
	}

	// Resolve the attack with the ability's damage multiplier
	ar := combat.ResolveAttack(actor, target, ability.DamageMult)
	events = append(events, narrateAttack(actor.Name, target.Name, ar))

	// Apply status effect if the ability hits
	if ability.AppliesStatus != "" && (ar.Outcome == combat.OutcomeHit || ar.Outcome == combat.OutcomeCriticalHit) {
		target.AddStatus(ability.AppliesStatus, ability.StatusDuration, ability.StatusPower)
		events = append(events, TurnEvent{
			Actor:   actor.Name,
			Message: fmt.Sprintf("→ %s recebe o status [%s] por %d turno(s).", target.Name, ability.AppliesStatus, ability.StatusDuration),
		})
	}

	return events
}

// ── End-of-Turn Housekeeping ──────────────────────────────────────────────────

func (b *Battle) endOfTurn(actor *characters.Character, result *TurnResult) {
	// Reset defend stance after their turn resolves
	actor.Defending = false

	// Tick status effects
	for _, msg := range actor.TickStatuses() {
		result.Events = append(result.Events, TurnEvent{Actor: actor.Name, Message: msg})
	}

	// Tick ability cooldown
	actor.TickAbilityCooldown()

	// Remove dead combatants from queue
	b.Queue.RemoveDead()
}

func (b *Battle) endOfTurnEnemy(enemy *enemies.Enemy, result *TurnResult) {
	enemy.Defending = false
	enemy.IncrementTurnCounter()

	for _, msg := range enemy.TickStatuses() {
		result.Events = append(result.Events, TurnEvent{Actor: enemy.Name, Message: msg})
	}

	enemy.TickAbilityCooldown()
	b.Queue.RemoveDead()
}

// ── Battle End Check ──────────────────────────────────────────────────────────

func (b *Battle) checkBattleEnd(result *TurnResult) {
	players, enemies := b.Queue.CountByTeam()

	if players == 0 && enemies == 0 {
		b.Status = BattleEnemyWon // everyone died — enemies "win" by default
	} else if enemies == 0 {
		b.Status = BattlePlayerWon
	} else if players == 0 {
		b.Status = BattleEnemyWon
	}

	result.BattleState = b.Status
}

// Conclude tallies rewards when the battle is over
func (b *Battle) Conclude() BattleResult {
	var xp, gold int
	survivors := make([]*characters.Character, 0)

	for _, e := range b.Enemies {
		if !e.IsAlive() {
			xp += e.XPReward
			gold += e.GoldReward
		}
	}

	for _, p := range b.Players {
		if p.IsAlive() {
			survivors = append(survivors, p)
		}
	}

	// Distribute XP only to survivors
	if b.Status == BattlePlayerWon {
		share := xp
		if len(survivors) > 0 {
			share = xp / len(survivors)
		}
		for _, p := range survivors {
			p.GainXP(share)
		}
	}

	return BattleResult{
		Status:     b.Status,
		TurnCount:  b.TurnNumber,
		Log:        b.Log,
		XPGained:   xp,
		GoldGained: gold,
		Survivors:  survivors,
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (b *Battle) livingPlayers() []*characters.Character {
	live := make([]*characters.Character, 0)
	for _, p := range b.Players {
		if p.IsAlive() {
			live = append(live, p)
		}
	}
	return live
}

func (b *Battle) randomLivingPlayer() *characters.Character {
	live := b.livingPlayers()
	if len(live) == 0 {
		return nil
	}
	return live[rand.Intn(len(live))]
}

func (b *Battle) randomLivingEnemy() *characters.Character {
	live := make([]*characters.Character, 0)
	for _, e := range b.Enemies {
		if e.IsAlive() {
			live = append(live, e.Character)
		}
	}
	if len(live) == 0 {
		return nil
	}
	return live[rand.Intn(len(live))]
}

// narrateAttack converts an AttackResult into a human-readable TurnEvent
func narrateAttack(attackerName, defenderName string, ar combat.AttackResult) TurnEvent {
	switch ar.Outcome {
	case combat.OutcomeCriticalMiss:
		return TurnEvent{
			Actor:   attackerName,
			Message: fmt.Sprintf("💀 FALHA CRÍTICA! %s rolou 1 e errou feio.", attackerName),
			IsError: true,
		}
	case combat.OutcomeMiss:
		return TurnEvent{
			Actor:   attackerName,
			Message: fmt.Sprintf("↷ %s ataca %s (rolou %d+mod=%d) mas erra. CA do alvo: alta demais.",
				attackerName, defenderName, ar.Roll, ar.HitValue),
		}
	case combat.OutcomeHit:
		return TurnEvent{
			Actor:   attackerName,
			Message: fmt.Sprintf("⚔ %s acerta %s por %d de dano! (roll:%d, fator_cripto:%.2f)",
				attackerName, defenderName, ar.Damage, ar.Roll, ar.CryptoFactor),
			Damage: ar.Damage,
		}
	case combat.OutcomeCriticalHit:
		return TurnEvent{
			Actor:   attackerName,
			Message: fmt.Sprintf("💥 ACERTO CRÍTICO! %s destroça %s por %d de dano!! (roll:20)",
				attackerName, defenderName, ar.Damage),
			Damage: ar.Damage,
		}
	}
	return TurnEvent{Actor: attackerName, Message: "ação desconhecida"}
}
