package handler

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/gin-gonic/gin"
)

func getUserID(c *gin.Context) int64 {
	return c.GetInt64(middleware.UserIDKey)
}
