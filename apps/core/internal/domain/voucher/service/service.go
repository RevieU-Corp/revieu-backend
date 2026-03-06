package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrVoucherNotFound      = errors.New("voucher not found")
	ErrVoucherForbidden     = errors.New("voucher forbidden")
	ErrVoucherNotRedeemable = errors.New("voucher not redeemable")
	ErrVoucherExpired       = errors.New("voucher expired")
)

type CreateVoucherRequest struct {
	CouponID string `json:"couponId"`
	UserID   string `json:"userId"`
	Code     string `json:"code"`
}

type VoucherService struct {
	db *gorm.DB
}

func NewVoucherService(db *gorm.DB) *VoucherService {
	if db == nil {
		db = database.DB
	}
	return &VoucherService{db: db}
}

func (s *VoucherService) Create(ctx context.Context, req CreateVoucherRequest) (model.Voucher, error) {
	couponID, _ := strconv.ParseInt(req.CouponID, 10, 64)
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)

	// Look up coupon to get merchant ID
	var coupon model.Coupon
	var merchantID *int64
	if err := s.db.WithContext(ctx).First(&coupon, couponID).Error; err == nil {
		merchantID = &coupon.MerchantID
	}

	scanToken, err := generateVoucherScanToken()
	if err != nil {
		return model.Voucher{}, err
	}

	v := model.Voucher{
		Code:       req.Code,
		ScanToken:  scanToken,
		CouponID:   couponID,
		UserID:     userID,
		MerchantID: merchantID,
		Status:     "active",
	}
	return v, s.db.WithContext(ctx).Create(&v).Error
}

func (s *VoucherService) List(ctx context.Context, userID int64) ([]model.Voucher, error) {
	var list []model.Voucher
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *VoucherService) Detail(ctx context.Context, id int64) (*model.Voucher, error) {
	var v model.Voucher
	if err := s.db.WithContext(ctx).First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *VoucherService) ByCode(ctx context.Context, code string) (*model.Voucher, error) {
	var v model.Voucher
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *VoucherService) Use(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Model(&model.Voucher{}).Where("id = ?", id).UpdateColumn("status", "used").Error
}

func (s *VoucherService) UpdateStatus(ctx context.Context, id int64, status string) error {
	return s.db.WithContext(ctx).Model(&model.Voucher{}).Where("id = ?", id).UpdateColumn("status", status).Error
}

func (s *VoucherService) RedeemByMerchant(ctx context.Context, userID, voucherID int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var merchant model.Merchant
		if err := tx.Where("user_id = ?", userID).First(&merchant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVoucherForbidden
			}
			return err
		}

		var voucher model.Voucher
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&voucher, voucherID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVoucherNotFound
			}
			return err
		}

		var coupon model.Coupon
		if err := tx.Unscoped().First(&coupon, voucher.CouponID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVoucherNotFound
			}
			return err
		}
		if coupon.StoreID == nil || coupon.MerchantID != merchant.ID {
			return ErrVoucherForbidden
		}

		now := time.Now()
		if voucher.Status != "active" {
			return ErrVoucherNotRedeemable
		}
		if voucher.ValidUntil != nil && voucher.ValidUntil.Before(now) {
			return ErrVoucherExpired
		}
		if !voucher.ExpiryDate.IsZero() && voucher.ExpiryDate.Before(now) {
			return ErrVoucherExpired
		}

		if err := tx.Model(&model.Voucher{}).
			Where("id = ?", voucher.ID).
			Updates(map[string]interface{}{
				"status":      "used",
				"redeemed_at": now,
				"redeemed_by": userID,
			}).Error; err != nil {
			return err
		}

		return tx.Unscoped().Model(&model.Coupon{}).
			Where("id = ?", coupon.ID).
			UpdateColumn("redeemed_count", gorm.Expr("redeemed_count + 1")).Error
	})
}

func generateVoucherScanToken() (string, error) {
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate voucher scan token: %w", err)
	}
	return "vst_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
