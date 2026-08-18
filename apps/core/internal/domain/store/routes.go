package store

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/store/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/store/service"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/middleware"
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
		merchantStores.GET("", h.ListMine)
		merchantStores.POST("", h.Create)
		merchantStores.POST("/:id/activate", h.Activate)
		merchantStores.POST("/:id/deactivate", h.Deactivate)
		merchantStores.PATCH("/:id", h.Update)
		merchantStores.DELETE("/:id", h.Delete)
	}
}
