package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
	"github.com/RevieU-Corp/revieu-backend/apps/core/internal/token"
	"github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
	"github.com/gin-gonic/gin"
)

func setupAPITest(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	database.DB = db

	cfg := &config.Config{
		Server: config.ServerConfig{APIBasePath: "/api/v1"},
		JWT:    config.JWTConfig{Secret: "test-secret", ExpireHour: 24},
	}

	r := gin.New()
	Setup(r, cfg)

	user := model.User{Role: "user", Status: 0}
	_ = db.Create(&user).Error
	auth := model.UserAuth{UserID: user.ID, IdentityType: "email", Identifier: "user@example.com"}
	_ = db.Create(&auth).Error
	tok, _ := token.New(cfg.JWT).GenerateToken(&user, &auth)

	return r, tok
}

func TestFeedHome(t *testing.T) {
	r, _ := setupAPITest(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/feed/home", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMerchantsList(t *testing.T) {
	r, _ := setupAPITest(t)

	db := database.DB
	_ = db.Create(&model.Merchant{Name: "Cafe", Category: "food"}).Error

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/merchants", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMerchantDetail(t *testing.T) {
	r, _ := setupAPITest(t)

	db := database.DB
	m := model.Merchant{Name: "Shop"}
	_ = db.Create(&m).Error

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/merchants/%d", m.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMerchantReviews(t *testing.T) {
	r, _ := setupAPITest(t)

	db := database.DB
	user := model.User{Role: "user", Status: 0}
	_ = db.Create(&user).Error
	m := model.Merchant{Name: "Shop"}
	_ = db.Create(&m).Error
	_ = db.Create(&model.Review{UserID: user.ID, MerchantID: m.ID, Rating: 4, Content: "nice"}).Error

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/merchants/%d/reviews", m.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReviewsList(t *testing.T) {
	r, tok := setupAPITest(t)

	db := database.DB
	m := model.Merchant{Name: "Cafe"}
	_ = db.Create(&m).Error
	u := model.User{Role: "user", Status: 0}
	_ = db.Create(&u).Error
	_ = db.Create(&model.Review{UserID: u.ID, MerchantID: m.ID, Rating: 5, Content: "great"}).Error

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/reviews", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReviewsCreateAndDetail(t *testing.T) {
	r, tok := setupAPITest(t)

	db := database.DB
	m := model.Merchant{Name: "Cafe"}
	_ = db.Create(&m).Error

	body := strings.NewReader(fmt.Sprintf(`{"merchantId":"%d","rating":4.5,"text":"nice"}`, m.ID))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/reviews", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestReviewsCreateReturnsNotFoundWhenMerchantMissing(t *testing.T) {
	r, tok := setupAPITest(t)

	body := strings.NewReader(`{"merchantId":"99999","rating":4.5,"text":"nice"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/reviews", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestReviewsCreateReturnsNotFoundWhenStoreMissing(t *testing.T) {
	r, tok := setupAPITest(t)
	db := database.DB

	merchant := model.Merchant{Name: "Cafe"}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	body := strings.NewReader(fmt.Sprintf(`{"merchantId":"%d","storeId":"99999","rating":4.5,"text":"nice"}`, merchant.ID))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/reviews", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestReviewsCreateReturnsUnprocessableWhenStoreMerchantMismatch(t *testing.T) {
	r, tok := setupAPITest(t)
	db := database.DB

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

	body := strings.NewReader(fmt.Sprintf(`{"merchantId":"%d","storeId":"%d","rating":4.5,"text":"nice"}`, merchantA.ID, store.ID))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/reviews", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestReviewDetailIncludesMerchantFields(t *testing.T) {
	r, _ := setupAPITest(t)
	db := database.DB

	u := model.User{Role: "user", Status: 0}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	m := model.Merchant{Name: "Cafe", BusinessName: "Cafe Business", Address: "San Francisco"}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	review := model.Review{UserID: u.ID, MerchantID: m.ID, VenueID: m.ID, Rating: 4.5, Content: "ok"}
	if err := db.Create(&review).Error; err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/reviews/%d", review.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["businessName"] != "Cafe Business" {
		t.Fatalf("expected businessName Cafe Business, got %v", resp["businessName"])
	}
	if resp["location"] != "San Francisco" {
		t.Fatalf("expected location San Francisco, got %v", resp["location"])
	}
}

func TestReviewLikeAndComment(t *testing.T) {
	r, tok := setupAPITest(t)

	db := database.DB
	m := model.Merchant{Name: "Cafe"}
	_ = db.Create(&m).Error
	u := model.User{Role: "user", Status: 0}
	_ = db.Create(&u).Error
	review := model.Review{UserID: u.ID, MerchantID: m.ID, Rating: 4, Content: "ok"}
	_ = db.Create(&review).Error

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/reviews/%d/like", review.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	body := strings.NewReader(`{"text":"nice"}`)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/reviews/%d/comments", review.ID), body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCouponValidateAndRedeem(t *testing.T) {
	r, tok := setupAPITest(t)

	db := database.DB
	m := model.Merchant{Name: "Shop"}
	_ = db.Create(&m).Error
	coupon := model.Coupon{MerchantID: m.ID, Title: "Free", Type: "free"}
	_ = db.Create(&coupon).Error

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/coupons/%d/validate", coupon.ID), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/coupons/%d/redeem", coupon.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestVoucherCreateAndList(t *testing.T) {
	r, tok := setupAPITest(t)

	body := strings.NewReader(`{"couponId":"1","userId":"1","code":"ABC"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vouchers", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/vouchers", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPaymentsCreateAndDetail(t *testing.T) {
	r, tok := setupAPITest(t)

	body := strings.NewReader(`{"amount":10,"currency":"USD","status":"pending"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/payments", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestMediaUploadAndAnalysis(t *testing.T) {
	r, tok := setupAPITest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/media/uploads", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAISuggestions(t *testing.T) {
	r, tok := setupAPITest(t)

	body := strings.NewReader(`{"overallRating":4.5,"merchantName":"Cafe"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/ai/reviews/suggestions", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthForgotPassword(t *testing.T) {
	r, _ := setupAPITest(t)

	body := strings.NewReader(`{"email":"user@example.com"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestStoreUpdateOwnStore(t *testing.T) {
	r, tok := setupAPITest(t)
	db := database.DB

	var ownerAuth model.UserAuth
	if err := db.Where("identifier = ?", "user@example.com").First(&ownerAuth).Error; err != nil {
		t.Fatalf("failed to load owner auth: %v", err)
	}
	ownerID := ownerAuth.UserID

	merchant := model.Merchant{Name: "Owner Merchant", UserID: &ownerID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Old Store", City: "Old City"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	body := strings.NewReader(`{"name":"Updated Store","city":"San Francisco"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/merchant/stores/%d", store.ID), body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var refreshed model.Store
	if err := db.First(&refreshed, store.ID).Error; err != nil {
		t.Fatalf("failed to load updated store: %v", err)
	}
	if refreshed.Name != "Updated Store" {
		t.Fatalf("unexpected updated name: got %q", refreshed.Name)
	}
	if refreshed.City != "San Francisco" {
		t.Fatalf("unexpected updated city: got %q", refreshed.City)
	}
}

func TestStoreUpdateForbiddenForNonOwner(t *testing.T) {
	r, _ := setupAPITest(t)
	db := database.DB

	var ownerAuth model.UserAuth
	if err := db.Where("identifier = ?", "user@example.com").First(&ownerAuth).Error; err != nil {
		t.Fatalf("failed to load owner auth: %v", err)
	}
	ownerID := ownerAuth.UserID

	merchant := model.Merchant{Name: "Owner Merchant", UserID: &ownerID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Protected Store"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	other := model.User{Role: "user", Status: 0}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("failed to create other user: %v", err)
	}
	otherAuth := model.UserAuth{UserID: other.ID, IdentityType: "email", Identifier: "other@example.com"}
	if err := db.Create(&otherAuth).Error; err != nil {
		t.Fatalf("failed to create other auth: %v", err)
	}
	otherToken, err := token.New(config.JWTConfig{Secret: "test-secret", ExpireHour: 24}).GenerateToken(&other, &otherAuth)
	if err != nil {
		t.Fatalf("failed to generate token for other user: %v", err)
	}

	body := strings.NewReader(`{"name":"Hijacked Name"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/merchant/stores/%d", store.ID), body)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	var refreshed model.Store
	if err := db.First(&refreshed, store.ID).Error; err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if refreshed.Name != "Protected Store" {
		t.Fatalf("store should not be updated by non-owner, got %q", refreshed.Name)
	}

}

func TestStoreUpdateUnauthorizedWithoutJWT(t *testing.T) {
	r, _ := setupAPITest(t)
	db := database.DB

	var ownerAuth model.UserAuth
	if err := db.Where("identifier = ?", "user@example.com").First(&ownerAuth).Error; err != nil {
		t.Fatalf("failed to load owner auth: %v", err)
	}
	ownerID := ownerAuth.UserID

	merchant := model.Merchant{Name: "Owner Merchant", UserID: &ownerID}
	if err := db.Create(&merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}
	store := model.Store{MerchantID: merchant.ID, Name: "Protected Store"}
	if err := db.Create(&store).Error; err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	body := strings.NewReader(`{"name":"No Auth Update"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/merchant/stores/%d", store.ID), body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
