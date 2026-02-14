package service

import (
	"context"
	"strconv"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"gorm.io/gorm"
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

	v := model.Voucher{Code: req.Code, CouponID: couponID, UserID: userID, MerchantID: merchantID, Status: "active"}
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
