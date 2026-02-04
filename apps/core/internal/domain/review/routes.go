package review

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers review routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewReviewService(nil)
	h := handler.NewReviewHandler(svc)

	reviews := r.Group("/reviews")
	{
		reviews.GET("", middleware.JWTAuth(cfg.JWT), h.ListMyReviews)
		reviews.POST("", middleware.JWTAuth(cfg.JWT), h.Create)
		reviews.GET("/:id", h.Detail)
		reviews.POST("/:id/like", middleware.JWTAuth(cfg.JWT), h.Like)
		reviews.POST("/:id/comments", middleware.JWTAuth(cfg.JWT), h.Comment)
	}
}
