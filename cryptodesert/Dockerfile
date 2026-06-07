# ── Estágio 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Dependências para compilar modernc.org/sqlite (sem CGO)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copia go.mod e go.sum e baixa dependências (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila o binário — CGO desabilitado, binário estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o crypto-desert ./cmd/server/main.go

# ── Estágio 2: Imagem final ───────────────────────────────────────────────────
FROM alpine:3.20

# Certificados SSL (necessário para chamadas HTTPS à CoinGecko)
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copia o binário compilado
COPY --from=builder /app/crypto-desert .

# Copia o frontend estático
COPY web/ ./web/

# Cria pasta para o banco de dados SQLite
RUN mkdir -p /app/data

# Porta exposta
EXPOSE 8080

# Variáveis de ambiente com valores padrão
ENV PORT=8080
ENV WEB_DIR=./web
ENV DB_PATH=/app/data/game.db
ENV JWT_SECRET=troque-este-segredo-em-producao

# Volume para persistir o banco de dados entre reinicializações
VOLUME ["/app/data"]

# Comando de inicialização
CMD ["./crypto-desert"]
