package combat

import "crypto-desert/internal/characters"

// ── Combatant interface ───────────────────────────────────────────────────────
// Any participant in battle must satisfy this interface.
// Both *characters.Character (players) and *enemies.Enemy satisfy it.

type Combatant interface {
	GetCharacter() *characters.Character
	GetName() string
	GetTeam() Team
}

type Team int

const (
	TeamPlayer Team = iota
	TeamEnemy
)

// ── Initiative Entry ──────────────────────────────────────────────────────────

// InitiativeEntry holds one combatant's position in the turn order
type InitiativeEntry struct {
	Combatant  Combatant
	Initiative int // d20 + speed modifier
	Speed      int // used as tiebreaker
}

// ── Initiative Queue ──────────────────────────────────────────────────────────
// Implemented as a sorted slice (insertion-sorted on enqueue).
// This is a circular queue: after every combatant acts, we rotate to the next.
// Complexity:
//   Enqueue  — O(n) due to insertion sort (n is small in RPG context, ≤ ~10)
//   Peek/Dequeue — O(1)
//   Remove dead — O(n)

type InitiativeQueue struct {
	entries []InitiativeEntry
	current int // index of the combatant whose turn it is
}

// NewInitiativeQueue creates an empty queue
func NewInitiativeQueue() *InitiativeQueue {
	return &InitiativeQueue{}
}

// Enqueue adds a combatant with their initiative roll.
// Maintains descending order by initiative (highest goes first).
// Ties broken by Speed; if still tied, insertion order is preserved.
func (q *InitiativeQueue) Enqueue(c Combatant, initiative, speed int) {
	entry := InitiativeEntry{
		Combatant:  c,
		Initiative: initiative,
		Speed:      speed,
	}

	// Find insertion point (descending initiative)
	pos := len(q.entries)
	for i, e := range q.entries {
		if initiative > e.Initiative || (initiative == e.Initiative && speed > e.Speed) {
			pos = i
			break
		}
	}

	// Insert at pos
	q.entries = append(q.entries, InitiativeEntry{})
	copy(q.entries[pos+1:], q.entries[pos:])
	q.entries[pos] = entry
}

// Len returns the number of combatants in the queue
func (q *InitiativeQueue) Len() int {
	return len(q.entries)
}

// Current returns the combatant whose turn it is without advancing
func (q *InitiativeQueue) Current() Combatant {
	if len(q.entries) == 0 {
		return nil
	}
	return q.entries[q.current].Combatant
}

// Advance moves to the next combatant in the rotation.
// Skips dead combatants automatically.
// Returns the next live combatant, or nil if the queue is empty.
func (q *InitiativeQueue) Advance() Combatant {
	if len(q.entries) == 0 {
		return nil
	}

	// Cycle through to find the next alive combatant
	for range q.entries {
		q.current = (q.current + 1) % len(q.entries)
		c := q.entries[q.current].Combatant
		if c.GetCharacter().IsAlive() {
			return c
		}
	}

	return nil // everyone is dead
}

// RemoveDead removes all dead combatants from the queue, adjusting current index.
func (q *InitiativeQueue) RemoveDead() {
	alive := q.entries[:0]
	currentName := ""
	if len(q.entries) > 0 {
		currentName = q.entries[q.current].Combatant.GetName()
	}

	for _, e := range q.entries {
		if e.Combatant.GetCharacter().IsAlive() {
			alive = append(alive, e)
		}
	}
	q.entries = alive

	// Re-find current position
	q.current = 0
	for i, e := range q.entries {
		if e.Combatant.GetName() == currentName {
			q.current = i
			break
		}
	}
}

// TurnOrder returns the current ordered list (for display purposes)
func (q *InitiativeQueue) TurnOrder() []InitiativeEntry {
	result := make([]InitiativeEntry, len(q.entries))
	copy(result, q.entries)
	return result
}

// CountByTeam returns how many live combatants each team has
func (q *InitiativeQueue) CountByTeam() (players int, enemies int) {
	for _, e := range q.entries {
		if !e.Combatant.GetCharacter().IsAlive() {
			continue
		}
		if e.Combatant.GetTeam() == TeamPlayer {
			players++
		} else {
			enemies++
		}
	}
	return
}
