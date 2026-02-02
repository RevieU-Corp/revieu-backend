package follow

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers follow routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewFollowService(nil)
	userHandler := handler.NewUserHandler(svc)
	merchantHandler := handler.NewMerchantHandler(svc)

	auth := r.Group("", middleware.JWTAuth(cfg.JWT))
	{
		auth.POST("/users/:id/follow", userHandler.FollowUser)
		auth.DELETE("/users/:id/follow", userHandler.UnfollowUser)
		auth.POST("/merchants/:id/follow", merchantHandler.FollowMerchant)
		auth.DELETE("/merchants/:id/follow", merchantHandler.UnfollowMerchant)
	}
}
