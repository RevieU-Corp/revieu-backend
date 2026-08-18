package service

import (
	"context"

	"github.com/revieu-corp/revieu-core-api-go/apps/core/internal/domain/feed/dto"
	"github.com/revieu-corp/revieu-core-api-go/apps/core/pkg/database"
	"gorm.io/gorm"
)

type FeedService struct {
	db *gorm.DB
}

func NewFeedService(db *gorm.DB) *FeedService {
	if db == nil {
		db = database.DB
	}
	return &FeedService{db: db}
}

func (s *FeedService) Home(_ context.Context) ([]dto.FeedItem, error) {
	return []dto.FeedItem{}, nil
}
