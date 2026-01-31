package middleware

import (
	"net/http"
	"strings"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UserEmailKey        = "user_email"
	UserRoleKey         = "user_role"
)

func JWTAuth(jwtCfg config.JWTConfig) gin.HandlerFunc {
	tokenService := service.NewTokenService(jwtCfg)

	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// Set user info in context
		if sub, ok := claims["sub"].(float64); ok {
			c.Set(UserIDKey, int64(sub))
		}
		if email, ok := claims["email"].(string); ok {
			c.Set(UserEmailKey, email)
		}
		if role, ok := claims["role"].(string); ok {
			c.Set(UserRoleKey, role)
		}

		c.Next()
	}
}
