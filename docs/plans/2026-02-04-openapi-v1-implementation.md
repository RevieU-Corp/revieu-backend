# OpenAPI V1 Missing Endpoints Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Issue:** #96

**Goal:** Implement all missing OpenAPI v1 endpoints (feed, merchants, reviews, coupons, vouchers, payments, media, AI suggestions, and auth forgot-password) without changing existing auth/user/profile behavior.

**Architecture:** New domain packages per tag with Gin routes, handlers, and services using GORM. Reuse existing models where possible and add new models for coupons/vouchers/payments/media/review comments. Register new routes in `router.Setup` only.

**Tech Stack:** Go 1.21, Gin, GORM, Postgres, sqlite (tests).

**Note on imports:** As tests and DTOs expand, update import blocks with `fmt`, `strings`, and `time` where used.

---

### Task 1: Add API test harness + failing feed endpoint test

**Files:**
- Create: `apps/core/internal/router/openapi_v1_test.go`

**Step 1: Write the failing test**
```go
package router

import (
    "net/http"
    "net/http/httptest"
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
```

**Step 2: Run test to verify it fails**
Run: `go test ./internal/router -run TestFeedHome`
Expected: FAIL with 404 status

**Step 3: Write minimal implementation**
- Create `apps/core/internal/domain/feed/routes.go`
```go
package feed

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed/service"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, _ *config.Config) {
    svc := service.NewFeedService(nil)
    h := handler.NewFeedHandler(svc)

    feed := r.Group("/feed")
    {
        feed.GET("/home", h.Home)
    }
}
```
- Create `apps/core/internal/domain/feed/handler/feed.go`
```go
package handler

import (
    "net/http"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed/service"
    "github.com/gin-gonic/gin"
)

type FeedHandler struct {
    svc *service.FeedService
}

func NewFeedHandler(svc *service.FeedService) *FeedHandler {
    if svc == nil {
        svc = service.NewFeedService(nil)
    }
    return &FeedHandler{svc: svc}
}

func (h *FeedHandler) Home(c *gin.Context) {
    items, err := h.svc.Home(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": items})
}
```
- Create `apps/core/internal/domain/feed/service/service.go`
```go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type FeedService struct {
    db *gorm.DB
}

func NewFeedService(db *gorm.DB) *FeedService {
    if db == nil {
        db = database.DB
    }
    return &FeedService{db: db}
}

func (s *FeedService) Home(_ context.Context) ([]dto.FeedItem, error) {
    return []dto.FeedItem{}, nil
}
```
- Create `apps/core/internal/domain/feed/dto/feed.go`
```go
package dto

type FeedItem struct {
    ID    string `json:"id"`
    Type  string `json:"type"`
    Title string `json:"title"`
    Image string `json:"image"`
}
```
- Update `apps/core/internal/router/router.go`
```go
import "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/feed"
// ...
feed.RegisterRoutes(api, cfg)
```

**Step 4: Run test to verify it passes**
Run: `go test ./internal/router -run TestFeedHome`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/feed apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add feed home endpoint"
```

---

### Task 2: Merchants list endpoint (test → implement)

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Create: `apps/core/internal/domain/merchant/routes.go`
- Create: `apps/core/internal/domain/merchant/handler/merchant.go`
- Create: `apps/core/internal/domain/merchant/service/service.go`
- Create: `apps/core/internal/domain/merchant/dto/merchant.go`
- Modify: `apps/core/internal/router/router.go`

**Step 1: Write the failing test**
```go
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
```

**Step 2: Run test to verify it fails**
Run: `go test ./internal/router -run TestMerchantsList`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- `apps/core/internal/domain/merchant/routes.go`
```go
package merchant

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/service"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, _ *config.Config) {
    svc := service.NewMerchantService(nil)
    h := handler.NewMerchantHandler(svc)

    merchants := r.Group("/merchants")
    {
        merchants.GET("", h.List)
        merchants.GET("/:id", h.Detail)
        merchants.GET("/:id/reviews", h.Reviews)
    }
}
```
- `apps/core/internal/domain/merchant/service/service.go`
```go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type MerchantService struct {
    db *gorm.DB
}

func NewMerchantService(db *gorm.DB) *MerchantService {
    if db == nil {
        db = database.DB
    }
    return &MerchantService{db: db}
}

