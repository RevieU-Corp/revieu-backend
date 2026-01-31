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
	userHandler := NewUserHandler(nil, nil, nil, nil)

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

		// User routes (JWT required)
		user := api.Group("/user", middleware.JWTAuth(cfg.JWT))
		{
			profile := user.Group("/profile")
			{
				profile.GET("", userHandler.GetProfile)
				profile.PATCH("", userHandler.UpdateProfile)
			}

			privacy := user.Group("/privacy")
			{
				privacy.GET("", userHandler.GetPrivacy)
				privacy.PATCH("", userHandler.UpdatePrivacy)
			}

			notifications := user.Group("/notifications")
			{
				notifications.GET("", userHandler.GetNotifications)
				notifications.PATCH("", userHandler.UpdateNotifications)
			}

			addresses := user.Group("/addresses")
			{
				addresses.GET("", userHandler.ListAddresses)
				addresses.POST("", userHandler.CreateAddress)
				addresses.PATCH("/:id", userHandler.UpdateAddress)
				addresses.DELETE("/:id", userHandler.DeleteAddress)
				addresses.POST("/:id/default", userHandler.SetDefaultAddress)
			}

			account := user.Group("/account")
			{
				account.POST("/export", userHandler.RequestAccountExport)
				account.DELETE("", userHandler.RequestAccountDeletion)
			}

			user.GET("/posts", userHandler.ListMyPosts)
			user.GET("/reviews", userHandler.ListMyReviews)
			user.GET("/favorites", userHandler.ListMyFavorites)
			user.GET("/likes", userHandler.ListMyLikes)

			following := user.Group("/following")
			{
				following.GET("/users", userHandler.ListFollowingUsers)
				following.GET("/merchants", userHandler.ListFollowingMerchants)
			}

			followers := user.Group("/followers")
			{
				followers.GET("", userHandler.ListFollowers)
			}
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
