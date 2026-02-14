package conversation

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/conversation/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/conversation/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers conversation routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewConversationService(nil)
	h := handler.NewConversationHandler(svc)

	convos := r.Group("/conversations", middleware.JWTAuth(cfg.JWT))
	{
		convos.GET("", h.List)
		convos.POST("", h.Create)
		convos.GET("/:id/messages", h.Messages)
		convos.POST("/:id/messages", h.SendMessage)
		convos.PATCH("/:id/settings", h.UpdateSettings)
	}
}