func (s *MerchantService) List(ctx context.Context, category string) ([]model.Merchant, error) {
    q := s.db.WithContext(ctx).Model(&model.Merchant{})
    if category != "" {
        q = q.Where("category = ?", category)
    }
    var merchants []model.Merchant
    if err := q.Order("id desc").Find(&merchants).Error; err != nil {
        return nil, err
    }
    return merchants, nil
}
```
- `apps/core/internal/domain/merchant/handler/merchant.go`
```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/merchant/service"
    "github.com/gin-gonic/gin"
)

type MerchantHandler struct {
    svc *service.MerchantService
}

func NewMerchantHandler(svc *service.MerchantService) *MerchantHandler {
    if svc == nil {
        svc = service.NewMerchantService(nil)
    }
    return &MerchantHandler{svc: svc}
}

func (h *MerchantHandler) List(c *gin.Context) {
    category := c.Query("category")
    merchants, err := h.svc.List(c.Request.Context(), category)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    items := make([]dto.Merchant, 0, len(merchants))
    for _, m := range merchants {
        items = append(items, dto.FromModel(m))
    }
    c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *MerchantHandler) Detail(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    merchant, err := h.svc.Detail(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, dto.FromModel(merchant))
}

func (h *MerchantHandler) Reviews(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    reviews, err := h.svc.Reviews(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": reviews})
}
```
- `apps/core/internal/domain/merchant/dto/merchant.go`
```go
package dto

import "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"

type Merchant struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Category    string   `json:"category"`
    Rating      float32  `json:"rating"`
    ReviewCount int      `json:"reviewCount"`
    Distance    string   `json:"distance"`
    Tags        []string `json:"tags"`
    CoverImage  string   `json:"coverImage"`
}

func FromModel(m model.Merchant) Merchant {
    return Merchant{
        ID:          fmt.Sprintf("%d", m.ID),
        Name:        m.Name,
        Category:    m.Category,
        Rating:      m.AvgRating,
        ReviewCount: m.ReviewCount,
        Distance:    "",
        Tags:        []string{},
        CoverImage:  m.CoverImage,
    }
}
```
- Update `apps/core/internal/router/router.go` to register `merchant.RegisterRoutes`.

**Step 4: Run test to verify it passes**
Run: `go test ./internal/router -run TestMerchantsList`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/merchant apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add merchants list endpoint"
```

---

### Task 3: Merchant detail + reviews endpoints

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Modify: `apps/core/internal/domain/merchant/service/service.go`

**Step 1: Write the failing tests**
```go
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
```

**Step 2: Run tests to verify they fail**
Run: `go test ./internal/router -run TestMerchantDetail`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- Extend `apps/core/internal/domain/merchant/service/service.go`
```go
func (s *MerchantService) Detail(ctx context.Context, id int64) (*model.Merchant, error) {
    var m model.Merchant
    if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
        return nil, err
    }
    return &m, nil
}

func (s *MerchantService) Reviews(ctx context.Context, merchantID int64) ([]model.Review, error) {
    var reviews []model.Review
    if err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Order("id desc").Find(&reviews).Error; err != nil {
        return nil, err
    }
    return reviews, nil
}
```

**Step 4: Run tests to verify they pass**
Run: `go test ./internal/router -run TestMerchantDetail`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/merchant/service/service.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add merchant detail and reviews"
```

---

### Task 4: Reviews list/create/detail endpoints

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Create: `apps/core/internal/domain/review/routes.go`
- Create: `apps/core/internal/domain/review/handler/review.go`
- Create: `apps/core/internal/domain/review/service/service.go`
- Create: `apps/core/internal/domain/review/dto/review.go`
- Create: `apps/core/internal/model/review_comment.go`
- Modify: `apps/core/internal/router/router.go`

**Step 1: Write the failing tests**
```go
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
```

**Step 2: Run tests to verify they fail**
Run: `go test ./internal/router -run TestReviewsList`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- `apps/core/internal/model/review_comment.go`
```go
package model

import "time"

type ReviewComment struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    ReviewID  int64     `gorm:"not null;index" json:"review_id"`
    UserID    int64     `gorm:"not null;index" json:"user_id"`
    Content   string    `gorm:"type:text;not null" json:"content"`
    CreatedAt time.Time `json:"created_at"`
}

