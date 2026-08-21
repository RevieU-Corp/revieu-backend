package notification

import (
	"github.com/gin-gonic/gin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/authorization"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/notification/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/notification/service"
)

// RegisterRoutes registers notification routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewNotificationService(nil)
	h := handler.NewNotificationHandler(svc)

	notifs := r.Group("/notifications", authorization.JWTAuth(cfg.JWT))
	{
		notifs.GET("", h.List)
		notifs.PATCH("/:id/read", h.MarkRead)
		notifs.POST("/read-all", h.ReadAll)
	}
}
