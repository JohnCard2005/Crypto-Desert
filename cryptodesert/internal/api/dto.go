package api

import (
	"crypto-desert/internal/characters"
	"crypto-desert/internal/enemies"
	"crypto-desert/internal/game"
)

// ── Character DTOs ────────────────────────────────────────────────────────────

type CreateCharacterRequest struct {
	Name  string `json:"name"`
	Class string `json:"class"`
}

type CharacterResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Faction     string `json:"faction"`
	FactionName string `json:"faction_name"`
	Level       int    `json:"level"`
	XP          int    `json:"xp"`
	XPToNext    int    `json:"xp_to_next"`
	Gold        int    `json:"gold"`
	HP          int    `json:"hp"`
	MaxHP       int    `json:"max_hp"`
	Mana        int    `json:"mana"`
	MaxMana     int    `json:"max_mana"`
	AttackMod   int    `json:"attack_mod"`
	StrengthMod int    `json:"strength_mod"`
	CA          int    `json:"ca"`
	Defense     int    `json:"defense"`
	Speed       int    `json:"speed"`
	DamageDice  int    `json:"damage_dice"`
	CryptoID    string `json:"crypto_id"`

	// Bônus de equipamento ativos
	BonusAttackMod    int     `json:"bonus_attack_mod"`
	BonusStrengthMod  int     `json:"bonus_strength_mod"`
	BonusCA           int     `json:"bonus_ca"`
	BonusDefense      int     `json:"bonus_defense"`
	BonusSpeed        int     `json:"bonus_speed"`
	CryptoFactorBonus float64 `json:"crypto_factor_bonus"`

	// Stats efetivos (base + bônus) — conveniência para o frontend
	EffectiveAttackMod int `json:"effective_attack_mod"`
	EffectiveCA        int `json:"effective_ca"`
	EffectiveDefense   int `json:"effective_defense"`

	// Dados crypto ao vivo
	CryptoVariation float64 `json:"crypto_variation"` // variação 7d em %
	CryptoFactor    float64 `json:"crypto_factor"`    // fator de dano [0.5, 2.0]

	Ability  AbilityResponse  `json:"ability"`

	// Habilidade 1 — campos flat para o inventário/roster
	AbilityName      string `json:"ability_name"`
	AbilityDesc      string `json:"ability_desc"`
	AbilityAvailable bool   `json:"ability_available"`

	// Habilidade 2
	Ability2Name     string `json:"ability2_name"`
	Ability2Desc     string `json:"ability2_desc"`
	Ability2Unlocked bool   `json:"ability2_unlocked"`

	// Habilidade 3
	Ability3Name     string `json:"ability3_name"`
	Ability3Desc     string `json:"ability3_desc"`
	Ability3Unlocked bool   `json:"ability3_unlocked"`

	// Passivas
	Passive1Name     string `json:"passive1_name"`
	Passive1Desc     string `json:"passive1_desc"`
	Passive1Unlocked bool   `json:"passive1_unlocked"`
	Passive2Name     string `json:"passive2_name"`
	Passive2Desc     string `json:"passive2_desc"`
	Passive2Unlocked bool   `json:"passive2_unlocked"`
	Passive3Name     string `json:"passive3_name"`
	Passive3Desc     string `json:"passive3_desc"`
	Passive3Unlocked bool   `json:"passive3_unlocked"`

	Statuses []StatusResponse `json:"statuses"`
	Alive    bool             `json:"alive"`
}

type AbilityResponse struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	ManaCost     int     `json:"mana_cost"`
	Cooldown     int     `json:"cooldown"`
	DamageMult   float64 `json:"damage_mult"`
	CooldownLeft int     `json:"cooldown_left"`
	Used         bool    `json:"used"`
	Available    bool    `json:"available"`
}

type StatusResponse struct {
	Effect   string `json:"effect"`
	Duration int    `json:"duration"`
	Power    int    `json:"power"`
}

// ── Class Info DTO ────────────────────────────────────────────────────────────

