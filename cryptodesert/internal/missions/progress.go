package missions

import (
	"fmt"
	"time"
)

type WaveStatus string

const (
	WaveStatusLocked    WaveStatus = "locked"
	WaveStatusAvailable WaveStatus = "available"
	WaveStatusCleared   WaveStatus = "cleared"
)

type WaveProgress struct {
	WaveID    string     `json:"wave_id"`
	Status    WaveStatus `json:"status"`
	ClearedAt time.Time  `json:"cleared_at,omitempty"`
	Attempts  int        `json:"attempts"`
}

type MissionProgress struct {
	MissionID  string                   `json:"mission_id"`
	CityID     string                   `json:"city_id"`
	Cleared    bool                     `json:"cleared"`
	ClearedAt  time.Time                `json:"cleared_at,omitempty"`
	Waves      map[string]*WaveProgress `json:"waves"`
	ActiveWave string                   `json:"active_wave"`
	TotalXP    int                      `json:"total_xp"`
	TotalGold  int                      `json:"total_gold"`
}

func (mp *MissionProgress) CurrentWaveIndex(city City) int {
	for i, w := range city.Mission.Waves {
		if w.ID == mp.ActiveWave { return i }
	}
	return 0
}

func (mp *MissionProgress) WavesCleared() int {
	n := 0
	for _, wp := range mp.Waves {
		if wp.Status == WaveStatusCleared { n++ }
	}
	return n
}

type PlayerProgress struct {
	CharacterID     int                               `json:"character_id"`
	CharacterName   string                            `json:"character_name"`
	CurrentDifficulty Difficulty                      `json:"current_difficulty"`
	HighestClear    Difficulty                        `json:"highest_clear"`
	MissionProgress map[Difficulty]map[string]*MissionProgress `json:"mission_progress"`
	BattleHistory   []BattleRecord                    `json:"battle_history"`
	CreatedAt       time.Time                         `json:"created_at"`
	UpdatedAt       time.Time                         `json:"updated_at"`
}

type BattleRecord struct {
	CityID     string    `json:"city_id"`
	WaveID     string    `json:"wave_id"`
	EnemyName  string    `json:"enemy_name"`
	Won        bool      `json:"won"`
	TurnCount  int       `json:"turn_count"`
	XPGained   int       `json:"xp_gained"`
	GoldGained int       `json:"gold_gained"`
	At         time.Time `json:"at"`
}

// CitySummary — json tags adicionadas para serialização correta
type CitySummary struct {
	City         City `json:"city"`
	Unlocked     bool `json:"unlocked"`
	Cleared      bool `json:"cleared"`
	WavesTotal   int  `json:"waves_total"`
	WavesCleared int  `json:"waves_cleared"`
	TotalXP      int  `json:"total_xp"`
	TotalGold    int  `json:"total_gold"`
	Difficulty   Difficulty `json:"difficulty"`
}

