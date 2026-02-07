package media

import (
	"log"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/handler"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/service"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/storage"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers media routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
	// Initialize R2 client
	r2Client, err := storage.NewR2Client(storage.R2Config{
		AccountID:       cfg.R2.AccountID,
		AccessKeyID:     cfg.R2.AccessKeyID,
		SecretAccessKey: cfg.R2.SecretAccessKey,
		BucketName:      cfg.R2.BucketName,
		PublicURL:       cfg.R2.PublicURL,
	})
	if err != nil {
		log.Printf("Warning: Failed to initialize R2 client: %v", err)
	}

	svc := service.NewMediaService(nil, r2Client)
	h := handler.NewMediaHandler(svc)

	media := r.Group("/media", middleware.JWTAuth(cfg.JWT))
	{
		media.POST("/uploads", h.CreateUpload)
		media.POST("/presigned-urls", h.CreatePresignedURLs)
		media.POST("/:id/analysis", h.Analyze)
	}
}
