package service

import (
	"context"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type InteractionService struct {
	db *gorm.DB
}

func NewInteractionService(db *gorm.DB) *InteractionService {
	if db == nil {
		db = database.DB
	}
	return &InteractionService{db: db}
}

func (s *InteractionService) Like(ctx context.Context, userID int64, targetType string, targetID int64) error {
	like := model.Like{UserID: userID, TargetType: targetType, TargetID: targetID}
	return s.db.WithContext(ctx).FirstOrCreate(&like, like).Error
}

func (s *InteractionService) Unlike(ctx context.Context, userID int64, targetType string, targetID int64) error {
	return s.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).Delete(&model.Like{}).Error
}

func (s *InteractionService) Favorite(ctx context.Context, userID int64, targetType string, targetID int64) error {
	fav := model.Favorite{UserID: userID, TargetType: targetType, TargetID: targetID}
	return s.db.WithContext(ctx).FirstOrCreate(&fav, fav).Error
}

func (s *InteractionService) Unfavorite(ctx context.Context, userID int64, targetType string, targetID int64) error {
	return s.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).Delete(&model.Favorite{}).Error
}
