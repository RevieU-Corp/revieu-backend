package profile

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/profile/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers profile routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	_ = cfg
	svc := service.NewService(nil)
	handler := NewHandler(svc)

	users := r.Group("/users")
	{
		users.GET("/:id", handler.GetPublicProfile)
	}
}
