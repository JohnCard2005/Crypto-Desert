package characters

// Ability represents a character's special skill
type Ability struct {
	Name        string
	Description string
	ManaCost    int
	Cooldown    int  // turns between uses (0 = once per battle)
	DamageMult  float64
	AppliesStatus StatusEffect
	StatusDuration int
	StatusPower    int
	Targeting   AbilityTarget
}

type AbilityTarget string

const (
	TargetEnemy AbilityTarget = "enemy"
	TargetSelf  AbilityTarget = "self"
	TargetAll   AbilityTarget = "all_enemies"
)

// ClassAbilities maps class names to their special abilities
var ClassAbilities = map[string]Ability{
	"warrior": {
		Name:        "Fúria do Bloco",
		Description: "Canaliza o poder do bloco genesis — dano triplo em um único golpe devastador.",
		ManaCost:    0,
		Cooldown:    0, // once per battle
		DamageMult:  3.0,
		Targeting:   TargetEnemy,
	},
	"mage": {
		Name:           "Contrato Inteligente",
		Description:    "Executa um contrato que congela o alvo, travando suas ações por 1 turno.",
		ManaCost:       20,
		Cooldown:       3,
		DamageMult:     1.5,
		AppliesStatus:  StatusFrozen,
		StatusDuration: 1,
		Targeting:      TargetEnemy,
	},
	"archer": {
		Name:        "Snipe Veloz",
		Description: "Ataca antes da ordem de turno com precisão cirúrgica. Ignora metade da defesa do alvo.",
		ManaCost:    15,
		Cooldown:    2,
		DamageMult:  1.8,
		Targeting:   TargetEnemy,
	},
	"rogue": {
		Name:           "Hack & Slash",
		Description:    "Injeta veneno nos sistemas do alvo. Causa dano por 3 turnos após o golpe.",
		ManaCost:       10,
		Cooldown:       4,
		DamageMult:     1.2,
		AppliesStatus:  StatusPoisoned,
		StatusDuration: 3,
		StatusPower:    5,
		Targeting:      TargetEnemy,
	},
	"shaman": {
		Name:           "Lua de Meme",
		Description:    "Invoca o caos lunar — buffa o próprio ataque e aplica BearTrap no inimigo.",
		ManaCost:       25,
		Cooldown:       3,
		DamageMult:     1.0,
		AppliesStatus:  StatusBearTrap,
		StatusDuration: 2,
		Targeting:      TargetEnemy,
	},
}