func (r *ReviewComment) TableName() string {
    return "review_comments"
}
```
- `apps/core/internal/domain/review/routes.go`
```go
package review

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewReviewService(nil)
    h := handler.NewReviewHandler(svc)

    reviews := r.Group("/reviews")
    {
        reviews.GET("", middleware.JWTAuth(cfg.JWT), h.ListMyReviews)
        reviews.POST("", middleware.JWTAuth(cfg.JWT), h.Create)
        reviews.GET("/:id", h.Detail)
        reviews.POST("/:id/like", middleware.JWTAuth(cfg.JWT), h.Like)
        reviews.POST("/:id/comments", middleware.JWTAuth(cfg.JWT), h.Comment)
    }
}
```
- `apps/core/internal/domain/review/handler/review.go`
```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/service"
    "github.com/gin-gonic/gin"
)

type ReviewHandler struct {
    svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
    if svc == nil {
        svc = service.NewReviewService(nil)
    }
    return &ReviewHandler{svc: svc}
}

func (h *ReviewHandler) ListMyReviews(c *gin.Context) {
    userID := c.GetInt64("user_id")
    reviews, err := h.svc.ListByUser(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": dto.FromModels(reviews)})
}

func (h *ReviewHandler) Create(c *gin.Context) {
    userID := c.GetInt64("user_id")
    var req dto.Review
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    created, err := h.svc.Create(c.Request.Context(), userID, req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, dto.FromModel(created))
}

func (h *ReviewHandler) Detail(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    review, err := h.svc.Detail(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, dto.FromModel(*review))
}

func (h *ReviewHandler) Like(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    userID := c.GetInt64("user_id")
    if err := h.svc.Like(c.Request.Context(), userID, id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{})
}

func (h *ReviewHandler) Comment(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    userID := c.GetInt64("user_id")
    var req struct{ Text string `json:"text"` }
    if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid text"})
        return
    }
    if err := h.svc.Comment(c.Request.Context(), userID, id, req.Text); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{})
}
```
- `apps/core/internal/domain/review/service/service.go`
```go
package service

import (
    "context"
    "encoding/json"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/review/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type ReviewService struct {
    db *gorm.DB
}

func NewReviewService(db *gorm.DB) *ReviewService {
    if db == nil {
        db = database.DB
    }
    return &ReviewService{db: db}
}

func (s *ReviewService) ListByUser(ctx context.Context, userID int64) ([]model.Review, error) {
    var reviews []model.Review
    if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc").Find(&reviews).Error; err != nil {
        return nil, err
    }
    return reviews, nil
}

func (s *ReviewService) Detail(ctx context.Context, id int64) (*model.Review, error) {
    var review model.Review
    if err := s.db.WithContext(ctx).First(&review, id).Error; err != nil {
        return nil, err
    }
    return &review, nil
}

func (s *ReviewService) Create(ctx context.Context, userID int64, req dto.Review) (model.Review, error) {
    merchantID, err := req.MerchantID()
    if err != nil {
        return model.Review{}, err
    }
    imagesJSON, _ := json.Marshal(req.Images)
    review := model.Review{
        UserID:     userID,
        MerchantID: merchantID,
        Rating:     float32(req.Rating),
        Content:    req.Text,
        Images:     string(imagesJSON),
    }
    if err := s.db.WithContext(ctx).Create(&review).Error; err != nil {
        return model.Review{}, err
    }
    return review, nil
}

func (s *ReviewService) Like(ctx context.Context, userID, reviewID int64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var review model.Review
        if err := tx.First(&review, reviewID).Error; err != nil {
            return err
        }
        like := model.Like{UserID: userID, TargetType: "review", TargetID: reviewID}
        if err := tx.FirstOrCreate(&like, like).Error; err != nil {
            return err
        }
        return tx.Model(&review).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
    })
}

func (s *ReviewService) Comment(ctx context.Context, userID, reviewID int64, text string) error {
    comment := model.ReviewComment{ReviewID: reviewID, UserID: userID, Content: text}
    return s.db.WithContext(ctx).Create(&comment).Error
}
```
- `apps/core/internal/domain/review/dto/review.go`
```go
package dto

