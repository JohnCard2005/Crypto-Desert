package enemies

import (
	"crypto-desert/internal/characters"
	"fmt"
)

// EnemyTemplate is a reusable blueprint for spawning enemies
type EnemyTemplate struct {
	Name        string
	Class       string     // reuses player class stats as base
	Level       int
	Behavior    AIBehavior
	Tier        EnemyTier
	Description string
	XPReward    int
	GoldReward  int

	// Stat overrides (0 = use class default)
	HPOverride          int
	AttackModOverride   int
	StrengthModOverride int
	CAOverride          int
	SpeedOverride       int
}

// Catalogue is the full list of spawnable enemies, organized by zone/tier
var Catalogue = []EnemyTemplate{

	// ── COMMON ───────────────────────────────────────────────────────────────

	{
		Name:        "Especulador Novato",
		Class:       "warrior",
		Level:       1,
		Behavior:    BehaviorAggressive,
		Tier:        TierCommon,
		Description: "Um apostador desesperado que perdeu tudo no crash de 2085. Ataca sem estratégia.",
		XPReward:    80,
		GoldReward:  10,
		HPOverride:  60,
	},
	{
		Name:        "Bot de Pump",
		Class:       "rogue",
		Level:       2,
		Behavior:    BehaviorRandom,
		Tier:        TierCommon,
		Description: "Script automatizado corrompido, seus ataques são caóticos e imprevisíveis.",
		XPReward:    100,
		GoldReward:  15,
	},
	{
		Name:        "Minerador Fantasma",
		Class:       "warrior",
		Level:       3,
		Behavior:    BehaviorDefensive,
		Tier:        TierCommon,
		Description: "Antigo minerador de BTC que sobreviveu ao deserto digital. Resistente e cauteloso.",
		XPReward:    120,
		GoldReward:  20,
		HPOverride:  150,
		CAOverride:  16,
	},
	{
		Name:        "Fomo Cultist",
		Class:       "mage",
		Level:       2,
		Behavior:    BehaviorAggressive,
		Tier:        TierCommon,
		Description: "Seguidor fanático que ataca em ondas de pânico. Fraco mas perigoso em grupo.",
		XPReward:    90,
		GoldReward:  12,
		HPOverride:  55,
		AttackModOverride: 7,
	},
	{
		Name:        "Dust Raider",
		Class:       "archer",
		Level:       3,
		Behavior:    BehaviorAggressive,
		Tier:        TierCommon,
		Description: "Saqueador das ruínas de exchanges falidas. Prefere alvos com pouca defesa.",
		XPReward:    110,
		GoldReward:  18,
	},

	// ── ELITE ────────────────────────────────────────────────────────────────

	{
		Name:        "Whale Corrupta",
		Class:       "warrior",
		Level:       8,
		Behavior:    BehaviorBerserker,
		Tier:        TierElite,
		Description: "Uma baleia do mercado que virou predadora. Usa habilidades sem hesitação e devora os fracos.",
		XPReward:    500,
		GoldReward:  120,
		HPOverride:  300,
		StrengthModOverride: 6,
	},
	{
		Name:        "Oráculo Corrompido",
		Class:       "mage",
		Level:       7,
		Behavior:    BehaviorSupport,
		Tier:        TierElite,
		Description: "Um nó de oráculo blockchain que enlouqueceu. Aplica debuffs antes de atacar.",
		XPReward:    480,
		GoldReward:  100,
		HPOverride:  180,
		AttackModOverride: 9,
	},
	{
		Name:        "Sombra do Mempool",
		Class:       "rogue",
		Level:       9,
		Behavior:    BehaviorSupport,
		Tier:        TierElite,
		Description: "Entidade que habita o espaço entre transações. Envenena e desaparece.",
		XPReward:    550,
		GoldReward:  130,
		SpeedOverride: 6,
	},
	{
		Name:        "Validador Traidor",
		Class:       "archer",
		Level:       8,
		Behavior:    BehaviorDefensive,
		Tier:        TierElite,
		Description: "Um validador de rede que vendeu seu nó. Atira à distância e se protege quando pressionado.",
		XPReward:    490,
		GoldReward:  110,
	},

	// ── BOSS ─────────────────────────────────────────────────────────────────

	{
		Name:        "Satoshi das Trevas",
		Class:       "warrior",
		Level:       15,
		Behavior:    BehaviorBerserker,
		Tier:        TierBoss,
		Description: "A entidade que corrompeu o bloco genesis. Implacável, com fúria do bloco sempre disponível.",
		XPReward:    2000,
		GoldReward:  500,
		HPOverride:  600,
		AttackModOverride:   10,
		StrengthModOverride: 8,
		CAOverride:          18,
	},
	{
		Name:        "Vitalik Void",
		Class:       "mage",
		Level:       14,
		Behavior:    BehaviorSupport,
		Tier:        TierBoss,
		Description: "O arquiteto de um contrato inteligente que escapou do controle. Congela e destrói.",
		XPReward:    1800,
		GoldReward:  450,
		HPOverride:  400,
		AttackModOverride: 12,
	},
	{
		Name:        "O Liquidador",
		Class:       "rogue",
		Level:       16,
		Behavior:    BehaviorAggressive,
		Tier:        TierBoss,
		Description: "Protocolo de liquidação forçada que ganhou consciência. Elimina posições — e vidas.",
		XPReward:    2200,
		GoldReward:  600,
		HPOverride:  450,
		SpeedOverride:       7,
		AttackModOverride:   9,
		StrengthModOverride: 7,
	},
	{
		Name:        "DOGE Primordial",
		Class:       "shaman",
		Level:       12,
		Behavior:    BehaviorRandom,
		Tier:        TierBoss,
		Description: "O meme original, manifestado como entidade. Seus ataques são caóticos mas devastadores.",
		XPReward:    1600,
		GoldReward:  400,
		HPOverride:  380,
	},
}

