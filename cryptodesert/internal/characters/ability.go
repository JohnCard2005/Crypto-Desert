package characters

// ── Ability ───────────────────────────────────────────────────────────────────

type Ability struct {
	Name           string
	Description    string
	ManaCost       int
	Cooldown       int         // 0 = once per battle, >0 = turns between uses
	DamageMult     float64
	AppliesStatus  StatusEffect
	StatusDuration int
	StatusPower    int
	Targeting      AbilityTarget
	Unlocked       bool        // true se o personagem já tem acesso
}

type AbilityTarget string

const (
	TargetEnemy AbilityTarget = "enemy"
	TargetSelf  AbilityTarget = "self"
	TargetAll   AbilityTarget = "all_enemies"
)

// ── Passive ───────────────────────────────────────────────────────────────────

// PassiveEffect é o efeito permanente de uma passiva
type PassiveEffect string

const (
	PassiveExtraHP          PassiveEffect = "extra_hp"           // +N de HP máximo permanente
	PassiveExtraAtk         PassiveEffect = "extra_atk"          // +N de ataque permanente
	PassiveExtraDef         PassiveEffect = "extra_def"          // +N de defesa permanente
	PassiveExtraCA          PassiveEffect = "extra_ca"           // +N de CA permanente
	PassiveRegenMana        PassiveEffect = "regen_mana"         // regenera N mana por turno
	PassiveLifesteal        PassiveEffect = "lifesteal"          // cura N% do dano causado
	PassiveFirstBlood       PassiveEffect = "first_blood"        // +30% dano no primeiro ataque
	PassiveCryptoAmplify    PassiveEffect = "crypto_amplify"     // +0.15 no fator crypto
	PassivePoisonMastery    PassiveEffect = "poison_mastery"     // veneno causa +50% dano
	PassiveIronWill         PassiveEffect = "iron_will"          // imune a 1 derrota por batalha (sobrevive com 1HP)
	PassiveSpiritGuard      PassiveEffect = "spirit_guard"       // +2 DEF e +2 SPD permanentes
	PassiveBloodPrice       PassiveEffect = "blood_price"        // cada kill restaura 20HP
)

type Passive struct {
	Name        string
	Description string
	Effect      PassiveEffect
	Value       float64 // valor numérico do efeito (HP, ATK, % etc)
	Unlocked    bool
}

// ── Ability Slots ─────────────────────────────────────────────────────────────

// ClassAbilityTree define todas as habilidades e passivas de cada classe por nível
type ClassAbilityTree struct {
	// Habilidades (desbloqueadas nos níveis 1, 10, 20, 30)
	Ability1       Ability // Lv.1  — habilidade inicial
	Ability2       Ability // Lv.10 — segunda habilidade
	Ability3       Ability // Lv.20 — terceira habilidade
	Ability1Evo    Ability // Lv.30 — evolução da primeira habilidade

	// Passivas (desbloqueadas nos níveis 5, 15, 25)
	Passive1       Passive // Lv.5
	Passive2       Passive // Lv.15
	Passive3       Passive // Lv.25
}

// AbilityUnlockLevel mapeia o slot para o nível de desbloqueio
var AbilityUnlockLevel = map[string]int{
	"ability1":    1,
	"passive1":    5,
	"ability2":    10,
	"passive2":    15,
	"ability3":    20,
	"passive3":    25,
	"ability1evo": 30,
}

