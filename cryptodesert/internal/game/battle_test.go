package game_test

import (
	"fmt"
	"testing"

	"crypto-desert/internal/characters"
	"crypto-desert/internal/combat"
	"crypto-desert/internal/enemies"
	"crypto-desert/internal/game"
)

func makePlayer(name, class string) *characters.Character {
	c, err := characters.NewCharacter(name, class)
	if err != nil { panic(err) }
	return c
}

func makeEnemy(templateName string) *enemies.Enemy {
	e, err := enemies.Spawn(templateName)
	if err != nil { panic(err) }
	return e
}

func logTurn(t *testing.T, result game.TurnResult) {
	t.Helper()
	for _, ev := range result.Events {
		t.Logf("  %s", ev.Message)
	}
}

func TestInitiativeOrder(t *testing.T) {
	warrior := makePlayer("Kabom", "warrior")
	mage := makePlayer("Zeta", "mage")
	enemy := makeEnemy("Especulador Novato")

	battle := game.NewBattle([]*characters.Character{warrior, mage}, []*enemies.Enemy{enemy})
	order := battle.InitiativeOrder()

	if len(order) != 3 {
		t.Fatalf("expected 3 combatants, got %d", len(order))
	}
	for i := 1; i < len(order); i++ {
		if order[i].Initiative > order[i-1].Initiative {
			t.Errorf("initiative not sorted at index %d", i)
		}
	}
	t.Log("Turn order:")
	for i, e := range order {
		t.Logf("  %d. %s — initiative %d", i+1, e.Combatant.GetName(), e.Initiative)
	}
}

func TestPlayerDefend(t *testing.T) {
	p := makePlayer("Guard", "warrior")
	e := makeEnemy("Especulador Novato")
	battle := game.NewBattle([]*characters.Character{p}, []*enemies.Enemy{e})

	baseCA := p.CA
	result := battle.ProcessPlayerTurn(p, game.PlayerAction{Type: game.ActionDefend})
	logTurn(t, result)

	if !p.HasStatus(characters.StatusDefending) {
		t.Error("player should have StatusDefending after defend action")
	}
	if p.EffectiveCA() != baseCA+4 {
		t.Errorf("expected CA %d, got %d", baseCA+4, p.EffectiveCA())
	}
}

func TestPlayerFlee(t *testing.T) {
	successes, failures := 0, 0
	for i := 0; i < 100; i++ {
		p, _ := characters.NewCharacter("Coward", "rogue")
		e := makeEnemy("Especulador Novato")
		b := game.NewBattle([]*characters.Character{p}, []*enemies.Enemy{e})
		result := b.ProcessPlayerTurn(p, game.PlayerAction{Type: game.ActionFlee})
		if result.BattleState == game.BattlePlayerFled {
			successes++
		} else {
			failures++
		}
	}
	t.Logf("Flee: %d successes, %d failures over 100 attempts", successes, failures)
	if successes == 0 { t.Error("never succeeded fleeing") }
	if failures == 0 { t.Error("never failed to flee") }
}

func TestPlayerAbility(t *testing.T) {
	warrior, _ := characters.NewCharacterAtLevel("Kabom", "warrior", 3)
	e := makeEnemy("Especulador Novato")
	battle := game.NewBattle([]*characters.Character{warrior}, []*enemies.Enemy{e})

	result := battle.ProcessPlayerTurn(warrior, game.PlayerAction{
		Type: game.ActionAbility, Target: e.Character,
	})
	logTurn(t, result)

	if warrior.CanUseAbility() {
		t.Error("once-per-battle ability should be unavailable after use")
	}

	result2 := battle.ProcessPlayerTurn(warrior, game.PlayerAction{
		Type: game.ActionAbility, Target: e.Character,
	})
	hasError := false
	for _, ev := range result2.Events {
		if ev.IsError { hasError = true }
	}
	if !hasError {
		t.Error("second ability use should produce an error event")
	}
}

func TestAIBehaviors(t *testing.T) {
	tests := []struct {
		name     string
		behavior enemies.AIBehavior
	}{
		{"Bot de Pump", enemies.BehaviorRandom},
		{"Minerador Fantasma", enemies.BehaviorDefensive},
		{"Whale Corrupta", enemies.BehaviorBerserker},
		{"Oráculo Corrompido", enemies.BehaviorSupport},
		{"Dust Raider", enemies.BehaviorAggressive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := characters.NewCharacter("Tester", "warrior")
			p.HP = 9999; p.MaxHP = 9999
			e := makeEnemy(tt.name)
			if e.Behavior != tt.behavior {
				t.Errorf("expected behavior %s, got %s", tt.behavior, e.Behavior)
			}
			battle := game.NewBattle([]*characters.Character{p}, []*enemies.Enemy{e})
			for i := 0; i < 5; i++ {
				result := battle.ProcessEnemyTurn(e)
				if len(result.Events) == 0 {
					t.Errorf("turn %d produced no events", i+1)
				}
				t.Logf("[%s] T%d: %s", tt.name, i+1, result.Events[0].Message)
			}
		})
	}
}

