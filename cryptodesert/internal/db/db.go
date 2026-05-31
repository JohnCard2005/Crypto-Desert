package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// DB é o banco de dados da aplicação.
// Usa SQLite via modernc.org/sqlite (sem CGO).
type DB struct {
	*sql.DB
}

// New abre (ou cria) o banco de dados SQLite no caminho indicado.
// Passa ":memory:" para banco em memória (testes).
func New(path string) (*DB, error) {
	sqldb, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}

	// SQLite funciona melhor com uma conexão por vez para escritas
	sqldb.SetMaxOpenConns(1)

	db := &DB{sqldb}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrar banco: %w", err)
	}

	log.Printf("[db] banco inicializado em %s", path)
	return db, nil
}

// migrate cria todas as tabelas necessárias se não existirem.
func (db *DB) migrate() error {
	schema := `
	-- Usuários (sistema de login)
	CREATE TABLE IF NOT EXISTS users (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		username   TEXT    NOT NULL UNIQUE,
		password   TEXT    NOT NULL,  -- bcrypt hash
		created_at TEXT    NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
	);

	-- Personagens (vinculados a um usuário)
	CREATE TABLE IF NOT EXISTS characters (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name       TEXT    NOT NULL,
		class      TEXT    NOT NULL,
		level      INTEGER NOT NULL DEFAULT 1,
		gold       INTEGER NOT NULL DEFAULT 0,
		data       TEXT    NOT NULL,  -- Character completo em JSON
		created_at TEXT    NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT    NOT NULL DEFAULT (datetime('now'))
	);

	-- Índice para listar personagens por usuário rapidamente
	CREATE INDEX IF NOT EXISTS idx_characters_user ON characters(user_id);

	-- Inventários (um por personagem)
	CREATE TABLE IF NOT EXISTS inventories (
		character_id INTEGER PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
		data         TEXT NOT NULL DEFAULT '{}',
		updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
	);

	-- Progresso de missões (um por personagem)
	CREATE TABLE IF NOT EXISTS progress (
		character_id INTEGER PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
		data         TEXT NOT NULL DEFAULT '{}',
		updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
	);

	-- Ranking (view calculada — atualizada a cada mudança de nível/gold)
	CREATE TABLE IF NOT EXISTS ranking (
		character_id INTEGER PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
		user_id      INTEGER NOT NULL,
		name         TEXT    NOT NULL,
		class        TEXT    NOT NULL,
		level        INTEGER NOT NULL DEFAULT 1,
		gold         INTEGER NOT NULL DEFAULT 0,
		battles_won  INTEGER NOT NULL DEFAULT 0,
		updated_at   TEXT    NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_ranking_level ON ranking(level DESC);
	CREATE INDEX IF NOT EXISTS idx_ranking_gold  ON ranking(gold  DESC);

	-- WAL mode para melhor performance de escrita concorrente
	PRAGMA journal_mode=WAL;
	PRAGMA foreign_keys=ON;
	`

	_, err := db.Exec(schema)
	return err
}
