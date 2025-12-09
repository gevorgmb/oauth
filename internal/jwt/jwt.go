package jwtutil

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type Manager struct {
	secret []byte
	accTTL time.Duration
	rfrTTL time.Duration
}

func NewManagerFromEnv() *Manager {
	secret := []byte(getenv("JWT_SECRET", "dev_secret_4321"))
	tokenLifetime, _ := strconv.Atoi(getenv("ACCESS_TOKEN_LIFETIME", "900"))
	refreshTokenLifetime, _ := strconv.Atoi(getenv("REFRESH_TOKEN_LIFETIME", "604800"))
	return &Manager{
		secret: secret,
		accTTL: time.Duration(tokenLifetime * int(time.Second)),
		rfrTTL: time.Duration(refreshTokenLifetime * int(time.Second)),
	}
}

func getenv(k string, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func (m *Manager) GenerateAccessToken(email string, role string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(m.accTTL).Unix()
	claims := jwt.MapClaims{
		"sub":  email,
		"role": role,
		"exp":  exp,
		"iat":  now.Unix(),
		"typ":  "access",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}
	return s, exp, nil
}

func (m *Manager) GenerateRefreshToken(email string, role string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(m.rfrTTL).Unix()
	claims := jwt.MapClaims{
		"sub":  email,
		"role": role,
		"exp":  exp,
		"iat":  now.Unix(),
		"typ":  "refresh",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}
	return s, exp, nil
}

func (m *Manager) Parse(tokenStr string) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithLeeway(5 * time.Second))
	tok, err := parser.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return m.secret, nil })
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !tok.Valid {
		return nil, ErrInvalidToken
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	return nil, ErrInvalidToken
}
