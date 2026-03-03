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

func TestSetupRegistersAuthRefreshRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:            "test-secret",
			ExpireHour:        24,
			RefreshExpireHour: 168,
		},
	}

	Setup(r, cfg)

	found := false
	for _, route := range r.Routes() {
		if route.Method == "POST" && route.Path == "/auth/refresh" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected POST /auth/refresh to be registered")
	}
}
