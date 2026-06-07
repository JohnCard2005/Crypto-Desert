# ⚔ Crypto Desert

> RPG de turnos ambientado em 2087, onde o poder dos personagens flutua com o mercado de criptomoedas.

[![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)](https://hub.docker.com/)

---

## Sumário

- [Descrição](#descrição)
- [Stack](#stack)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Pré-requisitos](#pré-requisitos)
- [Instalação e Execução](#instalação-e-execução)
  - [Docker (recomendado)](#docker-recomendado)
  - [Local (sem Docker)](#local-sem-docker)
- [Endpoints da API](#endpoints-da-api)
- [Variáveis de Ambiente](#variáveis-de-ambiente)

---

## Descrição

**Crypto Desert** é um RPG de turnos onde personagens de 5 facções (BTC, ETH, SOL, BNB, DOGE) combatem em um deserto digital pós-apocalíptico. O diferencial: a variação de preço real de cada criptomoeda nas últimas 7 dias afeta diretamente o dano dos personagens em batalha.

- 🔥 **Bull run** → fator de dano até ×2.0
- 📉 **Bear market** → fator de dano mínimo de ×0.5
- Os dados são buscados em tempo real via **CoinGecko API**

### Funcionalidades implementadas

| Sistema | Status |
|---|---|
| Criação de personagens (5 classes) | ✅ |
| Combate por turnos com d20 | ✅ |
| Fator de dano baseado em crypto | ✅ |
| IA inimiga (5 comportamentos) | ✅ |
| Sistema de missões com 5 cidades | ✅ |
| Progressão de nível e XP | ✅ |
| Inventário e equipamentos | ✅ |
| Loja com preços dinâmicos | ✅ |
| Campfire (descanso Dark Souls-style) | ✅ |
| NG+ após completar a campanha | ✅ |

---

## Stack

| Camada | Tecnologia |
|---|---|
| **Backend / Lógica** | Go 1.24 (stdlib apenas) |
| **Frontend** | HTML + CSS + JS (vanilla) |
| **API de Crypto** | CoinGecko (free tier) |
| **Persistência** | In-memory (store com mutex) |
| **Container** | Docker + docker-compose |

---

## Estrutura do Projeto

```
crypto-desert/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── crypto.go            # Serviço CoinGecko com cache
│   │   ├── dto.go               # DTOs de request/response JSON
│   │   ├── handler.go           # Handlers HTTP de todos os endpoints
│   │   └── routes.go            # Registro de rotas + middlewares
│   ├── characters/
│   │   ├── character.go         # Struct principal do personagem
│   │   ├── factory.go           # NewCharacter(), NewCharacterAtLevel()
│   │   ├── faction.go           # 5 facções com lore e cores
│   │   ├── ability.go           # Habilidades especiais por classe
│   │   ├── status.go            # 8 status effects com modificadores
│   │   ├── xp.go                # Curva de XP e scaling por nível
│   │   └── methods.go           # Todos os métodos do personagem
│   ├── combat/
│   │   ├── damage.go            # d20, CryptoFactor, ResolveAttack
│   │   └── queue.go             # Fila de iniciativa ordenada
│   ├── enemies/
│   │   ├── enemy.go             # Struct Enemy com metadata de IA
│   │   └── catalogue.go         # 13 inimigos (5 common, 4 elite, 4 boss)
│   ├── game/
│   │   ├── battle.go            # Motor de batalha por turnos
│   │   ├── ai.go                # 5 comportamentos de IA
│   │   └── combatant.go         # Wrappers player/enemy para a fila
│   ├── items/
│   │   ├── item.go              # Tipos, categorias, efeitos
│   │   ├── catalogue.go         # 28 itens com lore crypto
│   │   ├── inventory.go         # Inventário com equip/use/drop
│   │   ├── shop.go              # Loja com preços dinâmicos por crypto
│   │   └── campfire.go          # Nó de descanso Dark Souls-style
│   ├── missions/
│   │   ├── world.go             # 5 cidades com lore e waves
│   │   ├── progress.go          # Save game, desbloqueios, NG+
│   │   └── runner.go            # Máquina de estado Pokémon-style
│   └── store/
│       └── store.go             # Stores in-memory com mutex
├── web/
│   └── index.html               # Frontend single-page
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

---

## Pré-requisitos

- **Docker** ≥ 24.0 e **docker-compose** ≥ 2.0 (para execução via container)
- **Go** ≥ 1.24 (para execução local)

---

## Instalação e Execução

### Docker (recomendado)

```bash
# Clone o repositório
git clone https://github.com/seu-usuario/crypto-desert.git
cd crypto-desert

# Build e sobe o servidor
docker-compose up --build

# O servidor estará disponível em:
# http://localhost:8080
```

Para rodar em background:

```bash
docker-compose up --build -d
docker-compose logs -f   # acompanhar logs
docker-compose down      # parar
```

### Local (sem Docker)

```bash
# Clone o repositório
git clone https://github.com/seu-usuario/crypto-desert.git
cd crypto-desert

# Executa o servidor (Go 1.24+ necessário)
go run ./cmd/server

# O servidor estará disponível em:
# http://localhost:8080
```

Variáveis opcionais:

```bash
PORT=9090 WEB_DIR=./web go run ./cmd/server
```

Para executar os testes:

```bash
go test ./...                    # todos os testes
go test ./internal/characters/... -v
go test ./internal/game/... -v
go test ./internal/items/... -v
go test ./internal/missions/... -v
```

---

## Endpoints da API

A API base é `http://localhost:8080/api`.

### Crypto
| Método | Endpoint | Descrição |
|---|---|---|
| GET | `/api/crypto` | Cotações atuais das 5 cryptos com fator de dano |

### Classes
| Método | Endpoint | Descrição |
|---|---|---|
| GET | `/api/classes` | Definições de todas as classes com fator crypto ao vivo |

### Personagens
| Método | Endpoint | Descrição |
|---|---|---|
| GET | `/api/characters` | Lista todos os personagens |
| POST | `/api/characters` | Cria personagem `{"name":"...","class":"..."}` |
| GET | `/api/characters/{id}` | Busca por ID |
| DELETE | `/api/characters/{id}` | Remove personagem |
| GET | `/api/characters/{id}/inventory` | Inventário do personagem |
| POST | `/api/characters/{id}/inventory/use` | Usa item `{"item_id":"..."}` |
| POST | `/api/characters/{id}/inventory/equip` | Equipa item `{"item_id":"..."}` |

### Inimigos
| Método | Endpoint | Descrição |
|---|---|---|
| GET | `/api/enemies` | Catálogo completo com fator crypto atual |

### Batalhas (standalone)
| Método | Endpoint | Descrição |
|---|---|---|
| POST | `/api/battles` | Inicia `{"character_id":1,"enemy_name":"..."}` |
| GET | `/api/battles/{session}` | Estado atual da batalha |
| POST | `/api/battles/{session}/action` | Executa ação `{"action":"attack"}` |

### Missões
| Método | Endpoint | Descrição |
|---|---|---|
| POST | `/api/missions/session` | Cria sessão `{"character_id":1}` |
| GET | `/api/missions/session/{id}` | Snapshot do estado atual |
| POST | `/api/missions/session/{id}/enter` | Entra na cidade `{"city_id":"..."}` |
| POST | `/api/missions/session/{id}/start` | Inicia a missão |
| POST | `/api/missions/session/{id}/battle/begin` | Começa a batalha da wave |
| POST | `/api/missions/session/{id}/battle/action` | Ação do jogador `{"action":"..."}` |
| POST | `/api/missions/session/{id}/confirm` | Confirma transição `{"action":"next_wave"}` |

### Loja
| Método | Endpoint | Descrição |
|---|---|---|
| GET | `/api/shop/{city_id}` | Estoque com preços dinâmicos |
| POST | `/api/shop/{city_id}/buy` | Compra `{"character_id":1,"item_id":"...","quantity":1}` |
| POST | `/api/shop/{city_id}/sell` | Vende `{"character_id":1,"item_id":"...","quantity":1}` |

### Campfire
| Método | Endpoint | Descrição |
|---|---|---|
| GET | `/api/campfire/{city_id}?character_id=1` | Serviços disponíveis |
| POST | `/api/campfire/{city_id}/rest` | Usa serviço `{"character_id":1,"service":"rest_full"}` |

---

## Variáveis de Ambiente

| Variável | Default | Descrição |
|---|---|---|
| `PORT` | `8080` | Porta do servidor HTTP |
| `WEB_DIR` | `./web` | Diretório do frontend estático |

---

## Exemplo de uso da API

```bash
# 1. Ver cotações das cryptos
curl http://localhost:8080/api/crypto

# 2. Criar um personagem
curl -X POST http://localhost:8080/api/characters \
  -H "Content-Type: application/json" \
  -d '{"name":"Kabom","class":"warrior"}'

# 3. Iniciar uma sessão de missão
curl -X POST http://localhost:8080/api/missions/session \
  -H "Content-Type: application/json" \
  -d '{"character_id":1}'

# 4. Entrar na primeira cidade
curl -X POST http://localhost:8080/api/missions/session/m-1/enter \
  -H "Content-Type: application/json" \
  -d '{"city_id":"genesis_block"}'

# 5. Iniciar a missão
curl -X POST http://localhost:8080/api/missions/session/m-1/start

# 6. Começar a batalha
curl -X POST http://localhost:8080/api/missions/session/m-1/battle/begin

# 7. Atacar
curl -X POST http://localhost:8080/api/missions/session/m-1/battle/action \
  -H "Content-Type: application/json" \
  -d '{"action":"attack"}'
```