import (
    "errors"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

type Review struct {
    ID         string   `json:"id"`
    MerchantID string   `json:"merchantId"`
    UserID     string   `json:"userId"`
    Rating     float64  `json:"rating"`
    Text       string   `json:"text"`
    Images     []string `json:"images"`
    Tags       []string `json:"tags"`
    CreatedAt  string   `json:"createdAt"`
}

func (r Review) MerchantID() (int64, error) {
    if r.MerchantID == "" {
        return 0, errors.New("merchantId required")
    }
    return strconv.ParseInt(r.MerchantID, 10, 64)
}

func FromModel(m model.Review) Review {
    return Review{
        ID:         strconv.FormatInt(m.ID, 10),
        MerchantID: strconv.FormatInt(m.MerchantID, 10),
        UserID:     strconv.FormatInt(m.UserID, 10),
        Rating:     float64(m.Rating),
        Text:       m.Content,
        Images:     []string{},
        Tags:       []string{},
        CreatedAt:  m.CreatedAt.Format(time.RFC3339),
    }
}

func FromModels(items []model.Review) []Review {
    out := make([]Review, 0, len(items))
    for _, r := range items {
        out = append(out, FromModel(r))
    }
    return out
}
```
- Update `apps/core/internal/router/router.go` to register `review.RegisterRoutes`.

**Step 4: Run tests to verify they pass**
Run: `go test ./internal/router -run TestReviewsList`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/review apps/core/internal/model/review_comment.go apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add reviews endpoints"
```

---

### Task 5: Review like/comment endpoints tests

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`

**Step 1: Write the failing tests**
```go
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
```

**Step 2: Run test to verify it fails**
Run: `go test ./internal/router -run TestReviewLikeAndComment`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
Implementation already included in Task 4 (like/comment handlers). If failing, adjust handler/service to return correct status codes.

**Step 4: Run test to verify it passes**
Run: `go test ./internal/router -run TestReviewLikeAndComment`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/router/openapi_v1_test.go
git commit -m "test: cover review like and comment"
```

---

### Task 6: Coupons endpoints

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Create: `apps/core/internal/model/coupon.go`
- Create: `apps/core/internal/domain/coupon/routes.go`
- Create: `apps/core/internal/domain/coupon/handler/coupon.go`
- Create: `apps/core/internal/domain/coupon/service/service.go`
- Modify: `apps/core/internal/router/router.go`

**Step 1: Write the failing tests**
```go
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
```

**Step 2: Run tests to verify they fail**
Run: `go test ./internal/router -run TestCouponValidateAndRedeem`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- `apps/core/internal/model/coupon.go`
```go
package model

import "time"

type Coupon struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    MerchantID int64     `gorm:"not null;index" json:"merchant_id"`
    Title      string    `gorm:"type:varchar(100);not null" json:"title"`
    Type       string    `gorm:"type:varchar(20);not null" json:"type"`
    Value      string    `gorm:"type:varchar(50)" json:"value"`
    Price      float64   `json:"price"`
    ExpiryDate time.Time `json:"expiry_date"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

func (c *Coupon) TableName() string { return "coupons" }
```
- `apps/core/internal/domain/coupon/routes.go`
```go
package coupon

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/coupon/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/coupon/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewCouponService(nil)
    h := handler.NewCouponHandler(svc)

    coupons := r.Group("/coupons")
    {
        coupons.POST("/:id/validate", h.Validate)
        coupons.POST("/:id/payment/initiate", h.InitiatePayment)
        coupons.POST("/:id/redeem", middleware.JWTAuth(cfg.JWT), h.Redeem)
    }
}
```
- `apps/core/internal/domain/coupon/handler/coupon.go`
```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/coupon/service"
    "github.com/gin-gonic/gin"
)

type CouponHandler struct { svc *service.CouponService }

func NewCouponHandler(svc *service.CouponService) *CouponHandler {
    if svc == nil { svc = service.NewCouponService(nil) }
    return &CouponHandler{svc: svc}
}

func (h *CouponHandler) Validate(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := h.svc.Validate(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{})
}

func (h *CouponHandler) InitiatePayment(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    var req struct{ UserID string `json:"userId"` }
    _ = c.ShouldBindJSON(&req)
    if err := h.svc.InitiatePayment(c.Request.Context(), id, req.UserID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{})
}

func (h *CouponHandler) Redeem(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    userID := c.GetInt64("user_id")
    if err := h.svc.Redeem(c.Request.Context(), id, userID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{})
}
```
- `apps/core/internal/domain/coupon/service/service.go`
```go
package service

