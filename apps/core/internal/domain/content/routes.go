package content

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers content routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	contentSvc := service.NewContentService(nil)

	postHandler := handler.NewPostHandler(contentSvc)
	reviewHandler := handler.NewReviewHandler(contentSvc)
	favHandler := handler.NewFavoriteHandler(contentSvc)
	likeHandler := handler.NewLikeHandler(contentSvc)

	users := r.Group("/users")
	{
		users.GET("/:id/posts", postHandler.ListUserPosts)
		users.GET("/:id/reviews", reviewHandler.ListUserReviews)
	}

	user := r.Group("/user", middleware.JWTAuth(cfg.JWT))
	{
		user.GET("/posts", postHandler.ListMyPosts)
		user.GET("/reviews", reviewHandler.ListMyReviews)
		user.GET("/favorites", favHandler.ListMyFavorites)
		user.GET("/likes", likeHandler.ListMyLikes)
	}
}