type ClassInfoResponse struct {
	Key             string  `json:"key"`
	Name            string  `json:"name"`
	Icon            string  `json:"icon"`
	Faction         string  `json:"faction"`
	FactionName     string  `json:"faction_name"`
	Lore            string  `json:"lore"`
	HP              int     `json:"hp"`
	Mana            int     `json:"mana"`
	AttackMod       int     `json:"attack_mod"`
	StrengthMod     int     `json:"strength_mod"`
	CA              int     `json:"ca"`
	Defense         int     `json:"defense"`
	Speed           int     `json:"speed"`
	DamageDice      int     `json:"damage_dice"`
	CryptoFactor    float64 `json:"crypto_factor"`
	CryptoVariation float64 `json:"crypto_variation"`
}

// ── Enemy DTOs ────────────────────────────────────────────────────────────────

type EnemyResponse struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Class           string           `json:"class"`
	Tier            string           `json:"tier"`
	Behavior        string           `json:"behavior"`
	Description     string           `json:"description"`
	Icon            string           `json:"icon"`
	Level           int              `json:"level"`
	HP              int              `json:"hp"`
	MaxHP           int              `json:"max_hp"`
	AttackMod       int              `json:"attack_mod"`
	StrengthMod     int              `json:"strength_mod"`
	CA              int              `json:"ca"`
	Defense         int              `json:"defense"`
	Speed           int              `json:"speed"`
	DamageDice      int              `json:"damage_dice"`
	Faction         string           `json:"faction"`
	XPReward        int              `json:"xp_reward"`
	GoldReward      int              `json:"gold_reward"`
	CryptoFactor    float64          `json:"crypto_factor"`
	CryptoVariation float64          `json:"crypto_variation"`
	Statuses        []StatusResponse `json:"statuses"`
	Alive           bool             `json:"alive"`
}

// ── Battle DTOs ───────────────────────────────────────────────────────────────

type StartBattleRequest struct {
	CharacterID int    `json:"character_id"`
	EnemyName   string `json:"enemy_name"`
}

type BattleStateResponse struct {
	SessionID    string            `json:"session_id"`
	Status       string            `json:"status"`
	TurnNumber   int               `json:"turn_number"`
	Player       CharacterResponse `json:"player"`
	Enemy        EnemyResponse     `json:"enemy"`
	Initiative   []InitEntry       `json:"initiative"`
	CurrentActor string            `json:"current_actor"` // "player" or "enemy"
	Events       []EventResponse   `json:"events"`
}

type InitEntry struct {
	Name       string `json:"name"`
	Initiative int    `json:"initiative"`
	IsPlayer   bool   `json:"is_player"`
	IsCurrent  bool   `json:"is_current"`
	Alive      bool   `json:"alive"`
}

type EventResponse struct {
	Actor   string `json:"actor"`
	Message string `json:"message"`
	Damage  int    `json:"damage"`
	IsError bool   `json:"is_error"`
}

type TakeActionRequest struct {
	SessionID string `json:"session_id"`
	Action    string `json:"action"` // attack / defend / ability / flee
	TargetID  string `json:"target_id,omitempty"`
}

type BattleConclusionResponse struct {
	Status     string `json:"status"`
	TurnCount  int    `json:"turn_count"`
	XPGained   int    `json:"xp_gained"`
	GoldGained int    `json:"gold_gained"`
	NewLevel   int    `json:"new_level,omitempty"`
	LeveledUp  bool   `json:"leveled_up"`
}

// ── Converters ────────────────────────────────────────────────────────────────

