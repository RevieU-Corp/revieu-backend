package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

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

	if err := db.AutoMigrate(
		&model.User{},
		&model.UserProfile{},
		&model.Merchant{},
		&model.Store{},
		&model.StoreHour{},
		&model.Category{},
		&model.StoreCategory{},
		&model.Review{},
	); err != nil {
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

func TestStoreServiceDetailPublishedPreloadsHoursAndCategories(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	merchant := model.Merchant{Name: "Detail Merchant"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	store := model.Store{MerchantID: merchant.ID, Name: "Published Detail", Status: StoreStatusPublished}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	hour := model.StoreHour{StoreID: store.ID, DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"}
	if err := db.Create(&hour).Error; err != nil {
		t.Fatalf("failed to create store hour: %v", err)
	}

	category := model.Category{Name: "Cafe"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	storeCategory := model.StoreCategory{StoreID: store.ID, CategoryID: category.ID}
	if err := db.Create(&storeCategory).Error; err != nil {
		t.Fatalf("failed to create store category relation: %v", err)
	}

	detail, err := svc.DetailPublished(context.Background(), store.ID)
	if err != nil {
		t.Fatalf("detail published returned error: %v", err)
	}

	if len(detail.Hours) != 1 {
		t.Fatalf("expected 1 hour preloaded, got %d", len(detail.Hours))
	}
	if detail.Hours[0].DayOfWeek != 1 {
		t.Fatalf("unexpected preloaded hour day: %d", detail.Hours[0].DayOfWeek)
	}

	if len(detail.Categories) != 1 {
		t.Fatalf("expected 1 category preloaded, got %d", len(detail.Categories))
	}
	if detail.Categories[0].Name != "Cafe" {
		t.Fatalf("unexpected preloaded category: %q", detail.Categories[0].Name)
	}
}

func TestStoreServiceListPublishedFilteredByCategoryRatingLocationAndCursor(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	merchant := model.Merchant{Name: "Filter Merchant"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	cafe := model.Category{Name: "Cafe"}
	bakery := model.Category{Name: "Bakery"}
	if err := db.Create(&cafe).Error; err != nil {
		t.Fatalf("failed to create cafe category: %v", err)
	}
	if err := db.Create(&bakery).Error; err != nil {
		t.Fatalf("failed to create bakery category: %v", err)
	}

	createStore := func(name string, status int16, rating float32, lat, lng float64, categoryID int64) model.Store {
		store := model.Store{
			MerchantID: merchant.ID,
			Name:       name,
			Status:     status,
			AvgRating:  rating,
			Latitude:   lat,
			Longitude:  lng,
		}
		if err := db.Create(&store).Error; err != nil {
			t.Fatalf("failed to create store %s: %v", name, err)
		}
		if categoryID != 0 {
			if err := db.Create(&model.StoreCategory{StoreID: store.ID, CategoryID: categoryID}).Error; err != nil {
				t.Fatalf("failed to create store-category relation for %s: %v", name, err)
			}
		}
		return store
	}

	matchOld := createStore("Match Old", StoreStatusPublished, 4.1, 37.7750, -122.4190, cafe.ID)
	matchNew := createStore("Match New", StoreStatusPublished, 4.8, 37.7752, -122.4188, cafe.ID)
	_ = createStore("Low Rating", StoreStatusPublished, 3.0, 37.7751, -122.4189, cafe.ID)
	_ = createStore("Far Away", StoreStatusPublished, 4.9, 39.0000, -120.0000, cafe.ID)
	_ = createStore("Other Category", StoreStatusPublished, 4.9, 37.7753, -122.4187, bakery.ID)
	_ = createStore("Draft Hidden", StoreStatusDraft, 4.9, 37.7751, -122.4189, cafe.ID)

	lat := 37.7750
	lng := -122.4190
	rating := float32(4.0)
	limit := 1
	radiusKm := 5.0
	firstPage, cursor, err := svc.ListPublishedFiltered(context.Background(), dto.StoreListQuery{
		Category: categoryNamePtr("Cafe"),
		Lat:      &lat,
		Lng:      &lng,
		Rating:   &rating,
		RadiusKM: &radiusKm,
		Limit:    &limit,
	})
	if err != nil {
		t.Fatalf("list filtered first page returned error: %v", err)
	}
	if len(firstPage) != 1 {
		t.Fatalf("expected first page size 1, got %d", len(firstPage))
	}
	if firstPage[0].ID != matchNew.ID {
		t.Fatalf("unexpected first-page store id: got %d, want %d", firstPage[0].ID, matchNew.ID)
	}
	if cursor == nil {
		t.Fatalf("expected cursor for second page")
	}

	secondPage, nextCursor, err := svc.ListPublishedFiltered(context.Background(), dto.StoreListQuery{
		Category: categoryNamePtr("Cafe"),
		Lat:      &lat,
		Lng:      &lng,
		Rating:   &rating,
		RadiusKM: &radiusKm,
		Limit:    &limit,
		Cursor:   cursor,
	})
	if err != nil {
		t.Fatalf("list filtered second page returned error: %v", err)
	}
	if len(secondPage) != 1 {
		t.Fatalf("expected second page size 1, got %d", len(secondPage))
	}
	if secondPage[0].ID != matchOld.ID {
		t.Fatalf("unexpected second-page store id: got %d, want %d", secondPage[0].ID, matchOld.ID)
	}
	if nextCursor != nil {
		t.Fatalf("expected no further cursor after second page, got %d", *nextCursor)
	}
}

func TestStoreServiceReviewsPublishedPaginatedPreloadsUser(t *testing.T) {
	db := setupStoreTestDB(t)
	svc := NewStoreService(db)

	merchant := model.Merchant{Name: "Review Merchant"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Review Store", Status: StoreStatusPublished}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	makeUserWithProfile := func(nickname string) model.User {
		user := model.User{Role: "user", Status: 0}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("failed to create user %s: %v", nickname, err)
		}
		profile := model.UserProfile{UserID: user.ID, Nickname: nickname}
		if err := db.Create(&profile).Error; err != nil {
			t.Fatalf("failed to create profile %s: %v", nickname, err)
		}
		return user
	}

	u1 := makeUserWithProfile("u1")
	u2 := makeUserWithProfile("u2")
	u3 := makeUserWithProfile("u3")

	addReview := func(userID int64, content string) model.Review {
		storeID := store.ID
		review := model.Review{
			UserID:     userID,
			MerchantID: merchant.ID,
			VenueID:    merchant.ID,
			StoreID:    &storeID,
			Rating:     4.5,
			Content:    content,
			VisitDate:  time.Now().UTC(),
		}
		if err := db.Create(&review).Error; err != nil {
			t.Fatalf("failed to create review %s: %v", content, err)
		}
		return review
	}

	oldest := addReview(u1.ID, "first")
	_ = addReview(u2.ID, "second")
	_ = addReview(u3.ID, "third")

	limit := 2
	firstPage, cursor, err := svc.ReviewsPublishedPaginated(context.Background(), store.ID, dto.StoreReviewListQuery{
		Limit: &limit,
	})
	if err != nil {
		t.Fatalf("reviews first page returned error: %v", err)
	}
	if len(firstPage) != 2 {
		t.Fatalf("expected first page size 2, got %d", len(firstPage))
	}
	if cursor == nil {
		t.Fatalf("expected cursor for second review page")
	}
	if firstPage[0].User == nil || firstPage[0].User.Profile == nil {
		t.Fatalf("expected user/profile to be preloaded on first review page")
	}

	secondPage, nextCursor, err := svc.ReviewsPublishedPaginated(context.Background(), store.ID, dto.StoreReviewListQuery{
		Limit:  &limit,
		Cursor: cursor,
	})
	if err != nil {
		t.Fatalf("reviews second page returned error: %v", err)
	}
	if len(secondPage) != 1 {
		t.Fatalf("expected second page size 1, got %d", len(secondPage))
	}
	if secondPage[0].ID != oldest.ID {
		t.Fatalf("unexpected second-page review id: got %d, want %d", secondPage[0].ID, oldest.ID)
	}
	if nextCursor != nil {
		t.Fatalf("expected no further cursor after second page, got %d", *nextCursor)
	}
	if secondPage[0].User == nil || secondPage[0].User.Profile == nil {
		t.Fatalf("expected user/profile to be preloaded on second review page")
	}
}

func categoryNamePtr(v string) *string {
	return &v
}

func strPtr(v string) *string {
	return &v
}
