package feed

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers feed routes.
func RegisterRoutes(r *gin.RouterGroup, _ *config.Config) {
	svc := service.NewFeedService(nil)
	h := handler.NewFeedHandler(svc)

	feed := r.Group("/feed")
	{
		feed.GET("/home", h.Home)
	}
}
