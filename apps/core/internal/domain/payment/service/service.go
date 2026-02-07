package service

import (
	"context"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type CreatePaymentRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
}

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	if db == nil {
		db = database.DB
	}
	return &PaymentService{db: db}
}

func (s *PaymentService) Create(ctx context.Context, req CreatePaymentRequest) (model.Payment, error) {
	p := model.Payment{Amount: req.Amount, Currency: req.Currency, Status: req.Status}
	return p, s.db.WithContext(ctx).Create(&p).Error
}

func (s *PaymentService) Detail(ctx context.Context, id int64) (*model.Payment, error) {
	var p model.Payment
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
