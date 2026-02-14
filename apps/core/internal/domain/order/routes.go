package order

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/order/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/order/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers order routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewOrderService(nil)
	h := handler.NewOrderHandler(svc)

	orders := r.Group("/orders", middleware.JWTAuth(cfg.JWT))
	{
		orders.POST("", h.Create)
		orders.GET("", h.List)
		orders.GET("/:id", h.Detail)
	}
}
