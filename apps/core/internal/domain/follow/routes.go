package follow

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/follow/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/follow/service"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/middleware"
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
