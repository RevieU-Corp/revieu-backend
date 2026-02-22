package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/verification/service"
	"github.com/gin-gonic/gin"
)

type VerificationHandler struct {
	svc *service.VerificationService
}

func NewVerificationHandler(svc *service.VerificationService) *VerificationHandler {
	if svc == nil {
		svc = service.NewVerificationService(nil)
	}
	return &VerificationHandler{svc: svc}
}

// SubmitVerification godoc
// @Summary Submit merchant verification
// @Description Submits verification documents for the authenticated merchant
// @Tags verification
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /merchant/verification [post]
func (h *VerificationHandler) Submit(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// VerificationStatus godoc
// @Summary Get verification status
// @Description Returns the verification status for the authenticated merchant
// @Tags verification
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /merchant/verification [get]
func (h *VerificationHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}
