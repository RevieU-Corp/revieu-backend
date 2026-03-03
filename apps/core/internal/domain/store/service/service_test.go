package service

import (
	"context"
	"encoding/json"
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

	if err := db.AutoMigrate(&model.User{}, &model.Merchant{}, &model.Store{}, &model.StoreHour{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func TestStoreServiceCreate(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(101)
	if err := db.Create(&model.User{ID: userID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
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
		Hours: []dto.StoreHourRequest{
			{
				DayOfWeek: 1,
				OpenTime:  "09:00",
				CloseTime: "18:00",
				IsClosed:  false,
			},
			{
				DayOfWeek: 2,
				IsClosed:  true,
			},
		},
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

	var hours []model.StoreHour
	if err := db.Where("store_id = ?", store.ID).Order("day_of_week asc").Find(&hours).Error; err != nil {
		t.Fatalf("failed to query store hours: %v", err)
	}
	if len(hours) != 2 {
		t.Fatalf("unexpected store_hours count: got %d, want 2", len(hours))
	}
	if hours[0].DayOfWeek != 1 || hours[0].OpenTime != "09:00" || hours[0].CloseTime != "18:00" || hours[0].IsClosed {
		t.Fatalf("unexpected first hour row: %+v", hours[0])
	}
	if hours[1].DayOfWeek != 2 || !hours[1].IsClosed {
		t.Fatalf("unexpected second hour row: %+v", hours[1])
	}
}

func TestStoreServiceCreateAutoCreatesMerchantPlaceholder(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(303)
	if err := db.Create(&model.User{ID: userID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	store, err := svc.Create(context.Background(), userID, dto.CreateStoreRequest{
		Name: "No Merchant Yet",
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if store.ID == 0 {
		t.Fatalf("expected store id to be set")
	}

	var merchant model.Merchant
	if err := db.Where("user_id = ?", userID).First(&merchant).Error; err != nil {
		t.Fatalf("failed to query placeholder merchant: %v", err)
	}
	if merchant.ID == 0 {
		t.Fatalf("expected placeholder merchant id to be set")
	}
	if merchant.TotalStores != 1 {
		t.Fatalf("unexpected total_stores for placeholder merchant: got %d, want 1", merchant.TotalStores)
	}
	if store.MerchantID != merchant.ID {
		t.Fatalf("unexpected merchant id on store: got %d, want %d", store.MerchantID, merchant.ID)
	}
}

func TestStoreServiceCreateDefaultName(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(202)
	if err := db.Create(&model.User{ID: userID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
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

func TestStoreServiceCreateUserNotFound(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	_, err := svc.Create(context.Background(), 9999, dto.CreateStoreRequest{Name: "Missing User"})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestStoreServiceUpdateOwnStore(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	ownerID := int64(401)
	if err := db.Create(&model.User{ID: ownerID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create owner user: %v", err)
	}
	merchant := model.Merchant{Name: "Owner Merchant", UserID: &ownerID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	store := model.Store{
		MerchantID:  merchant.ID,
		Name:        "Old Store",
		Description: "keep-me",
		City:        "Old City",
		Images:      `["https://img.example/old.jpg"]`,
	}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	originalHour := model.StoreHour{
		StoreID:   store.ID,
		DayOfWeek: 1,
		OpenTime:  "09:00",
		CloseTime: "18:00",
	}
	if err := db.Create(&originalHour).Error; err != nil {
		t.Fatalf("failed to create original store hour: %v", err)
	}

	newImages := []string{"https://img.example/new-1.jpg", "https://img.example/new-2.jpg"}
	updated, err := svc.Update(context.Background(), ownerID, store.ID, dto.UpdateStoreRequest{
		Name:   strPtr("New Store"),
		City:   strPtr("New City"),
		Images: &newImages,
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	if updated.Name != "New Store" {
		t.Fatalf("unexpected updated name: got %q, want %q", updated.Name, "New Store")
	}
	if updated.City != "New City" {
		t.Fatalf("unexpected updated city: got %q, want %q", updated.City, "New City")
	}
	if updated.Description != "keep-me" {
		t.Fatalf("description should remain unchanged: got %q", updated.Description)
	}

	var parsedImages []string
	if err := json.Unmarshal([]byte(updated.Images), &parsedImages); err != nil {
		t.Fatalf("failed to unmarshal updated images: %v", err)
	}
	if len(parsedImages) != 2 || parsedImages[0] != newImages[0] || parsedImages[1] != newImages[1] {
		t.Fatalf("unexpected updated images: %+v", parsedImages)
	}

	var hours []model.StoreHour
	if err := db.Where("store_id = ?", store.ID).Find(&hours).Error; err != nil {
		t.Fatalf("failed to query store hours: %v", err)
	}
	if len(hours) != 1 {
		t.Fatalf("expected existing store hours unchanged, got %d", len(hours))
	}
	if hours[0].ID != originalHour.ID {
		t.Fatalf("expected original store hour to remain, got id=%d want=%d", hours[0].ID, originalHour.ID)
	}
}

func TestStoreServiceUpdateOwnStoreReplacesHours(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	ownerID := int64(402)
	if err := db.Create(&model.User{ID: ownerID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create owner user: %v", err)
	}
	merchant := model.Merchant{Name: "Owner Merchant", UserID: &ownerID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	store := model.Store{MerchantID: merchant.ID, Name: "Hours Store"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	if err := db.Create(&model.StoreHour{StoreID: store.ID, DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"}).Error; err != nil {
		t.Fatalf("failed to seed old store hour #1: %v", err)
	}
	if err := db.Create(&model.StoreHour{StoreID: store.ID, DayOfWeek: 2, OpenTime: "09:00", CloseTime: "18:00"}).Error; err != nil {
		t.Fatalf("failed to seed old store hour #2: %v", err)
	}

	newHours := []dto.StoreHourRequest{
		{
			DayOfWeek: 5,
			OpenTime:  "10:00",
			CloseTime: "20:00",
			IsClosed:  false,
		},
	}
	_, err := svc.Update(context.Background(), ownerID, store.ID, dto.UpdateStoreRequest{Hours: &newHours})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	var hours []model.StoreHour
	if err := db.Where("store_id = ?", store.ID).Order("day_of_week asc").Find(&hours).Error; err != nil {
		t.Fatalf("failed to query store hours: %v", err)
	}
	if len(hours) != 1 {
		t.Fatalf("expected replaced store hours count=1, got %d", len(hours))
	}
	if hours[0].DayOfWeek != 5 || hours[0].OpenTime != "10:00" || hours[0].CloseTime != "20:00" || hours[0].IsClosed {
		t.Fatalf("unexpected replaced hour row: %+v", hours[0])
	}
}

func TestStoreServiceUpdateForbiddenWhenNotOwner(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	ownerID := int64(403)
	otherID := int64(404)
	if err := db.Create(&model.User{ID: ownerID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create owner user: %v", err)
	}
	if err := db.Create(&model.User{ID: otherID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create other user: %v", err)
	}
	merchant := model.Merchant{Name: "Owner Merchant", UserID: &ownerID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Private Store"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	_, err := svc.Update(context.Background(), otherID, store.ID, dto.UpdateStoreRequest{Name: strPtr("Hacked Name")})
	if !errors.Is(err, ErrStoreForbidden) {
		t.Fatalf("expected ErrStoreForbidden, got %v", err)
	}
}

func TestStoreServiceUpdateStoreNotFound(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(405)
	if err := db.Create(&model.User{ID: userID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, err := svc.Update(context.Background(), userID, 9999, dto.UpdateStoreRequest{Name: strPtr("Missing Store")})
	if !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("expected ErrStoreNotFound, got %v", err)
	}
}

func strPtr(v string) *string {
	return &v
}
