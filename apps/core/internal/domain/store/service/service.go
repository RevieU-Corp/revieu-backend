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

var ErrMerchantNotFound = errors.New("merchant not found")

func NewStoreService(db *gorm.DB) *StoreService {
	if db == nil {
		db = database.DB
	}
	return &StoreService{db: db}
}

func (s *StoreService) Create(ctx context.Context, userID int64, req dto.CreateStoreRequest) (*model.Store, error) {
	var merchant model.Merchant
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMerchantNotFound
		}
		return nil, err
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
		return tx.Model(&model.Merchant{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", merchant.ID).
			UpdateColumn("total_stores", gorm.Expr("total_stores + 1")).Error
	}); err != nil {
		return nil, err
	}

	return &store, nil
}