import (
    "context"
    "errors"
    "time"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type CouponService struct { db *gorm.DB }

func NewCouponService(db *gorm.DB) *CouponService {
    if db == nil { db = database.DB }
    return &CouponService{db: db}
}

func (s *CouponService) Validate(ctx context.Context, id int64) error {
    var cpn model.Coupon
    if err := s.db.WithContext(ctx).First(&cpn, id).Error; err != nil {
        return err
    }
    if !cpn.ExpiryDate.IsZero() && cpn.ExpiryDate.Before(time.Now()) {
        return errors.New("expired")
    }
    return nil
}

func (s *CouponService) InitiatePayment(ctx context.Context, couponID int64, userID string) error {
    _ = userID
    var cpn model.Coupon
    if err := s.db.WithContext(ctx).First(&cpn, couponID).Error; err != nil {
        return err
    }
    return nil
}

func (s *CouponService) Redeem(ctx context.Context, couponID, userID int64) error {
    _ = userID
    var cpn model.Coupon
    if err := s.db.WithContext(ctx).First(&cpn, couponID).Error; err != nil {
        return err
    }
    return nil
}
```
- Update `apps/core/internal/router/router.go` to register `coupon.RegisterRoutes`.

**Step 4: Run tests to verify they pass**
Run: `go test ./internal/router -run TestCouponValidateAndRedeem`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/coupon apps/core/internal/model/coupon.go apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add coupon endpoints"
```

---

### Task 7: Voucher endpoints

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Create: `apps/core/internal/model/voucher.go`
- Create: `apps/core/internal/domain/voucher/routes.go`
- Create: `apps/core/internal/domain/voucher/handler/voucher.go`
- Create: `apps/core/internal/domain/voucher/service/service.go`
- Modify: `apps/core/internal/router/router.go`

**Step 1: Write the failing tests**
```go
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
```

**Step 2: Run tests to verify they fail**
Run: `go test ./internal/router -run TestVoucherCreateAndList`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- `apps/core/internal/model/voucher.go`
```go
package model

import "time"

type Voucher struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Code       string    `gorm:"type:varchar(50);uniqueIndex" json:"code"`
    CouponID   int64     `gorm:"not null;index" json:"coupon_id"`
    UserID     int64     `gorm:"not null;index" json:"user_id"`
    Status     string    `gorm:"type:varchar(20);not null" json:"status"`
    ExpiryDate time.Time `json:"expiry_date"`
    QRCode     string    `gorm:"type:varchar(255)" json:"qr_code"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

func (v *Voucher) TableName() string { return "vouchers" }
```
- `apps/core/internal/domain/voucher/routes.go`
```go
package voucher

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/voucher/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/voucher/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewVoucherService(nil)
    h := handler.NewVoucherHandler(svc)

    vouchers := r.Group("/vouchers", middleware.JWTAuth(cfg.JWT))
    {
        vouchers.POST("", h.Create)
        vouchers.GET("", h.List)
        vouchers.GET("/:id", h.Detail)
        vouchers.GET("/code/:code", h.ByCode)
        vouchers.PATCH("/:id/use", h.Use)
        vouchers.PATCH("/:id/status", h.UpdateStatus)
        vouchers.POST("/share/email", h.ShareEmail)
        vouchers.POST("/share/sms", h.ShareSMS)
    }
}
```
- `apps/core/internal/domain/voucher/handler/voucher.go`
```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/voucher/service"
    "github.com/gin-gonic/gin"
)

type VoucherHandler struct { svc *service.VoucherService }

func NewVoucherHandler(svc *service.VoucherService) *VoucherHandler {
    if svc == nil { svc = service.NewVoucherService(nil) }
    return &VoucherHandler{svc: svc}
}

func (h *VoucherHandler) Create(c *gin.Context) {
    var req service.CreateVoucherRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    v, err := h.svc.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, v)
}

func (h *VoucherHandler) List(c *gin.Context) {
    userID := c.GetInt64("user_id")
    list, err := h.svc.List(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *VoucherHandler) Detail(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    v, err := h.svc.Detail(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, v)
}

func (h *VoucherHandler) ByCode(c *gin.Context) {
    v, err := h.svc.ByCode(c.Request.Context(), c.Param("code"))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, v)
}

func (h *VoucherHandler) Use(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := h.svc.Use(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{})
}

func (h *VoucherHandler) UpdateStatus(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := h.svc.UpdateStatus(c.Request.Context(), id, "used"); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{})
}

func (h *VoucherHandler) ShareEmail(c *gin.Context) { c.JSON(http.StatusOK, gin.H{}) }
func (h *VoucherHandler) ShareSMS(c *gin.Context)   { c.JSON(http.StatusOK, gin.H{}) }
```
- `apps/core/internal/domain/voucher/service/service.go`
```go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type CreateVoucherRequest struct {
    CouponID string `json:"couponId"`
    UserID   string `json:"userId"`
    Code     string `json:"code"`
}

