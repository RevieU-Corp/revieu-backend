package handler

import (
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/service"
	reviewdto "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/dto"
	"github.com/gin-gonic/gin"
)

type MerchantHandler struct {
	svc *service.MerchantService
}

func NewMerchantHandler(svc *service.MerchantService) *MerchantHandler {
	if svc == nil {
		svc = service.NewMerchantService(nil)
	}
	return &MerchantHandler{svc: svc}
}

// ListMerchants godoc
// @Summary List merchants
// @Description Returns merchants, optionally filtered by category
// @Tags merchant
// @Produce json
// @Param category query string false "Category filter"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /merchants [get]
func (h *MerchantHandler) List(c *gin.Context) {
	category := c.Query("category")
	merchants, err := h.svc.List(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]dto.Merchant, 0, len(merchants))
	for _, m := range merchants {
		items = append(items, dto.FromModel(m))
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// MerchantDetail godoc
// @Summary Get merchant detail
// @Description Returns a merchant by ID
// @Tags merchant
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} dto.Merchant
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /merchants/{id} [get]
func (h *MerchantHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	merchant, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, dto.FromModel(*merchant))
}

// MerchantReviews godoc
// @Summary List merchant reviews
// @Description Returns reviews for a merchant
// @Tags merchant
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /merchants/{id}/reviews [get]
func (h *MerchantHandler) Reviews(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	reviews, err := h.svc.Reviews(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reviewdto.FromModels(reviews)})
}
