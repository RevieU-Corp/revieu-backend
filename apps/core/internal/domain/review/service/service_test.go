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

func TestCreateReviewSyncsStoreAndMerchantAggregates(t *testing.T) {
	svc, db, userID := setupReviewServiceTest(t)

	merchant := model.Merchant{Name: "Cafe"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Downtown"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	create := func(rating float64, text string) {
		_, err := svc.Create(context.Background(), userID, dto.Review{
			MerchantID: fmt.Sprintf("%d", merchant.ID),
			StoreID:    fmt.Sprintf("%d", store.ID),
			Rating:     rating,
			Text:       text,
		})
		if err != nil {
			t.Fatalf("create review failed: %v", err)
		}
	}

	create(4.0, "good")
	create(2.0, "ok")

	var refreshedStore model.Store
	if err := db.First(&refreshedStore, store.ID).Error; err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}
	if refreshedStore.ReviewCount != 2 {
		t.Fatalf("unexpected store review_count: got %d want 2", refreshedStore.ReviewCount)
	}
	if refreshedStore.AvgRating != float32(3.0) {
		t.Fatalf("unexpected store avg_rating: got %.2f want 3.00", refreshedStore.AvgRating)
	}

	var refreshedMerchant model.Merchant
	if err := db.First(&refreshedMerchant, merchant.ID).Error; err != nil {
		t.Fatalf("failed to reload merchant: %v", err)
	}
	if refreshedMerchant.ReviewCount != 2 {
		t.Fatalf("unexpected merchant review_count: got %d want 2", refreshedMerchant.ReviewCount)
	}
	if refreshedMerchant.TotalReviews != 2 {
		t.Fatalf("unexpected merchant total_reviews: got %d want 2", refreshedMerchant.TotalReviews)
	}
	if refreshedMerchant.AvgRating != float32(3.0) {
		t.Fatalf("unexpected merchant avg_rating: got %.2f want 3.00", refreshedMerchant.AvgRating)
	}
}

func TestLikeIsIdempotentPerUser(t *testing.T) {
	svc, db, userID := setupReviewServiceTest(t)

	merchant := model.Merchant{Name: "Cafe"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	review := model.Review{
		UserID:     userID,
		MerchantID: merchant.ID,
		VenueID:    merchant.ID,
		Rating:     4.0,
		Content:    "nice",
		VisitDate:  time.Now().UTC(),
	}
	if err := db.Create(&review).Error; err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	if err := svc.Like(context.Background(), userID, review.ID); err != nil {
		t.Fatalf("first like failed: %v", err)
	}
	if err := svc.Like(context.Background(), userID, review.ID); err != nil {
		t.Fatalf("second like failed: %v", err)
	}

	var likes int64
	if err := db.Model(&model.Like{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "review", review.ID).
		Count(&likes).Error; err != nil {
		t.Fatalf("failed to count likes: %v", err)
	}
	if likes != 1 {
		t.Fatalf("unexpected like rows: got %d want 1", likes)
	}

	var refreshed model.Review
	if err := db.First(&refreshed, review.ID).Error; err != nil {
		t.Fatalf("failed to reload review: %v", err)
	}
	if refreshed.LikeCount != 1 {
		t.Fatalf("unexpected like_count: got %d want 1", refreshed.LikeCount)
	}
}

func TestCommentIncrementsCommentCount(t *testing.T) {
	svc, db, userID := setupReviewServiceTest(t)

	merchant := model.Merchant{Name: "Cafe"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	review := model.Review{
		UserID:     userID,
		MerchantID: merchant.ID,
		VenueID:    merchant.ID,
		Rating:     4.0,
		Content:    "nice",
		VisitDate:  time.Now().UTC(),
	}
	if err := db.Create(&review).Error; err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	if err := svc.Comment(context.Background(), userID, review.ID, "first"); err != nil {
		t.Fatalf("first comment failed: %v", err)
	}
	if err := svc.Comment(context.Background(), userID, review.ID, "second"); err != nil {
		t.Fatalf("second comment failed: %v", err)
	}

	var refreshed model.Review
	if err := db.First(&refreshed, review.ID).Error; err != nil {
		t.Fatalf("failed to reload review: %v", err)
	}
	if refreshed.CommentCount != 2 {
		t.Fatalf("unexpected comment_count: got %d want 2", refreshed.CommentCount)
	}
}
