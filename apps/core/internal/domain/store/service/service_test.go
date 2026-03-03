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
	if store.Status != StoreStatusDraft {
		t.Fatalf("unexpected store status: got %d, want %d", store.Status, StoreStatusDraft)
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

func TestStoreServiceActivateOwnStore(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(1001)
	if err := db.Create(&model.User{ID: userID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	merchant := model.Merchant{Name: "Owner Merchant", UserID: &userID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Draft Store", Status: StoreStatusDraft}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	if err := svc.Activate(context.Background(), userID, store.ID); err != nil {
		t.Fatalf("activate returned error: %v", err)
	}

	var refreshed model.Store
	if err := db.First(&refreshed, store.ID).Error; err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}
	if refreshed.Status != StoreStatusPublished {
		t.Fatalf("unexpected status after activation: got %d, want %d", refreshed.Status, StoreStatusPublished)
	}
}

func TestStoreServiceActivateNonOwnerForbidden(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	ownerID := int64(2001)
	otherID := int64(2002)
	for _, id := range []int64{ownerID, otherID} {
		if err := db.Create(&model.User{ID: id, Role: "user", Status: 0}).Error; err != nil {
			t.Fatalf("failed to create user %d: %v", id, err)
		}
	}
	merchant := model.Merchant{Name: "Owner Merchant", UserID: &ownerID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Draft Store", Status: StoreStatusDraft}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	err := svc.Activate(context.Background(), otherID, store.ID)
	if !errors.Is(err, ErrStoreForbidden) {
		t.Fatalf("expected ErrStoreForbidden, got %v", err)
	}
}

func TestStoreServiceDeactivateOwnStore(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(3001)
	if err := db.Create(&model.User{ID: userID, Role: "user", Status: 0}).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	merchant := model.Merchant{Name: "Owner Merchant", UserID: &userID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Published Store", Status: StoreStatusPublished}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	if err := svc.Deactivate(context.Background(), userID, store.ID); err != nil {
		t.Fatalf("deactivate returned error: %v", err)
	}

	var refreshed model.Store
	if err := db.First(&refreshed, store.ID).Error; err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}
	if refreshed.Status != StoreStatusHidden {
		t.Fatalf("unexpected status after deactivation: got %d, want %d", refreshed.Status, StoreStatusHidden)
	}
}

func TestStoreServiceListPublished(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	merchant := model.Merchant{Name: "Demo"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	draft := model.Store{MerchantID: merchant.ID, Name: "Draft", Status: StoreStatusDraft}
	published := model.Store{MerchantID: merchant.ID, Name: "Published", Status: StoreStatusPublished}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatalf("failed to create draft store: %v", err)
	}
	if err := db.Create(&published).Error; err != nil {
		t.Fatalf("failed to create published store: %v", err)
	}

	stores, err := svc.ListPublished(context.Background())
	if err != nil {
		t.Fatalf("list published returned error: %v", err)
	}
	if len(stores) != 1 {
		t.Fatalf("unexpected published stores count: got %d, want 1", len(stores))
	}
	if stores[0].ID != published.ID {
		t.Fatalf("unexpected published store id: got %d, want %d", stores[0].ID, published.ID)
	}
}

func TestStoreServiceListMine(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	userID := int64(4001)
	otherID := int64(4002)
	for _, id := range []int64{userID, otherID} {
		if err := db.Create(&model.User{ID: id, Role: "user", Status: 0}).Error; err != nil {
			t.Fatalf("failed to create user %d: %v", id, err)
		}
	}

	myMerchant := model.Merchant{Name: "Mine", UserID: &userID}
	otherMerchant := model.Merchant{Name: "Other", UserID: &otherID}
	if err := db.Create(&myMerchant).Error; err != nil {
		t.Fatalf("failed to create my merchant: %v", err)
	}
	if err := db.Create(&otherMerchant).Error; err != nil {
		t.Fatalf("failed to create other merchant: %v", err)
	}

	myStore := model.Store{MerchantID: myMerchant.ID, Name: "My Store", Status: StoreStatusDraft}
	otherStore := model.Store{MerchantID: otherMerchant.ID, Name: "Other Store", Status: StoreStatusPublished}
	if err := db.Create(&myStore).Error; err != nil {
		t.Fatalf("failed to create my store: %v", err)
	}
	if err := db.Create(&otherStore).Error; err != nil {
		t.Fatalf("failed to create other store: %v", err)
	}

	stores, err := svc.ListMine(context.Background(), userID)
	if err != nil {
		t.Fatalf("list mine returned error: %v", err)
	}
	if len(stores) != 1 {
		t.Fatalf("unexpected mine stores count: got %d, want 1", len(stores))
	}
	if stores[0].ID != myStore.ID {
		t.Fatalf("unexpected mine store id: got %d, want %d", stores[0].ID, myStore.ID)
	}
}
