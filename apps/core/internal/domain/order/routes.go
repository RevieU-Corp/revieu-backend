package order

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/order/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/order/service"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/middleware"
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
		orders.POST("/:id/pay", h.Pay)
	}
}
