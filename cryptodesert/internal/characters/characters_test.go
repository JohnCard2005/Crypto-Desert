package characters_test

import (
	"crypto-desert/internal/characters"
	"fmt"
	"testing"
)

func TestNewCharacter(t *testing.T) {
	tests := []struct {
		class   string
		wantHP  int
		faction characters.Faction
	}{
		{"warrior", 120, characters.FactionBTC},
		{"mage", 80, characters.FactionETH},
		{"archer", 95, characters.FactionSOL},
		{"rogue", 90, characters.FactionBNB},
		{"shaman", 100, characters.FactionDOGE},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			c, err := characters.NewCharacter("Test", tt.class)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.HP != tt.wantHP {
				t.Errorf("HP: got %d, want %d", c.HP, tt.wantHP)
			}
			if c.Faction != tt.faction {
				t.Errorf("Faction: got %q, want %q", c.Faction, tt.faction)
			}
			if c.Level != 1 {
				t.Errorf("Level: got %d, want 1", c.Level)
			}
			if !c.Alive {
				t.Error("character should be alive")
			}
		})
	}
}

func TestUnknownClass(t *testing.T) {
	_, err := characters.NewCharacter("Ghost", "ninja")
	if err == nil {
		t.Error("expected error for unknown class, got nil")
	}
}

func TestTakeDamage(t *testing.T) {
	c, _ := characters.NewCharacter("Tank", "warrior")
	c.TakeDamage(50)
	if c.HP != 70 {
		t.Errorf("HP after 50 damage: got %d, want 70", c.HP)
	}
	if !c.IsAlive() {
		t.Error("should still be alive")
	}

	c.TakeDamage(200)
	if c.HP != 0 {
		t.Errorf("HP should floor at 0, got %d", c.HP)
	}
	if c.IsAlive() {
		t.Error("should be dead after lethal damage")
	}
}

func TestHeal(t *testing.T) {
	c, _ := characters.NewCharacter("Healer", "warrior")
	c.TakeDamage(50)
	c.Heal(30)
	if c.HP != 100 {
		t.Errorf("HP after heal: got %d, want 100", c.HP)
	}
	c.Heal(9999)
	if c.HP != c.MaxHP {
		t.Errorf("HP should cap at MaxHP %d, got %d", c.MaxHP, c.HP)
	}
}

func TestXPAndLeveling(t *testing.T) {
	c, _ := characters.NewCharacter("Hero", "warrior")
	startHP := c.MaxHP
	startAttack := c.AttackMod

	// Give enough XP to reach level 2 (needs 100 XP)
	levels := c.GainXP(100)
	if levels != 1 {
		t.Errorf("expected 1 level gain, got %d", levels)
	}
	if c.Level != 2 {
		t.Errorf("expected level 2, got %d", c.Level)
	}
	if c.MaxHP <= startHP {
		t.Errorf("MaxHP should have increased from %d", startHP)
	}
	if c.AttackMod <= startAttack {
		t.Errorf("AttackMod should have increased from %d", startAttack)
	}
}

func TestNewCharacterAtLevel(t *testing.T) {
	c, err := characters.NewCharacterAtLevel("Veteran", "mage", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Level != 10 {
		t.Errorf("expected level 10, got %d", c.Level)
	}
	if c.HP != c.MaxHP {
		t.Errorf("HP should be full after creation, got %d/%d", c.HP, c.MaxHP)
	}
}

func TestStatusEffects(t *testing.T) {
	c, _ := characters.NewCharacter("Mage", "mage")

	c.AddStatus(characters.StatusFrozen, 2, 0)
	if !c.HasStatus(characters.StatusFrozen) {
		t.Error("should have frozen status")
	}

	// Frozen character should skip turn
	if !c.ShouldSkipTurn() {
		t.Error("frozen character should always skip turn")
	}

	// Tick once — should still have 1 duration left
	c.TickStatuses()
	if !c.HasStatus(characters.StatusFrozen) {
		t.Error("frozen status should still be active after 1 tick")
	}

	// Tick again — should expire
	c.TickStatuses()
	if c.HasStatus(characters.StatusFrozen) {
		t.Error("frozen status should have expired after 2 ticks")
	}
}

func TestPoisonDamage(t *testing.T) {
	c, _ := characters.NewCharacter("Victim", "warrior")
	c.AddStatus(characters.StatusPoisoned, 3, 10)

	hpBefore := c.HP
	c.TickStatuses()
	if c.HP != hpBefore-10 {
		t.Errorf("poison should deal 10 damage, HP went from %d to %d", hpBefore, c.HP)
	}
}

func TestAbilityUse(t *testing.T) {
	c, _ := characters.NewCharacter("Fighter", "warrior")

	// Warrior ability: once per battle, no mana cost
	if !c.CanUseAbility() {
		t.Error("ability should be available at start")
	}
	if !c.UseAbility() {
		t.Error("first use should succeed")
	}
	if c.CanUseAbility() {
		t.Error("once-per-battle ability should not be reusable")
	}

	c.ResetForNewBattle()
	if !c.CanUseAbility() {
		t.Error("ability should reset for new battle")
	}
}

func TestDefendBonus(t *testing.T) {
	c, _ := characters.NewCharacter("Guard", "warrior")
	baseCA := c.EffectiveCA()
	c.Defending = true
	if c.EffectiveCA() != baseCA+4 {
		t.Errorf("defending should add +4 CA, got %d (base %d)", c.EffectiveCA(), baseCA)
	}
}

func TestFactionInfo(t *testing.T) {
	for _, faction := range []characters.Faction{
		characters.FactionBTC, characters.FactionETH,
		characters.FactionSOL, characters.FactionBNB, characters.FactionDOGE,
	} {
		info, ok := characters.Factions[faction]
		if !ok {
			t.Errorf("faction %q has no info", faction)
		}
		if info.Name == "" {
			t.Errorf("faction %q has empty name", faction)
		}
	}
}

// Quick smoke test — prints character sheets for visual inspection
func TestPrintCharacterSheet(t *testing.T) {
	for _, class := range characters.ValidClasses() {
		c, err := characters.NewCharacter("Test_"+class, class)
		if err != nil {
			t.Errorf("class %q failed: %v", class, err)
			continue
		}
		faction := characters.Factions[c.Faction]
		fmt.Printf("\n[%s] %s — %s (%s)\n", c.Class, c.Name, faction.Name, c.Faction)
		fmt.Printf("  HP:%d  Mana:%d  CA:%d  Def:%d  Spd:%d  d%d\n",
			c.MaxHP, c.MaxMana, c.CA, c.Defense, c.Speed, c.DamageDice)
		fmt.Printf("  ATK:+%d  STR:+%d  Crypto:%s\n", c.AttackMod, c.StrengthMod, c.CryptoID)
		fmt.Printf("  Ability: %s — %s\n", c.Ability.Name, c.Ability.Description)
		fmt.Printf("  XP to lv2: %d\n", c.XPToNext)
	}
}
