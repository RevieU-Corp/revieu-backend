package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/service"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	svc *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

// Suggestions godoc
// @Summary Get review suggestions
// @Description Generates AI suggestions for improving a review
// @Tags ai
// @Accept json
// @Produce json
// @Param request body service.SuggestionsRequest true "Suggestions request"
// @Success 200 {object} service.SuggestionsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /ai/reviews/suggestions [post]
func (h *AIHandler) Suggestions(c *gin.Context) {
	var req service.SuggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.svc.Suggestions(req)
	c.JSON(http.StatusOK, resp)
}
