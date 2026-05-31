package store

import (
	"database/sql"
	"time"
	"encoding/json"
	"fmt"
	"sync"

	"crypto-desert/internal/characters"
	"crypto-desert/internal/db"
	"crypto-desert/internal/game"
	"crypto-desert/internal/items"
	"crypto-desert/internal/missions"
)

// ── CharacterStore ────────────────────────────────────────────────────────────
// Suporta dois modos:
//   - SQLite (db != nil): persiste em disco, multi-usuário
//   - In-memory (db == nil): perde dados ao reiniciar

type CharacterStore struct {
	db     *db.DB
	mu     sync.RWMutex
	chars  map[int]*characters.Character
	nextID int
}

func NewCharacterStore(database *db.DB) *CharacterStore {
	return &CharacterStore{
		db:     database,
		chars:  make(map[int]*characters.Character),
		nextID: 1,
	}
}

func (s *CharacterStore) Create(c *characters.Character) error {
	if s.db != nil {
		// Primeiro INSERT com data temporária para obter o ID gerado
		err := s.db.QueryRow(
			`INSERT INTO characters (user_id, name, class, level, gold, data)
			 VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
			c.UserID, c.Name, c.Class, c.Level, c.Gold, "{}",
		).Scan(&c.ID)
		if err != nil {
			return fmt.Errorf("criar personagem: %w", err)
		}
		// Agora que temos o ID real, serializa e atualiza o data com ID correto
		data, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("serializar personagem: %w", err)
		}
		s.db.Exec(`UPDATE characters SET data = ? WHERE id = ?`, string(data), c.ID)
		s.db.Exec(`INSERT OR IGNORE INTO inventories (character_id, data) VALUES (?, '{}')`, c.ID)
		s.db.Exec(`INSERT OR IGNORE INTO progress (character_id, data) VALUES (?, '{}')`, c.ID)
		s.db.Exec(
			`INSERT OR REPLACE INTO ranking (character_id, user_id, name, class, level, gold, battles_won)
			 VALUES (?, ?, ?, ?, ?, ?, 0)`,
			c.ID, c.UserID, c.Name, c.Class, c.Level, c.Gold,
		)
		return nil
	}
	// In-memory
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = s.nextID
	s.nextID++
	clone := *c
	s.chars[c.ID] = &clone
	return nil
}

func (s *CharacterStore) Get(id int) (*characters.Character, error) {
	if s.db != nil {
		var data string
		err := s.db.QueryRow(`SELECT data FROM characters WHERE id = ?`, id).Scan(&data)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("personagem %d não encontrado", id)
		}
		if err != nil {
			return nil, fmt.Errorf("buscar personagem: %w", err)
		}
		var c characters.Character
		if err := json.Unmarshal([]byte(data), &c); err != nil {
			return nil, fmt.Errorf("deserializar personagem: %w", err)
		}
		// Garante que o ID vem do banco (não do JSON que pode estar desatualizado)
		c.ID = id
		return &c, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.chars[id]
	if !ok {
		return nil, fmt.Errorf("personagem %d não encontrado", id)
	}
	clone := *c
	return &clone, nil
}

func (s *CharacterStore) ListByUser(userID int) ([]*characters.Character, error) {
	if s.db != nil {
		rows, err := s.db.Query(
			`SELECT id, data FROM characters WHERE user_id = ? ORDER BY id`, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("listar personagens: %w", err)
		}
		defer rows.Close()
		var list []*characters.Character
		for rows.Next() {
			var dbID int
			var data string
			if err := rows.Scan(&dbID, &data); err != nil {
				continue
			}
			var c characters.Character
			if err := json.Unmarshal([]byte(data), &c); err != nil {
				continue
			}
			// Garante que o ID vem do banco
			c.ID = dbID
			list = append(list, &c)
		}
		return list, nil
	}
	// In-memory: retorna todos
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*characters.Character, 0, len(s.chars))
	for _, c := range s.chars {
		clone := *c
		list = append(list, &clone)
	}
	return list, nil
}

func (s *CharacterStore) List() []*characters.Character {
	list, _ := s.ListByUser(0)
	return list
}

func (s *CharacterStore) Update(c *characters.Character) error {
	if s.db != nil {
		data, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("serializar personagem: %w", err)
		}
		_, err = s.db.Exec(
			`UPDATE characters SET name=?, class=?, level=?, gold=?, data=?, updated_at=datetime('now')
			 WHERE id=?`,
			c.Name, c.Class, c.Level, c.Gold, string(data), c.ID,
		)
		if err != nil {
			return fmt.Errorf("atualizar personagem: %w", err)
		}
		s.db.Exec(
			`UPDATE ranking SET name=?, class=?, level=?, gold=?, updated_at=datetime('now')
			 WHERE character_id=?`,
			c.Name, c.Class, c.Level, c.Gold, c.ID,
		)
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := *c
	s.chars[c.ID] = &clone
	return nil
}

func (s *CharacterStore) Delete(id int) error {
	if s.db != nil {
		_, err := s.db.Exec(`DELETE FROM characters WHERE id = ?`, id)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.chars, id)
	return nil
}

func (s *CharacterStore) IncrementBattlesWon(charID int) {
	if s.db != nil {
		s.db.Exec(
			`UPDATE ranking SET battles_won = battles_won + 1 WHERE character_id = ?`, charID,
		)
	}
}

// ── RankingEntry ──────────────────────────────────────────────────────────────

type RankingEntry struct {
	CharacterID int    `json:"character_id"`
	UserID      int    `json:"user_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Level       int    `json:"level"`
	Gold        int    `json:"gold"`
	BattlesWon  int    `json:"battles_won"`
}

// ── RankingStore ──────────────────────────────────────────────────────────────

type RankingStore struct {
	db    *db.DB
	mu    sync.RWMutex
	cache []RankingEntry
}

func NewRankingStore(database *db.DB) *RankingStore {
	rs := &RankingStore{db: database}
	if database != nil {
		rs.refresh()
	}
	return rs
}

func (s *RankingStore) TopByLevel(n int) []RankingEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if n > len(s.cache) {
		n = len(s.cache)
	}
	return s.cache[:n]
}

func (s *RankingStore) TopByGold(n int) []RankingEntry {
	if s.db == nil {
		return s.TopByLevel(n)
	}
	rows, err := s.db.Query(
		`SELECT character_id, user_id, name, class, level, gold, battles_won
		 FROM ranking ORDER BY gold DESC LIMIT ?`, n,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanRanking(rows)
}

func (s *RankingStore) TopByBattles(n int) []RankingEntry {
	if s.db == nil {
		return s.TopByLevel(n)
	}
	rows, err := s.db.Query(
		`SELECT character_id, user_id, name, class, level, gold, battles_won
		 FROM ranking ORDER BY battles_won DESC LIMIT ?`, n,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return scanRanking(rows)
}

func (s *RankingStore) Refresh() {
	if s.db != nil {
		s.refresh()
	}
}

func (s *RankingStore) refresh() {
	rows, err := s.db.Query(
		`SELECT character_id, user_id, name, class, level, gold, battles_won
		 FROM ranking ORDER BY level DESC, gold DESC LIMIT 50`,
	)
	if err != nil {
		return
	}
	defer rows.Close()
	entries := scanRanking(rows)
	s.mu.Lock()
	s.cache = entries
	s.mu.Unlock()
}

func scanRanking(rows *sql.Rows) []RankingEntry {
	var entries []RankingEntry
	for rows.Next() {
		var e RankingEntry
		if err := rows.Scan(
			&e.CharacterID, &e.UserID, &e.Name, &e.Class,
			&e.Level, &e.Gold, &e.BattlesWon,
		); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries
}

// ── InventoryStore ────────────────────────────────────────────────────────────

type InventoryStore struct {
	db   *db.DB
	mu   sync.RWMutex
	invs map[int]*items.Inventory
}

func NewInventoryStore(database *db.DB) *InventoryStore {
	return &InventoryStore{db: database, invs: make(map[int]*items.Inventory)}
}

func (s *InventoryStore) Get(charID int) *items.Inventory {
	if s.db != nil {
		var data string
		err := s.db.QueryRow(`SELECT data FROM inventories WHERE character_id = ?`, charID).Scan(&data)
		if err != nil {
			return items.NewInventory(charID)
		}
		var inv items.Inventory
		if err := json.Unmarshal([]byte(data), &inv); err != nil {
			return items.NewInventory(charID)
		}
		return &inv
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if inv, ok := s.invs[charID]; ok {
		return inv
	}
	return nil
}

func (s *InventoryStore) Set(charID int, inv *items.Inventory) {
	if s.db != nil {
		data, err := json.Marshal(inv)
		if err != nil {
			return
		}
		s.db.Exec(`INSERT OR REPLACE INTO inventories (character_id, data, updated_at) VALUES (?, ?, datetime('now'))`, charID, string(data))
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invs[charID] = inv
}

func (s *InventoryStore) Delete(charID int) {
	if s.db != nil {
		s.db.Exec(`DELETE FROM inventories WHERE character_id = ?`, charID)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.invs, charID)
}

// ── ProgressStore ─────────────────────────────────────────────────────────────

type ProgressStore struct {
	db       *db.DB
	mu       sync.RWMutex
	progress map[int]*missions.PlayerProgress
}

func NewProgressStore(database *db.DB) *ProgressStore {
	return &ProgressStore{db: database, progress: make(map[int]*missions.PlayerProgress)}
}

func (s *ProgressStore) Get(charID int) *missions.PlayerProgress {
	if s.db != nil {
		var data string
		err := s.db.QueryRow(`SELECT data FROM progress WHERE character_id = ?`, charID).Scan(&data)
		if err != nil || data == "{}" || data == "" {
			return nil
		}
		var pp missions.PlayerProgress
		if err := json.Unmarshal([]byte(data), &pp); err != nil {
			return nil
		}
		return &pp
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.progress[charID]
}

func (s *ProgressStore) Save(charID int, pp *missions.PlayerProgress) {
	if s.db != nil {
		data, err := json.Marshal(pp)
		if err != nil {
			return
		}
		s.db.Exec(`INSERT OR REPLACE INTO progress (character_id, data, updated_at) VALUES (?, ?, datetime('now'))`, charID, string(data))
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progress[charID] = pp
}

func (s *ProgressStore) Delete(charID int) {
	if s.db != nil {
		s.db.Exec(`DELETE FROM progress WHERE character_id = ?`, charID)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.progress, charID)
}

// ── BattleStore (sempre in-memory) ───────────────────────────────────────────

type BattleSession struct {
	Battle    *game.Battle
	PlayerID  int
	EnemyName string
}

type BattleStore struct {
	mu       sync.RWMutex
	sessions map[string]*BattleSession
}

func NewBattleStore() *BattleStore {
	return &BattleStore{sessions: make(map[string]*BattleSession)}
}

func (s *BattleStore) Set(id string, sess *BattleSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = sess
}

func (s *BattleStore) Get(id string) (*BattleSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("battle session %q not found", id)
	}
	return sess, nil
}

func (s *BattleStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

// ── RunnerStore (sempre in-memory) ───────────────────────────────────────────

type RunnerStore struct {
	mu      sync.RWMutex
	runners map[string]*missions.Runner
}

func NewRunnerStore() *RunnerStore {
	return &RunnerStore{runners: make(map[string]*missions.Runner)}
}

func (s *RunnerStore) Get(id string) (*missions.Runner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.runners[id]
	if !ok {
		return nil, fmt.Errorf("runner %q não encontrado", id)
	}
	return r, nil
}

func (s *RunnerStore) Set(id string, r *missions.Runner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runners[id] = r
}

func (s *RunnerStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.runners, id)
}

// CleanOldRunners remove sessões inativas há mais de maxAge.
func CleanOldRunners(runners *RunnerStore, maxAge time.Duration) {
	runners.mu.Lock()
	defer runners.mu.Unlock()
	for id, r := range runners.runners {
		if time.Since(r.UpdatedAt()) > maxAge {
			delete(runners.runners, id)
		}
	}
}