type VoucherService struct { db *gorm.DB }

func NewVoucherService(db *gorm.DB) *VoucherService {
    if db == nil { db = database.DB }
    return &VoucherService{db: db}
}

func (s *VoucherService) Create(ctx context.Context, req CreateVoucherRequest) (model.Voucher, error) {
    v := model.Voucher{Code: req.Code, Status: "active"}
    return v, s.db.WithContext(ctx).Create(&v).Error
}

func (s *VoucherService) List(ctx context.Context, userID int64) ([]model.Voucher, error) {
    var list []model.Voucher
    if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&list).Error; err != nil {
        return nil, err
    }
    return list, nil
}

func (s *VoucherService) Detail(ctx context.Context, id int64) (*model.Voucher, error) {
    var v model.Voucher
    if err := s.db.WithContext(ctx).First(&v, id).Error; err != nil { return nil, err }
    return &v, nil
}

func (s *VoucherService) ByCode(ctx context.Context, code string) (*model.Voucher, error) {
    var v model.Voucher
    if err := s.db.WithContext(ctx).Where("code = ?", code).First(&v).Error; err != nil { return nil, err }
    return &v, nil
}

func (s *VoucherService) Use(ctx context.Context, id int64) error {
    return s.db.WithContext(ctx).Model(&model.Voucher{}).Where("id = ?", id).UpdateColumn("status", "used").Error
}

func (s *VoucherService) UpdateStatus(ctx context.Context, id int64, status string) error {
    return s.db.WithContext(ctx).Model(&model.Voucher{}).Where("id = ?", id).UpdateColumn("status", status).Error
}
```
- Update `apps/core/internal/router/router.go` to register `voucher.RegisterRoutes`.

**Step 4: Run tests to verify they pass**
Run: `go test ./internal/router -run TestVoucherCreateAndList`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/voucher apps/core/internal/model/voucher.go apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add voucher endpoints"
```

---

### Task 8: Payments endpoints

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Create: `apps/core/internal/model/payment.go`
- Create: `apps/core/internal/domain/payment/routes.go`
- Create: `apps/core/internal/domain/payment/handler/payment.go`
- Create: `apps/core/internal/domain/payment/service/service.go`
- Modify: `apps/core/internal/router/router.go`

**Step 1: Write the failing tests**
```go
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
```

**Step 2: Run tests to verify they fail**
Run: `go test ./internal/router -run TestPaymentsCreateAndDetail`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- `apps/core/internal/model/payment.go`
```go
package model

import "time"

type Payment struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Amount     float64   `json:"amount"`
    Currency   string    `gorm:"type:varchar(10)" json:"currency"`
    Status     string    `gorm:"type:varchar(20)" json:"status"`
    CouponID   *int64    `json:"coupon_id"`
    MerchantID *int64    `json:"merchant_id"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

func (p *Payment) TableName() string { return "payments" }
```
- `apps/core/internal/domain/payment/routes.go`
```go
package payment

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/payment/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/payment/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewPaymentService(nil)
    h := handler.NewPaymentHandler(svc)

    payments := r.Group("/payments", middleware.JWTAuth(cfg.JWT))
    {
        payments.POST("", h.Create)
        payments.GET("/:id", h.Detail)
    }
}
```
- `apps/core/internal/domain/payment/handler/payment.go`
```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/payment/service"
    "github.com/gin-gonic/gin"
)

type PaymentHandler struct { svc *service.PaymentService }

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
    if svc == nil { svc = service.NewPaymentService(nil) }
    return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) Create(c *gin.Context) {
    var req service.CreatePaymentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    p, err := h.svc.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, p)
}

func (h *PaymentHandler) Detail(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    p, err := h.svc.Detail(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, p)
}
```
- `apps/core/internal/domain/payment/service/service.go`
```go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type CreatePaymentRequest struct {
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
    Status   string  `json:"status"`
}

type PaymentService struct { db *gorm.DB }

func NewPaymentService(db *gorm.DB) *PaymentService {
    if db == nil { db = database.DB }
    return &PaymentService{db: db}
}

func (s *PaymentService) Create(ctx context.Context, req CreatePaymentRequest) (model.Payment, error) {
    p := model.Payment{Amount: req.Amount, Currency: req.Currency, Status: req.Status}
    return p, s.db.WithContext(ctx).Create(&p).Error
}

