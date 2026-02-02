package auth

import (
	"fmt"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

// TokenService issues and validates JWTs.
type TokenService struct {
	secret     []byte
	expireHour int
}

// NewTokenService creates a TokenService for middleware and auth logic.
func NewTokenService(cfg config.JWTConfig) *TokenService {
	return &TokenService{
		secret:     []byte(cfg.Secret),
		expireHour: cfg.ExpireHour,
	}
}

func (s *TokenService) GenerateToken(user *model.User, auth *model.UserAuth) (string, error) {
	claims := jwt.MapClaims{
		"sub":           user.ID,
		"email":         auth.Identifier,
		"identity_type": auth.IdentityType,
		"role":          user.Role,
		"exp":           time.Now().Add(time.Hour * time.Duration(s.expireHour)).Unix(),
		"iat":           time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *TokenService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
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
