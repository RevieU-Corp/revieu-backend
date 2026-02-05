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

// HomeFeed godoc
// @Summary Get home feed
// @Description Returns the home feed items
// @Tags feed
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /feed/home [get]
func (h *FeedHandler) Home(c *gin.Context) {
	items, err := h.svc.Home(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
