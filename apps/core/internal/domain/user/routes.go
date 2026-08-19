package user

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	contentHandler "github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/content/handler"
	contentService "github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/content/service"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/user/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/user/service"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers user routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	userSvc := service.NewUserService(nil)
	userHandler := handler.NewUserHandler(userSvc)

	contentSvc := contentService.NewContentService(nil)
	postHandler := contentHandler.NewPostHandler(contentSvc)
	reviewHandler := contentHandler.NewReviewHandler(contentSvc)
	favHandler := contentHandler.NewFavoriteHandler(contentSvc)
	likeHandler := contentHandler.NewLikeHandler(contentSvc)

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

		// Current user's content, delegated to the content domain's handlers.
		user.GET("/posts", postHandler.ListMyPosts)
		user.GET("/reviews", reviewHandler.ListMyReviews)
		user.GET("/favorites", favHandler.ListMyFavorites)
		user.GET("/likes", likeHandler.ListMyLikes)
	}
}
