package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TestResponse represents the test endpoint response
type TestResponse struct {
	Message string `json:"message" example:"Hello, World!"`
	Status  string `json:"status" example:"success"`
}

// TestHandler handles test endpoint requests
type TestHandler struct{}

// NewTestHandler creates a new test handler
func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

// GetTest godoc
// @Summary Test endpoint
// @Description Returns a test message to verify API is working
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {object} TestResponse
// @Router /api/v1/test [get]
func (h *TestHandler) GetTest(c *gin.Context) {
	response := TestResponse{
		Message: "Hello, World!",
		Status:  "success",
	}
	c.JSON(http.StatusOK, response)
}

// PostTest godoc
// @Summary Test POST endpoint
// @Description Echoes back the received message
// @Tags test
// @Accept json
// @Produce json
// @Param message body string true "Message to echo"
// @Success 200 {object} TestResponse
// @Router /api/v1/test [post]
func (h *TestHandler) PostTest(c *gin.Context) {
	var request struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := TestResponse{
		Message: request.Message,
		Status:  "success",
	}
	c.JSON(http.StatusOK, response)
}
