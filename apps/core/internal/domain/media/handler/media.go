package handler

import (
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/service"
	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	svc *service.MediaService
}

func NewMediaHandler(svc *service.MediaService) *MediaHandler {
	if svc == nil {
		svc = service.NewMediaService(nil)
	}
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) CreateUpload(c *gin.Context) {
	upload, err := h.svc.CreateUpload(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, upload)
}

func (h *MediaHandler) Analyze(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Analyze(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
