package media

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers media routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewMediaService(nil)
	h := handler.NewMediaHandler(svc)

	media := r.Group("/media", middleware.JWTAuth(cfg.JWT))
	{
		media.POST("/uploads", h.CreateUpload)
		media.POST("/:id/analysis", h.Analyze)
	}
}
