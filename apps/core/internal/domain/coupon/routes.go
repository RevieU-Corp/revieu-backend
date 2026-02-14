package coupon

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/coupon/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/coupon/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers coupon and package routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewCouponService(nil)
	h := handler.NewCouponHandler(svc)

	coupons := r.Group("/coupons")
	{
		coupons.POST("/:id/validate", h.Validate)
		coupons.POST("/:id/payment/initiate", h.InitiatePayment)
		coupons.POST("/:id/redeem", middleware.JWTAuth(cfg.JWT), h.Redeem)
	}

	packages := r.Group("/packages")
	{
		packages.GET("", h.ListPackages)
		packages.GET("/:id", h.PackageDetail)
	}
}
