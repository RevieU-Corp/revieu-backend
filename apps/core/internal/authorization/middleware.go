package authorization

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/token"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UserEmailKey        = "user_email"
	UserRoleKey         = "user_role"
)

// JWTAuth authenticates a bearer token and installs its principal in Gin context.
func JWTAuth(jwtCfg config.JWTConfig) gin.HandlerFunc {
	tokenService := token.New(jwtCfg)
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}
		claims, err := tokenService.ValidateToken(strings.TrimPrefix(authHeader, BearerPrefix))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
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
