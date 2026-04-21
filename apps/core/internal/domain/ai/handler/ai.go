package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/service"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	svc *service.AIService
}

// SuggestionsRequest is the request body for generating suggestions.
type SuggestionsRequest struct {
	OverallRating    float64 `json:"overallRating"`
	BusinessCategory string  `json:"businessCategory"`
	CurrentText      string  `json:"currentText"`
	MerchantName     string  `json:"merchantName"`
}

// SuggestionsResponse is the response body for generated suggestions.
type SuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
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
// @Param request body SuggestionsRequest true "Suggestions request"
// @Success 200 {object} SuggestionsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /ai/reviews/suggestions [post]
func (h *AIHandler) Suggestions(c *gin.Context) {
	var req SuggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.svc.Suggestions(service.SuggestionsRequest{
		OverallRating:    req.OverallRating,
		BusinessCategory: req.BusinessCategory,
		CurrentText:      req.CurrentText,
		MerchantName:     req.MerchantName,
	})
	c.JSON(http.StatusOK, SuggestionsResponse{Suggestions: resp.Suggestions})
}
