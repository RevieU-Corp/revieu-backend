package payment

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/payment/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/payment/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers payment routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewPaymentService(nil)
	h := handler.NewPaymentHandler(svc)

	payments := r.Group("/payments", middleware.JWTAuth(cfg.JWT))
	{
		payments.POST("", h.Create)
		payments.GET("/:id", h.Detail)
	}
}
