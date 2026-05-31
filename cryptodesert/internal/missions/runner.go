package missions

import (
	"fmt"
	"math"
	"time"

	"crypto-desert/internal/characters"
	"crypto-desert/internal/enemies"
	"crypto-desert/internal/game"
)

// ── Runner State ──────────────────────────────────────────────────────────────

// RunnerPhase descreve em que passo do fluxo o jogador está
type RunnerPhase string

const (
	PhaseWorldMap    RunnerPhase = "world_map"    // tela do mapa — escolhe cidade
	PhaseCityScreen  RunnerPhase = "city_screen"  // lore + problema da cidade
	PhaseWaveIntro   RunnerPhase = "wave_intro"   // texto de introdução da wave
	PhaseBattle      RunnerPhase = "battle"       // batalha ativa (1 inimigo por vez)
	PhaseWaveCleared RunnerPhase = "wave_cleared" // wave concluída, pausa antes da próxima
	PhaseMissionEnd  RunnerPhase = "mission_end"  // missão concluída (boss morto)
	PhaseCityCleared RunnerPhase = "city_cleared" // tela de vitória da cidade
	PhaseGameOver    RunnerPhase = "game_over"    // personagem morto
)

// EnemyQueueEntry representa um inimigo na fila da wave atual
type EnemyQueueEntry struct {
	Enemy       *enemies.Enemy
	TemplateName string
	Defeated    bool
}

// Runner é a máquina de estado central do sistema de missões.
// Ele mantém: qual cidade, qual wave, qual inimigo está sendo enfrentado agora.
// O combate propriamente dito é delegado ao game.Battle existente.
type Runner struct {
	// Quem está jogando
	Player   *characters.Character
	Progress *PlayerProgress

	// Estado de navegação
	Phase      RunnerPhase
	ActiveCity *City
	ActiveWave *Wave

	// Fila de inimigos da wave atual (Pokémon-style: um por vez)
	EnemyQueue   []*EnemyQueueEntry
	CurrentEnemy int // índice no EnemyQueue

	// Batalha em andamento
	Battle *game.Battle

	// Acumuladores da wave/missão
	WaveXP   int
	WaveGold int

	// Mensagem contextual para a UI exibir
	Message string

	// Eventos do último turno para o frontend
	LastEvents []BattleEvent

	// Log de eventos desta sessão
	Log []RunnerEvent
}

// RunnerEvent é um registro auditável de tudo que aconteceu
type RunnerEvent struct {
	Phase   RunnerPhase
	Message string
	At      time.Time
}

// ── Constructor ───────────────────────────────────────────────────────────────

// NewRunner cria um Runner para um personagem com seu progresso salvo.
func NewRunner(player *characters.Character, progress *PlayerProgress) *Runner {
	return &Runner{
		Player:   player,
		Progress: progress,
		Phase:    PhaseWorldMap,
	}
}

// ── Navegação ─────────────────────────────────────────────────────────────────



// StartNG inicia o New Game Plus, aumentando a dificuldade global.
func (r *Runner) StartNG() error {
	if err := r.Progress.StartNewGame(); err != nil {
		return err
	}
	// Reseta estado de batalha
	r.ActiveCity = nil
	r.ActiveWave = nil
	r.EnemyQueue = nil
	r.Battle = nil
	r.Phase = PhaseWorldMap
	r.Message = fmt.Sprintf("🔥 %s ativado! Inimigos mais fortes, recompensas maiores.",
		r.Progress.CurrentDifficulty.Label())
	return nil
}

