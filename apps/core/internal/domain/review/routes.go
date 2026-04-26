package review

import (
	aisvc "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers review routes. styleSvc is forwarded into the review service
// so review submissions can fan out into the writing-style derive pipeline. It may be
// nil when the AI feature is disabled at boot — the review service treats nil as "no
// style learning" and behaves identically to the un-personalized path.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config, styleSvc *aisvc.StyleService) {
	svc := service.NewReviewService(nil).WithStyle(styleSvc)
	h := handler.NewReviewHandler(svc)

	reviews := r.Group("/reviews")
	{
		reviews.POST("", middleware.JWTAuth(cfg.JWT), h.Create)
		reviews.GET("/:id", h.Detail)
		reviews.POST("/:id/like", middleware.JWTAuth(cfg.JWT), h.Like)
		reviews.POST("/:id/comments", middleware.JWTAuth(cfg.JWT), h.Comment)
	}
}
