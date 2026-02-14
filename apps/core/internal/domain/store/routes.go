package store

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/store/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/store/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers store routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewStoreService(nil)
	h := handler.NewStoreHandler(svc)

	stores := r.Group("/stores")
	{
		stores.GET("", h.List)
		stores.GET("/:id", h.Detail)
		stores.GET("/:id/reviews", h.Reviews)
		stores.GET("/:id/hours", h.Hours)
	}

	merchantStores := r.Group("/merchant/stores", middleware.JWTAuth(cfg.JWT))
	{
		merchantStores.POST("", h.Create)
		merchantStores.PATCH("/:id", h.Update)
	}
}
