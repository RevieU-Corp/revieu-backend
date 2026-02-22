package service

import (
	"context"
	"errors"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
)

type CouponService struct {
	db *gorm.DB
}

func NewCouponService(db *gorm.DB) *CouponService {
	if db == nil {
		db = database.DB
	}
	return &CouponService{db: db}
}

func (s *CouponService) Validate(ctx context.Context, id int64) error {
	var coupon model.Coupon
	if err := s.db.WithContext(ctx).First(&coupon, id).Error; err != nil {
		return err
	}
	if !coupon.ExpiryDate.IsZero() && coupon.ExpiryDate.Before(time.Now()) {
		return errors.New("expired")
	}
	if coupon.ValidUntil != nil && coupon.ValidUntil.Before(time.Now()) {
		return errors.New("expired")
	}
	return nil
}

func (s *CouponService) InitiatePayment(ctx context.Context, couponID int64, userID string) error {
	_ = userID
	var coupon model.Coupon
	if err := s.db.WithContext(ctx).First(&coupon, couponID).Error; err != nil {
		return err
	}
	return nil
}

func (s *CouponService) Redeem(ctx context.Context, couponID, userID int64) error {
	_ = userID
	var coupon model.Coupon
	if err := s.db.WithContext(ctx).First(&coupon, couponID).Error; err != nil {
		return err
	}
	return nil
}
