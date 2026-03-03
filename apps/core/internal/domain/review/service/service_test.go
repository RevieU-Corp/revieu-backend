package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
	"gorm.io/gorm"
)

func setupReviewServiceTest(t *testing.T) (*ReviewService, *gorm.DB, int64) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	svc := NewReviewService(db)

	user := model.User{Role: "user", Status: 0}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	return svc, db, user.ID
}

func TestCreateReviewMerchantMustExist(t *testing.T) {
	svc, _, userID := setupReviewServiceTest(t)

	_, err := svc.Create(context.Background(), userID, dto.Review{
		MerchantID: "99999",
		Rating:     4.5,
		Text:       "great",
	})
	if err == nil || err.Error() != "merchant not found" {
		t.Fatalf("expected merchant not found, got %v", err)
	}
}

func TestCreateReviewStoreMustExist(t *testing.T) {
	svc, db, userID := setupReviewServiceTest(t)

	merchant := model.Merchant{Name: "Cafe"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	_, err := svc.Create(context.Background(), userID, dto.Review{
		MerchantID: fmt.Sprintf("%d", merchant.ID),
		StoreID:    "99999",
		Rating:     4.5,
		Text:       "great",
	})
	if err == nil || err.Error() != "store not found" {
		t.Fatalf("expected store not found, got %v", err)
	}
}

func TestCreateReviewStoreMustBelongToMerchant(t *testing.T) {
	svc, db, userID := setupReviewServiceTest(t)

	merchantA := model.Merchant{Name: "A"}
	if err := db.Create(&merchantA).Error; err != nil {
		t.Fatalf("failed to create merchantA: %v", err)
	}
	merchantB := model.Merchant{Name: "B"}
	if err := db.Create(&merchantB).Error; err != nil {
		t.Fatalf("failed to create merchantB: %v", err)
	}

	store := model.Store{MerchantID: merchantB.ID, Name: "B-Store"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	_, err := svc.Create(context.Background(), userID, dto.Review{
		MerchantID: fmt.Sprintf("%d", merchantA.ID),
		StoreID:    fmt.Sprintf("%d", store.ID),
		Rating:     4.5,
		Text:       "great",
	})
	if err == nil || err.Error() != "store does not belong to merchant" {
		t.Fatalf("expected store ownership error, got %v", err)
	}
}

func TestReviewDetailPreloadsMerchantAndStore(t *testing.T) {
	svc, db, userID := setupReviewServiceTest(t)

	merchant := model.Merchant{Name: "Cafe", BusinessName: "Cafe Business", Address: "SF"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	store := model.Store{MerchantID: merchant.ID, Name: "Downtown"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	storeID := store.ID
	review := model.Review{
		UserID:     userID,
		VenueID:    merchant.ID,
		MerchantID: merchant.ID,
		StoreID:    &storeID,
		Rating:     5,
		Content:    "nice",
		VisitDate:  time.Now(),
	}
	if err := db.Create(&review).Error; err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	got, err := svc.Detail(context.Background(), review.ID)
	if err != nil {
		t.Fatalf("detail failed: %v", err)
	}
	if got.Merchant == nil {
		t.Fatalf("expected merchant preloaded")
	}
	if got.Store == nil {
		t.Fatalf("expected store preloaded")
	}
}
