package review

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/review/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/review/service"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers review routes: public reads and authenticated writes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewReviewService(nil)
	h := handler.NewReviewHandler(svc)

	// Public: anyone can read review details
	reviewsPublic := r.Group("/reviews")
	{
		reviewsPublic.GET("/:id", h.Detail)
	}

	// Authenticated: create reviews, like, and comment
	reviewsAuth := r.Group("/reviews", middleware.JWTAuth(cfg.JWT))
	{
		reviewsAuth.POST("", h.Create)
		reviewsAuth.POST("/:id/like", h.Like)
		reviewsAuth.POST("/:id/comments", h.Comment)
	}
}