func TestXPDistributionOnWin(t *testing.T) {
	warrior, _ := characters.NewCharacterAtLevel("Kabom", "warrior", 5)
	startXP := warrior.XP

	e := makeEnemy("Especulador Novato")
	e.HP = 1

	battle := game.NewBattle([]*characters.Character{warrior}, []*enemies.Enemy{e})
	logTurn(t, battle.ProcessPlayerTurn(warrior, game.PlayerAction{
		Type: game.ActionAttack, Target: e.Character,
	}))

	concluded := battle.Conclude()
	if concluded.Status == game.BattlePlayerWon && warrior.XP <= startXP {
		t.Errorf("XP should have increased: was %d, now %d", startXP, warrior.XP)
	}
	t.Logf("XP gained: %d → level %d", concluded.XPGained, warrior.Level)
}

func TestEnemyCatalogue(t *testing.T) {
	for _, tmpl := range enemies.Catalogue {
		t.Run(tmpl.Name, func(t *testing.T) {
			e, err := enemies.Spawn(tmpl.Name)
			if err != nil { t.Fatalf("spawn failed: %v", err) }
			if !e.IsAlive() { t.Error("should be alive") }
			if e.HP <= 0 { t.Errorf("HP should be > 0, got %d", e.HP) }
			t.Logf("[%s] %s — HP:%d CA:%d ATK:+%d BEH:%s", tmpl.Tier, e.Name, e.HP, e.CA, e.AttackMod, e.Behavior)
		})
	}
}

func TestSpawnGroup(t *testing.T) {
	group, err := enemies.SpawnGroup(5, 3)
	if err != nil { t.Fatalf("SpawnGroup failed: %v", err) }
	if len(group) != 3 { t.Errorf("expected 3 enemies, got %d", len(group)) }
	for _, e := range group {
		t.Logf("Spawned: %s (level %d, tier %s)", e.Name, e.Level, e.Tier)
	}
}

func TestFullBattleSmoke(t *testing.T) {
	warrior, _ := characters.NewCharacterAtLevel("Kabom", "warrior", 5)
	mage, _ := characters.NewCharacterAtLevel("Zeta", "mage", 4)
	enemy1 := makeEnemy("Whale Corrupta")
	enemy2 := makeEnemy("Bot de Pump")

	players := []*characters.Character{warrior, mage}
	battleEnemies := []*enemies.Enemy{enemy1, enemy2}

	battle := game.NewBattle(players, battleEnemies)

	fmt.Println("\n═══ CRYPTO DESERT BATTLE START ═══")
	for i, e := range battle.InitiativeOrder() {
		fmt.Printf("  %d. %s (initiative %d)\n", i+1, e.Combatant.GetName(), e.Initiative)
	}

	for battle.Status == game.BattleOngoing {
		if battle.TurnNumber >= 80 { t.Log("max turns reached"); break }

		current := battle.Queue.Current()
		if current == nil { break }

		var result game.TurnResult

		if current.GetTeam() == combat.TeamPlayer {
			var actor *characters.Character
			for _, p := range players {
				if p.Name == current.GetName() && p.IsAlive() { actor = p; break }
			}
			if actor == nil { battle.Queue.Advance(); continue }

			var target *characters.Character
			for _, e := range battleEnemies {
				if e.IsAlive() { target = e.Character; break }
			}
			result = battle.ProcessPlayerTurn(actor, game.PlayerAction{
				Type: game.ActionAttack, Target: target,
			})
		} else {
			for _, e := range battleEnemies {
				if e.Name == current.GetName() && e.IsAlive() {
					result = battle.ProcessEnemyTurn(e); break
				}
			}
		}

		for _, ev := range result.Events {
			prefix := "· "
			if ev.Damage > 0 { prefix = "⚔ " }
			fmt.Printf("  %s%s\n", prefix, ev.Message)
		}
		battle.Queue.Advance()
	}

	concluded := battle.Conclude()
	fmt.Printf("\n═══ RESULTADO: %s ═══\n", concluded.Status)
	fmt.Printf("Turnos: %d | XP: %d | Ouro: %d\n", concluded.TurnCount, concluded.XPGained, concluded.GoldGained)
	for _, s := range concluded.Survivors {
		fmt.Printf("Sobrevivente: %s — HP %d/%d | Nível %d\n", s.Name, s.HP, s.MaxHP, s.Level)
	}
}