func NewPlayerProgress(charID int, charName string) *PlayerProgress {
	pp := &PlayerProgress{
		CharacterID:       charID,
		CharacterName:     charName,
		CurrentDifficulty: DifficultyNormal,
		HighestClear:      Difficulty(-1),
		MissionProgress:   make(map[Difficulty]map[string]*MissionProgress),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	pp.initDifficulty(DifficultyNormal)
	return pp
}

func (pp *PlayerProgress) initDifficulty(d Difficulty) {
	if _, ok := pp.MissionProgress[d]; ok { return }
	pp.MissionProgress[d] = make(map[string]*MissionProgress)
	for _, city := range Campaign {
		mp := &MissionProgress{
			MissionID:  city.Mission.ID,
			CityID:     city.ID,
			Waves:      make(map[string]*WaveProgress),
			ActiveWave: city.Mission.Waves[0].ID,
		}
		for i, wave := range city.Mission.Waves {
			status := WaveStatusLocked
			// Primeira wave da primeira cidade sempre disponível em qualquer dificuldade
			if city.UnlockedBy == "" && i == 0 {
				status = WaveStatusAvailable
			}
			mp.Waves[wave.ID] = &WaveProgress{WaveID: wave.ID, Status: status}
		}
		pp.MissionProgress[d][city.Mission.ID] = mp
	}
}


// EnsureDifficulty garante que a dificuldade atual está inicializada.
// Chamado ao carregar progresso do banco para garantir consistência.
func (pp *PlayerProgress) EnsureDifficulty() {
	pp.initDifficulty(pp.CurrentDifficulty)
}

func (pp *PlayerProgress) MissionState(missionID string) (*MissionProgress, error) {
	missions, ok := pp.MissionProgress[pp.CurrentDifficulty]
	if !ok { return nil, fmt.Errorf("dificuldade %s não inicializada", pp.CurrentDifficulty.Label()) }
	mp, ok := missions[missionID]
	if !ok { return nil, fmt.Errorf("missão %q não encontrada", missionID) }
	return mp, nil
}

func (pp *PlayerProgress) CityUnlocked(city City) bool {
	if city.UnlockedBy == "" { return true }
	for _, c := range Campaign {
		if c.ID == city.UnlockedBy {
			mp, err := pp.MissionState(c.Mission.ID)
			if err != nil { return false }
			return mp.Cleared
		}
	}
	return false
}

func (pp *PlayerProgress) WaveUnlocked(missionID, waveID string) bool {
	mp, err := pp.MissionState(missionID)
	if err != nil { return false }
	wp, ok := mp.Waves[waveID]
	if !ok { return false }
	return wp.Status == WaveStatusAvailable || wp.Status == WaveStatusCleared
}

func (pp *PlayerProgress) ActiveWaveFor(city City) (*Wave, *WaveProgress, error) {
	mp, err := pp.MissionState(city.Mission.ID)
	if err != nil { return nil, nil, err }
	for i := range city.Mission.Waves {
		w := &city.Mission.Waves[i]
		wp := mp.Waves[w.ID]
		if wp != nil && wp.Status == WaveStatusAvailable { return w, wp, nil }
	}
	// Nenhuma wave Available — missão concluída normalmente (não em replay)
	if mp.Cleared { return nil, nil, fmt.Errorf("missão já concluída. Use replay") }
	return nil, nil, fmt.Errorf("nenhuma wave disponível")
}

func (pp *PlayerProgress) CampaignSummary() []CitySummary {
	result := make([]CitySummary, 0, len(Campaign))
	for _, city := range Campaign {
		mp, _ := pp.MissionState(city.Mission.ID)
		cs := CitySummary{
			City:       city,
			Unlocked:   pp.CityUnlocked(city),
			Difficulty: pp.CurrentDifficulty,
		}
		if mp != nil {
			cs.Cleared = mp.Cleared
			cs.WavesTotal = len(city.Mission.Waves)
			cs.WavesCleared = mp.WavesCleared()
			cs.TotalXP = mp.TotalXP
			cs.TotalGold = mp.TotalGold
		}
		result = append(result, cs)
	}
	return result
}

func (pp *PlayerProgress) RecordWaveCleared(city City, waveID string, xp, gold int) {
	mp, err := pp.MissionState(city.Mission.ID)
	if err != nil { return }
	if wp, ok := mp.Waves[waveID]; ok {
		wp.Status = WaveStatusCleared
		wp.ClearedAt = time.Now()
	}
	mp.TotalXP += xp
	mp.TotalGold += gold
	pp.UpdatedAt = time.Now()
	for i, w := range city.Mission.Waves {
		if w.ID == waveID {
			if i+1 < len(city.Mission.Waves) {
				nextID := city.Mission.Waves[i+1].ID
				if wp, ok := mp.Waves[nextID]; ok {
					// Só desbloqueia a próxima se ainda estiver Locked
					// Em replay, waves já concluídas (Cleared/Available) não devem ser rebaixadas
					if wp.Status == WaveStatusLocked {
						wp.Status = WaveStatusAvailable
					}
				}
				// Só atualiza ActiveWave se a atual não era mais avançada
				if mp.ActiveWave == waveID {
					mp.ActiveWave = nextID
				}
			} else {
				// Só registra mission cleared se ainda não tinha sido concluída antes
				if !mp.Cleared {
					pp.recordMissionCleared(city, mp)
				}
			}
			break
		}
	}
}

func (pp *PlayerProgress) RecordWaveAttempt(missionID, waveID string) {
	mp, err := pp.MissionState(missionID)
	if err != nil { return }
	if wp, ok := mp.Waves[waveID]; ok { wp.Attempts++ }
}

func (pp *PlayerProgress) recordMissionCleared(city City, mp *MissionProgress) {
	mp.Cleared = true
	mp.ClearedAt = time.Now()
	for _, nextCity := range Campaign {
		if nextCity.UnlockedBy == city.ID {
			nextMP, err := pp.MissionState(nextCity.Mission.ID)
			if err != nil { continue }
			if len(nextCity.Mission.Waves) > 0 {
				firstWaveID := nextCity.Mission.Waves[0].ID
				if wp, ok := nextMP.Waves[firstWaveID]; ok {
					wp.Status = WaveStatusAvailable
					nextMP.ActiveWave = firstWaveID
				}
			}
		}
	}
	pp.checkCampaignClear()
}

func (pp *PlayerProgress) checkCampaignClear() {
	missions := pp.MissionProgress[pp.CurrentDifficulty]
	for _, mp := range missions {
		if !mp.Cleared { return }
	}
	if pp.CurrentDifficulty > pp.HighestClear { pp.HighestClear = pp.CurrentDifficulty }
	next := pp.CurrentDifficulty + 1
	if next <= DifficultyNGPPP { pp.initDifficulty(next) }
}

func (pp *PlayerProgress) StartNewGame() error {
	next := pp.CurrentDifficulty + 1
	if next > DifficultyNGPPP { return fmt.Errorf("já está no nível máximo") }
	pp.CurrentDifficulty = next
	pp.initDifficulty(next)
	pp.UpdatedAt = time.Now()
	return nil
}

func (pp *PlayerProgress) ReplayMission(city City) error {
	mp, err := pp.MissionState(city.Mission.ID)
	if err != nil { return err }
	// Permite replay se ao menos 1 wave foi concluída (não precisa ter matado o boss)
	hasCleared := false
	for _, wp := range mp.Waves {
		if wp.Status == WaveStatusCleared {
			hasCleared = true
			break
		}
	}
	if !hasCleared {
		return fmt.Errorf("complete ao menos uma wave antes de repetir a missão")
	}
	// Replay: todas as waves ficam Available para o jogador escolher.
	// mp.Cleared permanece true — isso garante que cidades desbloqueadas
	// pela conclusão desta missão continuam acessíveis.
	for _, wave := range city.Mission.Waves {
		mp.Waves[wave.ID] = &WaveProgress{WaveID: wave.ID, Status: WaveStatusAvailable}
	}
	// mp.Cleared NÃO é alterado — preserva desbloqueio de cidades seguintes
	mp.ActiveWave = city.Mission.Waves[0].ID
	pp.UpdatedAt = time.Now()
	return nil
}

func (pp *PlayerProgress) AddBattleRecord(r BattleRecord) {
	pp.BattleHistory = append(pp.BattleHistory, r)
	if len(pp.BattleHistory) > 100 {
		pp.BattleHistory = pp.BattleHistory[len(pp.BattleHistory)-100:]
	}
}

func (pp *PlayerProgress) AvailableDifficulties() []Difficulty {
	max := pp.HighestClear + 1
	if max > DifficultyNGPPP { max = DifficultyNGPPP }
	result := make([]Difficulty, 0)
	for d := Difficulty(0); d <= max; d++ { result = append(result, d) }
	return result
}