// ReplayMission reinicia uma missão já concluída (ou parcialmente feita) para replay.
// Mantém o personagem no mesmo estado (HP, gold, items) — só reseta o progresso da missão.
func (r *Runner) ReplayMission() error {
	if r.ActiveCity == nil {
		return fmt.Errorf("nenhuma cidade ativa")
	}
	if err := r.Progress.ReplayMission(*r.ActiveCity); err != nil {
		return err
	}
	// Começa do início — primeira wave
	wave, _, err := r.Progress.ActiveWaveFor(*r.ActiveCity)
	if err != nil {
		return fmt.Errorf("falha ao obter primeira wave: %w", err)
	}
	r.ActiveWave = wave
	r.Phase = PhaseWaveIntro
	r.Message = wave.Intro
	r.EnemyQueue = nil
	r.Battle = nil
	r.WaveXP = 0
	r.WaveGold = 0
	r.logEvent(PhaseWaveIntro, fmt.Sprintf("Replay: %s — Wave: %s", r.ActiveCity.Name, wave.Title))
	return nil
}


// StartWave inicia uma wave específica da cidade ativa pelo ID.
// Usado no modo replay para o jogador escolher qual wave jogar.
func (r *Runner) StartWave(waveID string) error {
	if r.ActiveCity == nil {
		return fmt.Errorf("nenhuma cidade ativa")
	}
	var targetWave *Wave
	for i := range r.ActiveCity.Mission.Waves {
		w := &r.ActiveCity.Mission.Waves[i]
		if w.ID == waveID {
			targetWave = w
			break
		}
	}
	if targetWave == nil {
		return fmt.Errorf("wave %q não encontrada", waveID)
	}
	mp, err := r.Progress.MissionState(r.ActiveCity.Mission.ID)
	if err != nil {
		return err
	}
	wp, ok := mp.Waves[waveID]
	if !ok || wp.Status == WaveStatusLocked {
		return fmt.Errorf("wave ainda bloqueada")
	}
	r.ActiveWave = targetWave
	r.Phase = PhaseWaveIntro
	r.Message = targetWave.Intro
	r.EnemyQueue = nil
	r.Battle = nil
	r.WaveXP = 0
	r.WaveGold = 0
	return nil
}

// EnterCity move o runner para a tela de uma cidade.
// Retorna erro se a cidade estiver bloqueada.
func (r *Runner) EnterCity(cityID string) error {
	city, ok := CityByID(cityID)
	if !ok {
		return fmt.Errorf("cidade %q não encontrada", cityID)
	}
	if !r.Progress.CityUnlocked(city) {
		prereq := city.UnlockedBy
		return fmt.Errorf("cidade %q bloqueada — conclua %q primeiro", city.Name, prereq)
	}

	r.ActiveCity = &city
	r.Phase = PhaseCityScreen
	r.Message = city.Lore
	r.logEvent(PhaseCityScreen, fmt.Sprintf("Entrou em %s", city.Name))
	return nil
}

// StartMission avança da tela da cidade para a primeira wave disponível.
func (r *Runner) StartMission() error {
	if r.ActiveCity == nil {
		return fmt.Errorf("nenhuma cidade ativa")
	}

	wave, _, err := r.Progress.ActiveWaveFor(*r.ActiveCity)
	if err != nil {
		// Missão já concluída — pergunta se quer replay
		return fmt.Errorf("missão já concluída. Use ReplayMission para jogar novamente")
	}

	r.ActiveWave = wave
	r.Phase = PhaseWaveIntro
	r.Message = wave.Intro
	r.logEvent(PhaseWaveIntro, fmt.Sprintf("Wave: %s", wave.Title))
	return nil
}

// BeginWaveBattle spawna os inimigos da wave e começa a primeira batalha.
// Chamado após o jogador confirmar o WaveIntro.
func (r *Runner) BeginWaveBattle() error {
	if r.ActiveWave == nil {
		return fmt.Errorf("nenhuma wave ativa")
	}

	diff := r.Progress.CurrentDifficulty
	r.Progress.RecordWaveAttempt(r.ActiveCity.Mission.ID, r.ActiveWave.ID)

	// Spawna todos os inimigos da wave com scaling de dificuldade
	queue := make([]*EnemyQueueEntry, 0, len(r.ActiveWave.EnemyNames))
	for _, name := range r.ActiveWave.EnemyNames {
		e, err := spawnScaled(name, diff)
		if err != nil {
			return fmt.Errorf("falha ao spawnar %q: %w", name, err)
		}
		queue = append(queue, &EnemyQueueEntry{
			Enemy:        e,
			TemplateName: name,
		})
	}

	r.EnemyQueue = queue
	r.CurrentEnemy = 0
	r.WaveXP = 0
	r.WaveGold = 0

	return r.startNextBattle()
}