// ClassTrees é a árvore completa de progressão por classe
var ClassTrees = map[string]ClassAbilityTree{

	// ── WARRIOR (BTC) ──────────────────────────────────────────────────────────
	"warrior": {
		Ability1: Ability{
			Name:        "Fúria do Bloco",
			Description: "Canaliza o poder do bloco genesis — dano triplo em um único golpe devastador.",
			ManaCost:    0, Cooldown: 0, DamageMult: 2.5, Targeting: TargetEnemy, Unlocked: true,
		},
		Passive1: Passive{
			Name:        "Pele de Ferro",
			Description: "+15 HP Máximo e +2 Defesa permanentes. O bloco não cede.",
			Effect: PassiveExtraHP, Value: 15,
		},
		Ability2: Ability{
			Name:        "Avalanche de Hash",
			Description: "Ataque em área — golpeia o inimigo e libera uma onda que causa 50% do dano como dano bônus fixo.",
			ManaCost:    0, Cooldown: 2, DamageMult: 1.8, Targeting: TargetEnemy,
		},
		Passive2: Passive{
			Name:        "Vontade de Ferro",
			Description: "Uma vez por batalha, sobrevive com 1 HP ao invés de morrer.",
			Effect: PassiveIronWill, Value: 1,
		},
		Ability3: Ability{
			Name:        "Punho do Genesis",
			Description: "Golpe massivo — dano ×2.5 e aplica Burning por 2 turnos.",
			ManaCost:    0, Cooldown: 3, DamageMult: 2.5,
			AppliesStatus: StatusBurning, StatusDuration: 2, StatusPower: 8,
			Targeting: TargetEnemy,
		},
		Passive3: Passive{
			Name:        "Sangue do Bloco",
			Description: "Cada inimigo derrotado restaura 20 HP.",
			Effect: PassiveBloodPrice, Value: 20,
		},
		Ability1Evo: Ability{
			Name:        "Fúria Primordial",
			Description: "Evolução da Fúria do Bloco — dano ×5.0 e ignora toda a defesa do alvo.",
			ManaCost:    0, Cooldown: 0, DamageMult: 4.0, Targeting: TargetEnemy,
		},
	},

	// ── MAGE (ETH) ────────────────────────────────────────────────────────────
	"mage": {
		Ability1: Ability{
			Name:        "Contrato Inteligente",
			Description: "Executa um contrato que congela o alvo, travando suas ações por 1 turno.",
			ManaCost: 20, Cooldown: 3, DamageMult: 1.5,
			AppliesStatus: StatusFrozen, StatusDuration: 1,
			Targeting: TargetEnemy, Unlocked: true,
		},
		Passive1: Passive{
			Name:        "Canal Aberto",
			Description: "Regenera 8 de Mana por turno passivamente.",
			Effect: PassiveRegenMana, Value: 8,
		},
		Ability2: Ability{
			Name:        "Overflow de Gas",
			Description: "Lança uma explosão de gas fee — dano ×2.2 e aplica Burning por 2 turnos.",
			ManaCost: 30, Cooldown: 3, DamageMult: 2.2,
			AppliesStatus: StatusBurning, StatusDuration: 2, StatusPower: 6,
			Targeting: TargetEnemy,
		},
		Passive2: Passive{
			Name:        "Amplificador Crypto",
			Description: "+0.20 no fator de dano crypto. O mercado trabalha para você.",
			Effect: PassiveCryptoAmplify, Value: 0.20,
		},
		Ability3: Ability{
			Name:        "Cascata de Contratos",
			Description: "Executa 3 contratos simultâneos — cada um causa dano ×1.2 e tem chance de congelar.",
			ManaCost: 50, Cooldown: 4, DamageMult: 1.2,
			AppliesStatus: StatusFrozen, StatusDuration: 1,
			Targeting: TargetEnemy,
		},
		Passive3: Passive{
			Name:        "Sangue de Gas",
			Description: "+25 Mana Máximo e +3 Ataque permanentes.",
			Effect: PassiveExtraAtk, Value: 3,
		},
		Ability1Evo: Ability{
			Name:        "Contrato Eterno",
			Description: "Evolução do Contrato Inteligente — dano ×2.5, congela por 2 turnos e paralisa.",
			ManaCost: 20, Cooldown: 3, DamageMult: 2.5,
			AppliesStatus: StatusFrozen, StatusDuration: 2,
			Targeting: TargetEnemy,
		},
	},

	// ── ARCHER (SOL) ─────────────────────────────────────────────────────────
	"archer": {
		Ability1: Ability{
			Name:        "Snipe Veloz",
			Description: "Ataca com precisão cirúrgica. Dano ×1.8 e ignora metade da defesa do alvo.",
			ManaCost: 15, Cooldown: 2, DamageMult: 1.8, Targeting: TargetEnemy, Unlocked: true,
		},
		Passive1: Passive{
			Name:        "Olho de Falcão",
			Description: "+4 Ataque permanente. Nunca erra o que mira.",
			Effect: PassiveExtraAtk, Value: 4,
		},
		Ability2: Ability{
			Name:        "Chuva de Flechas",
			Description: "Dispara uma sequência rápida — dano ×1.5 com 2 hits consecutivos.",
			ManaCost: 25, Cooldown: 3, DamageMult: 1.5, Targeting: TargetEnemy,
		},
		Passive2: Passive{
			Name:        "Passo do Vento",
			Description: "+3 Velocidade e +10 HP Máximo permanentes.",
			Effect: PassiveExtraHP, Value: 10,
		},
		Ability3: Ability{
			Name:        "Protocolo Sniper",
			Description: "O tiro definitivo — dano ×3.0 e ignora toda CA do alvo.",
			ManaCost: 35, Cooldown: 4, DamageMult: 3.0, Targeting: TargetEnemy,
		},
		Passive3: Passive{
			Name:        "Amplificador Solar",
			Description: "+0.15 no fator de dano crypto. A luz do SOL potencializa cada golpe.",
			Effect: PassiveCryptoAmplify, Value: 0.15,
		},
		Ability1Evo: Ability{
			Name:        "Snipe Lendário",
			Description: "Evolução do Snipe Veloz — dano ×3.5, ignora toda defesa e aplica Paralisia.",
			ManaCost: 15, Cooldown: 2, DamageMult: 3.5,
			AppliesStatus: StatusParalyzed, StatusDuration: 2,
			Targeting: TargetEnemy,
		},
	},

	// ── ROGUE (BNB) ──────────────────────────────────────────────────────────
	"rogue": {
		Ability1: Ability{
			Name:        "Hack & Slash",
			Description: "Injeta veneno nos sistemas do alvo. Causa dano por 3 turnos após o golpe.",
			ManaCost: 10, Cooldown: 4, DamageMult: 1.2,
			AppliesStatus: StatusPoisoned, StatusDuration: 3, StatusPower: 5,
			Targeting: TargetEnemy, Unlocked: true,
		},
		Passive1: Passive{
			Name:        "Mestre do Veneno",
			Description: "Seu veneno causa +50% mais dano. Sistemas nunca se recuperam rápido o suficiente.",
			Effect: PassivePoisonMastery, Value: 1.5,
		},
		Ability2: Ability{
			Name:        "Golpe Sombrio",
			Description: "Ataque rápido das sombras — dano ×2.0 e aplica Paralisia por 1 turno.",
			ManaCost: 20, Cooldown: 3, DamageMult: 2.0,
			AppliesStatus: StatusParalyzed, StatusDuration: 1, StatusPower: 0,
			Targeting: TargetEnemy,
		},
		Passive2: Passive{
			Name:        "Roubo de Vida",
			Description: "Cada ataque cura 15% do dano causado.",
			Effect: PassiveLifesteal, Value: 0.15,
		},
		Ability3: Ability{
			Name:        "Injeção Zero-Day",
			Description: "Explora uma vulnerabilidade crítica — dano ×2.5 e aplica veneno + paralisia simultâneos.",
			ManaCost: 30, Cooldown: 4, DamageMult: 2.5,
			AppliesStatus: StatusPoisoned, StatusDuration: 4, StatusPower: 8,
			Targeting: TargetEnemy,
		},
		Passive3: Passive{
			Name:        "Sombra Perpétua",
			Description: "+5 Ataque e +0.10 fator crypto permanentes.",
			Effect: PassiveCryptoAmplify, Value: 0.10,
		},
		Ability1Evo: Ability{
			Name:        "Colapso de Sistema",
			Description: "Evolução do Hack & Slash — dano ×2.0, veneno por 5 turnos com potência dobrada.",
			ManaCost: 10, Cooldown: 4, DamageMult: 2.0,
			AppliesStatus: StatusPoisoned, StatusDuration: 5, StatusPower: 12,
			Targeting: TargetEnemy,
		},
	},

	// ── SHAMAN (DOGE) ────────────────────────────────────────────────────────
	"shaman": {
		Ability1: Ability{
			Name:        "Lua de Meme",
			Description: "Invoca o caos lunar — aplica BearTrap no inimigo reduzindo seu fator crypto.",
			ManaCost: 25, Cooldown: 3, DamageMult: 1.0,
			AppliesStatus: StatusBearTrap, StatusDuration: 2,
			Targeting: TargetEnemy, Unlocked: true,
		},
		Passive1: Passive{
			Name:        "Espírito Guardião",
			Description: "+2 Defesa e +2 Velocidade permanentes. A Horda protege os seus.",
			Effect: PassiveSpiritGuard, Value: 2,
		},
		Ability2: Ability{
			Name:        "Ritual do Bull Run",
			Description: "Invoca o Bull Run — aplica BullRun em si mesmo por 2 turnos duplicando o fator crypto.",
			ManaCost: 35, Cooldown: 4, DamageMult: 1.0,
			AppliesStatus: StatusBullRun, StatusDuration: 2,
			Targeting: TargetSelf,
		},
		Passive2: Passive{
			Name:        "Amplificador Lunar",
			Description: "+0.20 no fator de dano crypto. A lua amplifica cada ataque.",
			Effect: PassiveCryptoAmplify, Value: 0.20,
		},
		Ability3: Ability{
			Name:        "Tempestade de Memes",
			Description: "Libera o caos total — dano ×2.0 e aplica BearTrap + Burning simultaneamente.",
			ManaCost: 45, Cooldown: 4, DamageMult: 2.0,
			AppliesStatus: StatusBearTrap, StatusDuration: 2,
			Targeting: TargetEnemy,
		},
		Passive3: Passive{
			Name:        "Preço do Sangue",
			Description: "Cada inimigo derrotado restaura 25 HP.",
			Effect: PassiveBloodPrice, Value: 25,
		},
		Ability1Evo: Ability{
			Name:        "Eclipse Primordial",
			Description: "Evolução da Lua de Meme — BearTrap por 4 turnos, dano ×1.8 e aplica Burning.",
			ManaCost: 25, Cooldown: 3, DamageMult: 1.8,
			AppliesStatus: StatusBearTrap, StatusDuration: 4,
			Targeting: TargetEnemy,
		},
	},
}

// ClassAbilities mantém compatibilidade com o código existente
// Retorna a habilidade do slot 1 (inicial) de cada classe
var ClassAbilities = map[string]Ability{
	"warrior": ClassTrees["warrior"].Ability1,
	"mage":    ClassTrees["mage"].Ability1,
	"archer":  ClassTrees["archer"].Ability1,
	"rogue":   ClassTrees["rogue"].Ability1,
	"shaman":  ClassTrees["shaman"].Ability1,
}
