package service

import (
	"context"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type MerchantService struct {
	db *gorm.DB
}

func NewMerchantService(db *gorm.DB) *MerchantService {
	if db == nil {
		db = database.DB
	}
	return &MerchantService{db: db}
}

func (s *MerchantService) List(ctx context.Context, category string) ([]model.Merchant, error) {
	q := s.db.WithContext(ctx).Model(&model.Merchant{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var merchants []model.Merchant
	if err := q.Order("id desc").Find(&merchants).Error; err != nil {
		return nil, err
	}
	return merchants, nil
}

func (s *MerchantService) Detail(ctx context.Context, id int64) (*model.Merchant, error) {
	var merchant model.Merchant
	if err := s.db.WithContext(ctx).First(&merchant, id).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (s *MerchantService) Reviews(ctx context.Context, merchantID int64) ([]model.Review, error) {
	var reviews []model.Review
	if err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("id desc").Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}
