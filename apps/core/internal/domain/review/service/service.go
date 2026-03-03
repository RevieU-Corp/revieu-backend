package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type ReviewService struct {
	db *gorm.DB
}

var ErrMerchantNotFound = errors.New("merchant not found")
var ErrStoreNotFound = errors.New("store not found")
var ErrStoreMerchantMismatch = errors.New("store does not belong to merchant")

func NewReviewService(db *gorm.DB) *ReviewService {
	if db == nil {
		db = database.DB
	}
	return &ReviewService{db: db}
}

func (s *ReviewService) ListByUser(ctx context.Context, userID int64) ([]model.Review, error) {
	var reviews []model.Review
	if err := s.db.WithContext(ctx).Preload("Merchant").Where("user_id = ?", userID).Order("id desc").Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

func (s *ReviewService) Detail(ctx context.Context, id int64) (*model.Review, error) {
	var review model.Review
	if err := s.db.WithContext(ctx).Preload("Merchant").Preload("Store").First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (s *ReviewService) Create(ctx context.Context, userID int64, req dto.Review) (model.Review, error) {
	merchantID, err := req.MerchantIDValue()
	if err != nil {
		return model.Review{}, err
	}

	var merchant model.Merchant
	if err := s.db.WithContext(ctx).Select("id").First(&merchant, merchantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Review{}, ErrMerchantNotFound
		}
		return model.Review{}, err
	}

	venueID := merchantID
	if req.VenueID != "" {
		venueID, err = req.VenueIDValue()
		if err != nil {
			return model.Review{}, err
		}
	}
	storeID, err := req.StoreIDValue()
	if err != nil {
		return model.Review{}, err
	}
	if storeID != nil {
		var store model.Store
		if err := s.db.WithContext(ctx).Select("id", "merchant_id").First(&store, *storeID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.Review{}, ErrStoreNotFound
			}
			return model.Review{}, err
		}
		if store.MerchantID != merchantID {
			return model.Review{}, ErrStoreMerchantMismatch
		}
	}
	visitDate, err := req.VisitDateValue()
	if err != nil {
		return model.Review{}, err
	}
	imagesJSON, _ := json.Marshal(req.Images)
	review := model.Review{
		UserID:     userID,
		MerchantID: merchantID,
		VenueID:    venueID,
		StoreID:    storeID,
		Rating:     float32(req.Rating),
		Content:    req.Text,
		Images:     string(imagesJSON),
		VisitDate:  visitDate,
	}
	if err := s.db.WithContext(ctx).Create(&review).Error; err != nil {
		return model.Review{}, err
	}
	return review, nil
}

func (s *ReviewService) Like(ctx context.Context, userID, reviewID int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var review model.Review
		if err := tx.First(&review, reviewID).Error; err != nil {
			return err
		}
		like := model.Like{UserID: userID, TargetType: "review", TargetID: reviewID}
		if err := tx.FirstOrCreate(&like, like).Error; err != nil {
			return err
		}
		return tx.Model(&review).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

func (s *ReviewService) Comment(ctx context.Context, userID, reviewID int64, text string) error {
	var review model.Review
	if err := s.db.WithContext(ctx).First(&review, reviewID).Error; err != nil {
		return err
	}
	comment := model.ReviewComment{ReviewID: reviewID, UserID: userID, Content: text}
	return s.db.WithContext(ctx).Create(&comment).Error
}
