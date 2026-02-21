package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

// Service issues and validates JWTs.
type Service struct {
	secret            []byte
	expireHour        int
	refreshExpireHour int
}

// New creates a JWT service.
func New(cfg config.JWTConfig) *Service {
	return &Service{
		secret:            []byte(cfg.Secret),
		expireHour:        cfg.ExpireHour,
		refreshExpireHour: cfg.RefreshExpireHour,
	}
}

func (s *Service) GenerateToken(user *model.User, auth *model.UserAuth) (string, error) {
	claims := jwt.MapClaims{
		"sub":           user.ID,
		"email":         auth.Identifier,
		"identity_type": auth.IdentityType,
		"role":          user.Role,
		"exp":           time.Now().Add(time.Hour * time.Duration(s.expireHour)).Unix(),
		"iat":           time.Now().Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashToken returns a deterministic hash representation of a token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// GenerateRefreshToken creates a random token and its hash for DB persistence.
func (s *Service) GenerateRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(raw)
	return token, HashToken(token), nil
}

// RefreshTokenTTL returns configured refresh token time-to-live.
func (s *Service) RefreshTokenTTL() time.Duration {
	hours := s.refreshExpireHour
	if hours <= 0 {
		hours = 24 * 7
	}
	return time.Hour * time.Duration(hours)
}
