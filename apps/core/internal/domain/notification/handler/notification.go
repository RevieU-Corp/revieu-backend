package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/notification/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	if svc == nil {
		svc = service.NewNotificationService(nil)
	}
	return &NotificationHandler{svc: svc}
}

// ListNotifications godoc
// @Summary List notifications
// @Description Returns notifications for the authenticated user
// @Tags notification
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// MarkNotificationRead godoc
// @Summary Mark notification as read
// @Description Marks a single notification as read
// @Tags notification
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// ReadAllNotifications godoc
// @Summary Mark all notifications as read
// @Description Marks all notifications as read for the authenticated user
// @Tags notification
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /notifications/read-all [post]
func (h *NotificationHandler) ReadAll(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}
