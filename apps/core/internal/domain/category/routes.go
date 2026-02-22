package category

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/category/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/category/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers category routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	svc := service.NewCategoryService(nil)
	h := handler.NewCategoryHandler(svc)

	categories := r.Group("/categories")
	{
		categories.GET("", h.List)
	}
}
