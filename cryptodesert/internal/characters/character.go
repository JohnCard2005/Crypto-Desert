package characters

// Character is the core entity of the game.
// It holds all combat stats, progression state and active conditions.
type Character struct {
	// Identity
	ID      int
	UserID  int    // dono do personagem (usado pelo sistema de auth)
	Name    string
	Class   string
	Faction Faction

	// Progression
	Level    int
	XP       int
	XPToNext int

	// Economy
	Gold int // moeda ganha ao derrotar inimigos; gasta na loja e no campfire

	// Vitals
	HP      int
	MaxHP   int
	Mana    int
	MaxMana int

	// Base combat stats (sem bônus de equipamento)
	AttackMod   int
	StrengthMod int
	CA          int  // Armor Class — base defense threshold
	Defense     int  // damage reduction after hit
	Speed       int  // initiative modifier (dex equivalent)
	DamageDice  int  // number of sides on damage die (d6, d8, d10...)

	// Equipment bonuses — preenchidos pelo items.Inventory.TotalBonuses()
	// antes de cada batalha ou sempre que o equipamento muda.
	BonusAttackMod    int
	BonusStrengthMod  int
	BonusCA           int
	BonusDefense      int
	BonusSpeed        int
	BonusMaxHP        int
	BonusMaxMana      int
	CryptoFactorBonus float64 // somado ao CryptoVariation antes do clamp

	// Crypto link
	CryptoID        string  // CoinGecko API ID (e.g. "bitcoin")
	CryptoVariation float64 // cached 7d % change, populated before battle

	// Habilidade ativa (slot 1 — principal)
	Ability             Ability
	AbilityUsed         bool
	AbilityCooldownLeft int

	// Slots de habilidades adicionais (desbloqueados ao subir de nível)
	Ability2             Ability
	Ability2Used         bool
	Ability2CooldownLeft int

	Ability3             Ability
	Ability3Used         bool
	Ability3CooldownLeft int

	// Slots de passivas (desbloqueadas ao subir de nível)
	Passive1 Passive
	Passive2 Passive
	Passive3 Passive

	// Flags de efeitos de passivas aplicados permanentemente
	IronWillUsed bool // Vontade de Ferro — usada nessa batalha

	// Active status conditions
	Statuses []ActiveStatus

	// State
	Alive     bool
	Defending bool // true if player chose Defend action this turn
}

// StatSummary is a lightweight snapshot for UI display / serialization
type StatSummary struct {
	Name     string
	Class    string
	Faction  string
	Level    int
	HP       int
	MaxHP    int
	Mana     int
	MaxMana  int
	XP       int
	XPToNext int
	Gold     int
}

func (c *Character) Summary() StatSummary {
	return StatSummary{
		Name:     c.Name,
		Class:    c.Class,
		Faction:  string(c.Faction),
		Level:    c.Level,
		HP:       c.HP,
		MaxHP:    c.MaxHP,
		Mana:     c.Mana,
		MaxMana:  c.MaxMana,
		XP:       c.XP,
		XPToNext: c.XPToNext,
		Gold:     c.Gold,
	}
}

// EffectiveAttackMod retorna o attack modifier total incluindo equipamentos e status
func (c *Character) EffectiveAttackMod() int {
	mods := c.GetCombatModifiers()
	return c.AttackMod + c.BonusAttackMod + mods.AttackBonus
}

// EffectiveCA retorna a CA total incluindo equipamentos, status e postura defensiva
func (c *Character) EffectiveCA() int {
	mods := c.GetCombatModifiers()
	ca := c.CA + c.BonusCA + mods.DefenseBonus
	if c.Defending {
		ca += 4
	}
	return ca
}

// EffectiveDefense retorna a defesa total incluindo equipamentos
func (c *Character) EffectiveDefense() int {
	return c.Defense + c.BonusDefense
}

// EffectiveStrengthMod retorna a força total incluindo equipamentos
func (c *Character) EffectiveStrengthMod() int {
	return c.StrengthMod + c.BonusStrengthMod
}

// EffectiveSpeed retorna a velocidade total incluindo equipamentos
func (c *Character) EffectiveSpeed() int {
	return c.Speed + c.BonusSpeed
}
