package router

import (
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
