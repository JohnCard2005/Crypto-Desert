package game

import (
	"crypto-desert/internal/characters"
	"crypto-desert/internal/combat"
	"crypto-desert/internal/enemies"
)

// ── PlayerCombatant ───────────────────────────────────────────────────────────

// PlayerCombatant wraps a player Character to satisfy the combat.Combatant interface
type PlayerCombatant struct {
	Char *characters.Character
}

func (p *PlayerCombatant) GetCharacter() *characters.Character { return p.Char }
func (p *PlayerCombatant) GetName() string                     { return p.Char.Name }
func (p *PlayerCombatant) GetTeam() combat.Team                { return combat.TeamPlayer }

// ── EnemyCombatant ────────────────────────────────────────────────────────────

// EnemyCombatant wraps an Enemy to satisfy the combat.Combatant interface
type EnemyCombatant struct {
	Enemy *enemies.Enemy
}

func (e *EnemyCombatant) GetCharacter() *characters.Character { return e.Enemy.Character }
func (e *EnemyCombatant) GetName() string                     { return e.Enemy.Name }
func (e *EnemyCombatant) GetTeam() combat.Team                { return combat.TeamEnemy }
