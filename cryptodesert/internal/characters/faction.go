package characters

// Faction represents a crypto-backed faction in the desert
type Faction string

const (
	FactionBTC  Faction = "BTC"
	FactionETH  Faction = "ETH"
	FactionSOL  Faction = "SOL"
	FactionBNB  Faction = "BNB"
	FactionDOGE Faction = "DOGE"
)

type FactionInfo struct {
	Name        string
	Crypto      string  // CoinGecko ID
	Symbol      Faction
	Lore        string
	ColorCode   string // ANSI color for CLI display
}

var Factions = map[Faction]FactionInfo{
	FactionBTC: {
		Name:      "Ordem dos Blocos",
		Crypto:    "bitcoin",
		Symbol:    FactionBTC,
		Lore:      "Os primeiros. Guerreiros forjados na origem da cadeia, lentos mas devastadores.",
		ColorCode: "\033[33m", // yellow
	},
	FactionETH: {
		Name:      "Conclave dos Contratos",
		Crypto:    "ethereum",
		Symbol:    FactionETH,
		Lore:      "Magos que manipulam a realidade através de contratos inteligentes. Poder e fragilidade.",
		ColorCode: "\033[35m", // magenta
	},
	FactionSOL: {
		Name:      "Rastreadores Solares",
		Crypto:    "solana",
		Symbol:    FactionSOL,
		Lore:      "Arqueiros velozes que operam na velocidade das transações de alta frequência.",
		ColorCode: "\033[36m", // cyan
	},
	FactionBNB: {
		Name:      "Guilda das Taxas",
		Crypto:    "binancecoin",
		Symbol:    FactionBNB,
		Lore:      "Ladinos que lucram no caos, cobrando pedágio em cada troca do mercado.",
		ColorCode: "\033[32m", // green
	},
	FactionDOGE: {
		Name:      "Horda Lunar",
		Crypto:    "dogecoin",
		Symbol:    FactionDOGE,
		Lore:      "Caóticos e imprevisíveis. Sua força vem de memes e movimentos irracionais do mercado.",
		ColorCode: "\033[31m", // red
	},
}
