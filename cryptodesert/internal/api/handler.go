package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"crypto-desert/internal/characters"
	"crypto-desert/internal/enemies"
	"crypto-desert/internal/game"
	"crypto-desert/internal/items"
	"crypto-desert/internal/missions"
	"crypto-desert/internal/auth"
	"crypto-desert/internal/store"
)

// Handler agrupa todas as dependências dos handlers HTTP
type Handler struct {
	chars       *store.CharacterStore
	battles     *store.BattleStore
	runners     *store.RunnerStore
	inventories *store.InventoryStore
	progress    *store.ProgressStore
	ranking     *store.RankingStore
	auth        *auth.Service
	crypto    *CryptoService
}

func NewHandler(
	chars *store.CharacterStore,
	battles *store.BattleStore,
	runners *store.RunnerStore,
	inventories *store.InventoryStore,
	progress *store.ProgressStore,
	ranking *store.RankingStore,
	authSvc *auth.Service,
	crypto *CryptoService,
) *Handler {
	return &Handler{chars, battles, runners, inventories, progress, ranking, authSvc, crypto}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[handler] encode error: %v", err)
	}
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func pathID(r *http.Request, prefix string) (int, error) {
	raw := strings.TrimPrefix(r.URL.Path, prefix)
	// Pega só o primeiro segmento (antes de qualquer "/" adicional)
	// ex: "1/inventory" → "1"
	if idx := strings.Index(raw, "/"); idx != -1 {
		raw = raw[:idx]
	}
	raw = strings.TrimSuffix(raw, "/")
	return strconv.Atoi(raw)
}

func pathStr(r *http.Request, prefix string) string {
	return strings.TrimPrefix(r.URL.Path, prefix)
}

// ── Crypto ────────────────────────────────────────────────────────────────────

// GET /api/crypto
// Retorna as cotações atuais de todas as cryptos monitoradas.
func (h *Handler) GetCryptoPrices(w http.ResponseWriter, r *http.Request) {
	h.crypto.MaybeRefresh()
	writeJSON(w, http.StatusOK, h.crypto.GetAll())
}

// ── Classes ───────────────────────────────────────────────────────────────────

// GET /api/classes
// Retorna as definições estáticas de todas as classes jogáveis.
func (h *Handler) GetClasses(w http.ResponseWriter, r *http.Request) {
	classBases := map[string]struct {
		hp, mana, atk, str, ca, def, spd, dice int
		faction                                  string
	}{
		"warrior": {120, 0, 4, 3, 14, 2, 1, 10, "BTC"},
		"mage":    {80, 80, 6, 1, 11, 0, 2, 6, "ETH"},
		"archer":  {95, 40, 5, 2, 13, 1, 3, 8, "SOL"},
		"rogue":   {90, 50, 5, 2, 12, 1, 4, 8, "BNB"},
		"shaman":  {100, 60, 3, 3, 12, 1, 2, 8, "DOGE"},
	}

	result := make([]ClassInfoResponse, 0, len(classBases))
	for key, base := range classBases {
		faction := characters.Factions[characters.Faction(base.faction)]
		cryptoFactor := h.crypto.GetFactor(faction.Crypto)

		result = append(result, ClassInfoResponse{
			Key:             key,
			Faction:         base.faction,
			FactionName:     faction.Name,
			Lore:            faction.Lore,
			HP:              base.hp,
			Mana:            base.mana,
			AttackMod:       base.atk,
			StrengthMod:     base.str,
			CA:              base.ca,
			Defense:         base.def,
			Speed:           base.spd,
			DamageDice:      base.dice,
			CryptoFactor:    cryptoFactor,
			CryptoVariation: h.crypto.GetChange7d(faction.Crypto),
		})
	}

	// Ordem consistente
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	writeJSON(w, http.StatusOK, result)
}

// ── Characters ────────────────────────────────────────────────────────────────