// startNextBattle inicia a batalha contra o inimigo atual da fila.
func (r *Runner) startNextBattle() error {
	if r.CurrentEnemy >= len(r.EnemyQueue) {
		return r.resolveWaveCleared()
	}

	entry := r.EnemyQueue[r.CurrentEnemy]
	r.Battle = game.NewBattle(
		[]*characters.Character{r.Player},
		[]*enemies.Enemy{entry.Enemy},
	)
	r.Phase = PhaseBattle

	r.logEvent(PhaseBattle, fmt.Sprintf(
		"Batalha: %s vs %s", r.Player.Name, entry.Enemy.Name,
	))
	return nil
}

// ── Ações do jogador durante batalha ─────────────────────────────────────────

// PlayerAct executa uma ação do jogador no turno atual.
// Retorna o TurnResult para a UI renderizar, e avança o estado.
func (r *Runner) PlayerAct(action game.PlayerAction) (*game.TurnResult, error) {
	if r.Phase != PhaseBattle {
		return nil, fmt.Errorf("não está em batalha (fase atual: %s)", r.Phase)
	}
	if r.Battle == nil {
		return nil, fmt.Errorf("sem batalha ativa")
	}

	// Limpa eventos do turno anterior
	r.LastEvents = nil

	result := r.Battle.ProcessPlayerTurn(r.Player, action)

	// Armazena eventos do turno do jogador
	for _, e := range result.Events {
		r.LastEvents = append(r.LastEvents, BattleEvent{
			Actor: e.Actor, Message: e.Message, Damage: e.Damage, IsError: e.IsError,
		})
	}

	// Avança a fila para o próximo combatente (inimigo)
	if r.Battle != nil && r.Battle.Status == game.BattleOngoing {
		r.Battle.Queue.Advance()
	}

	// Verifica se a batalha terminou
	if err := r.handleBattleEnd(result.BattleState); err != nil {
		return &result, err
	}

	return &result, nil
}

// EnemyAct executa o turno da IA automaticamente.
// Deve ser chamado quando o turno atual pertence ao inimigo.
func (r *Runner) EnemyAct() (*game.TurnResult, error) {
	if r.Phase != PhaseBattle || r.Battle == nil {
		return nil, fmt.Errorf("não está em batalha")
	}

	entry := r.EnemyQueue[r.CurrentEnemy]
	result := r.Battle.ProcessEnemyTurn(entry.Enemy)

	// Armazena eventos do turno do inimigo (acumula com os do jogador)
	for _, e := range result.Events {
		r.LastEvents = append(r.LastEvents, BattleEvent{
			Actor: e.Actor, Message: e.Message, Damage: e.Damage, IsError: e.IsError,
		})
	}

	// Avança a fila de volta para o jogador
	if r.Battle != nil && r.Battle.Status == game.BattleOngoing {
		r.Battle.Queue.Advance()
	}

	if err := r.handleBattleEnd(result.BattleState); err != nil {
		return &result, err
	}

	return &result, nil
}

// handleBattleEnd reage ao resultado de uma batalha individual
func (r *Runner) handleBattleEnd(status game.BattleStatus) error {
	switch status {
	case game.BattleOngoing:
		return nil // continua

	case game.BattlePlayerWon:
		return r.onEnemyDefeated()

	case game.BattleEnemyWon:
		return r.onPlayerDefeated()

	case game.BattlePlayerFled:
		// Fuga reseta para o inicio da wave
		r.Phase = PhaseCityScreen
		r.ActiveWave = nil
		r.EnemyQueue = nil
		r.Battle = nil
		r.Message = fmt.Sprintf("%s fugiu da batalha. A wave foi reiniciada.", r.Player.Name)
		r.logEvent(PhaseCityScreen, "Jogador fugiu — wave reiniciada")
		return nil
	}

	return nil
}

