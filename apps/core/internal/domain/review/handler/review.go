package handler

import (
	"net/http"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/service"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	if svc == nil {
		svc = service.NewReviewService(nil)
	}
	return &ReviewHandler{svc: svc}
}

func (h *ReviewHandler) ListMyReviews(c *gin.Context) {
	userID := c.GetInt64("user_id")
	reviews, err := h.svc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dto.FromModels(reviews)})
}

func (h *ReviewHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req dto.Review
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.FromModel(created))
}

func (h *ReviewHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	review, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, dto.FromModel(*review))
}

func (h *ReviewHandler) Like(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.Like(c.Request.Context(), userID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *ReviewHandler) Comment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetInt64("user_id")
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid text"})
		return
	}
	if err := h.svc.Comment(c.Request.Context(), userID, id, req.Text); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}