// POST /api/characters
// Cria um novo personagem. Body: {"name":"...","class":"..."}
func (h *Handler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	userID, _, err := h.userFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "faça login para criar um personagem")
		return
	}
	var req CreateCharacterRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "nome é obrigatório")
		return
	}

	c, err := characters.NewCharacter(req.Name, req.Class)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Injeta variação crypto antes de salvar
	faction := characters.Factions[c.Faction]
	c.CryptoVariation = h.crypto.GetChange7d(faction.Crypto)

	// Vincula ao usuário logado
	c.UserID = userID

	if err := h.chars.Create(c); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar personagem: "+err.Error())
		return
	}

	// Cria inventário vazio para o personagem
	h.inventories.Set(c.ID, items.NewInventory(c.ID))

	writeJSON(w, http.StatusCreated, CharToResponse(c, h.crypto))
}

// GET /api/characters
// Lista todos os personagens com dados crypto ao vivo.
func (h *Handler) ListCharacters(w http.ResponseWriter, r *http.Request) {
	userID, _, err := h.userFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "faça login para ver seus personagens")
		return
	}
	list, _ := h.chars.ListByUser(userID)
	result := make([]CharacterResponse, 0, len(list))
	for _, c := range list {
		// Aplica bônus de equipamento antes de serializar
		if inv := h.inventories.Get(c.ID); inv != nil {
			b := inv.TotalBonuses()
			c.BonusAttackMod = b.AttackMod
			c.BonusStrengthMod = b.StrengthMod
			c.BonusCA = b.CA
			c.BonusDefense = b.Defense
			c.BonusSpeed = b.Speed
			c.BonusMaxHP = b.MaxHP
			c.BonusMaxMana = b.MaxMana
			c.CryptoFactorBonus = b.CryptoFactorBonus
		}
		result = append(result, CharToResponse(c, h.crypto))
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/characters/{id}
// Retorna um personagem por ID.
func (h *Handler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "/api/characters/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}
	c, err := h.chars.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, CharToResponse(c, h.crypto))
}

// DELETE /api/characters/{id}
func (h *Handler) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "/api/characters/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}
	h.chars.Delete(id)
	h.inventories.Delete(id)
	w.WriteHeader(http.StatusNoContent)
}

// ── Enemies ───────────────────────────────────────────────────────────────────

// GET /api/enemies
// Lista todos os inimigos do catálogo com fator crypto atual.
func (h *Handler) ListEnemies(w http.ResponseWriter, r *http.Request) {
	result := make([]EnemyResponse, 0, len(enemies.Catalogue))
	for _, tmpl := range enemies.Catalogue {
		e, err := enemies.Spawn(tmpl.Name)
		if err != nil {
			continue
		}
		result = append(result, EnemyToResponse(e, h.crypto))
	}
	writeJSON(w, http.StatusOK, result)
}

// ── Battle ────────────────────────────────────────────────────────────────────

// POST /api/battles
// Inicia uma batalha standalone (fora do sistema de missões).
// Body: {"character_id": 1, "enemy_name": "Especulador Novato"}
func (h *Handler) StartBattle(w http.ResponseWriter, r *http.Request) {
	var req StartBattleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	c, err := h.chars.Get(req.CharacterID)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	e, err := enemies.Spawn(req.EnemyName)
	if err != nil {
		writeError(w, http.StatusBadRequest, "inimigo não encontrado: "+req.EnemyName)
		return
	}

	// Injeta fator crypto nos combatentes
	factionC := characters.Factions[c.Faction]
	c.CryptoVariation = h.crypto.GetChange7d(factionC.Crypto)
	factionE := characters.Factions[e.Faction]
	e.CryptoVariation = h.crypto.GetChange7d(factionE.Crypto)

	// Aplica bônus de equipamento
	if inv := h.inventories.Get(c.ID); inv != nil {
		bonuses := inv.TotalBonuses()
		c.BonusAttackMod = bonuses.AttackMod
		c.BonusStrengthMod = bonuses.StrengthMod
		c.BonusCA = bonuses.CA
		c.BonusDefense = bonuses.Defense
		c.BonusSpeed = bonuses.Speed
		c.BonusMaxHP = bonuses.MaxHP
		c.BonusMaxMana = bonuses.MaxMana
		c.CryptoFactorBonus = bonuses.CryptoFactorBonus
	}

	battle := game.NewBattle([]*characters.Character{c}, []*enemies.Enemy{e})

	sessionID := fmt.Sprintf("b-%d-%d", c.ID, time.Now().UnixNano())
	h.battles.Set(sessionID, &store.BattleSession{
		Battle:    battle,
		PlayerID:  c.ID,
		EnemyName: req.EnemyName,
	})

	writeJSON(w, http.StatusCreated, h.buildBattleState(sessionID, battle, c, e, nil))
}

