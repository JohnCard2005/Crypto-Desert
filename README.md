# ⚔️ Crypto Desert

> RPG de turnos inspirado em RPGs clássicos, onde o poder dos personagens é influenciado em tempo real pelo mercado de criptomoedas.

![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)
![SQLite](https://img.shields.io/badge/SQLite-003B57?logo=sqlite)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)
![License](https://img.shields.io/badge/Academic_Project-AEMS-blue)

---

# 📚 Trabalho Acadêmico

**Disciplina:** Estrutura de Dados II  
**Curso:** Engenharia da Computação – 5º Período  
**Instituição:** Faculdade AEMS – Três Lagoas/MS  
**Professor:** Bruno Gabriel

## 👨‍💻 Integrantes

- Samuel Medeiros
- João Cardoso
- Robert
- Rafael
- Caio Lopes
- Mahgid Thomé
- Alberto Suave

---

# 📋 Especificação do Projeto

Este projeto foi desenvolvido como Trabalho Final da disciplina de **Estrutura de Dados II**, seguindo os requisitos e critérios definidos pelo professor.

📌 **Repositório oficial do enunciado:**
https://github.com/brunogabrielpk/crypto-desert-ed2

O documento oficial contém:

- Regras da mecânica de combate;
- Requisitos de integração com APIs de criptomoedas;
- Estruturas de dados obrigatórias;
- Requisitos de persistência;
- Critérios de avaliação;
- Entregáveis obrigatórios;
- Requisitos de Docker e Git.

Este repositório representa a implementação prática da proposta acadêmica apresentada no enunciado oficial.

---

# 📖 Sumário

- [Especificação do Projeto](#-especificação-do-projeto)
- [Sobre o Projeto](#-sobre-o-projeto)
- [Narrativa](#-narrativa)
- [Historia](#-historia)
- [Funcionalidades](#-funcionalidades)
- [Tecnologias Utilizadas](#️-tecnologias-utilizadas)
- [Arquitetura](#️-arquitetura)
- [Estruturas de Dados](#-estruturas-de-dados-utilizadas)
- [Mecânica de Combate](#️-mecânica-de-combate)
- [Facções](#️-facções)
- [Classes Jogáveis](#-classes-jogáveis)
- [Integração com API](#-integração-com-api)
- [Persistência](#-persistência)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Instalação](#️-instalação)
- [Docker](#-docker)
- [Endpoints da API](#-endpoints-da-api)
- [Variáveis de Ambiente](#-variáveis-de-ambiente)
- [Testes](#-testes)
- [Controle de Versão](#-controle-de-versão)
- [Equipe](#-equipe)
- [Licença](#-licença)

---

# 🎮 Sobre o Projeto

Crypto Desert é um RPG de turnos ambientado em um universo pós-apocalíptico digital, onde facções dominam blockchains privadas e guerreiros extraem poder diretamente da valorização de suas criptomoedas.

O principal diferencial do jogo é a integração em tempo real com o mercado de criptomoedas. A valorização ou desvalorização de cada ativo influencia diretamente os atributos de combate dos personagens.

---

# 🌍 Narrativa

No ano de 2087, os governos colapsaram e o controle global foi assumido por facções que dominam blockchains privadas espalhadas pelo Deserto Digital.

Cada facção possui sua própria moeda e recruta guerreiros cujo poder depende diretamente do desempenho de seus ativos no mercado.

> "Aqui, seu poder não vem do treinamento. Vem do mercado."

O jogador atravessa cidades, enfrenta ondas de inimigos, derrota chefes e evolui seu personagem em uma jornada onde economia e combate estão diretamente conectados.

---

# 📖 Historia



---
# 🚀 Funcionalidades

| Sistema | Status |
|----------|----------|
| Sistema de Login | ✅ |
| Criação de Personagens | ✅ |
| 5 Classes Jogáveis | ✅ |
| 5 Facções | ✅ |
| Combate baseado em d20 | ✅ |
| Sistema de Habilidades | ✅ |
| IA Inimiga | ✅ |
| Sistema de Missões | ✅ |
| Progressão por Nível | ✅ |
| Inventário | ✅ |
| Equipamentos | ✅ |
| Loja Dinâmica | ✅ |
| Campfire | ✅ |
| Ranking Global | ✅ |
| New Game Plus | ✅ |
| Integração com API de Criptomoedas | ✅ |
| Persistência SQLite | ✅ |

---

# 🛠️ Tecnologias Utilizadas

| Camada | Tecnologia |
|----------|----------|
| Backend | Go 1.24 |
| Frontend | HTML + CSS + JavaScript |
| Banco de Dados | SQLite |
| API Externa | CoinGecko |
| Containerização | Docker |
| Controle de Versão | Git + GitHub |

---

# 🏗️ Arquitetura

O projeto foi desenvolvido seguindo uma arquitetura em camadas.

```text
Frontend
    │
    ▼
API REST
    │
    ▼
Motor de Jogo
 ├─ Combate
 ├─ Missões
 ├─ Inventário
 ├─ IA
 │
 ▼
Persistência SQLite
 │
 ▼
CoinGecko API
```

---

# 📚 Estruturas de Dados Utilizadas

## Queue (Fila)

Arquivo: `internal/combat/queue.go`

Utilizada para gerenciar a ordem dos turnos dos combatentes.

### Justificativa

A fila representa naturalmente a sequência de ações durante uma batalha por turnos.

### Complexidade

| Operação | Big O |
|-----------|-----------|
| Inserção | O(n) |
| Avançar turno | O(1) |
| Consulta | O(1) |
| Remoção | O(n) |

---

## Hash Table

Implementação:

```go
map[string]CryptoPrice
```

Utilizada para cache das cotações da CoinGecko.

### Justificativa

Permite acesso rápido às cotações sem necessidade de novas consultas à API.

### Complexidade

| Operação | Big O |
|-----------|-----------|
| Busca | O(1) |
| Inserção | O(1) |

---

## Banco de Dados SQLite

Utilizado para persistência de:

- Usuários
- Personagens
- Inventários
- Ranking
- Progresso das missões

---

# ⚔️ Mecânica de Combate

## Iniciativa

```text
iniciativa = d20 + velocidade
```

## Ataque

```text
acerto = d20 + modificador_ataque
```

O ataque acerta quando:

```text
acerto >= CA do alvo
```

## Críticos

```text
20 = Acerto Crítico
1 = Falha Crítica
```

## Dano

```text
dano = (dado + força) × fator_crypto
```

Onde:

```text
fator_crypto = 1 + (variacao_7_dias / 100)
```

Limitado entre:

```text
0.5x e 2.0x
```

---

# 🏛️ Facções

## ₿ Ordem dos Blocos (BTC)

Guerreiros da blockchain original.

## Ξ Conclave dos Contratos (ETH)

Mestres dos contratos inteligentes.

## ◎ Rastreadores Solares (SOL)

Especialistas em velocidade.

## BNB Guilda das Taxas

Mercenários focados em eficiência.

## Ð Horda Lunar (DOGE)

Combatentes imprevisíveis movidos por especulação.

---

# 🧙 Classes Jogáveis

| Classe | Cripto | HP | CA | Dado | Habilidade |
|----------|----------|----------|----------|----------|----------|
| Warrior | BTC | 120 | 14 | d10 | Fúria do Bloco |
| Mage | ETH | 80 | 11 | d6 | Contrato Inteligente |
| Archer | SOL | 95 | 13 | d8 | Snipe Veloz |
| Rogue | BNB | 90 | 12 | d8 | Ataque Sombrio |
| Shaman | DOGE | 100 | 12 | d8 | Caos Lunar |

---

# 🌐 Integração com API

API utilizada: CoinGecko

Endpoint:

```text
https://api.coingecko.com/api/v3/simple/price
```

## Recursos implementados

- Cache local
- TTL de 5 minutos
- Timeout de 8 segundos
- Tratamento de Rate Limit
- Fallback automático

Caso a API esteja indisponível, o sistema utiliza a última cotação válida armazenada.

---

# 💾 Persistência

O projeto utiliza SQLite como banco de dados principal.

Entidades persistidas:

- Usuários
- Personagens
- Inventários
- Ranking Global
- Progresso de Missões

Banco:

```text
data/game.db
```

---

# 📂 Estrutura do Projeto

```text
crypto-desert/
├── cmd/
├── internal/
├── web/
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

---

# ⚙️ Instalação

```bash
git clone https://github.com/JohnCard2005/Crypto-Desert.git
cd Crypto-Desert
```

---

# 🐳 Docker

```bash
docker-compose up --build
```

## DockerHub

Adicionar após publicação:

```text
docker pull SEU-USUARIO/crypto-desert
```

---

# 🔌 Endpoints da API

Base URL:

```text
http://localhost:8080/api
```

- Crypto
- Classes
- Personagens
- Inimigos
- Batalhas
- Missões
- Loja
- Campfire

---

# 🔐 Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|----------|----------|
| PORT | 8080 | Porta do servidor |
| WEB_DIR | ./web | Diretório do frontend |

---

# 🧪 Testes

```bash
go test ./...
```

---

# 🌳 Controle de Versão

O desenvolvimento foi realizado utilizando Git e GitHub.

Todos os integrantes contribuíram com commits distribuídos durante o desenvolvimento do projeto.

Repositório:

https://github.com/JohnCard2005/Crypto-Desert

---

# 👨‍💻 Equipe

Projeto desenvolvido pelos alunos do 5º Período de Engenharia da Computação da Faculdade AEMS.

---

# 📄 Licença

Projeto acadêmico desenvolvido exclusivamente para fins educacionais para a disciplina de Estrutura de Dados II.
