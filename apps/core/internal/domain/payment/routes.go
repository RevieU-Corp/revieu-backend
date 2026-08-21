package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/authorization"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/payment/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/payment/service"
)

// RegisterRoutes registers payment routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewPaymentService(nil)
	h := handler.NewPaymentHandler(svc)

	payments := r.Group("/payments", authorization.JWTAuth(cfg.JWT))
	{
		payments.POST("", h.Create)
		payments.GET("/:id", h.Detail)
	}
}
