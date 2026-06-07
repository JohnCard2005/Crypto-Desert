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
	Change24h  float64 `json:"change_24h"`
	Change7d   float64 `json:"change_7d"`
	Factor     float64 `json:"factor"`
	LastUpdate time.Time `json:"last_update"`
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
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true&include_7d_change=true",
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

	now := time.Now()

	// Captura snapshot dos preços anteriores ANTES de bloquear
	cs.mu.RLock()
	prevPrices := make(map[string]CryptoPrice, len(cs.prices))
	for k, v := range cs.prices { prevPrices[k] = v }
	cs.mu.RUnlock()

	cs.mu.Lock()
	defer cs.mu.Unlock()

	for id, data := range raw {
		change7d  := data["usd_7d_change"]
		change24h := data["usd_24h_change"]
		currentPrice := data["usd"]

		// Log para debug — mostra o que a CoinGecko retornou
		log.Printf("[crypto] %s: price=%.2f 24h=%.2f%% 7d=%.2f%%",
			symbols[id], currentPrice, change24h, change7d)

		// Se a CoinGecko retornou 0 em ambas as variações mas temos preço anterior,
		// calcula a variação pelo preço atual vs preço anterior guardado
		if change7d == 0 && change24h == 0 {
			if prev, hasPrev := prevPrices[id]; hasPrev && prev.PriceUSD > 0 && currentPrice > 0 {
				// Usa variação desde o último preço conhecido como proxy
				estimated := ((currentPrice - prev.PriceUSD) / prev.PriceUSD) * 100.0
				change24h = estimated
				log.Printf("[crypto] %s: variação estimada pelo preço: %.2f%%", symbols[id], estimated)
			}
		}

		// Blend: 60% peso no 24h (mais imediato) + 40% no 7d (tendência)
		blend := (change24h * 0.6) + (change7d * 0.4)

		// Amplifica × 5 — queda de 10% real → fator 0.50 (impacto claro no jogo)
		amplified := blend * 5.0
		factor := 1.0 + (amplified / 100.0)
		if factor < 0.4 { factor = 0.4 }
		if factor > 2.5 { factor = 2.5 }

		log.Printf("[crypto] %s: blend=%.2f%% fator=%.3f", symbols[id], blend, factor)

		cs.prices[id] = CryptoPrice{
			ID:         id,
			Symbol:     symbols[id],
			PriceUSD:   currentPrice,
			Change24h:  change24h,
			Change7d:   change7d,
			Factor:     factor,
			LastUpdate: now,
		}
	}

	cs.lastFetch = time.Now()
	log.Printf("[crypto] prices updated at %s", now)
	return nil
}