// GET /api/battles/{session_id}
// Retorna o estado atual de uma batalha.
func (h *Handler) GetBattle(w http.ResponseWriter, r *http.Request) {
	sessionID := pathStr(r, "/api/battles/")
	sess, err := h.battles.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}
	c, _ := h.chars.Get(sess.PlayerID)
	e, _ := enemies.Spawn(sess.EnemyName)
	writeJSON(w, http.StatusOK, h.buildBattleState(sessionID, sess.Battle, c, e, nil))
}

// POST /api/battles/{session_id}/action
// Executa uma ação do jogador e processa o turno do inimigo automaticamente.
// Body: {"action": "attack"|"defend"|"ability"|"flee"}
func (h *Handler) TakeAction(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(
		strings.TrimPrefix(r.URL.Path, "/api/battles/"),
		"/action",
	)

	sess, err := h.battles.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}

	var req TakeActionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	battle := sess.Battle
	if battle.Status != game.BattleOngoing {
		writeError(w, http.StatusConflict, "batalha já encerrada")
		return
	}

	c, err := h.chars.Get(sess.PlayerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "personagem não encontrado")
		return
	}

	// Monta a ação do jogador
	action := game.PlayerAction{Type: game.ActionType(req.Action)}
	if req.Action == "attack" || req.Action == "ability" {
		// Em 1v1 o alvo é sempre o único inimigo
		if len(battle.Enemies) > 0 {
			action.Target = battle.Enemies[0].Character
		}
	}

	// Turno do jogador
	playerResult := battle.ProcessPlayerTurn(c, action)
	allEvents := playerResult.Events

	// Se a batalha continua, processa o turno do inimigo automaticamente
	if battle.Status == game.BattleOngoing && len(battle.Enemies) > 0 {
		enemyResult := battle.ProcessEnemyTurn(battle.Enemies[0])
		allEvents = append(allEvents, enemyResult.Events...)
		battle.Queue.Advance()
	}

	// Se a batalha terminou, aplica XP e gold no personagem
	if battle.Status != game.BattleOngoing {
		result := battle.Conclude()
		if battle.Status == game.BattlePlayerWon {
			c.EarnGold(result.GoldGained)
		}
		if err := h.chars.Update(c); err != nil {
			log.Printf("[handler] update character: %v", err)
		}
		h.battles.Delete(sessionID)
	}

	enemy := battle.Enemies[0]
	writeJSON(w, http.StatusOK, h.buildBattleState(sessionID, battle, c, enemy, allEvents))
}

func (h *Handler) buildBattleState(sessionID string, b *game.Battle, c *characters.Character, e *enemies.Enemy, events []game.TurnEvent) BattleStateResponse {
	order := b.InitiativeOrder()
	initEntries := make([]InitEntry, len(order))
	for i, entry := range order {
		initEntries[i] = InitEntry{
			Name:       entry.Combatant.GetName(),
			Initiative: entry.Initiative,
			IsPlayer:   entry.Combatant.GetTeam() == 0,
			IsCurrent:  i == 0,
			Alive:      entry.Combatant.GetCharacter().IsAlive(),
		}
	}

	currentActor := "enemy"
	if cur := b.Queue.Current(); cur != nil && cur.GetTeam() == 0 {
		currentActor = "player"
	}

	return BattleStateResponse{
		SessionID:    sessionID,
		Status:       string(b.Status),
		TurnNumber:   b.TurnNumber,
		Player:       CharToResponse(c, h.crypto),
		Enemy:        EnemyToResponse(e, h.crypto),
		Initiative:   initEntries,
		CurrentActor: currentActor,
		Events:       TurnEventsToResponse(events),
	}
}

