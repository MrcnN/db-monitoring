package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/dbplatform/api/internal/config"
	"github.com/dbplatform/api/internal/user"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	cfg config.JWTConfig
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

func (s *JWTService) GenerateAccessToken(u *user.User) (string, error) {
	claims := Claims{
		UserID:    u.ID.String(),
		Email:     u.Email,
		Role:      string(u.Role),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *JWTService) GenerateRefreshToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	raw := base64.URLEncoding.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(raw))
	hashStr := hex.EncodeToString(hash[:])
	return raw, hashStr, nil
}

func (s *JWTService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.TokenType != "access" {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
