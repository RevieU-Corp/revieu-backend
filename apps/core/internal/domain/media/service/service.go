package service

import (
	"context"
	"fmt"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type MediaService struct {
	db *gorm.DB
}

func NewMediaService(db *gorm.DB) *MediaService {
	if db == nil {
		db = database.DB
	}
	return &MediaService{db: db}
}

func (s *MediaService) CreateUpload(ctx context.Context) (model.MediaUpload, error) {
	upload := model.MediaUpload{
		UploadURL: "https://example.com/upload",
		FileURL:   fmt.Sprintf("https://example.com/files/%d", time.Now().UnixNano()),
	}
	return upload, s.db.WithContext(ctx).Create(&upload).Error
}

func (s *MediaService) Analyze(ctx context.Context, id int64) error {
	var upload model.MediaUpload
	return s.db.WithContext(ctx).First(&upload, id).Error
}
