package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// CryptoPrice holds the market data for one coin
type CryptoPrice struct {
	ID         string  `json:"id"`
	Symbol     string  `json:"symbol"`
	PriceUSD   float64 `json:"price_usd"`
	Change7d   float64 `json:"change_7d"`
	Factor     float64 `json:"factor"` // clamped [0.5, 2.0]
	LastUpdate string  `json:"last_update"`
}

// CryptoService fetches, caches and serves CoinGecko data.
// The factor is injected directly into Character.CryptoVariation before each battle.
type CryptoService struct {
	mu        sync.RWMutex
	prices    map[string]CryptoPrice // keyed by coingecko ID
	lastFetch time.Time
	ttl       time.Duration
}

const coinIDs = "bitcoin,ethereum,solana,binancecoin,dogecoin"

// defaultPrices are used as fallback when the API is unavailable
var defaultPrices = map[string]CryptoPrice{
	"bitcoin":     {ID: "bitcoin", Symbol: "BTC", Change7d: 0, Factor: 1.0},
	"ethereum":    {ID: "ethereum", Symbol: "ETH", Change7d: 0, Factor: 1.0},
	"solana":      {ID: "solana", Symbol: "SOL", Change7d: 0, Factor: 1.0},
	"binancecoin": {ID: "binancecoin", Symbol: "BNB", Change7d: 0, Factor: 1.0},
	"dogecoin":    {ID: "dogecoin", Symbol: "DOGE", Change7d: 0, Factor: 1.0},
}

func NewCryptoService() *CryptoService {
	cs := &CryptoService{
		prices: copyDefaults(),
		ttl:    5 * time.Minute,
	}
	// Fetch on startup in background
	go func() {
		if err := cs.fetch(); err != nil {
			log.Printf("[crypto] startup fetch failed, using defaults: %v", err)
		}
	}()
	return cs
}

func copyDefaults() map[string]CryptoPrice {
	m := make(map[string]CryptoPrice, len(defaultPrices))
	for k, v := range defaultPrices {
		m[k] = v
	}
	return m
}

// GetAll returns current prices for all tracked coins
func (cs *CryptoService) GetAll() []CryptoPrice {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	result := make([]CryptoPrice, 0, len(cs.prices))
	for _, p := range cs.prices {
		result = append(result, p)
	}
	return result
}

// GetFactor returns the crypto damage factor for a given CoinGecko ID.
// Falls back to 1.0 if unknown.
func (cs *CryptoService) GetFactor(coinID string) float64 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	if p, ok := cs.prices[coinID]; ok {
		return p.Factor
	}
	return 1.0
}

// GetChange7d returns the 7d % change for a given coin
func (cs *CryptoService) GetChange7d(coinID string) float64 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	if p, ok := cs.prices[coinID]; ok {
		return p.Change7d
	}
	return 0.0
}

// MaybeRefresh fetches fresh data if the TTL has expired
func (cs *CryptoService) MaybeRefresh() {
	cs.mu.RLock()
	stale := time.Since(cs.lastFetch) > cs.ttl
	cs.mu.RUnlock()
	if stale {
		// Roda em goroutine para não bloquear os handlers
		go func() {
			if err := cs.fetch(); err != nil {
				log.Printf("[crypto] refresh failed: %v", err)
			}
		}()
	}
}

func (cs *CryptoService) fetch() error {
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd&include_7d_change=true",
		coinIDs,
	)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		log.Printf("[crypto] rate limited by CoinGecko (429) — mantendo valores anteriores")
		// Atualiza lastFetch para evitar spam de requisições
		cs.mu.Lock()
		cs.lastFetch = time.Now()
		cs.mu.Unlock()
		return fmt.Errorf("rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	// CoinGecko response: {"bitcoin": {"usd": 60000, "usd_7d_change": 3.5}, ...}
	var raw map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	symbols := map[string]string{
		"bitcoin":     "BTC",
		"ethereum":    "ETH",
		"solana":      "SOL",
		"binancecoin": "BNB",
		"dogecoin":    "DOGE",
	}

	now := time.Now().Format("15:04:05")

	cs.mu.Lock()
	defer cs.mu.Unlock()

	for id, data := range raw {
		change7d := data["usd_7d_change"]

		// Amplifica a variação × 3 para ter impacto real no gameplay.
		// Ex: +5% real → fator 1.15 (+15% dano) em vez de 1.05 (+5%)
		// Ex: -8% real → fator 0.76 (-24% dano) — penalidade sentida
		amplified := change7d * 3.0
		factor := 1.0 + (amplified / 100.0)
		if factor < 0.5 {
			factor = 0.5
		}
		if factor > 2.0 {
			factor = 2.0
		}

		cs.prices[id] = CryptoPrice{
			ID:         id,
			Symbol:     symbols[id],
			PriceUSD:   data["usd"],
			Change7d:   change7d,
			Factor:     factor,
			LastUpdate: now,
		}
	}

	cs.lastFetch = time.Now()
	log.Printf("[crypto] prices updated at %s", now)
	return nil
}
