package api

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// RegisterRoutes monta todas as rotas da API no mux fornecido.
// Todas as rotas ficam sob o prefixo /api/.
// O frontend em /web/ é servido como arquivos estáticos.
func RegisterRoutes(mux *http.ServeMux, h *Handler, webDir string) {
	// ── Frontend estático ────────────────────────────────────────────────────
	// Serve index.html e assets de web/
	fs := http.FileServer(http.Dir(webDir))
	mux.Handle("/", fs)

	// ── Auth ────────────────────────────────────────────────────────────────
	mux.HandleFunc("/api/auth/register", chain(h.Register, onlyPOST))
	mux.HandleFunc("/api/auth/login",    chain(h.Login,    onlyPOST))
	mux.HandleFunc("/api/auth/logout",   chain(h.Logout,   onlyPOST))
	mux.HandleFunc("/api/auth/me",       chain(h.Me,       onlyGET))

	// ── Ranking ──────────────────────────────────────────────────────────────
	// GET /api/ranking?by=level|gold|battles&limit=10
	mux.HandleFunc("/api/ranking", chain(h.GetRanking, onlyGET))

	// ── Crypto ───────────────────────────────────────────────────────────────
	// GET /api/crypto
	mux.HandleFunc("/api/crypto", chain(h.GetCryptoPrices, onlyGET))

	// ── Classes ──────────────────────────────────────────────────────────────
	// GET /api/classes
	mux.HandleFunc("/api/classes", chain(h.GetClasses, onlyGET))

	// ── Characters ───────────────────────────────────────────────────────────
	// GET    /api/characters          → lista todos
	// POST   /api/characters          → cria novo
	// GET    /api/characters/{id}     → busca por ID
	// DELETE /api/characters/{id}     → remove
	mux.HandleFunc("/api/characters", chain(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListCharacters(w, r)
		case http.MethodPost:
			h.CreateCharacter(w, r)
		default:
			methodNotAllowed(w)
		}
	}))

	mux.HandleFunc("/api/characters/", chain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case strings.HasSuffix(path, "/inventory/unequip"):
			onlyPOST(h.UnequipItem)(w, r)
		case strings.HasSuffix(path, "/inventory/use"):
			onlyPOST(h.UseItem)(w, r)
		case strings.HasSuffix(path, "/inventory/equip"):
			onlyPOST(h.EquipItem)(w, r)
		case strings.HasSuffix(path, "/inventory"):
			onlyGET(h.GetInventory)(w, r)
		default:
			// /api/characters/{id}
			switch r.Method {
			case http.MethodGet:
				h.GetCharacter(w, r)
			case http.MethodDelete:
				h.DeleteCharacter(w, r)
			default:
				methodNotAllowed(w)
			}
		}
	}))

	// ── Items ───────────────────────────────────────────────────────────────
	// GET /api/items
	mux.HandleFunc("/api/items", chain(h.ListItems, onlyGET))

	// ── Enemies ──────────────────────────────────────────────────────────────
	// GET /api/enemies
	mux.HandleFunc("/api/enemies", chain(h.ListEnemies, onlyGET))

	// ── Battles (standalone) ─────────────────────────────────────────────────
	// POST /api/battles                      → inicia batalha
	// GET  /api/battles/{session_id}          → estado atual
	// POST /api/battles/{session_id}/action   → executa ação
	mux.HandleFunc("/api/battles", chain(h.StartBattle, onlyPOST))

	mux.HandleFunc("/api/battles/", chain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/action"):
			onlyPOST(h.TakeAction)(w, r)
		default:
			onlyGET(h.GetBattle)(w, r)
		}
	}))

	// ── Missions ─────────────────────────────────────────────────────────────
	// POST /api/missions/session                         → cria/recupera sessão
	// GET  /api/missions/session/{id}                    → snapshot
	// POST /api/missions/session/{id}/enter              → entra na cidade
	// POST /api/missions/session/{id}/start              → inicia missão
	// POST /api/missions/session/{id}/battle/begin       → começa batalha
	// POST /api/missions/session/{id}/battle/action      → ação do jogador
	// POST /api/missions/session/{id}/confirm            → confirma transição
	mux.HandleFunc("/api/missions/session", chain(h.StartMissionSession, onlyPOST))

	mux.HandleFunc("/api/missions/session/", chain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/enter"):
			onlyPOST(h.EnterCity)(w, r)
		case strings.HasSuffix(path, "/start"):
			onlyPOST(h.StartMission)(w, r)
		case strings.HasSuffix(path, "/battle/begin"):
			onlyPOST(h.BeginWaveBattle)(w, r)
		case strings.HasSuffix(path, "/battle/action"):
			onlyPOST(h.MissionAction)(w, r)
		case strings.HasSuffix(path, "/confirm"):
			onlyPOST(h.MissionConfirm)(w, r)
		case strings.HasSuffix(path, "/replay"):
			onlyPOST(h.ReplayMission)(w, r)
		case strings.HasSuffix(path, "/start-wave"):
			onlyPOST(h.StartWave)(w, r)
		case strings.HasSuffix(path, "/ng"):
			onlyPOST(h.StartNG)(w, r)
		case strings.HasSuffix(path, "/use-item"):
			onlyPOST(h.UseItemInBattle)(w, r)
		default:
			onlyGET(h.GetMissionSession)(w, r)
		}
	}))

	// ── Shop ─────────────────────────────────────────────────────────────────
	// GET  /api/shop/{city_id}       → lista itens com preços
	// POST /api/shop/{city_id}/buy   → compra
	// POST /api/shop/{city_id}/sell  → vende
	mux.HandleFunc("/api/shop/", chain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/buy"):
			onlyPOST(h.BuyItem)(w, r)
		case strings.HasSuffix(path, "/sell"):
			onlyPOST(h.SellItem)(w, r)
		default:
			onlyGET(h.GetShop)(w, r)
		}
	}))

	// ── Campfire ─────────────────────────────────────────────────────────────
	// GET  /api/campfire/{city_id}?character_id=N → lista serviços com preços
	// POST /api/campfire/{city_id}/rest            → usa serviço
	mux.HandleFunc("/api/campfire/", chain(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/rest"):
			onlyPOST(h.UseCampfire)(w, r)
		default:
			onlyGET(h.GetCampfire)(w, r)
		}
	}))
}

// ── Middleware ────────────────────────────────────────────────────────────────

type middleware func(http.HandlerFunc) http.HandlerFunc

// chain aplica middlewares em ordem: cors → logger → handler
func chain(h http.HandlerFunc, middlewares ...middleware) http.HandlerFunc {
	wrapped := corsMiddleware(loggerMiddleware(h))
	for _, m := range middlewares {
		wrapped = m(wrapped)
	}
	return wrapped
}

// corsMiddleware adiciona os headers de CORS necessários para o frontend.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// loggerMiddleware registra cada request com método, path e duração.
func loggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next(rw, r)
		log.Printf("[%s] %s %s → %d (%s)",
			r.Method, r.URL.Path,
			r.URL.RawQuery,
			rw.status,
			time.Since(start).Round(time.Millisecond),
		)
	}
}

// onlyGET rejeita qualquer método que não seja GET.
func onlyGET(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		next(w, r)
	}
}

// onlyPOST rejeita qualquer método que não seja POST.
func onlyPOST(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		next(w, r)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

// responseWriter intercepta o status code para o logger.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