func (s *PaymentService) Detail(ctx context.Context, id int64) (*model.Payment, error) {
    var p model.Payment
    if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
        return nil, err
    }
    return &p, nil
}
```
- Update `apps/core/internal/router/router.go` to register `payment.RegisterRoutes`.

**Step 4: Run tests to verify they pass**
Run: `go test ./internal/router -run TestPaymentsCreateAndDetail`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/payment apps/core/internal/model/payment.go apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add payment endpoints"
```

---

### Task 9: Media upload + analysis endpoints

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Create: `apps/core/internal/model/media_upload.go`
- Create: `apps/core/internal/domain/media/routes.go`
- Create: `apps/core/internal/domain/media/handler/media.go`
- Create: `apps/core/internal/domain/media/service/service.go`
- Modify: `apps/core/internal/router/router.go`

**Step 1: Write the failing tests**
```go
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
```

**Step 2: Run tests to verify they fail**
Run: `go test ./internal/router -run TestMediaUploadAndAnalysis`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- `apps/core/internal/model/media_upload.go`
```go
package model

import "time"

type MediaUpload struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UploadURL string    `gorm:"type:varchar(255)" json:"upload_url"`
    FileURL   string    `gorm:"type:varchar(255)" json:"file_url"`
    CreatedAt time.Time `json:"created_at"`
}

func (m *MediaUpload) TableName() string { return "media_uploads" }
```
- `apps/core/internal/domain/media/routes.go`
```go
package media

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewMediaService(nil)
    h := handler.NewMediaHandler(svc)

    media := r.Group("/media", middleware.JWTAuth(cfg.JWT))
    {
        media.POST("/uploads", h.CreateUpload)
        media.POST("/:id/analysis", h.Analyze)
    }
}
```
- `apps/core/internal/domain/media/handler/media.go`
```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/media/service"
    "github.com/gin-gonic/gin"
)

type MediaHandler struct { svc *service.MediaService }

func NewMediaHandler(svc *service.MediaService) *MediaHandler {
    if svc == nil { svc = service.NewMediaService(nil) }
    return &MediaHandler{svc: svc}
}

func (h *MediaHandler) CreateUpload(c *gin.Context) {
    upload, err := h.svc.CreateUpload(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, upload)
}

func (h *MediaHandler) Analyze(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := h.svc.Analyze(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{})
}
```
- `apps/core/internal/domain/media/service/service.go`
```go
package service

import (
    "context"
    "fmt"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type MediaService struct { db *gorm.DB }

func NewMediaService(db *gorm.DB) *MediaService {
    if db == nil { db = database.DB }
    return &MediaService{db: db}
}

func (s *MediaService) CreateUpload(ctx context.Context) (model.MediaUpload, error) {
    upload := model.MediaUpload{
        UploadURL: "https://example.com/upload",
        FileURL:   fmt.Sprintf("https://example.com/files/%d", time.Now().UnixNano()),
    }
    return upload, s.db.WithContext(ctx).Create(&upload).Error
}

func (s *MediaService) Analyze(ctx context.Context, id int64) error {
    var upload model.MediaUpload
    return s.db.WithContext(ctx).First(&upload, id).Error
}
```
- Update `apps/core/internal/router/router.go` to register `media.RegisterRoutes`.

**Step 4: Run tests to verify they pass**
Run: `go test ./internal/router -run TestMediaUploadAndAnalysis`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/media apps/core/internal/model/media_upload.go apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add media endpoints"
```

---

### Task 10: AI suggestions endpoint

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Create: `apps/core/internal/domain/ai/routes.go`
- Create: `apps/core/internal/domain/ai/handler/ai.go`
- Create: `apps/core/internal/domain/ai/service/service.go`
- Modify: `apps/core/internal/router/router.go`

**Step 1: Write the failing test**
```go
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
```

**Step 2: Run test to verify it fails**
Run: `go test ./internal/router -run TestAISuggestions`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- `apps/core/internal/domain/ai/routes.go`
```go
package ai

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewAIService()
    h := handler.NewAIHandler(svc)

    ai := r.Group("/ai", middleware.JWTAuth(cfg.JWT))
    {
        ai.POST("/reviews/suggestions", h.Suggestions)
    }
}
```
- `apps/core/internal/domain/ai/handler/ai.go`
```go
package handler

import (
    "net/http"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/ai/service"
    "github.com/gin-gonic/gin"
)

type AIHandler struct { svc *service.AIService }

