package service

import (
	"context"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVoucherTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Merchant{},
		&model.Store{},
		&model.Coupon{},
		&model.Voucher{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestVoucherServiceRedeemByMerchantAllowsSoftDeletedCoupon(t *testing.T) {
	db := setupVoucherTestDB(t)
	svc := NewVoucherService(db)

	merchantUserID := int64(901)
	customerUserID := int64(902)
	for _, id := range []int64{merchantUserID, customerUserID} {
		if err := db.Create(&model.User{ID: id, Role: "user", Status: 0}).Error; err != nil {
			t.Fatalf("failed to create user %d: %v", id, err)
		}
	}

	merchant := model.Merchant{Name: "Merchant Owner", UserID: &merchantUserID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Redeem Store", Status: 1}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	storeID := store.ID
	coupon := model.Coupon{
		MerchantID:    merchant.ID,
		StoreID:       &storeID,
		Title:         "Redeem Coupon",
		Type:          "cash",
		TotalQuantity: 100,
		MaxPerUser:    1,
		Status:        "active",
	}
	if err := db.Create(&coupon).Error; err != nil {
		t.Fatalf("failed to create coupon: %v", err)
	}

	voucher := model.Voucher{
		Code:       "voucher-soft-delete-test",
		CouponID:   coupon.ID,
		UserID:     customerUserID,
		MerchantID: &merchant.ID,
		Status:     "active",
	}
	if err := db.Create(&voucher).Error; err != nil {
		t.Fatalf("failed to create voucher: %v", err)
	}

	if err := db.Delete(&coupon).Error; err != nil {
		t.Fatalf("failed to soft-delete coupon: %v", err)
	}

	if err := svc.RedeemByMerchant(context.Background(), merchantUserID, voucher.ID); err != nil {
		t.Fatalf("redeem by merchant returned error: %v", err)
	}

	var refreshedVoucher model.Voucher
	if err := db.First(&refreshedVoucher, voucher.ID).Error; err != nil {
		t.Fatalf("failed to reload voucher: %v", err)
	}
	if refreshedVoucher.Status != "used" {
		t.Fatalf("expected voucher status used, got %q", refreshedVoucher.Status)
	}
	if refreshedVoucher.RedeemedBy == nil || *refreshedVoucher.RedeemedBy != merchantUserID {
		t.Fatalf("expected redeemed_by=%d, got %+v", merchantUserID, refreshedVoucher.RedeemedBy)
	}

	var refreshedCoupon model.Coupon
	if err := db.Unscoped().First(&refreshedCoupon, coupon.ID).Error; err != nil {
		t.Fatalf("failed to reload coupon unscoped: %v", err)
	}
	if refreshedCoupon.RedeemedCount != 1 {
		t.Fatalf("expected redeemed_count=1, got %d", refreshedCoupon.RedeemedCount)
	}
}
