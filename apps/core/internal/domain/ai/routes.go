package ai

import (
	"context"
	"log/slog"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers AI routes and returns the StyleService so the review domain
// can hook review-submission events into style derivation. The Gemini client is
// constructed eagerly so a missing or invalid API key surfaces at boot time. When
// construction fails, the route is still registered but every request returns a
// configuration error from the service — this keeps boot resilient when the AI feature
// is intentionally disabled. The returned StyleService may be nil in that case; callers
// must treat that as "no personalization, no learning."
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) *service.StyleService {
	client, err := service.NewGeminiClient(context.Background(), cfg.Gemini)
	if err != nil {
		if logger.Log != nil {
			logger.Log.Warn("ai: gemini client unavailable, /ai/reviews/suggestions will return errors", slog.String("error", err.Error()))
		}
		client = nil
	}

	var styleSvc *service.StyleService
	if styleClient, ok := client.(service.StyleClient); ok && styleClient != nil {
		styleSvc = service.NewStyleService(nil, styleClient)
	}

	svc := service.NewAIService(client, cfg.Gemini).WithStyle(styleSvc)
	h := handler.NewAIHandler(svc, cfg.Gemini)

	ai := r.Group("/ai", middleware.JWTAuth(cfg.JWT))
	{
		ai.POST("/reviews/suggestions", h.Suggestions)
	}
	return styleSvc
}
