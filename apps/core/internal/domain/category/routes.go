package category

import (
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/config"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/category/handler"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/category/service"
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
