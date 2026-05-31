package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"crypto-desert/internal/db"

	"golang.org/x/crypto/bcrypt"
)

// ── User ──────────────────────────────────────────────────────────────────────

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type Service struct {
	db     *db.DB
	secret []byte // segredo para assinar tokens JWT
}

func NewService(database *db.DB, jwtSecret string) *Service {
	return &Service{db: database, secret: []byte(jwtSecret)}
}

// ── Register ──────────────────────────────────────────────────────────────────

func (s *Service) Register(username, password string) (*User, error) {
	if len(username) < 3 {
		return nil, fmt.Errorf("nome de usuário deve ter ao menos 3 caracteres")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("senha deve ter ao menos 6 caracteres")
	}

	if s.db == nil {
		return nil, fmt.Errorf("banco de dados não disponível — rode 'go get modernc.org/sqlite' para habilitar contas de usuário")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash de senha: %w", err)
	}

	var id int
	err = s.db.QueryRow(
		`INSERT INTO users (username, password) VALUES (?, ?) RETURNING id`,
		username, string(hash),
	).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, fmt.Errorf("nome de usuário já existe")
		}
		return nil, fmt.Errorf("criar usuário: %w", err)
	}

	return &User{ID: id, Username: username}, nil
}

// ── Login ─────────────────────────────────────────────────────────────────────

func (s *Service) Login(username, password string) (*User, string, error) {
	if s.db == nil {
		// Modo in-memory: login automático como guest
		user := &User{ID: 1, Username: username}
		token, _ := s.generateToken(1, username)
		return user, token, nil
	}

	var user User
	var hash string

	err := s.db.QueryRow(
		`SELECT id, username, password, created_at FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &hash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("usuário ou senha inválidos")
	}
	if err != nil {
		return nil, "", fmt.Errorf("buscar usuário: %w", err)
	}

	// Verifica senha
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("usuário ou senha inválidos")
	}

	// Gera token JWT simples (header.payload.signature)
	token, err := s.generateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("gerar token: %w", err)
	}

	return &user, token, nil
}

// ── JWT simples (sem biblioteca externa) ─────────────────────────────────────

type jwtClaims struct {
	UserID   int    `json:"uid"`
	Username string `json:"usr"`
	Exp      int64  `json:"exp"`
}

func (s *Service) generateToken(userID int, username string) (string, error) {
	claims := jwtClaims{
		UserID:   userID,
		Username: username,
		Exp:      time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 dias
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header  := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body    := base64.RawURLEncoding.EncodeToString(payload)
	sig     := s.sign(header + "." + body)

	return header + "." + body + "." + sig, nil
}

func (s *Service) sign(data string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ValidateToken valida um token JWT e retorna as claims.
func (s *Service) ValidateToken(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token inválido")
	}

	// Verifica assinatura
	expected := s.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return nil, fmt.Errorf("assinatura inválida")
	}

	// Decodifica payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decodificar payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parsear claims: %w", err)
	}

	// Verifica expiração
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expirado")
	}

	return &claims, nil
}

// UserFromRequest extrai o usuário do cookie de sessão.
// Retorna erro se não autenticado.
func (s *Service) UserFromRequest(cookie string) (int, string, error) {
	if s.db == nil {
		// Modo in-memory: aceita qualquer token ou retorna usuário convidado
		if cookie == "" {
			return 1, "guest", nil
		}
	}
	claims, err := s.ValidateToken(cookie)
	if err != nil {
		if s.db == nil {
			return 1, "guest", nil // no-auth mode
		}
		return 0, "", err
	}
	return claims.UserID, claims.Username, nil
}