// onEnemyDefeated processa a morte de um inimigo
func (r *Runner) onEnemyDefeated() error {
	entry := r.EnemyQueue[r.CurrentEnemy]
	entry.Defeated = true

	// Coleta recompensas
	result := r.Battle.Conclude()
	r.WaveXP += result.XPGained
	r.WaveGold += result.GoldGained
	r.Player.EarnGold(result.GoldGained)

	// Registra no histórico
	r.Progress.AddBattleRecord(BattleRecord{
		CityID:     r.ActiveCity.ID,
		WaveID:     r.ActiveWave.ID,
		EnemyName:  entry.Enemy.Name,
		Won:        true,
		TurnCount:  result.TurnCount,
		XPGained:   result.XPGained,
		GoldGained: result.GoldGained,
		At:         time.Now(),
	})

	r.logEvent(PhaseBattle, fmt.Sprintf(
		"%s derrotado (+%dxp +%g)", entry.Enemy.Name, result.XPGained, float64(result.GoldGained),
	))

	// Próximo inimigo da fila
	r.CurrentEnemy++
	if r.CurrentEnemy < len(r.EnemyQueue) {
		r.Message = fmt.Sprintf(
			"%s foi derrotado! Próximo inimigo: %s",
			entry.Enemy.Name,
			r.EnemyQueue[r.CurrentEnemy].Enemy.Name,
		)
		return r.startNextBattle()
	}

	// Todos os inimigos da wave foram derrotados
	return r.resolveWaveCleared()
}

// onPlayerDefeated processa a morte do jogador
func (r *Runner) onPlayerDefeated() error {
	r.Phase = PhaseGameOver

	r.Progress.AddBattleRecord(BattleRecord{
		CityID:    r.ActiveCity.ID,
		WaveID:    r.ActiveWave.ID,
		EnemyName: r.EnemyQueue[r.CurrentEnemy].Enemy.Name,
		Won:       false,
		At:        time.Now(),
	})

	r.Message = fmt.Sprintf(
		"%s foi eliminado por %s. O deserto digital não perdoa.",
		r.Player.Name,
		r.EnemyQueue[r.CurrentEnemy].Enemy.Name,
	)

	r.logEvent(PhaseGameOver, r.Message)

	// Restaura o jogador com HP mínimo para não ficar preso
	r.Player.HP = 1
	r.Player.Alive = true
	r.Battle = nil

	return nil
}

// resolveWaveCleared processa a conclusão de uma wave inteira
func (r *Runner) resolveWaveCleared() error {
	r.Battle = nil
	r.EnemyQueue = nil // limpa fila para não mostrar inimigos antigos no próximo intro

	// Aplica XP ao personagem — aqui é onde o level up acontece de verdade
	if r.WaveXP > 0 {
		r.Player.GainXP(r.WaveXP)
	}

	// Registra o clear no progresso
	r.Progress.RecordWaveCleared(
		*r.ActiveCity,
		r.ActiveWave.ID,
		r.WaveXP,
		r.WaveGold,
	)

	isBossWave := r.ActiveWave.IsBossWave

	r.logEvent(PhaseWaveCleared, fmt.Sprintf(
		"Wave %s concluída (+%dxp +%dg)", r.ActiveWave.ID, r.WaveXP, r.WaveGold,
	))

	if isBossWave {
		return r.resolveMissionCleared()
	}

	// Verifica se há próxima wave
	r.Phase = PhaseWaveCleared
	r.Message = fmt.Sprintf(
		"Wave concluída! +%d XP, +%d Gold. Prepare-se para o próximo encontro.",
		r.WaveXP, r.WaveGold,
	)

	// Avança para a próxima wave
	wave, _, err := r.Progress.ActiveWaveFor(*r.ActiveCity)
	if err != nil {
		// Não deveria acontecer, mas se acontecer vai para missão concluída
		return r.resolveMissionCleared()
	}
	r.ActiveWave = wave
	return nil
}

