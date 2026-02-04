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
