package handler

import (
	"net/http"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/conversation/service"
	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	svc *service.ConversationService
}

func NewConversationHandler(svc *service.ConversationService) *ConversationHandler {
	if svc == nil {
		svc = service.NewConversationService(nil)
	}
	return &ConversationHandler{svc: svc}
}

// ListConversations godoc
// @Summary List conversations
// @Description Returns conversations for the authenticated user
// @Tags conversation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /conversations [get]
func (h *ConversationHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// CreateConversation godoc
// @Summary Create conversation
// @Description Creates a new conversation
// @Tags conversation
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /conversations [post]
func (h *ConversationHandler) Create(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// ConversationMessages godoc
// @Summary Get conversation messages
// @Description Returns messages for a conversation
// @Tags conversation
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /conversations/{id}/messages [get]
func (h *ConversationHandler) Messages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// SendMessage godoc
// @Summary Send message
// @Description Sends a message in a conversation
// @Tags conversation
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 201 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /conversations/{id}/messages [post]
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}

// UpdateConversationSettings godoc
// @Summary Update conversation settings
// @Description Updates settings for a conversation (e.g. mute)
// @Tags conversation
// @Accept json
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /conversations/{id}/settings [patch]
func (h *ConversationHandler) UpdateSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
}