// resolveMissionCleared processa a conclusão da missão (boss derrotado)
func (r *Runner) resolveMissionCleared() error {
	mp, _ := r.Progress.MissionState(r.ActiveCity.Mission.ID)
	totalXP := 0
	totalGold := 0
	if mp != nil {
		totalXP = mp.TotalXP
		totalGold = mp.TotalGold
	}

	r.Phase = PhaseMissionEnd
	r.Message = r.ActiveCity.Reward

	r.logEvent(PhaseMissionEnd, fmt.Sprintf(
		"Missão %s concluída! Total: %dxp %dg", r.ActiveCity.Mission.Title, totalXP, totalGold,
	))

	// Verifica se desbloqueou nova cidade
	if next, ok := NextCity(r.ActiveCity.ID); ok {
		r.Message += fmt.Sprintf("\n\n🔓 %s foi desbloqueada!", next.Name)
	}

	return nil
}

// ── Continuação após telas de transição ──────────────────────────────────────

// ConfirmWaveCleared avança da tela PhaseWaveCleared para a intro da próxima wave.
func (r *Runner) ConfirmWaveCleared() error {
	if r.Phase != PhaseWaveCleared {
		return fmt.Errorf("não está na tela de wave concluída")
	}
	r.Phase = PhaseWaveIntro
	r.Message = r.ActiveWave.Intro
	return nil
}

// ConfirmMissionEnd retorna ao mapa após concluir a missão.
func (r *Runner) ConfirmMissionEnd() {
	r.Phase = PhaseWorldMap
	r.ActiveCity = nil
	r.ActiveWave = nil
	r.EnemyQueue = nil
	r.Battle = nil
}

// RetryWave reinicia a wave após uma derrota (o player já tem HP=1 do onPlayerDefeated).
func (r *Runner) RetryWave() error {
	if r.Phase != PhaseGameOver {
		return fmt.Errorf("só é possível reiniciar após game over")
	}
	if r.ActiveCity == nil || r.ActiveWave == nil {
		return fmt.Errorf("estado inválido para retry")
	}

	// Restaura HP completo para o retry
	r.Player.HP = r.Player.MaxHP
	r.Player.Alive = true
	r.Player.ResetForNewBattle()

	r.Phase = PhaseWaveIntro
	r.Message = r.ActiveWave.Intro
	return nil
}

// ── Estado de apresentação ────────────────────────────────────────────────────

// UpdatedAt retorna quando o runner foi atualizado pela última vez.
// Usado para limpeza de sessões inativas.
func (r *Runner) UpdatedAt() time.Time {
	if r.Progress != nil {
		return r.Progress.UpdatedAt
	}
	return time.Time{}
}

// Snapshot retorna uma visão completa do estado atual para a UI renderizar.
func (r *Runner) Snapshot() RunnerSnapshot {
	canNG := false
	if r.Progress.CurrentDifficulty <= DifficultyNGPPP {
		// Verifica se todas as cidades estão concluídas
		allCleared := true
		for _, cs := range r.Progress.CampaignSummary() {
			if !cs.Cleared {
				allCleared = false
				break
			}
		}
		canNG = allCleared && int(r.Progress.CurrentDifficulty) < int(DifficultyNGPPP)
	}
	snap := RunnerSnapshot{
		Phase:           r.Phase,
		Message:         r.Message,
		Difficulty:      r.Progress.CurrentDifficulty,
		DifficultyLabel: r.Progress.CurrentDifficulty.Label(),
		CanStartNG:      canNG,
		HighestClear:    int(r.Progress.HighestClear),
		Player:          playerSnapshot(r.Player),
		Cities:          r.Progress.CampaignSummary(),
		Events:          r.LastEvents,
	}

	if r.ActiveCity != nil {
		snap.ActiveCityID = r.ActiveCity.ID
		snap.ActiveCityName = r.ActiveCity.Name
		snap.ActiveCityIcon = r.ActiveCity.Icon
		snap.ActiveFaction = r.ActiveCity.Faction
	}

	if r.ActiveWave != nil {
		snap.ActiveWaveTitle = r.ActiveWave.Title
		snap.ActiveWaveIsBoss = r.ActiveWave.IsBossWave
		snap.EnemyQueue = r.enemyQueueSnapshot()
	}

	if r.Battle != nil {
		snap.InitiativeOrder = r.initiativeSnapshot()
	}

	return snap
}


