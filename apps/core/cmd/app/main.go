package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/router"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
	"github.com/gin-gonic/gin"

	_ "github.com/RevieU-Corp/revieu-backend/apps/core/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title RevieU Core API
// @version 1.0
// @description This is the core backend service for RevieU
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error(ctx, "Failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(logger.Config{
		Service: "revieu-core",
		Version: "1.0.0",
		Level:   cfg.Logger.Level,
	})

	// Set Gin debug output to JSON format
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.Debug(ctx, "Route registered",
			"method", httpMethod,
			"path", absolutePath,
			"handler", handlerName,
			"handlers_count", nuHandlers,
		)
	}

	// Connect to database
	if err := database.Connect(cfg.Database); err != nil {
		logger.Error(ctx, "Failed to connect to database", "error", err.Error())
		os.Exit(1)
	}

	// Auto migrate database schema
	if err := database.DB.AutoMigrate(
		&model.User{},
		&model.UserAuth{},
		&model.UserProfile{},
		&model.EmailVerification{},
		&model.RefreshToken{},
		&model.Merchant{},
		&model.Tag{},
		&model.Post{},
		&model.Review{},
		&model.UserFollow{},
		&model.MerchantFollow{},
		&model.Like{},
		&model.Favorite{},
		&model.UserAddress{},
		&model.UserPrivacy{},
		&model.UserNotification{},
		&model.AccountDeletion{},
		&model.MediaUpload{},
	); err != nil {
		logger.Error(ctx, "Failed to migrate database", "error", err.Error())
		os.Exit(1)
	}

	// Initialize Gin router with JSON logging
	gin.SetMode(gin.ReleaseMode)
	router := buildRouter(cfg)

	// Start server
	addr := cfg.Server.Address
	logger.Info(ctx, "Starting server", "address", addr)
	logger.Info(ctx, "Swagger documentation available", "url", "http://"+addr+cfg.Server.APIBasePath+"/swagger/index.html")
	if err := router.Run(addr); err != nil {
		logger.Error(ctx, "Failed to start server", "error", err.Error())
		os.Exit(1)
	}
}

func buildRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(jsonLoggerMiddleware())

	apiGroup := r.Group(cfg.Server.APIBasePath)
	{
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		apiGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	router.Setup(r, cfg)
	return r
}

func jsonLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info(c.Request.Context(), "HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency_ms", float64(latency.Nanoseconds())/1e6,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
