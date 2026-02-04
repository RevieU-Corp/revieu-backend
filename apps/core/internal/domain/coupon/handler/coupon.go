package handler

import (
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/coupon/service"
	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	svc *service.CouponService
}

func NewCouponHandler(svc *service.CouponService) *CouponHandler {
	if svc == nil {
		svc = service.NewCouponService(nil)
	}
	return &CouponHandler{svc: svc}
}

func (h *CouponHandler) Validate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Validate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *CouponHandler) InitiatePayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		UserID string `json:"userId"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.InitiatePayment(c.Request.Context(), id, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *CouponHandler) Redeem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.Redeem(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
