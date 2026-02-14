package verification

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/verification/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/verification/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers merchant verification routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewVerificationService(nil)
	h := handler.NewVerificationHandler(svc)

	merchantVerify := r.Group("/merchant/verification", middleware.JWTAuth(cfg.JWT))
	{
		merchantVerify.POST("", h.Submit)
		merchantVerify.GET("", h.Status)
	}
}
