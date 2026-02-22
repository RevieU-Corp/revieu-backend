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

// InitiatePaymentRequest is the request body for initiating a coupon payment.
type InitiatePaymentRequest struct {
	UserID string `json:"userId" binding:"required"`
}

func NewCouponHandler(svc *service.CouponService) *CouponHandler {
	if svc == nil {
		svc = service.NewCouponService(nil)
	}
	return &CouponHandler{svc: svc}
}

// ValidateCoupon godoc
// @Summary Validate coupon
// @Description Validates a coupon by ID
// @Tags coupon
// @Produce json
// @Param id path int true "Coupon ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /coupons/{id}/validate [post]
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

// InitiateCouponPayment godoc
// @Summary Initiate coupon payment
// @Description Initiates payment flow for a coupon
// @Tags coupon
// @Accept json
// @Produce json
// @Param id path int true "Coupon ID"
// @Param request body InitiatePaymentRequest true "Initiate payment request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /coupons/{id}/payment/initiate [post]
func (h *CouponHandler) InitiatePayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req InitiatePaymentRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.InitiatePayment(c.Request.Context(), id, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// RedeemCoupon godoc
// @Summary Redeem coupon
// @Description Redeems a coupon for the authenticated user
// @Tags coupon
// @Produce json
// @Param id path int true "Coupon ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /coupons/{id}/redeem [post]
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

// ListPackages godoc
// @Summary List packages
// @Description Returns a list of available packages
// @Tags package
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /packages [get]
func (h *CouponHandler) ListPackages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// PackageDetail godoc
// @Summary Get package detail
// @Description Returns a package by ID
// @Tags package
// @Produce json
// @Param id path int true "Package ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /packages/{id} [get]
func (h *CouponHandler) PackageDetail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}
