package service

import (
	"context"
	"strconv"
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

func TestVoucherServiceCreateAssignsUniqueScanTokens(t *testing.T) {
	db := setupVoucherTestDB(t)
	svc := NewVoucherService(db)

	user := model.User{ID: 1001, Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	merchantOwner := model.User{ID: 1002, Role: "user", Status: 0}
	if err := db.Create(&merchantOwner).Error; err != nil {
		t.Fatalf("failed to create merchant owner: %v", err)
	}

	merchant := model.Merchant{Name: "Create Merchant", UserID: &merchantOwner.ID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	store := model.Store{MerchantID: merchant.ID, Name: "Create Store", Status: 1}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	coupon := model.Coupon{
		MerchantID:    merchant.ID,
		StoreID:       &store.ID,
		Title:         "Create Coupon",
		Type:          "cash",
		TotalQuantity: 100,
		MaxPerUser:    5,
		Status:        "active",
	}
	if err := db.Create(&coupon).Error; err != nil {
		t.Fatalf("failed to create coupon: %v", err)
	}

	first, err := svc.Create(context.Background(), CreateVoucherRequest{
		CouponID: strconv.FormatInt(coupon.ID, 10),
		UserID:   strconv.FormatInt(user.ID, 10),
		Code:     "CREATE-ONE",
	})
	if err != nil {
		t.Fatalf("failed to create first voucher: %v", err)
	}
	second, err := svc.Create(context.Background(), CreateVoucherRequest{
		CouponID: strconv.FormatInt(coupon.ID, 10),
		UserID:   strconv.FormatInt(user.ID, 10),
		Code:     "CREATE-TWO",
	})
	if err != nil {
		t.Fatalf("failed to create second voucher: %v", err)
	}

	if first.ScanToken == "" {
		t.Fatalf("expected first voucher scan token to be populated")
	}
	if second.ScanToken == "" {
		t.Fatalf("expected second voucher scan token to be populated")
	}
	if first.ScanToken == second.ScanToken {
		t.Fatalf("expected voucher scan tokens to be unique, got %q", first.ScanToken)
	}
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
