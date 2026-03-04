package service

import (
	"context"
	"errors"
	"testing"
	"time"

	userdto "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
	"gorm.io/gorm"
)

func TestUserServiceProfileAndSettings(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := NewUserService(db)
	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.UserProfile{UserID: user.ID, Nickname: "n"}).Error; err != nil {
		t.Fatal(err)
	}

	prof, err := svc.GetProfile(context.Background(), user.ID)
	if err != nil || prof.UserID != user.ID {
		t.Fatalf("profile failed: %v", err)
	}

	newName := "new"
	if err := svc.UpdateProfile(context.Background(), user.ID, userdto.UpdateProfileRequest{Nickname: &newName}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
}

func TestUserServiceAddressDefault(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := NewUserService(db)
	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	addr, err := svc.CreateAddress(context.Background(), user.ID, userdto.CreateAddressRequest{Name: "A", Phone: "1", Address: "X"})
	if err != nil || !addr.IsDefault {
		t.Fatalf("expected default address")
	}
}

func TestUserServiceExecuteDueAccountDeletionsCascadesMerchantData(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := NewUserService(db)

	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.UserAuth{
		UserID:       user.ID,
		IdentityType: "email",
		Identifier:   "cascade-user@example.com",
		Credential:   "pw",
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.UserProfile{
		UserID:   user.ID,
		Nickname: "cascade-user",
	}).Error; err != nil {
		t.Fatal(err)
	}

	merchant := model.Merchant{Name: "Cascade Merchant", UserID: &user.ID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatal(err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Cascade Store", Status: 1}
	if err := db.Create(&store).Error; err != nil {
		t.Fatal(err)
	}
	storeID := store.ID
	coupon := model.Coupon{
		MerchantID:    merchant.ID,
		StoreID:       &storeID,
		Title:         "Cascade Coupon",
		Type:          "single",
		TotalQuantity: 1,
		MaxPerUser:    1,
		Status:        "published",
	}
	if err := db.Create(&coupon).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.AccountDeletion{
		UserID:      user.ID,
		Reason:      "cleanup test",
		ScheduledAt: time.Now().UTC().Add(-1 * time.Minute),
	}).Error; err != nil {
		t.Fatal(err)
	}

	processed, err := svc.ExecuteDueAccountDeletions(context.Background(), time.Now().UTC(), 10)
	if err != nil {
		t.Fatalf("execute due deletions failed: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected processed=1, got %d", processed)
	}

	if err := db.First(&model.User{}, user.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected user deleted, got err=%v", err)
	}
	if err := db.First(&model.UserAuth{}, "user_id = ?", user.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected user auth deleted, got err=%v", err)
	}
	if err := db.First(&model.UserProfile{}, "user_id = ?", user.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected user profile deleted, got err=%v", err)
	}
	if err := db.First(&model.AccountDeletion{}, "user_id = ?", user.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected account deletion row cleared, got err=%v", err)
	}

	if err := db.First(&model.Merchant{}, merchant.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected merchant hidden by soft delete, got err=%v", err)
	}
	if err := db.First(&model.Store{}, store.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected store hidden by soft delete, got err=%v", err)
	}
	if err := db.First(&model.Coupon{}, coupon.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected coupon hidden by soft delete, got err=%v", err)
	}

	var deletedMerchant model.Merchant
	if err := db.Unscoped().First(&deletedMerchant, merchant.ID).Error; err != nil {
		t.Fatalf("expected soft-deleted merchant to exist unscoped: %v", err)
	}
	if !deletedMerchant.DeletedAt.Valid {
		t.Fatalf("expected merchant deleted_at to be set")
	}
	var deletedStore model.Store
	if err := db.Unscoped().First(&deletedStore, store.ID).Error; err != nil {
		t.Fatalf("expected soft-deleted store to exist unscoped: %v", err)
	}
	if !deletedStore.DeletedAt.Valid {
		t.Fatalf("expected store deleted_at to be set")
	}
	var deletedCoupon model.Coupon
	if err := db.Unscoped().First(&deletedCoupon, coupon.ID).Error; err != nil {
		t.Fatalf("expected soft-deleted coupon to exist unscoped: %v", err)
	}
	if !deletedCoupon.DeletedAt.Valid {
		t.Fatalf("expected coupon deleted_at to be set")
	}
}

func TestUserServiceExecuteDueAccountDeletionsIsIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := NewUserService(db)

	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.UserProfile{UserID: user.ID, Nickname: "retry"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.AccountDeletion{
		UserID:      user.ID,
		Reason:      "retry",
		ScheduledAt: time.Now().UTC().Add(-1 * time.Minute),
	}).Error; err != nil {
		t.Fatal(err)
	}

	processed, err := svc.ExecuteDueAccountDeletions(context.Background(), time.Now().UTC(), 10)
	if err != nil {
		t.Fatalf("first execution failed: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected first processed=1, got %d", processed)
	}

	processed, err = svc.ExecuteDueAccountDeletions(context.Background(), time.Now().UTC(), 10)
	if err != nil {
		t.Fatalf("second execution failed: %v", err)
	}
	if processed != 0 {
		t.Fatalf("expected second processed=0, got %d", processed)
	}
}
