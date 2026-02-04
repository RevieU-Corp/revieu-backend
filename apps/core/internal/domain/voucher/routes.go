package voucher

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/voucher/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/voucher/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers voucher routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewVoucherService(nil)
	h := handler.NewVoucherHandler(svc)

	vouchers := r.Group("/vouchers", middleware.JWTAuth(cfg.JWT))
	{
		vouchers.POST("", h.Create)
		vouchers.GET("", h.List)
		vouchers.GET("/:id", h.Detail)
		vouchers.GET("/code/:code", h.ByCode)
		vouchers.PATCH("/:id/use", h.Use)
		vouchers.PATCH("/:id/status", h.UpdateStatus)
		vouchers.POST("/share/email", h.ShareEmail)
		vouchers.POST("/share/sms", h.ShareSMS)
	}
}
