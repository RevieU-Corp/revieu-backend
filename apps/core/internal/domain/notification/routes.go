package notification

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/notification/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/notification/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers notification routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewNotificationService(nil)
	h := handler.NewNotificationHandler(svc)

	notifs := r.Group("/notifications", middleware.JWTAuth(cfg.JWT))
	{
		notifs.GET("", h.List)
		notifs.PATCH("/:id/read", h.MarkRead)
		notifs.POST("/read-all", h.ReadAll)
	}
}