func CharToResponse(c *characters.Character, cs *CryptoService) CharacterResponse {
	factor := cs.GetFactor(c.CryptoID)
	change := cs.GetChange7d(c.CryptoID)

	factionInfo := characters.Factions[c.Faction]

	statuses := make([]StatusResponse, len(c.Statuses))
	for i, s := range c.Statuses {
		statuses[i] = StatusResponse{
			Effect:   string(s.Effect),
			Duration: s.Duration,
			Power:    s.Power,
		}
	}

	return CharacterResponse{
		ID:               c.ID,
		Name:             c.Name,
		Class:            c.Class,
		Faction:          string(c.Faction),
		FactionName:      factionInfo.Name,
		Level:            c.Level,
		XP:               c.XP,
		XPToNext:         c.XPToNext,
		Gold:             c.Gold,
		HP:               c.HP,
		MaxHP:            c.MaxHP,
		Mana:             c.Mana,
		MaxMana:          c.MaxMana,
		AttackMod:        c.AttackMod,
		StrengthMod:      c.StrengthMod,
		CA:               c.CA,
		Defense:          c.Defense,
		Speed:            c.Speed,
		DamageDice:       c.DamageDice,
		CryptoID:         c.CryptoID,
		BonusAttackMod:   c.BonusAttackMod,
		BonusStrengthMod: c.BonusStrengthMod,
		BonusCA:          c.BonusCA,
		BonusDefense:     c.BonusDefense,
		BonusSpeed:       c.BonusSpeed,
		CryptoFactorBonus: c.CryptoFactorBonus,
		EffectiveAttackMod: c.EffectiveAttackMod(),
		EffectiveCA:        c.EffectiveCA(),
		EffectiveDefense:   c.EffectiveDefense(),
		CryptoVariation:  change,
		CryptoFactor:     factor,
		Ability: AbilityResponse{
			Name:         c.Ability.Name,
			Description:  c.Ability.Description,
			ManaCost:     c.Ability.ManaCost,
			Cooldown:     c.Ability.Cooldown,
			DamageMult:   c.Ability.DamageMult,
			CooldownLeft: c.AbilityCooldownLeft,
			Used:         c.AbilityUsed,
			Available:    c.CanUseAbility(),
		},
		// Habilidade 1 (flat fields para o inventário)
		AbilityName:      c.Ability.Name,
		AbilityDesc:      c.Ability.Description,
		AbilityAvailable: c.CanUseAbility(),
		// Habilidade 2
		Ability2Name:     c.Ability2.Name,
		Ability2Desc:     c.Ability2.Description,
		Ability2Unlocked: c.Ability2.Unlocked,
		// Habilidade 3
		Ability3Name:     c.Ability3.Name,
		Ability3Desc:     c.Ability3.Description,
		Ability3Unlocked: c.Ability3.Unlocked,
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
		Statuses: statuses,
		Alive:    c.Alive,
	}
}

func EnemyToResponse(e *enemies.Enemy, cs *CryptoService) EnemyResponse {
	factor := cs.GetFactor(e.CryptoID)
	change := cs.GetChange7d(e.CryptoID)

	statuses := make([]StatusResponse, len(e.Statuses))
	for i, s := range e.Statuses {
		statuses[i] = StatusResponse{
			Effect:   string(s.Effect),
			Duration: s.Duration,
			Power:    s.Power,
		}
	}

	return EnemyResponse{
		ID:              e.Name,
		Name:            e.Name,
		Class:           e.Class,
		Tier:            string(e.Tier),
		Behavior:        string(e.Behavior),
		Description:     e.Description,
		Icon:            enemyIcon(e.Name),
		Level:           e.Level,
		HP:              e.HP,
		MaxHP:           e.MaxHP,
		AttackMod:       e.AttackMod,
		StrengthMod:     e.StrengthMod,
		CA:              e.CA,
		Defense:         e.Defense,
		Speed:           e.Speed,
		DamageDice:      e.DamageDice,
		Faction:         string(e.Faction),
		XPReward:        e.XPReward,
		GoldReward:      e.GoldReward,
		CryptoFactor:    factor,
		CryptoVariation: change,
		Statuses:        statuses,
		Alive:           e.Alive,
	}
}

func TurnEventsToResponse(events []game.TurnEvent) []EventResponse {
	resp := make([]EventResponse, len(events))
	for i, e := range events {
		resp[i] = EventResponse{
			Actor:   e.Actor,
			Message: e.Message,
			Damage:  e.Damage,
			IsError: e.IsError,
		}
	}
	return resp
}

func enemyIcon(name string) string {
	icons := map[string]string{
		"Especulador Novato": "💸",
		"Bot de Pump":        "🤖",
		"Minerador Fantasma": "⛏",
		"Fomo Cultist":       "🙏",
		"Dust Raider":        "🏜",
		"Whale Corrupta":     "🐋",
		"Oráculo Corrompido": "👁",
		"Sombra do Mempool":  "👤",
		"Validador Traidor":  "🗡",
		"Satoshi das Trevas": "💀",
		"Vitalik Void":       "🌀",
		"O Liquidador":       "⚡",
		"DOGE Primordial":    "🐕",
	}
	if icon, ok := icons[name]; ok {
		return icon
	}
	return "👾"
}
