package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTooManyFiles       = errors.New("too many files, maximum 10 allowed")
	ErrInvalidContentType = errors.New("invalid content type, only image/jpeg, image/png, image/gif, image/webp allowed")
)

var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type MediaService struct {
	db       *gorm.DB
	r2Client *storage.R2Client
}

func NewMediaService(db *gorm.DB, r2Client *storage.R2Client) *MediaService {
	if db == nil {
		db = database.DB
	}
	return &MediaService{db: db, r2Client: r2Client}
}

type FileRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

type PresignedURLRequest struct {
	Files []FileRequest `json:"files"`
}

type UploadInfo struct {
	ID        string    `json:"id"`
	UploadURL string    `json:"upload_url"`
	FileURL   string    `json:"file_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PresignedURLResponse struct {
	Uploads []UploadInfo `json:"uploads"`
}

func (s *MediaService) CreatePresignedURLs(ctx context.Context, userID int64, req *PresignedURLRequest) (*PresignedURLResponse, error) {
	if len(req.Files) > 10 {
		return nil, ErrTooManyFiles
	}

	if len(req.Files) == 0 {
		return &PresignedURLResponse{Uploads: []UploadInfo{}}, nil
	}

	response := &PresignedURLResponse{
		Uploads: make([]UploadInfo, 0, len(req.Files)),
	}

	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	for _, file := range req.Files {
		ext, ok := allowedContentTypes[strings.ToLower(file.ContentType)]
		if !ok {
			return nil, ErrInvalidContentType
		}

		fileUUID := uuid.New().String()
		objectKey := fmt.Sprintf("uploads/%s/%s/%s%s", year, month, fileUUID, ext)

		result, err := s.r2Client.GeneratePresignedURL(ctx, objectKey, file.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL for %s: %w", file.Filename, err)
		}

		upload := model.MediaUpload{
			UUID:      fileUUID,
			UserID:    userID,
			ObjectKey: objectKey,
			FileURL:   result.FileURL,
			Status:    "pending",
		}

		if err := s.db.WithContext(ctx).Create(&upload).Error; err != nil {
			return nil, fmt.Errorf("failed to save upload record: %w", err)
		}

		response.Uploads = append(response.Uploads, UploadInfo{
			ID:        fileUUID,
			UploadURL: result.UploadURL,
			FileURL:   result.FileURL,
			ExpiresAt: result.ExpiresAt,
		})
	}

	return response, nil
}

func (s *MediaService) CreateUpload(ctx context.Context) (model.MediaUpload, error) {
	upload := model.MediaUpload{
		UUID:      uuid.New().String(),
		ObjectKey: fmt.Sprintf("uploads/%d", time.Now().UnixNano()),
		FileURL:   fmt.Sprintf("https://example.com/files/%d", time.Now().UnixNano()),
		Status:    "pending",
	}
	return upload, s.db.WithContext(ctx).Create(&upload).Error
}

func (s *MediaService) Analyze(ctx context.Context, id int64) error {
	var upload model.MediaUpload
	return s.db.WithContext(ctx).First(&upload, id).Error
}

// getExtensionFromFilename extracts the file extension from a filename
func getExtensionFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	return strings.ToLower(ext)
}
