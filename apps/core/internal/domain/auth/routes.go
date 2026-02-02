package auth

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers auth routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    _ = r
    _ = cfg
}
