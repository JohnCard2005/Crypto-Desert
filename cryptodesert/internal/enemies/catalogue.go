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
		XPReward:    280,
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
		XPReward:    350,
		GoldReward:  14,
	},
	{
		Name:        "Minerador Fantasma",
		Class:       "warrior",
		Level:       3,
		Behavior:    BehaviorDefensive,
		Tier:        TierCommon,
		Description: "Antigo minerador de BTC que sobreviveu ao deserto digital. Resistente e cauteloso.",
		XPReward:    420,
		GoldReward:  18,
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
		XPReward:    300,
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
		XPReward:    380,
		GoldReward:  16,
	},

	// ── ELITE ────────────────────────────────────────────────────────────────

	{
		Name:        "Whale Corrupta",
		Class:       "warrior",
		Level:       8,
		Behavior:    BehaviorBerserker,
		Tier:        TierElite,
		Description: "Uma baleia do mercado que virou predadora. Usa habilidades sem hesitação e devora os fracos.",
		XPReward:    1500,
		GoldReward:  80,
		HPOverride:  160,
		StrengthModOverride: 6,
	},
	{
		Name:        "Oráculo Corrompido",
		Class:       "mage",
		Level:       7,
		Behavior:    BehaviorSupport,
		Tier:        TierElite,
		Description: "Um nó de oráculo blockchain que enlouqueceu. Aplica debuffs antes de atacar.",
		XPReward:    1400,
		GoldReward:  70,
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
		XPReward:    1600,
		GoldReward:  90,
		SpeedOverride: 6,
	},
	{
		Name:        "Validador Traidor",
		Class:       "archer",
		Level:       8,
		Behavior:    BehaviorDefensive,
		Tier:        TierElite,
		Description: "Um validador de rede que vendeu seu nó. Atira à distância e se protege quando pressionado.",
		XPReward:    1450,
		GoldReward:  75,
	},

	// ── BOSS ─────────────────────────────────────────────────────────────────

	{
		Name:        "Satoshi das Trevas",
		Class:       "warrior",
		Level:       8,
		Behavior:    BehaviorBerserker,
		Tier:        TierBoss,
		Description: "A entidade que corrompeu o bloco genesis. Implacável, com fúria do bloco sempre disponível.",
		XPReward:    7000,
		GoldReward:  350,
		HPOverride:  180,
		AttackModOverride:   5,
		StrengthModOverride: 4,
		CAOverride:          13,
	},
	{
		Name:        "Vitalik Void",
		Class:       "mage",
		Level:       11,
		Behavior:    BehaviorSupport,
		Tier:        TierBoss,
		Description: "O arquiteto de um contrato inteligente que escapou do controle. Congela e destrói.",
		XPReward:    6500,
		GoldReward:  300,
		HPOverride:  160,
		AttackModOverride: 7,
	},
	{
		Name:        "O Liquidador",
		Class:       "rogue",
		Level:       13,
		Behavior:    BehaviorAggressive,
		Tier:        TierBoss,
		Description: "Protocolo de liquidação forçada que ganhou consciência. Elimina posições — e vidas.",
		XPReward:    8000,
		GoldReward:  400,
		HPOverride:  200,
		SpeedOverride:       5,
		AttackModOverride:   7,
		StrengthModOverride: 4,
	},
	{
		Name:        "DOGE Primordial",
		Class:       "shaman",
		Level:       18,
		Behavior:    BehaviorRandom,
		Tier:        TierBoss,
		Description: "O meme original, manifestado como entidade. Seus ataques são caóticos mas devastadores.",
		XPReward:    6000,
		GoldReward:  280,
		HPOverride:  250,
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

// SpawnFromTemplate is the exported version of spawnFromTemplate,
// used by the missions package for difficulty-scaled enemy creation.
func SpawnFromTemplate(tmpl EnemyTemplate) (*Enemy, error) {
	return spawnFromTemplate(tmpl)
}
