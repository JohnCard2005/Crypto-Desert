package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto-desert/internal/api"
	"crypto-desert/internal/auth"
	"crypto-desert/internal/db"
	"crypto-desert/internal/store"
)

func main() {
	port      := envOr("PORT",       "8080")
	webDir    := envOr("WEB_DIR",    "./web")
	dbPath    := envOr("DB_PATH",    "./data/game.db")
	jwtSecret := envOr("JWT_SECRET", "crypto-desert-secret-mude-em-producao")

	// ── Banco de dados (opcional) ──────────────────────────────────────────
	// Se DB_PATH estiver vazio ou o SQLite não estiver disponível,
	// o servidor roda em modo in-memory (dados perdidos ao reiniciar).
	var database *db.DB
	if dbPath != "" {
		if err := os.MkdirAll("./data", 0755); err != nil {
			log.Printf("[main] aviso: não foi possível criar pasta data: %v", err)
		} else {
			d, err := db.New(dbPath)
			if err != nil {
				log.Printf("[main] aviso: banco SQLite indisponível (%v) — modo in-memory", err)
			} else {
				database = d
				defer database.Close()
				log.Printf("[main] banco SQLite: %s", dbPath)
			}
		}
	}

	if database == nil {
		log.Printf("[main] ⚠ Rodando em modo IN-MEMORY — dados serão perdidos ao reiniciar")
		log.Printf("[main] Para persistência: rode 'go get modernc.org/sqlite' e reinicie")
	}

	// ── Stores ────────────────────────────────────────────────────────────
	chars       := store.NewCharacterStore(database)
	battles     := store.NewBattleStore()
	runners     := store.NewRunnerStore()
	inventories := store.NewInventoryStore(database)
	progress    := store.NewProgressStore(database)
	ranking     := store.NewRankingStore(database)

	// ── Auth ──────────────────────────────────────────────────────────────
	authSvc := auth.NewService(database, jwtSecret)

	// ── Limpeza periódica de sessões antigas ──────────────────────────────
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			store.CleanOldRunners(runners, 4*time.Hour)
		}
	}()

	// ── Serviços ──────────────────────────────────────────────────────────
	cryptoSvc := api.NewCryptoService()

	// ── Handler e Rotas ───────────────────────────────────────────────────
	handler := api.NewHandler(chars, battles, runners, inventories, progress, ranking, authSvc, cryptoSvc)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, handler, webDir)

	// ── Servidor HTTP ──────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("\n  ⚔  CRYPTO DESERT — RPG 2087\n")
		fmt.Printf("  ──────────────────────────────\n")
		fmt.Printf("  http://localhost:%s\n\n", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[server] fatal: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[main] encerrando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("[main] encerrado.")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
