package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/category/service"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	if svc == nil {
		svc = service.NewCategoryService(nil)
	}
	return &CategoryHandler{svc: svc}
}

// ListCategories godoc
// @Summary List categories
// @Description Returns a list of all categories
// @Tags category
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}