// BattleEvent é um evento individual do log de batalha
type BattleEvent struct {
	Actor   string `json:"actor"`
	Message string `json:"message"`
	Damage  int    `json:"damage"`
	IsError bool   `json:"is_error"`
}

// RunnerSnapshot é o DTO completo que a API/UI consome
type RunnerSnapshot struct {
	Phase      RunnerPhase  `json:"phase"`
	Message    string       `json:"message"`
	Difficulty      Difficulty `json:"difficulty"`
	DifficultyLabel string     `json:"difficulty_label"`
	CanStartNG      bool       `json:"can_start_ng"`
	HighestClear    int        `json:"highest_clear"`

	// Player
	Player PlayerSnapshot `json:"player"`

	// City
	ActiveCityID   string `json:"active_city_id,omitempty"`
	ActiveCityName string `json:"active_city_name,omitempty"`
	ActiveCityIcon string `json:"active_city_icon,omitempty"`
	ActiveFaction  string `json:"active_faction,omitempty"`

	// Wave
	ActiveWaveTitle  string `json:"active_wave_title,omitempty"`
	ActiveWaveIsBoss bool   `json:"active_wave_is_boss,omitempty"`

	// Battle
	EnemyQueue      []EnemyQueueSnapshot  `json:"enemy_queue,omitempty"`
	InitiativeOrder []InitiativeSnapshot  `json:"initiative_order,omitempty"`

	// World map
	Cities []CitySummary `json:"cities"`

	// Eventos do último turno (log de batalha)
	Events []BattleEvent `json:"events"`
}

type PlayerSnapshot struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Class            string   `json:"class"`
	Faction          string   `json:"faction"`
	Level            int      `json:"level"`
	Gold             int      `json:"gold"`
	HP               int      `json:"hp"`
	MaxHP            int      `json:"max_hp"`
	Mana             int      `json:"mana"`
	MaxMana          int      `json:"max_mana"`
	XP               int      `json:"xp"`
	XPToNext         int      `json:"xp_to_next"`
	Alive            bool     `json:"alive"`
	Statuses         []string `json:"statuses"`
	CryptoFactor     float64  `json:"crypto_factor"`
	CryptoVariation  float64  `json:"crypto_variation"`
	EffectiveAttackMod int    `json:"effective_attack_mod"`
	EffectiveCA        int    `json:"effective_ca"`
	EffectiveDefense   int    `json:"effective_defense"`
	AbilityName      string   `json:"ability_name"`
	AbilityAvailable bool     `json:"ability_available"`
	AbilityUsed      bool     `json:"ability_used"`
	AbilityCooldown  int      `json:"ability_cooldown_left"`
	AbilityDesc      string   `json:"ability_desc"`
	AbilityManaCost  int      `json:"ability_mana_cost"`

	// Habilidade 2
	Ability2Name      string  `json:"ability2_name"`
	Ability2Desc      string  `json:"ability2_desc"`
	Ability2Available bool    `json:"ability2_available"`
	Ability2Used      bool    `json:"ability2_used"`
	Ability2Cooldown  int     `json:"ability2_cooldown_left"`
	Ability2ManaCost  int     `json:"ability2_mana_cost"`
	Ability2Unlocked  bool    `json:"ability2_unlocked"`

	// Habilidade 3
	Ability3Name      string  `json:"ability3_name"`
	Ability3Desc      string  `json:"ability3_desc"`
	Ability3Available bool    `json:"ability3_available"`
	Ability3Used      bool    `json:"ability3_used"`
	Ability3Cooldown  int     `json:"ability3_cooldown_left"`
	Ability3ManaCost  int     `json:"ability3_mana_cost"`
	Ability3Unlocked  bool    `json:"ability3_unlocked"`

	// Passivas
	Passive1Name     string  `json:"passive1_name"`
	Passive1Desc     string  `json:"passive1_desc"`
	Passive1Unlocked bool    `json:"passive1_unlocked"`
	Passive2Name     string  `json:"passive2_name"`
	Passive2Desc     string  `json:"passive2_desc"`
	Passive2Unlocked bool    `json:"passive2_unlocked"`
	Passive3Name     string  `json:"passive3_name"`
	Passive3Desc     string  `json:"passive3_desc"`
	Passive3Unlocked bool    `json:"passive3_unlocked"`
}

