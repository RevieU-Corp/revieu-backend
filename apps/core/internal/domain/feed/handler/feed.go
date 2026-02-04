package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed/service"
	"github.com/gin-gonic/gin"
)

type FeedHandler struct {
	svc *service.FeedService
}

func NewFeedHandler(svc *service.FeedService) *FeedHandler {
	if svc == nil {
		svc = service.NewFeedService(nil)
	}
	return &FeedHandler{svc: svc}
}

func (h *FeedHandler) Home(c *gin.Context) {
	items, err := h.svc.Home(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
