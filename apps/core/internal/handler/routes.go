package handler

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	// Initialize handlers
	testHandler := NewTestHandler()
	authHandler := NewAuthHandler(cfg.JWT, cfg.OAuth, cfg.SMTP, cfg.FrontendURL)

	// API routes (prefix handled by Ingress)
	api := router.Group("/")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/login/google", authHandler.GoogleLogin)
			auth.GET("/callback/google", authHandler.GoogleCallback)
			auth.GET("/verify", authHandler.VerifyEmail)

			// Protected routes
			auth.GET("/me", middleware.JWTAuth(cfg.JWT), authHandler.Me)
		}

		// Test routes
		test := api.Group("/test")
		{
			test.GET("", testHandler.GetTest)
			test.POST("", testHandler.PostTest)
		}

		// Example: User routes
		// users := api.Group("/users")
		// {
		//     users.GET("", userHandler.List)
		//     users.GET("/:id", userHandler.Get)
		//     users.POST("", userHandler.Create)
		//     users.PUT("/:id", userHandler.Update)
		//     users.DELETE("/:id", userHandler.Delete)
		// }

		// Add your route groups here
	}
}
