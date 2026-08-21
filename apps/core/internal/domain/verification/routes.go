package verification

import (
	"github.com/gin-gonic/gin"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/authorization"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/verification/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/verification/service"
)

// RegisterRoutes registers merchant verification routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewVerificationService(nil)
	h := handler.NewVerificationHandler(svc)

	merchantVerify := r.Group("/merchant/verification", authorization.JWTAuth(cfg.JWT))
	{
		merchantVerify.POST("", h.Submit)
		merchantVerify.GET("", h.Status)
	}
}
