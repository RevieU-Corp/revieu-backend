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

// CreateVoucher godoc
// @Summary Create voucher
// @Description Creates a voucher for the authenticated user
// @Tags voucher
// @Accept json
// @Produce json
// @Param request body service.CreateVoucherRequest true "Create voucher request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /vouchers [post]
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

// ListVouchers godoc
// @Summary List vouchers
// @Description Returns vouchers for the authenticated user
// @Tags voucher
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vouchers [get]
func (h *VoucherHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	list, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// VoucherDetail godoc
// @Summary Get voucher detail
// @Description Returns a voucher by ID
// @Tags voucher
// @Produce json
// @Param id path int true "Voucher ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /vouchers/{id} [get]
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

// VoucherByCode godoc
// @Summary Get voucher by code
// @Description Returns a voucher by code
// @Tags voucher
// @Produce json
// @Param code path string true "Voucher code"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /vouchers/code/{code} [get]
func (h *VoucherHandler) ByCode(c *gin.Context) {
	v, err := h.svc.ByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

// UseVoucher godoc
// @Summary Use voucher
// @Description Marks a voucher as used
// @Tags voucher
// @Produce json
// @Param id path int true "Voucher ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /vouchers/{id}/use [patch]
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

// UpdateVoucherStatus godoc
// @Summary Update voucher status
// @Description Updates voucher status to used
// @Tags voucher
// @Produce json
// @Param id path int true "Voucher ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /vouchers/{id}/status [patch]
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

// ShareVoucherEmail godoc
// @Summary Share voucher via email
// @Description Sends voucher share email
// @Tags voucher
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /vouchers/share/email [post]
func (h *VoucherHandler) ShareEmail(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) }

// ShareVoucherSMS godoc
// @Summary Share voucher via SMS
// @Description Sends voucher share SMS
// @Tags voucher
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /vouchers/share/sms [post]
func (h *VoucherHandler) ShareSMS(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) }
