package characters

import "fmt"

// classBase holds the base template for each class
type classBase struct {
	faction     Faction
	cryptoID    string
	hp          int
	mana        int
	attackMod   int
	strengthMod int
	ca          int
	defense     int
	speed       int
	damageDice  int // sides on damage die
}

var classBases = map[string]classBase{
	"warrior": {
		faction:     FactionBTC,
		cryptoID:    "bitcoin",
		hp:          120,
		mana:        0,
		attackMod:   4,
		strengthMod: 3,
		ca:          14,
		defense:     2,
		speed:       1,
		damageDice:  10,
	},
	"mage": {
		faction:     FactionETH,
		cryptoID:    "ethereum",
		hp:          80,
		mana:        80,
		attackMod:   6,
		strengthMod: 1,
		ca:          11,
		defense:     0,
		speed:       2,
		damageDice:  6,
	},
	"archer": {
		faction:     FactionSOL,
		cryptoID:    "solana",
		hp:          95,
		mana:        40,
		attackMod:   5,
		strengthMod: 2,
		ca:          13,
		defense:     1,
		speed:       3,
		damageDice:  8,
	},
	"rogue": {
		faction:     FactionBNB,
		cryptoID:    "binancecoin",
		hp:          90,
		mana:        50,
		attackMod:   5,
		strengthMod: 2,
		ca:          12,
		defense:     1,
		speed:       4,
		damageDice:  8,
	},
	"shaman": {
		faction:     FactionDOGE,
		cryptoID:    "dogecoin",
		hp:          100,
		mana:        60,
		attackMod:   3,
		strengthMod: 3,
		ca:          12,
		defense:     1,
		speed:       2,
		damageDice:  8,
	},
}

// NewCharacter creates a level-1 character of the given class.
// Returns an error if the class is unknown.
func NewCharacter(name string, class string) (*Character, error) {
	base, ok := classBases[class]
	if !ok {
		return nil, fmt.Errorf("unknown class %q — valid classes: warrior, mage, archer, rogue, shaman", class)
	}

	ability, ok := ClassAbilities[class]
	if !ok {
		ability = Ability{Name: "None", Description: "No special ability."}
	}

	c := &Character{
		Name:        name,
		Class:       class,
		Faction:     base.faction,
		Level:       1,
		XP:          0,
		XPToNext:    XPToNextLevel(1),
		HP:          base.hp,
		MaxHP:       base.hp,
		Mana:        base.mana,
		MaxMana:     base.mana,
		AttackMod:   base.attackMod,
		StrengthMod: base.strengthMod,
		CA:          base.ca,
		Defense:     base.defense,
		Speed:       base.speed,
		DamageDice:  base.damageDice,
		CryptoID:    base.cryptoID,
		Ability:     ability,
		Alive:       true,
	}

	return c, nil
}

// NewCharacterAtLevel creates a character already leveled up to the target level.
func NewCharacterAtLevel(name string, class string, level int) (*Character, error) {
	c, err := NewCharacter(name, class)
	if err != nil {
		return nil, err
	}

	for i := 2; i <= level; i++ {
		c.applyLevelUp()
	}
	c.Level = level
	c.XP = TotalXPForLevel(level)
	c.XPToNext = XPToNextLevel(level)

	// Restore to full after leveling
	c.HP = c.MaxHP
	c.Mana = c.MaxMana

	return c, nil
}

// ValidClasses returns all available class names
func ValidClasses() []string {
	classes := make([]string, 0, len(classBases))
	for k := range classBases {
		classes = append(classes, k)
	}
	return classes
}
