package profile

import (
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/profile/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc service.Service
	db  *gorm.DB
}

func NewHandler(svc service.Service) *Handler {
	if svc == nil {
		svc = service.NewService(nil)
	}
	return &Handler{svc: svc, db: database.DB}
}

// GetPublicProfile godoc
// @Summary Get public user profile
// @Description Returns a user's public profile
// @Tags profile
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} service.PublicProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (h *Handler) GetPublicProfile(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	viewerID := c.GetInt64("user_id")

	var privacy model.UserPrivacy
	if err := h.db.WithContext(c.Request.Context()).First(&privacy, "user_id = ?", targetID).Error; err == nil {
		if !privacy.IsPublic && viewerID != targetID {
			c.JSON(http.StatusForbidden, gin.H{"error": "profile is private"})
			return
		}
	}

	profile, err := h.svc.GetPublicProfile(c.Request.Context(), targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if viewerID != 0 && viewerID != targetID {
		var follow model.UserFollow
		if err := h.db.WithContext(c.Request.Context()).Where("follower_id = ? AND following_id = ?", viewerID, targetID).First(&follow).Error; err == nil {
			profile.IsFollowing = true
		}
	}

	c.JSON(http.StatusOK, profile)
}
