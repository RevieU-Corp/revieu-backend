package admin

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/admin/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/admin/service"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers admin routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewAdminService(nil)
	h := handler.NewAdminHandler(svc)

	adminGroup := r.Group("/admin", middleware.JWTAuth(cfg.JWT))
	{
		adminGroup.GET("/reports", h.ListReports)
		adminGroup.PATCH("/reports/:id", h.UpdateReport)
		adminGroup.GET("/merchants", h.ListMerchants)
		adminGroup.PATCH("/merchants/:id", h.UpdateMerchant)
	}
}
