package handler

import (
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/voucher/service"
	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	svc *service.VoucherService
}

func NewVoucherHandler(svc *service.VoucherService) *VoucherHandler {
	if svc == nil {
		svc = service.NewVoucherService(nil)
	}
	return &VoucherHandler{svc: svc}
}

func (h *VoucherHandler) Create(c *gin.Context) {
	var req service.CreateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *VoucherHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	list, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *VoucherHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	v, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VoucherHandler) ByCode(c *gin.Context) {
	v, err := h.svc.ByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VoucherHandler) Use(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Use(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *VoucherHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, "used"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *VoucherHandler) ShareEmail(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) }
func (h *VoucherHandler) ShareSMS(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{}) }