type EnemyQueueSnapshot struct {
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	Tier     string `json:"tier"`
	Defeated bool   `json:"defeated"`
	IsCurrent bool  `json:"is_current"`
}

type InitiativeSnapshot struct {
	Name      string `json:"name"`
	Initiative int   `json:"initiative"`
	IsPlayer  bool   `json:"is_player"`
	IsCurrent bool   `json:"is_current"`
	Alive     bool   `json:"alive"`
}

func playerSnapshot(c *characters.Character) PlayerSnapshot {
	statuses := make([]string, len(c.Statuses))
	for i, s := range c.Statuses {
		statuses[i] = string(s.Effect)
	}
	factionInfo := characters.Factions[c.Faction]
	_ = factionInfo
	return PlayerSnapshot{
		ID: c.ID, Name: c.Name, Class: c.Class,
		Faction: string(c.Faction),
		Level: c.Level, Gold: c.Gold,
		HP: c.HP, MaxHP: c.MaxHP,
		Mana: c.Mana, MaxMana: c.MaxMana,
		XP: c.XP, XPToNext: c.XPToNext,
		Alive: c.Alive, Statuses: statuses,
		CryptoVariation:    c.CryptoVariation,
		EffectiveAttackMod: c.EffectiveAttackMod(),
		EffectiveCA:        c.EffectiveCA(),
		EffectiveDefense:   c.EffectiveDefense(),
		AbilityName:        c.Ability.Name,
		AbilityAvailable:   c.CanUseAbility(),
		AbilityUsed:        c.AbilityUsed,
		AbilityCooldown:    c.AbilityCooldownLeft,
		AbilityDesc:        c.Ability.Description,
		AbilityManaCost:    c.Ability.ManaCost,
		// Ability 2
		Ability2Name:      c.Ability2.Name,
		Ability2Desc:      c.Ability2.Description,
		Ability2Available: c.CanUseAbility2(),
		Ability2Used:      c.Ability2Used,
		Ability2Cooldown:  c.Ability2CooldownLeft,
		Ability2ManaCost:  c.Ability2.ManaCost,
		Ability2Unlocked:  c.Ability2.Unlocked,
		// Ability 3
		Ability3Name:      c.Ability3.Name,
		Ability3Desc:      c.Ability3.Description,
		Ability3Available: c.CanUseAbility3(),
		Ability3Used:      c.Ability3Used,
		Ability3Cooldown:  c.Ability3CooldownLeft,
		Ability3ManaCost:  c.Ability3.ManaCost,
		Ability3Unlocked:  c.Ability3.Unlocked,
		// Passivas
		Passive1Name:     c.Passive1.Name,
		Passive1Desc:     c.Passive1.Description,
		Passive1Unlocked: c.Passive1.Unlocked,
		Passive2Name:     c.Passive2.Name,
		Passive2Desc:     c.Passive2.Description,
		Passive2Unlocked: c.Passive2.Unlocked,
		Passive3Name:     c.Passive3.Name,
		Passive3Desc:     c.Passive3.Description,
		Passive3Unlocked: c.Passive3.Unlocked,
	}
}

