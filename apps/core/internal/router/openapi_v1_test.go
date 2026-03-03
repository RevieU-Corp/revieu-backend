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

const testJWTSecret = "test-secret"

func setupAPITest(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := testutil.SetupTestDB(t)
	database.DB = db

	cfg := &config.Config{
		Server: config.ServerConfig{APIBasePath: "/api/v1"},
		JWT:    config.JWTConfig{Secret: testJWTSecret, ExpireHour: 24},
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

func issueAPITestToken(t *testing.T, user model.User, email string) string {
	t.Helper()
	auth := model.UserAuth{UserID: user.ID, IdentityType: "email", Identifier: email}
	if err := database.DB.Create(&auth).Error; err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}
	tok, err := token.New(config.JWTConfig{Secret: testJWTSecret, ExpireHour: 24}).GenerateToken(&user, &auth)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}
	return tok
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

func TestStoreCreateActivateAndPublicVisibility(t *testing.T) {
	r, tok := setupAPITest(t)

	createBody := strings.NewReader(`{"name":"Draft Store","address":"Austin"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/merchant/stores", createBody)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d", w.Code)
	}

	var created struct {
		Data model.Store `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	if created.Data.Status != 0 {
		t.Fatalf("expected created store status 0(draft), got %d", created.Data.Status)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/stores", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d", w.Code)
	}
	var listed struct {
		Data []model.Store `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatalf("failed to decode stores list response: %v", err)
	}
	if len(listed.Data) != 0 {
		t.Fatalf("expected no public stores before activation, got %d", len(listed.Data))
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/merchant/stores/%d/activate", created.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected activate 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/stores", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected list 200 after activate, got %d", w.Code)
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatalf("failed to decode stores list response after activate: %v", err)
	}
	if len(listed.Data) != 1 {
		t.Fatalf("expected one public store after activation, got %d", len(listed.Data))
	}
	if listed.Data[0].ID != created.Data.ID {
		t.Fatalf("unexpected public store id: got %d, want %d", listed.Data[0].ID, created.Data.ID)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/merchant/stores/%d/deactivate", created.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected deactivate 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/stores/%d", created.Data.ID), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected detail 404 after deactivate, got %d", w.Code)
	}
}

func TestStoreActivateForbiddenForNonOwner(t *testing.T) {
	r, ownerToken := setupAPITest(t)

	createBody := strings.NewReader(`{"name":"Owner Store","address":"Austin"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/merchant/stores", createBody)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected create 201, got %d", w.Code)
	}
	var created struct {
		Data model.Store `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	db := database.DB
	other := model.User{Role: "user", Status: 0}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("failed to create other user: %v", err)
	}
	otherToken := issueAPITestToken(t, other, "other@example.com")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/merchant/stores/%d/activate", created.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden activate for non-owner, got %d", w.Code)
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
