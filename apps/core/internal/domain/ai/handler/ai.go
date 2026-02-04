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

func (h *AIHandler) Suggestions(c *gin.Context) {
	var req service.SuggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.svc.Suggestions(req)
	c.JSON(http.StatusOK, resp)
}
