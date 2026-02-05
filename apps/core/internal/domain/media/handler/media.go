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

// CreateMediaUpload godoc
// @Summary Create media upload
// @Description Creates a media upload and returns upload URLs
// @Tags media
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /media/uploads [post]
func (h *MediaHandler) CreateUpload(c *gin.Context) {
	upload, err := h.svc.CreateUpload(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, upload)
}

// AnalyzeMedia godoc
// @Summary Analyze media upload
// @Description Triggers analysis for a media upload
// @Tags media
// @Produce json
// @Param id path int true "Upload ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /media/{id}/analysis [post]
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
