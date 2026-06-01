# ⚔ Crypto Desert

> RPG de turnos ambientado em 2087, onde o poder dos personagens flutua conforme o mercado de criptomoedas.

### Trabalho Final — Estrutura de Dados II

**Curso:** Engenharia da Computação – 5º Período
**Instituição:** Faculdade AEMS – Três Lagoas/MS
**Professor:** Bruno Gabriel

### Equipe

* Samuel Medeiros
* João Cardoso
* Robert
* Rafael
* Caio Lopes
* Mahgid Thomé
* Alberto Suave

[![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)](https://hub.docker.com/)

---

# 📑 Sumário

* [Descrição](#-descrição)
* [Narrativa](#-narrativa)
* [Funcionalidades](#-funcionalidades)
* [Arquitetura](#-arquitetura)
* [Estruturas de Dados Utilizadas](#-estruturas-de-dados-utilizadas)
* [Stack](#-stack)
* [Estrutura do Projeto](#-estrutura-do-projeto)
* [Pré-requisitos](#-pré-requisitos)
* [Instalação e Execução](#-instalação-e-execução)

  * [Docker](#docker-recomendado)
  * [Execução Local](#local-sem-docker)
* [Endpoints da API](#-endpoints-da-api)
* [Variáveis de Ambiente](#-variáveis-de-ambiente)
* [Exemplo de Uso da API](#-exemplo-de-uso-da-api)
* [Equipe](#-equipe)
* [Licença](#-licença)

---

# 📖 Descrição

**Crypto Desert** é um RPG de turnos ambientado em um deserto digital pós-apocalíptico no ano de 2087.

Cinco facções disputam o controle das blockchains sobreviventes:

* ₿ BTC
* Ξ ETH
* ◎ SOL
* BNB
* Ð DOGE

O diferencial do jogo é que o dano dos personagens é influenciado pela valorização ou desvalorização real das criptomoedas associadas à sua facção.

As cotações são obtidas em tempo real através da CoinGecko API, tornando cada batalha única.

### Impacto do Mercado

* 🔥 **Bull Run:** fator de dano de até **2.0x**
* 📉 **Bear Market:** fator de dano mínimo de **0.5x**
* 🌐 Atualização automática via API em tempo real

---

# 🌍 Narrativa

No ano de 2087, os governos colapsaram e o controle do mundo foi transferido para facções que dominam blockchains privadas espalhadas pelo Deserto Digital.

Cada facção possui sua própria moeda e seus guerreiros extraem poder diretamente da valorização de seus ativos. Nesse mundo, o mercado dita a força dos combatentes.

> "Aqui, seu poder não vem do treinamento. Vem do mercado."

O jogador atravessa cidades, enfrenta inimigos, derrota chefes e evolui seu personagem enquanto explora um universo onde economia e combate são inseparáveis.

---

# 🚀 Funcionalidades

| Sistema                               | Status |
| ------------------------------------- | ------ |
| Criação de personagens (5 classes)    | ✅      |
| Combate por turnos com d20            | ✅      |
| Fator de dano baseado em criptomoedas | ✅      |
| IA inimiga com 5 comportamentos       | ✅      |
| Sistema de missões com 5 cidades      | ✅      |
| Progressão de nível e experiência     | ✅      |
| Inventário e equipamentos             | ✅      |
| Loja com preços dinâmicos             | ✅      |
| Campfire (sistema de descanso)        | ✅      |
| New Game Plus (NG+)                   | ✅      |

---

# 🏗 Arquitetura

O projeto foi desenvolvido utilizando arquitetura em camadas para garantir separação de responsabilidades, organização do código e facilidade de manutenção.

```text
Frontend (HTML/CSS/JS)
          │
          ▼
       API REST
          │
          ▼
    Motor de Jogo
          │
 ┌────────┴────────┐
 ▼                 ▼
Sistema       CoinGecko API
de Dados
```

### Camadas

| Camada             | Responsabilidade                     |
| ------------------ | ------------------------------------ |
| Interface          | Interação com o usuário              |
| API REST           | Comunicação entre frontend e backend |
| Motor de Jogo      | Combate, missões, IA e progressão    |
| Integração Externa | Consumo da CoinGecko API             |
| Persistência       | Armazenamento dos dados do jogo      |

---

# 📚 Estruturas de Dados Utilizadas

## Queue (Fila)

Utilizada para controlar a ordem de iniciativa durante as batalhas.

**Aplicação:**

* Turnos de combate
* Gerenciamento de ações

**Complexidade:**

* Inserção: O(log n)
* Remoção: O(log n)

---

## Hash Table

Utilizada para armazenar o cache das cotações de criptomoedas.

**Aplicação:**

* Cache da CoinGecko API
* Busca rápida de ativos

**Complexidade:**

* Inserção: O(1)
* Busca: O(1)

---

## Listas Estruturadas

Utilizadas para gerenciamento de personagens, inimigos, itens, missões e inventários.

**Complexidade:**

* Inserção: O(1)
* Busca: O(n)

---

## Store com Mutex

Estrutura utilizada para controle concorrente dos dados armazenados em memória.

**Objetivo:**

* Garantir integridade dos dados
* Evitar condições de corrida (Race Conditions)

---

# 🛠 Stack

| Camada              | Tecnologia                        |
| ------------------- | --------------------------------- |
| Backend / Lógica    | Go 1.24 (Standard Library)        |
| Frontend            | HTML + CSS + JavaScript (Vanilla) |
| API de Criptomoedas | CoinGecko                         |
| Persistência        | In-Memory Store                   |
| Containerização     | Docker + Docker Compose           |

### Justificativa da Stack

A linguagem Go foi escolhida por sua simplicidade, desempenho, suporte nativo à concorrência e facilidade para construção de APIs REST.

Docker foi utilizado para garantir portabilidade e padronização do ambiente de execução.

---

# 📂 Estrutura do Projeto

```text
crypto-desert/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   ├── characters/
│   ├── combat/
│   ├── enemies/
│   ├── game/
│   ├── items/
│   ├── missions/
│   └── store/
├── web/
│   └── index.html
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

---

# 📋 Pré-requisitos

### Docker

* Docker ≥ 24.0
* Docker Compose ≥ 2.0

### Execução Local

* Go ≥ 1.24

---

# ⚙ Instalação e Execução

## Docker (Recomendado)

```bash
git clone https://github.com/seu-usuario/crypto-desert.git

cd crypto-desert

docker-compose up --build
```

Acesse:

```text
http://localhost:8080
```

### Executar em segundo plano

```bash
docker-compose up --build -d

docker-compose logs -f

docker-compose down
```

---

## Local (Sem Docker)

```bash
git clone https://github.com/seu-usuario/crypto-desert.git

cd crypto-desert

go run ./cmd/server
```

Acesse:

```text
http://localhost:8080
```

### Utilizando variáveis de ambiente

```bash
PORT=9090 WEB_DIR=./web go run ./cmd/server
```

---

## Testes

```bash
go test ./...

go test ./internal/characters/... -v

go test ./internal/game/... -v

go test ./internal/items/... -v

go test ./internal/missions/... -v
```

---

# 🔌 Endpoints da API

Base URL:

```text
http://localhost:8080/api
```

## Crypto

| Método | Endpoint    | Descrição                        |
| ------ | ----------- | -------------------------------- |
| GET    | /api/crypto | Retorna cotações e fator de dano |

---

## Classes

| Método | Endpoint     | Descrição                          |
| ------ | ------------ | ---------------------------------- |
| GET    | /api/classes | Lista todas as classes disponíveis |

---

## Personagens

| Método | Endpoint                             |
| ------ | ------------------------------------ |
| GET    | /api/characters                      |
| POST   | /api/characters                      |
| GET    | /api/characters/{id}                 |
| DELETE | /api/characters/{id}                 |
| GET    | /api/characters/{id}/inventory       |
| POST   | /api/characters/{id}/inventory/use   |
| POST   | /api/characters/{id}/inventory/equip |

---

## Inimigos

| Método | Endpoint     |
| ------ | ------------ |
| GET    | /api/enemies |

---

## Batalhas

| Método | Endpoint                      |
| ------ | ----------------------------- |
| POST   | /api/battles                  |
| GET    | /api/battles/{session}        |
| POST   | /api/battles/{session}/action |

---

## Missões

| Método | Endpoint                                 |
| ------ | ---------------------------------------- |
| POST   | /api/missions/session                    |
| GET    | /api/missions/session/{id}               |
| POST   | /api/missions/session/{id}/enter         |
| POST   | /api/missions/session/{id}/start         |
| POST   | /api/missions/session/{id}/battle/begin  |
| POST   | /api/missions/session/{id}/battle/action |
| POST   | /api/missions/session/{id}/confirm       |

---

## Loja

| Método | Endpoint                 |
| ------ | ------------------------ |
| GET    | /api/shop/{city_id}      |
| POST   | /api/shop/{city_id}/buy  |
| POST   | /api/shop/{city_id}/sell |

---

## Campfire

| Método | Endpoint                     |
| ------ | ---------------------------- |
| GET    | /api/campfire/{city_id}      |
| POST   | /api/campfire/{city_id}/rest |

---

# 🔐 Variáveis de Ambiente

| Variável | Padrão | Descrição             |
| -------- | ------ | --------------------- |
| PORT     | 8080   | Porta do servidor     |
| WEB_DIR  | ./web  | Diretório do frontend |

---

# 💻 Exemplo de Uso da API

```bash
curl http://localhost:8080/api/crypto

curl -X POST http://localhost:8080/api/characters \
-H "Content-Type: application/json" \
-d '{"name":"Kabom","class":"warrior"}'

curl -X POST http://localhost:8080/api/missions/session \
-H "Content-Type: application/json" \
-d '{"character_id":1}'
```

---

# 👨‍💻 Equipe

Projeto desenvolvido pelos alunos do 5º Período de Engenharia da Computação da Faculdade AEMS.

| Integrante      |
| --------------- |
| Samuel Medeiros |
| João Cardoso    |
| Robert          |
| Rafael          |
| Caio Lopes      |
| Mahgid Thomé    |
| Alberto Suave   |

---

# 📄 Licença

Projeto acadêmico desenvolvido exclusivamente para fins educacionais como requisito da disciplina de Estrutura de Dados II.
