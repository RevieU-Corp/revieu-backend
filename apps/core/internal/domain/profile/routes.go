package profile

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/profile/service"
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