func NewAIHandler(svc *service.AIService) *AIHandler { return &AIHandler{svc: svc} }

func (h *AIHandler) Suggestions(c *gin.Context) {
    var req service.SuggestionsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    resp := h.svc.Suggestions(req)
    c.JSON(http.StatusOK, resp)
}
```
- `apps/core/internal/domain/ai/service/service.go`
```go
package service

type SuggestionsRequest struct {
    OverallRating  float64 `json:"overallRating"`
    BusinessCategory string `json:"businessCategory"`
    CurrentText    string  `json:"currentText"`
    MerchantName   string  `json:"merchantName"`
}

type SuggestionsResponse struct {
    Suggestions []string `json:"suggestions"`
}

type AIService struct{}

func NewAIService() *AIService { return &AIService{} }

func (s *AIService) Suggestions(req SuggestionsRequest) SuggestionsResponse {
    name := req.MerchantName
    if name == "" { name = "this place" }
    return SuggestionsResponse{Suggestions: []string{
        "Highlight what you liked about " + name + ".",
        "Mention any standout service or atmosphere details.",
        "Add one concrete example to support your rating.",
    }}
}
```
- Update `apps/core/internal/router/router.go` to register `ai.RegisterRoutes`.

**Step 4: Run test to verify it passes**
Run: `go test ./internal/router -run TestAISuggestions`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/ai apps/core/internal/router/router.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add ai suggestions endpoint"
```

---

### Task 11: Auth forgot-password endpoint

**Files:**
- Modify: `apps/core/internal/router/openapi_v1_test.go`
- Modify: `apps/core/internal/domain/auth/routes.go`
- Modify: `apps/core/internal/domain/auth/handler.go`

**Step 1: Write the failing test**
```go
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
```

**Step 2: Run test to verify it fails**
Run: `go test ./internal/router -run TestAuthForgotPassword`
Expected: FAIL with 404

**Step 3: Write minimal implementation**
- Update `apps/core/internal/domain/auth/routes.go`
```go
auth.POST("/forgot-password", handler.ForgotPassword)
```
- Update `apps/core/internal/domain/auth/handler.go`
```go
func (h *Handler) ForgotPassword(c *gin.Context) {
    var req struct{ Email string `json:"email"` }
    if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
        return
    }
    // Placeholder: do not reveal user existence.
    c.JSON(http.StatusOK, gin.H{})
}
```

**Step 4: Run test to verify it passes**
Run: `go test ./internal/router -run TestAuthForgotPassword`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/domain/auth/routes.go apps/core/internal/domain/auth/handler.go apps/core/internal/router/openapi_v1_test.go
git commit -m "feat: add forgot-password endpoint"
```

---

### Task 12: Update test DB migrations + init-db.sql

**Files:**
- Modify: `apps/core/internal/testutil/db.go`
- Modify: `apps/core/init-db.sql`

**Step 1: Write the failing test**
```go
func TestNewModelsAutoMigrate(t *testing.T) {
    db := testutil.SetupTestDB(t)
    if err := db.Exec("SELECT 1 FROM coupons").Error; err != nil {
        t.Fatalf("expected coupons table: %v", err)
    }
}
```

**Step 2: Run test to verify it fails**
Run: `go test ./internal/router -run TestNewModelsAutoMigrate`
Expected: FAIL with "no such table"

**Step 3: Write minimal implementation**
- Update `apps/core/internal/testutil/db.go` AutoMigrate list:
```go
&model.Coupon{},
&model.Voucher{},
&model.Payment{},
&model.MediaUpload{},
&model.ReviewComment{},
```
- Extend `apps/core/init-db.sql` with new tables and indexes matching models (`coupons`, `vouchers`, `payments`, `media_uploads`, `review_comments`).

**Step 4: Run test to verify it passes**
Run: `go test ./internal/router -run TestNewModelsAutoMigrate`
Expected: PASS

**Step 5: Commit**
```bash
git add apps/core/internal/testutil/db.go apps/core/init-db.sql apps/core/internal/router/openapi_v1_test.go
git commit -m "chore: add schema for new openapi models"
```

---

**Plan complete and saved to** `docs/plans/2026-02-04-openapi-v1-implementation.md`.

Two execution options:

1. Subagent-Driven (this session) - I dispatch a fresh subagent per task, review between tasks.
2. Parallel Session (separate) - Open a new session in the worktree and execute with `superpowers:executing-plans`.

Which approach?
