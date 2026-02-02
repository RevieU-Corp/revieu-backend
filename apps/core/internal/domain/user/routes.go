package user

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers user routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    _ = r
    _ = cfg
}