// Spawn creates a live Enemy from a template by name.
// Returns error if the template is not found.
func Spawn(templateName string) (*Enemy, error) {
	for _, tmpl := range Catalogue {
		if tmpl.Name == templateName {
			return spawnFromTemplate(tmpl)
		}
	}
	return nil, fmt.Errorf("enemy template %q not found", templateName)
}

// SpawnByTier creates a random enemy of the given tier
func SpawnByTier(tier EnemyTier) (*Enemy, error) {
	var candidates []EnemyTemplate
	for _, tmpl := range Catalogue {
		if tmpl.Tier == tier {
			candidates = append(candidates, tmpl)
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no enemies of tier %q found", tier)
	}

	// Use crypto-style deterministic-ish selection
	idx := len(candidates) / 2 // replaced with rand in battle package
	return spawnFromTemplate(candidates[idx])
}

// SpawnRandom picks a random enemy from the full catalogue
func SpawnRandom() (*Enemy, error) {
	if len(Catalogue) == 0 {
		return nil, fmt.Errorf("catalogue is empty")
	}
	return spawnFromTemplate(Catalogue[0]) // caller should shuffle; see battle.go
}

// spawnFromTemplate builds the Character and wraps it as Enemy
func spawnFromTemplate(tmpl EnemyTemplate) (*Enemy, error) {
	c, err := characters.NewCharacterAtLevel(tmpl.Name, tmpl.Class, tmpl.Level)
	if err != nil {
		return nil, fmt.Errorf("spawn %q: %w", tmpl.Name, err)
	}

	// Apply stat overrides
	if tmpl.HPOverride > 0 {
		c.HP = tmpl.HPOverride
		c.MaxHP = tmpl.HPOverride
	}
	if tmpl.AttackModOverride > 0 {
		c.AttackMod = tmpl.AttackModOverride
	}
	if tmpl.StrengthModOverride > 0 {
		c.StrengthMod = tmpl.StrengthModOverride
	}
	if tmpl.CAOverride > 0 {
		c.CA = tmpl.CAOverride
	}
	if tmpl.SpeedOverride > 0 {
		c.Speed = tmpl.SpeedOverride
	}

	return NewEnemy(c, tmpl.Behavior, tmpl.Tier, tmpl.Description, tmpl.XPReward, tmpl.GoldReward), nil
}

// SpawnGroup creates a group of enemies appropriate for a given player level
func SpawnGroup(playerLevel int, count int) ([]*Enemy, error) {
	var pool []EnemyTemplate
	for _, tmpl := range Catalogue {
		diff := tmpl.Level - playerLevel
		if diff >= -2 && diff <= 3 { // enemies within 2 below or 3 above player level
			if tmpl.Tier != TierBoss { // no bosses in random groups
				pool = append(pool, tmpl)
			}
		}
	}

	if len(pool) == 0 {
		// fallback: use any common enemy
		for _, tmpl := range Catalogue {
			if tmpl.Tier == TierCommon {
				pool = append(pool, tmpl)
			}
		}
	}

	result := make([]*Enemy, 0, count)
	for i := 0; i < count; i++ {
		tmpl := pool[i%len(pool)]
		e, err := spawnFromTemplate(tmpl)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
