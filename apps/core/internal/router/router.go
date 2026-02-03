package router

import (
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/auth"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/profile"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user"
	"github.com/gin-gonic/gin"
)

// Setup registers all domain routes under the API base path.
func Setup(router *gin.Engine, cfg *config.Config) {
	api := router.Group(cfg.Server.APIBasePath)

	auth.RegisterRoutes(api, cfg)
	user.RegisterRoutes(api, cfg)
	profile.RegisterRoutes(api, cfg)
	follow.RegisterRoutes(api, cfg)
	content.RegisterRoutes(api, cfg)
}
