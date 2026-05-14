package characters

// Character is the core entity of the game.
// It holds all combat stats, progression state and active conditions.
type Character struct {
	// Identity
	ID      int
	Name    string
	Class   string
	Faction Faction

	// Progression
	Level   int
	XP      int
	XPToNext int

	// Vitals
	HP      int
	MaxHP   int
	Mana    int
	MaxMana int

	// Combat stats
	AttackMod   int
	StrengthMod int
	CA          int  // Armor Class — base defense threshold
	Defense     int  // damage reduction after hit
	Speed       int  // initiative modifier (dex equivalent)
	DamageDice  int  // number of sides on damage die (d6, d8, d10...)

	// Crypto link
	CryptoID    string  // CoinGecko API ID (e.g. "bitcoin")
	CryptoVariation float64 // cached 7d % change, populated before battle

	// Special ability
	Ability             Ability
	AbilityUsed         bool // once-per-battle abilities
	AbilityCooldownLeft int  // turns remaining on cooldown

	// Active status conditions
	Statuses []ActiveStatus

	// State
	Alive     bool
	Defending bool // true if player chose Defend action this turn
}

// StatSummary is a lightweight snapshot for UI display / serialization
type StatSummary struct {
	Name    string
	Class   string
	Faction string
	Level   int
	HP      int
	MaxHP   int
	Mana    int
	MaxMana int
	XP      int
	XPToNext int
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
	}
}