// ── Missions ──────────────────────────────────────────────────────────────────

// POST /api/missions/session
// Cria ou recupera uma sessão de missão para um personagem.
// Body: {"character_id": 1}
func (h *Handler) StartMissionSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CharacterID int `json:"character_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	c, err := h.chars.Get(req.CharacterID)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	sessionID := fmt.Sprintf("m-%d", c.ID)

	// Reaproveita sessão existente se houver
	if _, err := h.runners.Get(sessionID); err != nil {
		progress := missions.NewPlayerProgress(c.ID, c.Name)
		runner := missions.NewRunner(c, progress)
		h.runners.Set(sessionID, runner)
	}

	runner, _ := h.runners.Get(sessionID)
	writeJSON(w, http.StatusOK, map[string]any{
		"session_id": sessionID,
		"snapshot":   h.enrichSnapshot(runner.Snapshot()),
	})
}

// GET /api/missions/session/{session_id}
// Retorna o snapshot atual da sessão de missão.
func (h *Handler) GetMissionSession(w http.ResponseWriter, r *http.Request) {
	sessionID := pathStr(r, "/api/missions/session/")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}

// POST /api/missions/session/{session_id}/enter
// Entra em uma cidade. Body: {"city_id": "genesis_block"}
func (h *Handler) EnterCity(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/enter")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}

	var req struct {
		CityID string `json:"city_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	if err := runner.EnterCity(req.CityID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}

// POST /api/missions/session/{session_id}/start
// Inicia a missão da cidade atual (avança para WaveIntro).
func (h *Handler) StartMission(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/start")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}
	if err := runner.StartMission(); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}

// POST /api/missions/session/{session_id}/battle/begin
// Começa a batalha da wave atual.
func (h *Handler) BeginWaveBattle(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/battle/begin")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}

	// Injeta variação crypto no player antes da batalha
	c := runner.Player
	faction := characters.Factions[c.Faction]
	c.CryptoVariation = h.crypto.GetChange7d(faction.Crypto)

	// Aplica bônus de equipamento
	if inv := h.inventories.Get(c.ID); inv != nil {
		b := inv.TotalBonuses()
		c.BonusAttackMod = b.AttackMod
		c.BonusStrengthMod = b.StrengthMod
		c.BonusCA = b.CA
		c.BonusDefense = b.Defense
		c.BonusSpeed = b.Speed
		c.CryptoFactorBonus = b.CryptoFactorBonus
	}

	if err := runner.BeginWaveBattle(); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}

// POST /api/missions/session/{session_id}/battle/action
// Executa uma ação do jogador na batalha de missão.
// Body: {"action": "attack"|"defend"|"ability"|"flee"}
func (h *Handler) MissionAction(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/battle/action")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	// Monta a ação
	pa := game.PlayerAction{Type: game.ActionType(req.Action)}
	if (req.Action == "attack" || req.Action == "ability") && runner.Battle != nil {
		if len(runner.Battle.Enemies) > 0 {
			pa.Target = runner.Battle.Enemies[0].Character
		}
	}

	// Turno do jogador
	if _, err := runner.PlayerAct(pa); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	// Turno do inimigo — executa automaticamente após o turno do jogador
	if runner.Phase == missions.PhaseBattle && runner.Battle != nil {
		if _, err := runner.EnemyAct(); err != nil {
			log.Printf("[mission] enemy act: %v", err)
		}
	}

	// Persiste HP do personagem
	if err := h.chars.Update(runner.Player); err != nil {
		log.Printf("[mission] update player: %v", err)
	}

	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}

