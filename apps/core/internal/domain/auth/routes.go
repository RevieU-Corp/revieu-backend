package auth

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers auth routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	handler := NewHandler(cfg.JWT, cfg.OAuth, cfg.SMTP, cfg.FrontendURL, cfg.Server.APIBasePath)

	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/forgot-password", handler.ForgotPassword)
		auth.GET("/login/google", handler.GoogleLogin)
		auth.GET("/callback/google", handler.GoogleCallback)
		auth.GET("/verify", handler.VerifyEmail)
		auth.GET("/me", middleware.JWTAuth(cfg.JWT), handler.Me)
	}
}
