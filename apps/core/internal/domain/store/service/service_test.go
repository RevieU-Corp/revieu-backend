package service

import (
	"context"
	"errors"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/store/dto"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&model.Merchant{}, &model.Store{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestStoreServiceCreate(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(101)
	merchant := model.Merchant{Name: "Demo Merchant", UserID: &userID, TotalStores: 0}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	store, err := svc.Create(context.Background(), userID, dto.CreateStoreRequest{
		Name:          "Central Branch",
		City:          "Austin",
		Country:       "US",
		Images:        []string{"https://img.example/1.jpg", "https://img.example/2.jpg"},
		CoverImageURL: "https://img.example/cover.jpg",
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	if store.ID == 0 {
		t.Fatalf("expected store id to be set")
	}
	if store.MerchantID != merchant.ID {
		t.Fatalf("unexpected merchant id: got %d, want %d", store.MerchantID, merchant.ID)
	}
	if store.Images != "[\"https://img.example/1.jpg\",\"https://img.example/2.jpg\"]" {
		t.Fatalf("unexpected images json: %s", store.Images)
	}

	var refreshed model.Merchant
	if err := db.First(&refreshed, merchant.ID).Error; err != nil {
		t.Fatalf("failed to reload merchant: %v", err)
	}
	if refreshed.TotalStores != 1 {
		t.Fatalf("unexpected total_stores: got %d, want 1", refreshed.TotalStores)
	}
}

func TestStoreServiceCreateMerchantNotFound(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	_, err := svc.Create(context.Background(), 9999, dto.CreateStoreRequest{Name: "Missing Merchant Store"})
	if !errors.Is(err, ErrMerchantNotFound) {
		t.Fatalf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestStoreServiceCreateDefaultName(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(202)
	merchant := model.Merchant{Name: "Fallback Merchant", UserID: &userID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	store, err := svc.Create(context.Background(), userID, dto.CreateStoreRequest{})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	if store.Name != "Fallback Merchant" {
		t.Fatalf("unexpected default name: got %q", store.Name)
	}
}