// POST /api/missions/session/{session_id}/confirm
// Confirma telas de transição (WaveCleared, MissionEnd, GameOver retry).
// Body: {"action": "next_wave"|"finish"|"retry"}
func (h *Handler) MissionConfirm(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/confirm")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	switch req.Action {
	case "next_wave":
		if err := runner.ConfirmWaveCleared(); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
	case "finish":
		runner.ConfirmMissionEnd()
	case "retry":
		if err := runner.RetryWave(); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("ação desconhecida: %q", req.Action))
		return
	}

	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}

// ── Inventory ─────────────────────────────────────────────────────────────────

// GET /api/characters/{id}/inventory
func (h *Handler) GetInventory(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "/api/characters/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}
	if _, err := h.chars.Get(id); err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}
	inv := h.inventories.Get(id)
	if inv == nil {
		inv = items.NewInventory(id)
		h.inventories.Set(id, inv)
	}
	writeJSON(w, http.StatusOK, inv)
}

// POST /api/characters/{id}/inventory/use
// Usa um item consumível. Body: {"item_id": "potion_small"}
func (h *Handler) UseItem(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/characters/"), "/inventory/use")
	id, err := strconv.Atoi(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	c, err := h.chars.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	inv := h.inventories.Get(id)
	if inv == nil {
		writeError(w, http.StatusNotFound, "inventário não encontrado")
		return
	}

	result, err := inv.UseItem(req.ItemID, c, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.chars.Update(c)
	h.inventories.Set(id, inv)

	writeJSON(w, http.StatusOK, map[string]any{
		"result":    result,
		"character": CharToResponse(c, h.crypto),
	})
}

// POST /api/characters/{id}/inventory/equip
// Equipa um item. Body: {"item_id": "weapon_plasma_sword"}
func (h *Handler) EquipItem(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/characters/"), "/inventory/equip")
	id, err := strconv.Atoi(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	c, err := h.chars.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	inv := h.inventories.Get(id)
	if inv == nil {
		writeError(w, http.StatusNotFound, "inventário não encontrado")
		return
	}

	result, err := inv.Equip(req.ItemID, c.Class)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.inventories.Set(id, inv)
	writeJSON(w, http.StatusOK, result)
}

// ── Shop ──────────────────────────────────────────────────────────────────────

// GET /api/shop/{city_id}?character_id=1
// Retorna o estoque da loja. Bloqueia se a cidade não foi desbloqueada pelo personagem.
func (h *Handler) GetShop(w http.ResponseWriter, r *http.Request) {
	cityID := pathStr(r, "/api/shop/")

	// Verifica se o personagem tem acesso à cidade desta loja
	charIDStr := r.URL.Query().Get("character_id")
	if charIDStr != "" {
		charID, _ := strconv.Atoi(charIDStr)
		if charID > 0 {
			if pp := h.progress.Get(charID); pp != nil {
				var city *missions.City
				for _, c := range missions.Campaign {
					if c.ID == cityID {
						city = &c
						break
					}
				}
				if city != nil && !pp.CityUnlocked(*city) {
					writeError(w, http.StatusForbidden, "cidade ainda bloqueada — complete a missão anterior primeiro")
					return
				}
			}
		}
	}

	// Descobre a facção da cidade para pegar o fator crypto correto
	cityFaction := cityFactionMap[cityID]
	var cryptoFactor float64
	if cityFaction != "" {
		faction := characters.Factions[characters.Faction(cityFaction)]
		cryptoFactor = h.crypto.GetFactor(faction.Crypto)
	} else {
		cryptoFactor = 1.0
	}

	shop := items.NewShop(cityID, cryptoFactor)
	writeJSON(w, http.StatusOK, map[string]any{
		"city_id":      cityID,
		"crypto_factor": cryptoFactor,
		"price_info":   shop.PriceInfo(),
		"listings":     shop.AvailableListings(),
	})
}

// POST /api/shop/{city_id}/buy
// Compra um item. Body: {"character_id": 1, "item_id": "potion_small", "quantity": 2}
func (h *Handler) BuyItem(w http.ResponseWriter, r *http.Request) {
	cityID := strings.TrimSuffix(pathStr(r, "/api/shop/"), "/buy")

	var req struct {
		CharacterID int    `json:"character_id"`
		ItemID      string `json:"item_id"`
		Quantity    int    `json:"quantity"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	c, err := h.chars.Get(req.CharacterID)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	inv := h.inventories.Get(req.CharacterID)
	if inv == nil {
		writeError(w, http.StatusNotFound, "inventário não encontrado")
		return
	}

	cityFaction := cityFactionMap[cityID]
	cryptoFactor := 1.0
	if cityFaction != "" {
		faction := characters.Factions[characters.Faction(cityFaction)]
		cryptoFactor = h.crypto.GetFactor(faction.Crypto)
	}

	shop := items.NewShop(cityID, cryptoFactor)
	result, err := shop.Buy(req.ItemID, req.Quantity, &c.Gold, inv, c.Class)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.chars.Update(c)
	h.inventories.Set(req.CharacterID, inv)

	writeJSON(w, http.StatusOK, map[string]any{
		"result":    result,
		"gold_left": c.Gold,
	})
}

// POST /api/shop/{city_id}/sell
// Vende um item. Body: {"character_id": 1, "item_id": "potion_small", "quantity": 1}
func (h *Handler) SellItem(w http.ResponseWriter, r *http.Request) {
	cityID := strings.TrimSuffix(pathStr(r, "/api/shop/"), "/sell")

	var req struct {
		CharacterID int    `json:"character_id"`
		ItemID      string `json:"item_id"`
		Quantity    int    `json:"quantity"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	c, err := h.chars.Get(req.CharacterID)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	inv := h.inventories.Get(req.CharacterID)
	if inv == nil {
		writeError(w, http.StatusNotFound, "inventário não encontrado")
		return
	}

	cityFaction := cityFactionMap[cityID]
	cryptoFactor := 1.0
	if cityFaction != "" {
		faction := characters.Factions[characters.Faction(cityFaction)]
		cryptoFactor = h.crypto.GetFactor(faction.Crypto)
	}

	shop := items.NewShop(cityID, cryptoFactor)
	result, err := shop.Sell(req.ItemID, req.Quantity, &c.Gold, inv)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.chars.Update(c)
	h.inventories.Set(req.CharacterID, inv)

	writeJSON(w, http.StatusOK, map[string]any{
		"result":    result,
		"gold_left": c.Gold,
	})
}

// ── Campfire ──────────────────────────────────────────────────────────────────

// GET /api/campfire/{city_id}?character_id=1
// Retorna os serviços disponíveis no campfire com preços calculados.
func (h *Handler) GetCampfire(w http.ResponseWriter, r *http.Request) {
	cityID := pathStr(r, "/api/campfire/")

	charIDStr := r.URL.Query().Get("character_id")
	charID, err := strconv.Atoi(charIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "character_id obrigatório")
		return
	}

	c, err := h.chars.Get(charID)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	cityFaction := cityFactionMap[cityID]
	cryptoFactor := 1.0
	cityName := cityID
	if cityFaction != "" {
		faction := characters.Factions[characters.Faction(cityFaction)]
		cryptoFactor = h.crypto.GetFactor(faction.Crypto)
	}

	cf := items.NewCampfire(cityID, cityName, cityFaction, c.Level, cryptoFactor)
	writeJSON(w, http.StatusOK, map[string]any{
		"city_id":      cityID,
		"crypto_factor": cryptoFactor,
		"price_info":   cf.PriceInfo(),
		"offers":       cf.Offers,
		"character":    CharToResponse(c, h.crypto),
	})
}

// POST /api/campfire/{city_id}/rest
// Usa um serviço do campfire. Body: {"character_id": 1, "service": "rest_full"}
func (h *Handler) UseCampfire(w http.ResponseWriter, r *http.Request) {
	cityID := strings.TrimSuffix(pathStr(r, "/api/campfire/"), "/rest")

	var req struct {
		CharacterID int    `json:"character_id"`
		Service     string `json:"service"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	c, err := h.chars.Get(req.CharacterID)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	cityFaction := cityFactionMap[cityID]
	cryptoFactor := 1.0
	if cityFaction != "" {
		faction := characters.Factions[characters.Faction(cityFaction)]
		cryptoFactor = h.crypto.GetFactor(faction.Crypto)
	}

	cf := items.NewCampfire(cityID, cityID, cityFaction, c.Level, cryptoFactor)
	result, err := cf.UseService(items.CampfireService(req.Service), &c.Gold, c)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.chars.Update(c)

	writeJSON(w, http.StatusOK, map[string]any{
		"result":    result,
		"character": CharToResponse(c, h.crypto),
	})
}


// POST /api/characters/{id}/inventory/unequip
// Desequipa um item de um slot. Body: {"slot": "weapon"|"armor"|"accessory"}
func (h *Handler) UnequipItem(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/characters/"), "/inventory/unequip")
	id, err := strconv.Atoi(strings.Split(raw, "/")[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	c, err := h.chars.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "personagem não encontrado")
		return
	}

	var req struct {
		Slot string `json:"slot"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	inv := h.inventories.Get(id)
	if inv == nil {
		writeError(w, http.StatusNotFound, "inventário não encontrado")
		return
	}

	slotMap := map[string]items.EquipSlot{
		"weapon":    items.SlotWeapon,
		"armor":     items.SlotArmor,
		"accessory": items.SlotAccessory,
	}
	slot, ok := slotMap[req.Slot]
	if !ok {
		writeError(w, http.StatusBadRequest, "slot inválido: use weapon, armor ou accessory")
		return
	}

	if err := inv.Unequip(slot); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Remove bônus de equipamento do personagem
	bonuses := inv.TotalBonuses()
	c.BonusAttackMod = bonuses.AttackMod
	c.BonusStrengthMod = bonuses.StrengthMod
	c.BonusCA = bonuses.CA
	c.BonusDefense = bonuses.Defense
	c.BonusSpeed = bonuses.Speed
	c.BonusMaxHP = bonuses.MaxHP
	c.BonusMaxMana = bonuses.MaxMana
	c.CryptoFactorBonus = bonuses.CryptoFactorBonus
	h.chars.Update(c)
	h.inventories.Set(id, inv)

	writeJSON(w, http.StatusOK, map[string]any{
		"slot":      req.Slot,
		"inventory": inv,
		"character": CharToResponse(c, h.crypto),
	})
}

// GET /api/items
// Retorna o catálogo completo de itens para o frontend exibir nomes/ícones.
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, items.AllItems())
}


// POST /api/missions/session/{session_id}/replay
// Reinicia uma missão já concluída. Body: {} (usa cidade ativa)
func (h *Handler) ReplayMission(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/replay")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}
	if err := runner.ReplayMission(); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}


// POST /api/missions/session/{session_id}/ng
// Inicia o New Game Plus após completar a campanha.
func (h *Handler) StartNG(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/ng")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}
	if err := runner.StartNG(); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}

// POST /api/missions/session/{session_id}/use-item
// Usa um item consumível DURANTE a batalha.
// Body: {"item_id": "potion_small"}
func (h *Handler) UseItemInBattle(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/use-item")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}
	if runner.Phase != missions.PhaseBattle {
		writeError(w, http.StatusConflict, "só é possível usar itens durante a batalha")
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	c, err := h.chars.Get(runner.Player.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "personagem não encontrado")
		return
	}

	inv := h.inventories.Get(c.ID)
	if inv == nil {
		writeError(w, http.StatusNotFound, "inventário não encontrado")
		return
	}

	// Usa o item (deve ser usável em batalha)
	result, err := inv.UseItem(req.ItemID, c, true)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.chars.Update(c)
	h.inventories.Set(c.ID, inv)

	// Atualiza runner com os novos vitals do jogador
	runner.Player.HP = c.HP
	runner.Player.Mana = c.Mana

	// Adiciona evento ao log de batalha
	runner.LastEvents = append(runner.LastEvents, missions.BattleEvent{
		Actor:   runner.Player.Name,
		Message: fmt.Sprintf("🧪 %s usou %s — %s", runner.Player.Name, req.ItemID, result.Message),
	})

	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}


// POST /api/missions/session/{session_id}/start-wave
// Inicia uma wave específica. Body: {"wave_id": "genesis_w1"}
// Usado para replay — o jogador escolhe qual wave jogar diretamente.
func (h *Handler) StartWave(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSuffix(pathStr(r, "/api/missions/session/"), "/start-wave")
	runner, err := h.runners.Get(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "sessão não encontrada")
		return
	}

	var req struct {
		WaveID string `json:"wave_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	if req.WaveID == "" {
		writeError(w, http.StatusBadRequest, "wave_id é obrigatório")
		return
	}

	if err := runner.StartWave(req.WaveID); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, h.enrichSnapshot(runner.Snapshot()))
}


// ── Auth ──────────────────────────────────────────────────────────────────────

// POST /api/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	user, err := h.auth.Register(req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user": user,
		"message": "Conta criada com sucesso! Faça login para continuar.",
	})
}

// POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	user, token, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	// Token no cookie HTTP-only (mais seguro que localStorage)
	http.SetCookie(w, &http.Cookie{
		Name:     "cd_token",
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 dias
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"user":    user,
		"token":   token, // também no body para o frontend guardar
		"message": "Login realizado com sucesso!",
	})
}

// POST /api/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "cd_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Logout realizado."})
}

// GET /api/auth/me — retorna o usuário logado
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, username, err := h.userFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":       userID,
		"username": username,
	})
}

// userFromRequest extrai o usuário do token (cookie ou header Authorization).
func (h *Handler) userFromRequest(r *http.Request) (int, string, error) {
	// Tenta cookie primeiro
	if cookie, err := r.Cookie("cd_token"); err == nil {
		return h.auth.UserFromRequest(cookie.Value)
	}
	// Fallback: header Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return h.auth.UserFromRequest(authHeader[7:])
	}
	return 0, "", fmt.Errorf("não autenticado")
}

// ── Ranking ───────────────────────────────────────────────────────────────────

// GET /api/ranking?by=level&limit=10
func (h *Handler) GetRanking(w http.ResponseWriter, r *http.Request) {
	by    := r.URL.Query().Get("by")
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscan(l, &limit)
	}
	if limit > 50 { limit = 50 }

	var entries []store.RankingEntry
	switch by {
	case "gold":
		entries = h.ranking.TopByGold(limit)
	case "battles":
		entries = h.ranking.TopByBattles(limit)
	default:
		entries = h.ranking.TopByLevel(limit)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"by":      by,
		"entries": entries,
	})
}

// cityFactionMap mapeia city_id → símbolo de facção
var cityFactionMap = map[string]string{
	"genesis_block":  "BTC",
	"ether_citadel":  "ETH",
	"sol_dunes":      "SOL",
	"bnb_quarter":    "BNB",
	"doge_wasteland": "DOGE",
}

// enrichSnapshot injects live crypto_factor into the player snapshot,
// since the runner doesn't have access to the CryptoService.
func (h *Handler) enrichSnapshot(snap missions.RunnerSnapshot) missions.RunnerSnapshot {
	if snap.Player.ID == 0 {
		return snap
	}
	// Find the character to get their crypto ID
	c, err := h.chars.Get(snap.Player.ID)
	if err != nil {
		return snap
	}
	snap.Player.CryptoFactor = h.crypto.GetFactor(c.CryptoID)
	snap.Player.CryptoVariation = h.crypto.GetChange7d(c.CryptoID)
	return snap
}
