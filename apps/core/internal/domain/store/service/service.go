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

var ErrUserNotFound = errors.New("user not found")

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
