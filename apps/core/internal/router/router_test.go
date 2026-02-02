package router

import (
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

func TestSetupExists(t *testing.T) {
    r := gin.New()
    cfg := &config.Config{}
    Setup(r, cfg)
}