func (r *Runner) enemyQueueSnapshot() []EnemyQueueSnapshot {
	snap := make([]EnemyQueueSnapshot, len(r.EnemyQueue))
	for i, e := range r.EnemyQueue {
		snap[i] = EnemyQueueSnapshot{
			Name:      e.Enemy.Name,
			Icon:      enemyIcon(e.Enemy.Name),
			HP:        e.Enemy.HP,
			MaxHP:     e.Enemy.MaxHP,
			Tier:      string(e.Enemy.Tier),
			Defeated:  e.Defeated,
			IsCurrent: i == r.CurrentEnemy,
		}
	}
	return snap
}

func (r *Runner) initiativeSnapshot() []InitiativeSnapshot {
	if r.Battle == nil {
		return nil
	}
	order := r.Battle.InitiativeOrder()
	snap := make([]InitiativeSnapshot, len(order))
	for i, e := range order {
		snap[i] = InitiativeSnapshot{
			Name:       e.Combatant.GetName(),
			Initiative: e.Initiative,
			IsPlayer:   e.Combatant.GetTeam() == 0, // TeamPlayer = 0
			IsCurrent:  i == 0,                     // approximation; refine in handler
			Alive:      e.Combatant.GetCharacter().IsAlive(),
		}
	}
	return snap
}

// ── Spawn com Scaling ─────────────────────────────────────────────────────────

// spawnScaled cria um inimigo do catálogo com os stats escalados pela dificuldade.
func spawnScaled(templateName string, d Difficulty) (*enemies.Enemy, error) {
	// Encontra o template
	var tmpl *enemies.EnemyTemplate
	for i := range enemies.Catalogue {
		if enemies.Catalogue[i].Name == templateName {
			tmpl = &enemies.Catalogue[i]
			break
		}
	}
	if tmpl == nil {
		return nil, fmt.Errorf("template %q não encontrado", templateName)
	}

	if d == DifficultyNormal {
		return enemies.Spawn(templateName)
	}

	// Cria uma cópia do template com stats escalados
	scaled := *tmpl
	mult := d.StatMultiplier()

	scaled.Level = int(math.Ceil(float64(tmpl.Level) * mult))
	if tmpl.HPOverride > 0 {
		scaled.HPOverride = int(math.Ceil(float64(tmpl.HPOverride) * mult))
	}
	if tmpl.AttackModOverride > 0 {
		scaled.AttackModOverride = int(math.Ceil(float64(tmpl.AttackModOverride) * mult))
	}
	if tmpl.StrengthModOverride > 0 {
		scaled.StrengthModOverride = int(math.Ceil(float64(tmpl.StrengthModOverride) * mult))
	}
	scaled.XPReward = int(math.Ceil(float64(tmpl.XPReward) * d.XPMultiplier()))
	scaled.GoldReward = int(math.Ceil(float64(tmpl.GoldReward) * mult))

	return enemies.SpawnFromTemplate(scaled)
}

// enemyIcon — mapeamento de nomes para emojis (duplicado do dto.go; isolado aqui)
func enemyIcon(name string) string {
	icons := map[string]string{
		"Especulador Novato": "💸", "Bot de Pump": "🤖",
		"Minerador Fantasma": "⛏",  "Fomo Cultist": "🙏",
		"Dust Raider": "🏜",         "Whale Corrupta": "🐋",
		"Oráculo Corrompido": "👁",  "Sombra do Mempool": "👤",
		"Validador Traidor": "🗡",   "Satoshi das Trevas": "💀",
		"Vitalik Void": "🌀",        "O Liquidador": "⚡",
		"DOGE Primordial": "🐕",
	}
	if icon, ok := icons[name]; ok {
		return icon
	}
	return "👾"
}

func (r *Runner) logEvent(phase RunnerPhase, msg string) {
	r.Log = append(r.Log, RunnerEvent{
		Phase:   phase,
		Message: msg,
		At:      time.Now(),
	})
}
