package user

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers user routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	userSvc := service.NewUserService(nil)
	userHandler := handler.NewUserHandler(userSvc)

	user := r.Group("/user", middleware.JWTAuth(cfg.JWT))
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

		following := user.Group("/following")
		{
			following.GET("/users", userHandler.ListFollowingUsers)
			following.GET("/merchants", userHandler.ListFollowingMerchants)
		}

		followers := user.Group("/followers")
		{
			followers.GET("", userHandler.ListFollowers)
		}

		account := user.Group("/account")
		{
			account.POST("/export", userHandler.RequestAccountExport)
			account.DELETE("", userHandler.RequestAccountDeletion)
		}
	}
}
