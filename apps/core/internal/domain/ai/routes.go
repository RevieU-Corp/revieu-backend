package ai

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers AI routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewAIService()
	h := handler.NewAIHandler(svc)

	ai := r.Group("/ai", middleware.JWTAuth(cfg.JWT))
	{
		ai.POST("/reviews/suggestions", h.Suggestions)
	}
}
