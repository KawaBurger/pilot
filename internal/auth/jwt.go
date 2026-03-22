package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret []byte
}

func NewJWTManager(pilotHome string) (*JWTManager, error) {
	secretPath := filepath.Join(pilotHome, "jwt_secret")

	data, err := os.ReadFile(secretPath)
	if err == nil {
		secret, err := hex.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("decode jwt secret: %w", err)
		}
		return &JWTManager{secret: secret}, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read jwt secret: %w", err)
	}

	// Auto-generate secret.
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate jwt secret: %w", err)
	}

	encoded := hex.EncodeToString(secret)
	if err := os.MkdirAll(pilotHome, 0700); err != nil {
		return nil, fmt.Errorf("create pilot home: %w", err)
	}
	if err := os.WriteFile(secretPath, []byte(encoded), 0600); err != nil {
		return nil, fmt.Errorf("write jwt secret: %w", err)
	}

	return &JWTManager{secret: secret}, nil
}

func (m *JWTManager) Sign(userID int, username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"iat":      now.Unix(),
		"exp":      now.Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Verify(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
