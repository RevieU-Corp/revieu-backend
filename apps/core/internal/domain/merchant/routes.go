package merchant

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers merchant routes.
func RegisterRoutes(r *gin.RouterGroup, _ *config.Config) {
	svc := service.NewMerchantService(nil)
	h := handler.NewMerchantHandler(svc)

	merchants := r.Group("/merchants")
	{
		merchants.GET("", h.List)
		merchants.GET("/:id", h.Detail)
		merchants.GET("/:id/reviews", h.Reviews)
	}
}
