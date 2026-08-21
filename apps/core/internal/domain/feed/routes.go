package feed

import (
	"github.com/gin-gonic/gin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/feed/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/feed/service"
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
