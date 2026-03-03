package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/store/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StoreService struct {
	db *gorm.DB
}

const (
	StoreStatusDraft     int16 = 0
	StoreStatusPublished int16 = 1
	StoreStatusHidden    int16 = 2
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrStoreNotFound  = errors.New("store not found")
	ErrStoreForbidden = errors.New("store forbidden")
)

func NewStoreService(db *gorm.DB) *StoreService {
	if db == nil {
		db = database.DB
	}
	return &StoreService{db: db}
}

func (s *StoreService) Create(ctx context.Context, userID int64, req dto.CreateStoreRequest) (*model.Store, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var merchant model.Merchant
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&merchant).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		merchant = model.Merchant{
			UserID:             &userID,
			Name:               fmt.Sprintf("merchant-%d", userID),
			VerificationStatus: "unverified",
		}
		if err := s.db.WithContext(ctx).Create(&merchant).Error; err != nil {
			return nil, err
		}
	}

	imagesRaw, err := json.Marshal(req.Images)
	if err != nil {
		return nil, err
	}

	storeName := req.Name
	if storeName == "" {
		if merchant.BusinessName != "" {
			storeName = merchant.BusinessName
		} else if merchant.Name != "" {
			storeName = merchant.Name
		} else {
			storeName = fmt.Sprintf("store-%d", merchant.ID)
		}
	}

	store := model.Store{
		MerchantID:    merchant.ID,
		Name:          storeName,
		Description:   req.Description,
		Address:       req.Address,
		City:          req.City,
		State:         req.State,
		ZipCode:       req.ZipCode,
		Country:       req.Country,
		Phone:         req.Phone,
		Website:       req.Website,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		CoverImageURL: req.CoverImageURL,
		Images:        string(imagesRaw),
		Status:        StoreStatusDraft,
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&store).Error; err != nil {
			return err
		}
		if len(req.Hours) > 0 {
			hours := make([]model.StoreHour, 0, len(req.Hours))
			for _, h := range req.Hours {
				hours = append(hours, model.StoreHour{
					StoreID:   store.ID,
					DayOfWeek: h.DayOfWeek,
					OpenTime:  h.OpenTime,
					CloseTime: h.CloseTime,
					IsClosed:  h.IsClosed,
				})
			}
			if err := tx.Create(&hours).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.Merchant{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", merchant.ID).
			UpdateColumn("total_stores", gorm.Expr("total_stores + 1")).Error
	}); err != nil {
		return nil, err
	}

	return &store, nil
}

func (s *StoreService) ListPublished(ctx context.Context) ([]model.Store, error) {
	var stores []model.Store
	if err := s.db.WithContext(ctx).
		Where("status = ?", StoreStatusPublished).
		Order("id desc").
		Find(&stores).Error; err != nil {
		return nil, err
	}
	return stores, nil
}

func (s *StoreService) DetailPublished(ctx context.Context, storeID int64) (*model.Store, error) {
	var store model.Store
	if err := s.db.WithContext(ctx).
		Where("id = ? AND status = ?", storeID, StoreStatusPublished).
		First(&store).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}
	return &store, nil
}

func (s *StoreService) ReviewsPublished(ctx context.Context, storeID int64) ([]model.Review, error) {
	if _, err := s.DetailPublished(ctx, storeID); err != nil {
		return nil, err
	}
	var reviews []model.Review
	if err := s.db.WithContext(ctx).
		Where("store_id = ?", storeID).
		Order("id desc").
		Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

func (s *StoreService) HoursPublished(ctx context.Context, storeID int64) ([]model.StoreHour, error) {
	if _, err := s.DetailPublished(ctx, storeID); err != nil {
		return nil, err
	}
	var hours []model.StoreHour
	if err := s.db.WithContext(ctx).
		Where("store_id = ?", storeID).
		Order("day_of_week asc").
		Find(&hours).Error; err != nil {
		return nil, err
	}
	return hours, nil
}

func (s *StoreService) ListMine(ctx context.Context, userID int64) ([]model.Store, error) {
	var stores []model.Store
	if err := s.db.WithContext(ctx).
		Model(&model.Store{}).
		Joins("JOIN merchants ON merchants.id = stores.merchant_id").
		Where("merchants.user_id = ?", userID).
		Order("stores.id desc").
		Find(&stores).Error; err != nil {
		return nil, err
	}
	return stores, nil
}

func (s *StoreService) Activate(ctx context.Context, userID, storeID int64) error {
	return s.updateStatusOwned(ctx, userID, storeID, StoreStatusPublished)
}

func (s *StoreService) Deactivate(ctx context.Context, userID, storeID int64) error {
	return s.updateStatusOwned(ctx, userID, storeID, StoreStatusHidden)
}

func (s *StoreService) updateStatusOwned(ctx context.Context, userID, storeID int64, toStatus int16) error {
	var merchant model.Merchant
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStoreForbidden
		}
		return err
	}

	var store model.Store
	if err := s.db.WithContext(ctx).First(&store, storeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStoreNotFound
		}
		return err
	}
	if store.MerchantID != merchant.ID {
		return ErrStoreForbidden
	}
	if store.Status == toStatus {
		return nil
	}
	return s.db.WithContext(ctx).
		Model(&model.Store{}).
		Where("id = ?", storeID).
		UpdateColumn("status", toStatus).Error
}
