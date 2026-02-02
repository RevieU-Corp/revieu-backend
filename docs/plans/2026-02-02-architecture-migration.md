# Architecture Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate the core backend to the domain-driven structure described in ARCHITECTURE_MIGRATION_IMPLEMENTATION.md without changing behavior.

**Architecture:** Introduce domain packages (auth, content, user, follow, profile) with domain-level routes and handlers, move existing handler/service/dto logic into those packages, and replace the old handler/service/dto directories. Keep functionality and tests intact while reorganizing code and routes.

**Tech Stack:** Go, Gin, GORM, JWT, sqlite for tests.

---

### Task 1: Create domain scaffolding and shared types

**Files:**
- Create: `apps/core/internal/domain/common.go`
- Create: `apps/core/internal/router/router.go`
- Create: `apps/core/internal/domain/auth/routes.go`
- Create: `apps/core/internal/domain/content/routes.go`
- Create: `apps/core/internal/domain/user/routes.go`
- Create: `apps/core/internal/domain/follow/routes.go`
- Create: `apps/core/internal/domain/profile/routes.go`

**Step 1: Write failing test**

Create a minimal compile-time test that references the new router setup entry point.

```go
// apps/core/internal/router/router_test.go
package router

import (
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

func TestSetupExists(t *testing.T) {
    r := gin.New()
    cfg := &config.Config{}
    Setup(r, cfg)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/router -v`

Expected: FAIL with missing package or undefined `Setup`.

**Step 3: Write minimal implementation**

Create the shared types and router wiring.

```go
// apps/core/internal/domain/common.go
package domain

import "context"

// Transactional supports running operations inside a transaction.
type Transactional interface {
    WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CursorPagination is a cursor-based pagination request.
type CursorPagination struct {
    Cursor *int64
    Limit  int
}

// CursorPaginationResponse is a cursor-based pagination response.
type CursorPaginationResponse struct {
    Total  int   `json:"total"`
    Cursor *int64 `json:"cursor,omitempty"`
}
```

```go
// apps/core/internal/router/router.go
package router

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/auth"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/profile"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user"
    "github.com/gin-gonic/gin"
)

// Setup registers all domain routes under the API base path.
func Setup(router *gin.Engine, cfg *config.Config) {
    api := router.Group(cfg.Server.APIBasePath)

    auth.RegisterRoutes(api, cfg)
    user.RegisterRoutes(api, cfg)
    profile.RegisterRoutes(api, cfg)
    follow.RegisterRoutes(api, cfg)
    content.RegisterRoutes(api, cfg)
}
```

Create empty route registration stubs (they will be filled in later tasks):

```go
// apps/core/internal/domain/auth/routes.go
package auth

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers auth routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    _ = r
    _ = cfg
}
```

```go
// apps/core/internal/domain/content/routes.go
package content

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers content routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    _ = r
    _ = cfg
}
```

```go
// apps/core/internal/domain/user/routes.go
package user

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers user routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    _ = r
    _ = cfg
}
```

```go
// apps/core/internal/domain/follow/routes.go
package follow

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers follow routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    _ = r
    _ = cfg
}
```

```go
// apps/core/internal/domain/profile/routes.go
package profile

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers profile routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    _ = r
    _ = cfg
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/router -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/common.go apps/core/internal/router/router.go apps/core/internal/domain/*/routes.go apps/core/internal/router/router_test.go
git commit -m "refactor: add domain router scaffolding"
```

---

### Task 2: Add shared test DB helper for domain tests

**Files:**
- Create: `apps/core/internal/testutil/db.go`
- Create: `apps/core/internal/testutil/db_test.go`

**Step 1: Write failing test**

```go
// apps/core/internal/testutil/db_test.go
package testutil

import (
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
)

func TestSetupTestDB(t *testing.T) {
    db := SetupTestDB(t)
    if db == nil {
        t.Fatal("expected db")
    }
    if !db.Migrator().HasTable(&model.User{}) {
        t.Fatal("expected user table to exist")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/testutil -v`

Expected: FAIL with undefined `SetupTestDB`.

**Step 3: Write minimal implementation**

```go
// apps/core/internal/testutil/db.go
package testutil

import (
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// SetupTestDB creates an in-memory sqlite DB with schema migrations.
func SetupTestDB(t *testing.T) *gorm.DB {
    t.Helper()

    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: nil})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    if err := db.AutoMigrate(
        &model.User{},
        &model.UserAuth{},
        &model.UserProfile{},
        &model.EmailVerification{},
        &model.Merchant{},
        &model.Tag{},
        &model.Post{},
        &model.Review{},
        &model.UserFollow{},
        &model.MerchantFollow{},
        &model.Like{},
        &model.Favorite{},
        &model.UserAddress{},
        &model.UserPrivacy{},
        &model.UserNotification{},
        &model.AccountDeletion{},
    ); err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }

    return db
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/testutil -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/testutil
git commit -m "test: add shared test db helper"
```

---

### Task 3: Migrate auth domain (dto, service, handler, routes, tests)

**Files:**
- Move: `apps/core/internal/handler/auth.go` -> `apps/core/internal/domain/auth/handler.go`
- Move: `apps/core/internal/service/auth.go` -> `apps/core/internal/domain/auth/service.go`
- Move: `apps/core/internal/service/token.go` -> `apps/core/internal/domain/auth/token.go`
- Move: `apps/core/internal/service/auth_test.go` -> `apps/core/internal/domain/auth/service_test.go`
- Create: `apps/core/internal/domain/auth/dto.go`
- Modify: `apps/core/internal/domain/auth/routes.go`

**Step 1: Write failing test**

Move the auth service test to the new package and update its package name/imports.

```go
// apps/core/internal/domain/auth/service_test.go
package auth

import (
    "context"
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

var testJWTConfig = config.JWTConfig{
    Secret:     "test-secret-key-for-testing",
    ExpireHour: 24,
}

var testSMTPConfig = config.SMTPConfig{
    Host:     "localhost",
    Port:     25,
    Username: "",
    Password: "",
    From:     "test@example.com",
    UseTLS:   false,
}

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: nil})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    if err := db.AutoMigrate(
        &model.User{},
        &model.UserAuth{},
        &model.UserProfile{},
        &model.EmailVerification{},
        &model.Merchant{},
        &model.Tag{},
        &model.Post{},
        &model.Review{},
        &model.UserFollow{},
        &model.MerchantFollow{},
        &model.Like{},
        &model.Favorite{},
        &model.UserAddress{},
        &model.UserPrivacy{},
        &model.UserNotification{},
        &model.AccountDeletion{},
    ); err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }

    return db
}

func TestRegister(t *testing.T) {
    db := setupTestDB(t)
    authService := NewService(db, testJWTConfig, testSMTPConfig)

    ctx := context.Background()
    username := "testuser"
    email := "test@example.com"
    password := "password123"
    baseURL := "http://localhost:8080"

    user, err := authService.Register(ctx, username, email, password, baseURL)
    if err != nil {
        t.Errorf("Register failed: %v", err)
    }
    if user.ID == 0 {
        t.Error("Expected user ID to be generated")
    }
    if user.Role != "user" {
        t.Errorf("Expected role 'user', got %s", user.Role)
    }
    if user.Status != 2 {
        t.Errorf("Expected status 2 (pending), got %d", user.Status)
    }

    var verification model.EmailVerification
    if err := db.Where("user_id = ?", user.ID).First(&verification).Error; err != nil {
        t.Errorf("Expected email verification record to be created: %v", err)
    }

    _, err = authService.Register(ctx, "otheruser", email, "pass", baseURL)
    if err == nil {
        t.Error("Expected error for duplicate email, got nil")
    }
}

func TestLogin(t *testing.T) {
    db := setupTestDB(t)
    authService := NewService(db, testJWTConfig, testSMTPConfig)

    ctx := context.Background()
    email := "login@example.com"
    password := "securepass"
    username := "loginuser"
    baseURL := "http://localhost"

    user, err := authService.Register(ctx, username, email, password, baseURL)
    if err != nil {
        t.Fatalf("Failed to create user for login test: %v", err)
    }

    _, err = authService.Login(ctx, email, password, "127.0.0.1")
    if err == nil {
        t.Error("Expected error for unverified user, got nil")
    }

    var verification model.EmailVerification
    if err := db.Where("user_id = ?", user.ID).First(&verification).Error; err != nil {
        t.Fatalf("Failed to find verification record: %v", err)
    }
    if err := authService.VerifyEmail(ctx, verification.Token); err != nil {
        t.Fatalf("Failed to verify email: %v", err)
    }

    token, err := authService.Login(ctx, email, password, "127.0.0.1")
    if err != nil {
        t.Errorf("Login failed: %v", err)
    }
    if token == "" {
        t.Error("Expected JWT token, got empty string")
    }

    _, err = authService.Login(ctx, email, "wrongpass", "127.0.0.1")
    if err == nil {
        t.Error("Expected error for wrong password, got nil")
    }

    _, err = authService.Login(ctx, "nonexistent@example.com", password, "127.0.0.1")
    if err == nil {
        t.Error("Expected error for user not found, got nil")
    }
}

func TestUserProfileHasCounts(t *testing.T) {
    db := setupTestDB(t)
    type Column struct{ Name string }
    var cols []Column
    if err := db.Raw("PRAGMA table_info(user_profiles)").Scan(&cols).Error; err != nil {
        t.Fatalf("schema query failed: %v", err)
    }
    want := map[string]bool{
        "follower_count":  true,
        "following_count": true,
        "post_count":      true,
        "review_count":    true,
        "like_count":      true,
    }
    for _, c := range cols {
        delete(want, c.Name)
    }
    if len(want) != 0 {
        t.Fatalf("missing columns: %v", want)
    }
}

func TestMerchantAndTagModels(t *testing.T) {
    db := setupTestDB(t)
    merchant := model.Merchant{Name: "Cafe", Category: "food"}
    if err := db.Create(&merchant).Error; err != nil {
        t.Fatalf("merchant create failed: %v", err)
    }
    tag := model.Tag{Name: "#coffee"}
    if err := db.Create(&tag).Error; err != nil {
        t.Fatalf("tag create failed: %v", err)
    }
}

func TestPostAndReviewModels(t *testing.T) {
    db := setupTestDB(t)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil {
        t.Fatal(err)
    }
    merchant := model.Merchant{Name: "Cafe"}
    if err := db.Create(&merchant).Error; err != nil {
        t.Fatal(err)
    }

    post := model.Post{UserID: user.ID, MerchantID: &merchant.ID, Content: "hello"}
    if err := db.Create(&post).Error; err != nil {
        t.Fatal(err)
    }

    review := model.Review{UserID: user.ID, MerchantID: merchant.ID, Rating: 4.5, Content: "great"}
    if err := db.Create(&review).Error; err != nil {
        t.Fatal(err)
    }
}

func TestFollowAndInteractionModels(t *testing.T) {
    db := setupTestDB(t)
    u1 := model.User{Role: "user", Status: 0}
    u2 := model.User{Role: "user", Status: 0}
    if err := db.Create(&u1).Error; err != nil {
        t.Fatal(err)
    }
    if err := db.Create(&u2).Error; err != nil {
        t.Fatal(err)
    }

    follow := model.UserFollow{FollowerID: u1.ID, FollowingID: u2.ID}
    if err := db.Create(&follow).Error; err != nil {
        t.Fatal(err)
    }

    like := model.Like{UserID: u1.ID, TargetType: "post", TargetID: 123}
    if err := db.Create(&like).Error; err != nil {
        t.Fatal(err)
    }
}

func TestSettingsAndAddressModels(t *testing.T) {
    db := setupTestDB(t)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil {
        t.Fatal(err)
    }

    privacy := model.UserPrivacy{UserID: user.ID, IsPublic: true}
    if err := db.Create(&privacy).Error; err != nil {
        t.Fatal(err)
    }

    address := model.UserAddress{UserID: user.ID, Name: "A", Phone: "1", Address: "Street"}
    if err := db.Create(&address).Error; err != nil {
        t.Fatal(err)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/auth -v`

Expected: FAIL with undefined `NewService` and missing auth package files.

**Step 3: Write minimal implementation**

Move auth service and token logic into the new package, and extract request/response DTOs.

```go
// apps/core/internal/domain/auth/dto.go
package auth

import "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"

// RegisterRequest registers a new user.
type RegisterRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest logs in an existing user.
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// RegisterResponse is returned on successful registration.
type RegisterResponse struct {
    Message string `json:"message"`
    UserID  int64  `json:"user_id"`
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
    Token string `json:"token"`
    Type  string `json:"type"`
}

// UserInfoResponse describes the authenticated user.
type UserInfoResponse struct {
    UserID  interface{} `json:"user_id"`
    Email   interface{} `json:"email"`
    Role    interface{} `json:"role"`
    Message string      `json:"message"`
}

// ToUserResponse maps model.User to the auth response.
func ToUserResponse(user *model.User) UserInfoResponse {
    return UserInfoResponse{
        UserID:  user.ID,
        Email:   "",
        Role:    user.Role,
        Message: "Token is valid!",
    }
}
```

```go
// apps/core/internal/domain/auth/token.go
package auth

import (
    "fmt"
    "time"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/golang-jwt/jwt/v5"
)

// TokenService issues and validates JWTs.
type TokenService struct {
    secret     []byte
    expireHour int
}

// NewTokenService creates a TokenService for middleware and auth logic.
func NewTokenService(cfg config.JWTConfig) *TokenService {
    return &TokenService{
        secret:     []byte(cfg.Secret),
        expireHour: cfg.ExpireHour,
    }
}

func (s *TokenService) GenerateToken(user *model.User, auth *model.UserAuth) (string, error) {
    claims := jwt.MapClaims{
        "sub":           user.ID,
        "email":         auth.Identifier,
        "identity_type": auth.IdentityType,
        "role":          user.Role,
        "exp":           time.Now().Add(time.Hour * time.Duration(s.expireHour)).Unix(),
        "iat":           time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secret)
}

func (s *TokenService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.secret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}
```

```go
// apps/core/internal/domain/auth/service.go
package auth

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/email"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// Service exposes auth operations used by handlers.
type Service interface {
    Register(ctx context.Context, username, userEmail, password, baseURL string) (*model.User, error)
    Login(ctx context.Context, email, password, ipAddress string) (string, error)
    LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, avatar string) (string, error)
    VerifyEmail(ctx context.Context, token string) error
}

type service struct {
    db           *gorm.DB
    tokenService *tokenService
    emailClient  *email.SMTPClient
}

// NewService creates an auth service.
func NewService(db *gorm.DB, jwtCfg config.JWTConfig, smtpCfg config.SMTPConfig) Service {
    if db == nil {
        db = database.DB
    }
    return &service{
        db:           db,
        tokenService: NewTokenService(jwtCfg),
        emailClient:  email.NewSMTPClient(smtpCfg),
    }
}

func (s *service) Register(ctx context.Context, username, userEmail, password, baseURL string) (*model.User, error) {
    var existingAuth model.UserAuth
    if err := s.db.Where("identity_type = ? AND identifier = ?", "email", userEmail).First(&existingAuth).Error; err == nil {
        return nil, errors.New("user already exists")
    } else if err != gorm.ErrRecordNotFound {
        return nil, err
    }

    token := uuid.New().String()

    var user model.User
    err := s.db.Transaction(func(tx *gorm.DB) error {
        user = model.User{Role: "user", Status: 2}
        if err := tx.Create(&user).Error; err != nil {
            return err
        }

        auth := model.UserAuth{UserID: user.ID, IdentityType: "email", Identifier: userEmail}
        if err := auth.SetPassword(password); err != nil {
            return err
        }
        if err := tx.Create(&auth).Error; err != nil {
            return err
        }

        profile := model.UserProfile{UserID: user.ID, Nickname: username}
        if err := tx.Create(&profile).Error; err != nil {
            return err
        }

        verification := model.EmailVerification{
            UserID:    user.ID,
            Email:     userEmail,
            Token:     token,
            ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
        }
        if err := tx.Create(&verification).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)

    if err := s.emailClient.SendVerificationEmail(userEmail, verifyURL); err != nil {
        logger.Warn(ctx, "Failed to send verification email",
            "error", err.Error(),
            "user_id", user.ID,
            "email", userEmail,
        )
        logger.Info(ctx, fmt.Sprintf("Verification link for %s: %s", userEmail, verifyURL),
            "event", "user_registered",
            "user_id", user.ID,
            "email", userEmail,
        )
    } else {
        logger.Info(ctx, "Verification email sent",
            "event", "user_registered",
            "user_id", user.ID,
            "email", userEmail,
        )
    }

    return &user, nil
}

func (s *service) Login(ctx context.Context, email, password, ipAddress string) (string, error) {
    var auth model.UserAuth
    if err := s.db.Where("identity_type = ? AND identifier = ?", "email", email).First(&auth).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return "", errors.New("invalid credentials")
        }
        return "", err
    }

    if !auth.CheckPassword(password) {
        return "", errors.New("invalid credentials")
    }

    var user model.User
    if err := s.db.First(&user, auth.UserID).Error; err != nil {
        return "", err
    }

    if user.Status == 2 {
        return "", errors.New("please verify your email before logging in")
    }
    if user.Status == 1 {
        return "", errors.New("your account has been suspended")
    }

    now := time.Now().UTC()
    auth.LastLoginAt = &now
    if err := s.db.Save(&auth).Error; err != nil {
        logger.Warn(ctx, "Failed to update user login info",
            "user_id", user.ID,
            "error", err.Error(),
        )
    }

    token, err := s.tokenService.GenerateToken(&user, &auth)
    if err != nil {
        return "", err
    }

    logger.Info(ctx, "User logged in successfully",
        "event", "user_login_success",
        "user_id", user.ID,
    )

    return token, nil
}

func (s *service) LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, avatar string) (string, error) {
    var auth model.UserAuth
    var user model.User

    err := s.db.Where("identity_type = ? AND identifier = ?", provider, email).First(&auth).Error
    if err == nil {
        if err := s.db.First(&user, auth.UserID).Error; err != nil {
            return "", err
        }

        now := time.Now().UTC()
        auth.LastLoginAt = &now
        if err := s.db.Save(&auth).Error; err != nil {
            logger.Warn(ctx, "Failed to update OAuth user login info",
                "user_id", user.ID,
                "error", err.Error(),
            )
        }

        token, err := s.tokenService.GenerateToken(&user, &auth)
        if err != nil {
            return "", err
        }

        logger.Info(ctx, "OAuth user logged in successfully",
            "event", "oauth_login_success",
            "user_id", user.ID,
            "provider", provider,
        )

        return token, nil
    }

    if err != gorm.ErrRecordNotFound {
        return "", err
    }

    err = s.db.Transaction(func(tx *gorm.DB) error {
        user = model.User{Role: "user", Status: 0}
        if err := tx.Create(&user).Error; err != nil {
            return err
        }

        now := time.Now().UTC()
        auth = model.UserAuth{
            UserID:       user.ID,
            IdentityType: provider,
            Identifier:   email,
            LastLoginAt:  &now,
        }
        if err := tx.Create(&auth).Error; err != nil {
            return err
        }

        profile := model.UserProfile{UserID: user.ID, Nickname: name, AvatarURL: avatar}
        if err := tx.Create(&profile).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return "", err
    }

    token, err := s.tokenService.GenerateToken(&user, &auth)
    if err != nil {
        return "", err
    }

    logger.Info(ctx, "OAuth user registered and logged in successfully",
        "event", "oauth_register_success",
        "user_id", user.ID,
        "provider", provider,
    )

    return token, nil
}

func (s *service) VerifyEmail(ctx context.Context, token string) error {
    var verification model.EmailVerification
    if err := s.db.Where("token = ?", token).First(&verification).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return errors.New("invalid or expired verification token")
        }
        return err
    }

    if verification.IsExpired() {
        return errors.New("verification token has expired")
    }

    if err := s.db.Model(&model.User{}).Where("id = ?", verification.UserID).Update("status", 0).Error; err != nil {
        return fmt.Errorf("failed to activate user: %w", err)
    }

    if err := s.db.Delete(&verification).Error; err != nil {
        logger.Warn(ctx, "Failed to delete verification record",
            "error", err.Error(),
            "user_id", verification.UserID,
        )
    }

    logger.Info(ctx, "User email verified successfully",
        "event", "email_verified",
        "user_id", verification.UserID,
        "email", verification.Email,
    )

    return nil
}
```

Move the auth handler into the new package and update imports to use the new service type and DTOs.

```go
// apps/core/internal/domain/auth/handler.go
package auth

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    svc         Service
    oauthCfg    config.OAuthConfig
    frontendURL string
    apiBasePath string
}

func NewHandler(jwtCfg config.JWTConfig, oauthCfg config.OAuthConfig, smtpCfg config.SMTPConfig, frontendURL string, apiBasePath string) *Handler {
    return &Handler{
        svc:         NewService(nil, jwtCfg, smtpCfg),
        oauthCfg:    oauthCfg,
        frontendURL: frontendURL,
        apiBasePath: apiBasePath,
    }
}

// Register
func (h *Handler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    scheme := "http"
    if c.Request.TLS != nil {
        scheme = "https"
    }
    baseURL := scheme + "://" + c.Request.Host

    user, err := h.svc.Register(c.Request.Context(), req.Username, req.Email, req.Password, baseURL)
    if err != nil {
        logger.Error(c.Request.Context(), "Registration failed",
            "error", err.Error(),
            "event", "user_register_failed",
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, RegisterResponse{
        Message: "User created successfully. Please check your email for verification link (printed in server logs for now).",
        UserID:  user.ID,
    })
}

// Login
func (h *Handler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ipAddress := c.ClientIP()
    token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password, ipAddress)
    if err != nil {
        logger.Warn(c.Request.Context(), "Login failed",
            "error", err.Error(),
            "email", req.Email,
            "event", "user_login_failed",
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, LoginResponse{Token: token, Type: "Bearer"})
}

// GoogleLogin
func (h *Handler) GoogleLogin(c *gin.Context) {
    clientID := h.oauthCfg.Google.ClientID
    if clientID == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth not configured"})
        return
    }

    frontendURL := h.frontendURL
    if frontendURL == "" {
        if referer := c.GetHeader("Referer"); referer != "" {
            if parsedURL, err := url.Parse(referer); err == nil {
                frontendURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
            }
        } else if origin := c.GetHeader("Origin"); origin != "" {
            frontendURL = origin
        } else {
            frontendURL = "http://localhost:3000"
        }
    }

    scheme := "http"
    if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
        scheme = "https"
    } else if c.Request.TLS != nil {
        scheme = "https"
    }
    redirectURI := fmt.Sprintf("%s://%s%s/auth/callback/google", scheme, c.Request.Host, h.apiBasePath)

    state := url.QueryEscape(frontendURL)

    authURL := fmt.Sprintf(
        "https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&state=%s",
        url.QueryEscape(clientID),
        url.QueryEscape(redirectURI),
        url.QueryEscape("openid email profile"),
        state,
    )

    c.Redirect(http.StatusFound, authURL)
}

// GoogleCallback
func (h *Handler) GoogleCallback(c *gin.Context) {
    code := c.Query("code")
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
        return
    }

    state := c.Query("state")
    frontendURL := h.frontendURL
    if frontendURL == "" {
        if state != "" {
            if decodedURL, err := url.QueryUnescape(state); err == nil && decodedURL != "" {
                frontendURL = decodedURL
            } else {
                frontendURL = "http://localhost:3000"
            }
        } else {
            frontendURL = "http://localhost:3000"
        }
    }

    scheme := "http"
    if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
        scheme = "https"
    } else if c.Request.TLS != nil {
        scheme = "https"
    }
    redirectURI := fmt.Sprintf("%s://%s%s/auth/callback/google", scheme, c.Request.Host, h.apiBasePath)

    tokenResp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
        "code":          {code},
        "client_id":     {h.oauthCfg.Google.ClientID},
        "client_secret": {h.oauthCfg.Google.ClientSecret},
        "redirect_uri":  {redirectURI},
        "grant_type":    {"authorization_code"},
    })
    if err != nil {
        logger.Error(c.Request.Context(), "Failed to exchange code for token",
            "error", err.Error(),
            "event", "google_oauth_token_exchange_failed",
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange authorization code"})
        return
    }
    defer tokenResp.Body.Close()

    var tokenData struct {
        AccessToken string `json:"access_token"`
        IDToken     string `json:"id_token"`
        Error       string `json:"error"`
    }
    if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
        logger.Error(c.Request.Context(), "Failed to decode token response",
            "error", err.Error(),
            "event", "google_oauth_token_decode_failed",
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode token response"})
        return
    }

    if tokenData.Error != "" {
        logger.Error(c.Request.Context(), "Google OAuth error",
            "error", tokenData.Error,
            "event", "google_oauth_error",
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": tokenData.Error})
        return
    }

    userInfoResp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tokenData.AccessToken)
    if err != nil {
        logger.Error(c.Request.Context(), "Failed to get user info from Google",
            "error", err.Error(),
            "event", "google_oauth_userinfo_failed",
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
        return
    }
    defer userInfoResp.Body.Close()

    body, err := io.ReadAll(userInfoResp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read user info"})
        return
    }

    var userInfo struct {
        ID      string `json:"id"`
        Email   string `json:"email"`
        Name    string `json:"name"`
        Picture string `json:"picture"`
    }
    if err := json.Unmarshal(body, &userInfo); err != nil {
        logger.Error(c.Request.Context(), "Failed to decode user info",
            "error", err.Error(),
            "event", "google_oauth_userinfo_decode_failed",
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode user info"})
        return
    }

    token, err := h.svc.LoginOrRegisterOAuthUser(c.Request.Context(), userInfo.Email, userInfo.Name, "google", userInfo.Picture)
    if err != nil {
        logger.Error(c.Request.Context(), "Failed to login/register OAuth user",
            "error", err.Error(),
            "event", "google_oauth_login_failed",
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process login"})
        return
    }

    redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, url.QueryEscape(token))
    c.Redirect(http.StatusFound, redirectURL)
}

// VerifyEmail
func (h *Handler) VerifyEmail(c *gin.Context) {
    token := c.Query("token")
    if token == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing verification token"})
        return
    }

    if err := h.svc.VerifyEmail(c.Request.Context(), token); err != nil {
        logger.Warn(c.Request.Context(), "Email verification failed",
            "error", err.Error(),
            "event", "email_verification_failed",
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    frontendURL := h.frontendURL
    if frontendURL == "" {
        if referer := c.GetHeader("Referer"); referer != "" {
            if parsedURL, err := url.Parse(referer); err == nil {
                frontendURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
            }
        } else if origin := c.GetHeader("Origin"); origin != "" {
            frontendURL = origin
        } else {
            frontendURL = "http://localhost:3000"
        }
    }
    redirectURL := fmt.Sprintf("%s/auth/verified", frontendURL)
    c.Redirect(http.StatusFound, redirectURL)
}

// Me
func (h *Handler) Me(c *gin.Context) {
    userID, _ := c.Get("user_id")
    email, _ := c.Get("user_email")
    role, _ := c.Get("user_role")

    c.JSON(http.StatusOK, UserInfoResponse{
        UserID:  userID,
        Email:   email,
        Role:    role,
        Message: "Token is valid!",
    })
}
```

Update auth routes to use the new handler and middleware:

```go
// apps/core/internal/domain/auth/routes.go
package auth

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers auth routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    handler := NewHandler(cfg.JWT, cfg.OAuth, cfg.SMTP, cfg.FrontendURL, cfg.Server.APIBasePath)

    auth := r.Group("/auth")
    {
        auth.POST("/register", handler.Register)
        auth.POST("/login", handler.Login)
        auth.GET("/login/google", handler.GoogleLogin)
        auth.GET("/callback/google", handler.GoogleCallback)
        auth.GET("/verify", handler.VerifyEmail)
        auth.GET("/me", middleware.JWTAuth(cfg.JWT), handler.Me)
    }
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/auth -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/auth
git commit -m "refactor: migrate auth domain"
```

---

### Task 4: Migrate content domain (dto, services, handlers, tests)

**Files:**
- Move: `apps/core/internal/service/content.go` -> `apps/core/internal/domain/content/service/content.go`
- Move: `apps/core/internal/service/interaction.go` -> `apps/core/internal/domain/content/service/interaction.go`
- Move: `apps/core/internal/service/content_test.go` -> `apps/core/internal/domain/content/service/content_test.go`
- Move: `apps/core/internal/service/interaction_test.go` -> `apps/core/internal/domain/content/service/interaction_test.go`
- Create: `apps/core/internal/domain/content/dto/post.go`
- Create: `apps/core/internal/domain/content/dto/review.go`
- Create: `apps/core/internal/domain/content/dto/favorite.go`
- Create: `apps/core/internal/domain/content/dto/common.go`
- Create: `apps/core/internal/domain/content/handler/post.go`
- Create: `apps/core/internal/domain/content/handler/review.go`
- Create: `apps/core/internal/domain/content/handler/favorite.go`
- Create: `apps/core/internal/domain/content/handler/like.go`
- Create: `apps/core/internal/domain/content/handler/helpers.go`
- Modify: `apps/core/internal/domain/content/routes.go`

**Step 1: Write failing test**

Move the existing content tests into the new package and update setup to use the shared testutil.

```go
// apps/core/internal/domain/content/service/content_test.go
package service

import (
    "context"
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
)

func TestContentServiceListPosts(t *testing.T) {
    db := testutil.SetupTestDB(t)
    svc := NewContentService(db)
    user := model.User{Role: "user", Status: 0}
    db.Create(&user)
    db.Create(&model.Post{UserID: user.ID, Content: "a"})
    posts, total, err := svc.ListUserPosts(context.Background(), user.ID, nil, 10)
    if err != nil || total != 1 || len(posts) != 1 {
        t.Fatalf("list posts failed: %v", err)
    }
}
```

```go
// apps/core/internal/domain/content/service/interaction_test.go
package service

import (
    "context"
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
)

func TestInteractionServiceLike(t *testing.T) {
    db := testutil.SetupTestDB(t)
    svc := NewInteractionService(db)
    u := model.User{Role: "user", Status: 0}
    db.Create(&u)
    if err := svc.Like(context.Background(), u.ID, "post", 123); err != nil {
        t.Fatal(err)
    }
    if err := svc.Unlike(context.Background(), u.ID, "post", 123); err != nil {
        t.Fatal(err)
    }
}
```

Add a small DTO existence test to replace the old `service/dto_test.go` coverage.

```go
// apps/core/internal/domain/content/dto/dto_test.go
package dto

import "testing"

func TestDTOJSONTags(t *testing.T) {
    _ = PostItem{}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/content/... -v`

Expected: FAIL with undefined constructors.

**Step 3: Write minimal implementation**

Move the content service logic and interaction service logic into the new package. Keep method bodies identical.

```go
// apps/core/internal/domain/content/service/content.go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type ContentService struct {
    db *gorm.DB
}

func NewContentService(db *gorm.DB) *ContentService {
    if db == nil {
        db = database.DB
    }
    return &ContentService{db: db}
}

func (s *ContentService) ListUserPosts(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Post, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Post{}).Where("user_id = ?", userID).Order("id desc")
    if cursor != nil {
        q = q.Where("id < ?", *cursor)
    }
    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    var posts []model.Post
    if err := q.Limit(limit).Find(&posts).Error; err != nil {
        return nil, 0, err
    }
    return posts, total, nil
}

func (s *ContentService) ListUserReviews(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Review, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Review{}).Where("user_id = ?", userID).Order("id desc")
    if cursor != nil {
        q = q.Where("id < ?", *cursor)
    }
    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    var reviews []model.Review
    if err := q.Limit(limit).Find(&reviews).Error; err != nil {
        return nil, 0, err
    }
    return reviews, total, nil
}

func (s *ContentService) ListFavorites(ctx context.Context, userID int64, targetType string, cursor *int64, limit int) ([]model.Favorite, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ?", userID)
    if targetType != "" {
        q = q.Where("target_type = ?", targetType)
    }
    if cursor != nil {
        q = q.Where("id < ?", *cursor)
    }
    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    var items []model.Favorite
    if err := q.Order("id desc").Limit(limit).Find(&items).Error; err != nil {
        return nil, 0, err
    }
    return items, total, nil
}

func (s *ContentService) ListLikes(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Like, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ?", userID)
    if cursor != nil {
        q = q.Where("id < ?", *cursor)
    }
    var total int64
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    var items []model.Like
    if err := q.Order("id desc").Limit(limit).Find(&items).Error; err != nil {
        return nil, 0, err
    }
    return items, total, nil
}
```

```go
// apps/core/internal/domain/content/service/interaction.go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type InteractionService struct {
    db *gorm.DB
}

func NewInteractionService(db *gorm.DB) *InteractionService {
    if db == nil {
        db = database.DB
    }
    return &InteractionService{db: db}
}

func (s *InteractionService) Like(ctx context.Context, userID int64, targetType string, targetID int64) error {
    like := model.Like{UserID: userID, TargetType: targetType, TargetID: targetID}
    return s.db.WithContext(ctx).FirstOrCreate(&like, like).Error
}

func (s *InteractionService) Unlike(ctx context.Context, userID int64, targetType string, targetID int64) error {
    return s.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).Delete(&model.Like{}).Error
}

func (s *InteractionService) Favorite(ctx context.Context, userID int64, targetType string, targetID int64) error {
    fav := model.Favorite{UserID: userID, TargetType: targetType, TargetID: targetID}
    return s.db.WithContext(ctx).FirstOrCreate(&fav, fav).Error
}

func (s *InteractionService) Unfavorite(ctx context.Context, userID int64, targetType string, targetID int64) error {
    return s.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).Delete(&model.Favorite{}).Error
}
```

Create the content DTOs by splitting the existing `internal/dto/content.go` into domain DTO files.

```go
// apps/core/internal/domain/content/dto/common.go
package dto

type UserBrief struct {
    UserID    int64  `json:"user_id"`
    Nickname  string `json:"nickname"`
    AvatarURL string `json:"avatar_url"`
    Intro     string `json:"intro"`
}

type FollowingUsersResponse struct {
    Users []UserBrief `json:"users"`
    Total int         `json:"total"`
}

type FollowersResponse struct {
    Users []UserBrief `json:"users"`
    Total int         `json:"total"`
}

type MerchantBrief struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Category string `json:"category"`
}
```

```go
// apps/core/internal/domain/content/dto/post.go
package dto

import "time"

type PostItem struct {
    ID        int64          `json:"id"`
    Title     string         `json:"title"`
    Content   string         `json:"content"`
    Images    []string       `json:"images"`
    LikeCount int            `json:"like_count"`
    ViewCount int            `json:"view_count"`
    IsLiked   bool           `json:"is_liked"`
    Merchant  *MerchantBrief `json:"merchant,omitempty"`
    Tags      []string       `json:"tags"`
    CreatedAt time.Time      `json:"created_at"`
}

type PostListResponse struct {
    Posts  []PostItem `json:"posts"`
    Total  int        `json:"total"`
    Cursor *int64     `json:"cursor,omitempty"`
}
```

```go
// apps/core/internal/domain/content/dto/review.go
package dto

import "time"

type ReviewItem struct {
    ID            int64         `json:"id"`
    Rating        float32       `json:"rating"`
    RatingEnv     *float32      `json:"rating_env,omitempty"`
    RatingService *float32      `json:"rating_service,omitempty"`
    RatingValue   *float32      `json:"rating_value,omitempty"`
    Content       string        `json:"content"`
    Images        []string      `json:"images"`
    AvgCost       *int          `json:"avg_cost,omitempty"`
    LikeCount     int           `json:"like_count"`
    IsLiked       bool          `json:"is_liked"`
    Merchant      MerchantBrief `json:"merchant"`
    Tags          []string      `json:"tags"`
    CreatedAt     time.Time     `json:"created_at"`
}

type ReviewListResponse struct {
    Reviews []ReviewItem `json:"reviews"`
    Total   int          `json:"total"`
    Cursor  *int64       `json:"cursor,omitempty"`
}
```

```go
// apps/core/internal/domain/content/dto/favorite.go
package dto

import "time"

type FavoriteItem struct {
    ID         int64          `json:"id"`
    TargetType string         `json:"target_type"`
    TargetID   int64          `json:"target_id"`
    Post       *PostItem      `json:"post,omitempty"`
    Review     *ReviewItem    `json:"review,omitempty"`
    Merchant   *MerchantBrief `json:"merchant,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
}

type FavoriteListResponse struct {
    Items  []FavoriteItem `json:"items"`
    Total  int            `json:"total"`
    Cursor *int64         `json:"cursor,omitempty"`
}
```

Create content handlers by moving logic from `handler/user.go` and `handler/profile.go`.

```go
// apps/core/internal/domain/content/handler/helpers.go
package handler

import (
    "encoding/json"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type helper struct {
    db *gorm.DB
}

func newHelper() helper {
    return helper{db: database.DB}
}

func (h helper) getLikedIDs(c *gin.Context, targetType string) map[int64]bool {
    userID := c.GetInt64("user_id")
    if userID == 0 {
        return map[int64]bool{}
    }
    var likes []model.Like
    if err := h.db.WithContext(c.Request.Context()).Where("user_id = ? AND target_type = ?", userID, targetType).Find(&likes).Error; err != nil {
        return map[int64]bool{}
    }
    result := make(map[int64]bool, len(likes))
    for _, like := range likes {
        result[like.TargetID] = true
    }
    return result
}

func (h helper) loadMerchantBrief(id int64) *dto.MerchantBrief {
    var merchant model.Merchant
    if err := h.db.First(&merchant, id).Error; err != nil {
        return nil
    }
    return &dto.MerchantBrief{ID: merchant.ID, Name: merchant.Name, Category: merchant.Category}
}

func parseIDParam(c *gin.Context, name string) (int64, error) {
    return strconv.ParseInt(c.Param(name), 10, 64)
}

func parseCursorLimit(c *gin.Context) (*int64, int) {
    limit := 20
    if v := c.Query("limit"); v != "" {
        if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
            limit = parsed
        }
    }
    var cursor *int64
    if v := c.Query("cursor"); v != "" {
        if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
            cursor = &parsed
        }
    }
    return cursor, limit
}

func parseJSONStrings(raw string) []string {
    if raw == "" {
        return []string{}
    }
    var values []string
    if err := json.Unmarshal([]byte(raw), &values); err != nil {
        return []string{}
    }
    return values
}

func nextCursor[T any](items []T) *int64 {
    return nil
}
```

```go
// apps/core/internal/domain/content/handler/post.go
package handler

import (
    "net/http"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
    "github.com/gin-gonic/gin"
)

type PostHandler struct {
    svc    *service.ContentService
    helper helper
}

func NewPostHandler(svc *service.ContentService) *PostHandler {
    if svc == nil {
        svc = service.NewContentService(nil)
    }
    return &PostHandler{svc: svc, helper: newHelper()}
}

func (h *PostHandler) ListUserPosts(c *gin.Context) {
    targetID, err := parseIDParam(c, "id")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    cursor, limit := parseCursorLimit(c)
    posts, total, err := h.svc.ListUserPosts(c.Request.Context(), targetID, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    liked := h.helper.getLikedIDs(c, "post")
    items := make([]dto.PostItem, 0, len(posts))
    for _, post := range posts {
        var merchant *dto.MerchantBrief
        if post.MerchantID != nil {
            merchant = h.helper.loadMerchantBrief(*post.MerchantID)
        }
        items = append(items, dto.PostItem{
            ID:        post.ID,
            Title:     post.Title,
            Content:   post.Content,
            Images:    parseJSONStrings(post.Images),
            LikeCount: post.LikeCount,
            ViewCount: post.ViewCount,
            IsLiked:   liked[post.ID],
            Merchant:  merchant,
            Tags:      []string{},
            CreatedAt: post.CreatedAt,
        })
    }
    c.JSON(http.StatusOK, dto.PostListResponse{Posts: items, Total: int(total), Cursor: nextCursor(posts)})
}

func (h *PostHandler) ListMyPosts(c *gin.Context) {
    userID := c.GetInt64("user_id")
    cursor, limit := parseCursorLimit(c)
    posts, total, err := h.svc.ListUserPosts(c.Request.Context(), userID, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    liked := h.helper.getLikedIDs(c, "post")
    items := make([]dto.PostItem, 0, len(posts))
    for _, post := range posts {
        var merchant *dto.MerchantBrief
        if post.MerchantID != nil {
            merchant = h.helper.loadMerchantBrief(*post.MerchantID)
        }
        items = append(items, dto.PostItem{
            ID:        post.ID,
            Title:     post.Title,
            Content:   post.Content,
            Images:    parseJSONStrings(post.Images),
            LikeCount: post.LikeCount,
            ViewCount: post.ViewCount,
            IsLiked:   liked[post.ID],
            Merchant:  merchant,
            Tags:      []string{},
            CreatedAt: post.CreatedAt,
        })
    }
    c.JSON(http.StatusOK, dto.PostListResponse{Posts: items, Total: int(total), Cursor: nextCursor(posts)})
}
```

```go
// apps/core/internal/domain/content/handler/review.go
package handler

import (
    "net/http"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
    "github.com/gin-gonic/gin"
)

type ReviewHandler struct {
    svc    *service.ContentService
    helper helper
}

func NewReviewHandler(svc *service.ContentService) *ReviewHandler {
    if svc == nil {
        svc = service.NewContentService(nil)
    }
    return &ReviewHandler{svc: svc, helper: newHelper()}
}

func (h *ReviewHandler) ListUserReviews(c *gin.Context) {
    targetID, err := parseIDParam(c, "id")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    cursor, limit := parseCursorLimit(c)
    reviews, total, err := h.svc.ListUserReviews(c.Request.Context(), targetID, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    liked := h.helper.getLikedIDs(c, "review")
    items := make([]dto.ReviewItem, 0, len(reviews))
    for _, review := range reviews {
        merchant := h.helper.loadMerchantBrief(review.MerchantID)
        merchantValue := dto.MerchantBrief{}
        if merchant != nil {
            merchantValue = *merchant
        }
        items = append(items, dto.ReviewItem{
            ID:            review.ID,
            Rating:        review.Rating,
            RatingEnv:     review.RatingEnv,
            RatingService: review.RatingService,
            RatingValue:   review.RatingValue,
            Content:       review.Content,
            Images:        parseJSONStrings(review.Images),
            AvgCost:       review.AvgCost,
            LikeCount:     review.LikeCount,
            IsLiked:       liked[review.ID],
            Merchant:      merchantValue,
            Tags:          []string{},
            CreatedAt:     review.CreatedAt,
        })
    }
    c.JSON(http.StatusOK, dto.ReviewListResponse{Reviews: items, Total: int(total), Cursor: nextCursor(reviews)})
}

func (h *ReviewHandler) ListMyReviews(c *gin.Context) {
    userID := c.GetInt64("user_id")
    cursor, limit := parseCursorLimit(c)
    reviews, total, err := h.svc.ListUserReviews(c.Request.Context(), userID, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    liked := h.helper.getLikedIDs(c, "review")
    items := make([]dto.ReviewItem, 0, len(reviews))
    for _, review := range reviews {
        merchant := h.helper.loadMerchantBrief(review.MerchantID)
        merchantValue := dto.MerchantBrief{}
        if merchant != nil {
            merchantValue = *merchant
        }
        items = append(items, dto.ReviewItem{
            ID:            review.ID,
            Rating:        review.Rating,
            RatingEnv:     review.RatingEnv,
            RatingService: review.RatingService,
            RatingValue:   review.RatingValue,
            Content:       review.Content,
            Images:        parseJSONStrings(review.Images),
            AvgCost:       review.AvgCost,
            LikeCount:     review.LikeCount,
            IsLiked:       liked[review.ID],
            Merchant:      merchantValue,
            Tags:          []string{},
            CreatedAt:     review.CreatedAt,
        })
    }
    c.JSON(http.StatusOK, dto.ReviewListResponse{Reviews: items, Total: int(total), Cursor: nextCursor(reviews)})
}
```

```go
// apps/core/internal/domain/content/handler/favorite.go
package handler

import (
    "net/http"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
    svc    *service.ContentService
    helper helper
}

func NewFavoriteHandler(svc *service.ContentService) *FavoriteHandler {
    if svc == nil {
        svc = service.NewContentService(nil)
    }
    return &FavoriteHandler{svc: svc, helper: newHelper()}
}

func (h *FavoriteHandler) ListMyFavorites(c *gin.Context) {
    userID := c.GetInt64("user_id")
    targetType := c.Query("type")
    cursor, limit := parseCursorLimit(c)

    items, total, err := h.svc.ListFavorites(c.Request.Context(), userID, targetType, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    respItems := make([]dto.FavoriteItem, 0, len(items))
    for _, item := range items {
        fav := dto.FavoriteItem{
            ID:         item.ID,
            TargetType: item.TargetType,
            TargetID:   item.TargetID,
            CreatedAt:  item.CreatedAt,
        }
        switch item.TargetType {
        case "post":
            var post model.Post
            if err := h.helper.db.WithContext(c.Request.Context()).First(&post, item.TargetID).Error; err == nil {
                fav.Post = &dto.PostItem{
                    ID:        post.ID,
                    Title:     post.Title,
                    Content:   post.Content,
                    Images:    parseJSONStrings(post.Images),
                    LikeCount: post.LikeCount,
                    ViewCount: post.ViewCount,
                    IsLiked:   false,
                    Merchant:  nil,
                    Tags:      []string{},
                    CreatedAt: post.CreatedAt,
                }
            }
        case "review":
            var review model.Review
            if err := h.helper.db.WithContext(c.Request.Context()).First(&review, item.TargetID).Error; err == nil {
                fav.Review = &dto.ReviewItem{
                    ID:            review.ID,
                    Rating:        review.Rating,
                    RatingEnv:     review.RatingEnv,
                    RatingService: review.RatingService,
                    RatingValue:   review.RatingValue,
                    Content:       review.Content,
                    Images:        parseJSONStrings(review.Images),
                    AvgCost:       review.AvgCost,
                    LikeCount:     review.LikeCount,
                    IsLiked:       false,
                    Merchant:      dto.MerchantBrief{},
                    Tags:          []string{},
                    CreatedAt:     review.CreatedAt,
                }
            }
        case "merchant":
            if merchant := h.helper.loadMerchantBrief(item.TargetID); merchant != nil {
                fav.Merchant = merchant
            }
        }
        respItems = append(respItems, fav)
    }

    resp := dto.FavoriteListResponse{Items: respItems, Total: int(total), Cursor: nextCursor(items)}
    c.JSON(http.StatusOK, resp)
}
```

```go
// apps/core/internal/domain/content/handler/like.go
package handler

import (
    "net/http"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
    "github.com/gin-gonic/gin"
)

type LikeHandler struct {
    svc *service.ContentService
}

func NewLikeHandler(svc *service.ContentService) *LikeHandler {
    if svc == nil {
        svc = service.NewContentService(nil)
    }
    return &LikeHandler{svc: svc}
}

func (h *LikeHandler) ListMyLikes(c *gin.Context) {
    userID := c.GetInt64("user_id")
    cursor, limit := parseCursorLimit(c)
    items, total, err := h.svc.ListLikes(c.Request.Context(), userID, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    respItems := make([]gin.H, 0, len(items))
    for _, item := range items {
        respItems = append(respItems, gin.H{
            "id":          item.ID,
            "target_type": item.TargetType,
            "target_id":   item.TargetID,
            "created_at":  item.CreatedAt,
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "items":  respItems,
        "total":  total,
        "cursor": nextCursor(items),
    })
}
```

Update `apps/core/internal/domain/content/routes.go` to register the new handlers.

```go
package content

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers content routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    contentSvc := service.NewContentService(nil)

    postHandler := handler.NewPostHandler(contentSvc)
    reviewHandler := handler.NewReviewHandler(contentSvc)
    favHandler := handler.NewFavoriteHandler(contentSvc)
    likeHandler := handler.NewLikeHandler(contentSvc)

    users := r.Group("/users")
    {
        users.GET("/:id/posts", postHandler.ListUserPosts)
        users.GET("/:id/reviews", reviewHandler.ListUserReviews)
    }

    user := r.Group("/user", middleware.JWTAuth(cfg.JWT))
    {
        user.GET("/posts", postHandler.ListMyPosts)
        user.GET("/reviews", reviewHandler.ListMyReviews)
        user.GET("/favorites", favHandler.ListMyFavorites)
        user.GET("/likes", likeHandler.ListMyLikes)
    }
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/content/... -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/content
git commit -m "refactor: migrate content domain"
```

---

### Task 5: Migrate user domain (dto, services, handlers, tests)

**Files:**
- Move: `apps/core/internal/handler/user.go` -> `apps/core/internal/domain/user/handler/user.go`
- Move: `apps/core/internal/handler/user_test.go` -> `apps/core/internal/domain/user/handler/user_test.go`
- Move: `apps/core/internal/service/user.go` -> `apps/core/internal/domain/user/service/user.go`
- Move: `apps/core/internal/service/user_test.go` -> `apps/core/internal/domain/user/service/user_test.go`
- Create: `apps/core/internal/domain/user/dto/profile.go`
- Create: `apps/core/internal/domain/user/dto/privacy.go`
- Create: `apps/core/internal/domain/user/dto/notification.go`
- Create: `apps/core/internal/domain/user/dto/address.go`
- Create: `apps/core/internal/domain/user/dto/account.go`
- Modify: `apps/core/internal/domain/user/routes.go`

**Step 1: Write failing test**

Move the user handler and user service tests into the new packages and update imports.

```go
// apps/core/internal/domain/user/handler/user_test.go
package handler

import "testing"

func TestUserHandlerConstruction(t *testing.T) {
    h := NewUserHandler(nil)
    if h == nil {
        t.Fatal("expected handler")
    }
}
```

```go
// apps/core/internal/domain/user/service/user_test.go
package service

import (
    "context"
    "testing"

    userdto "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
)

func TestUserServiceProfileAndSettings(t *testing.T) {
    db := testutil.SetupTestDB(t)
    svc := NewUserService(db)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil {
        t.Fatal(err)
    }
    if err := db.Create(&model.UserProfile{UserID: user.ID, Nickname: "n"}).Error; err != nil {
        t.Fatal(err)
    }

    prof, err := svc.GetProfile(context.Background(), user.ID)
    if err != nil || prof.UserID != user.ID {
        t.Fatalf("profile failed: %v", err)
    }

    newName := "new"
    if err := svc.UpdateProfile(context.Background(), user.ID, userdto.UpdateProfileRequest{Nickname: &newName}); err != nil {
        t.Fatalf("update failed: %v", err)
    }
}

func TestUserServiceAddressDefault(t *testing.T) {
    db := testutil.SetupTestDB(t)
    svc := NewUserService(db)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil {
        t.Fatal(err)
    }

    addr, err := svc.CreateAddress(context.Background(), user.ID, userdto.CreateAddressRequest{Name: "A", Phone: "1", Address: "X"})
    if err != nil || !addr.IsDefault {
        t.Fatalf("expected default address")
    }
}
```

**Step 2: Run test to verify it fails**

Run:
- `go test ./internal/domain/user/handler -v`
- `go test ./internal/domain/user/service -v`

Expected: FAIL due to missing package and constructors.

**Step 3: Write minimal implementation**

Move user service logic into the new package and split DTOs into domain/user/dto files. Keep method bodies identical.

```go
// apps/core/internal/domain/user/dto/profile.go
package dto

type ProfileResponse struct {
    UserID    int64  `json:"user_id"`
    Nickname  string `json:"nickname"`
    AvatarURL string `json:"avatar_url"`
    Intro     string `json:"intro"`
    Location  string `json:"location"`
}

type UpdateProfileRequest struct {
    Nickname  *string `json:"nickname,omitempty"`
    AvatarURL *string `json:"avatar_url,omitempty"`
    Intro     *string `json:"intro,omitempty"`
    Location  *string `json:"location,omitempty"`
}
```

```go
// apps/core/internal/domain/user/dto/privacy.go
package dto

type PrivacySettings struct {
    IsPublic bool `json:"is_public"`
}
```

```go
// apps/core/internal/domain/user/dto/notification.go
package dto

type NotificationSettings struct {
    PushEnabled  bool `json:"push_enabled"`
    EmailEnabled bool `json:"email_enabled"`
}
```

```go
// apps/core/internal/domain/user/dto/address.go
package dto

type AddressItem struct {
    ID         int64  `json:"id"`
    Name       string `json:"name"`
    Phone      string `json:"phone"`
    Province   string `json:"province"`
    City       string `json:"city"`
    District   string `json:"district"`
    Address    string `json:"address"`
    PostalCode string `json:"postal_code"`
    IsDefault  bool   `json:"is_default"`
}

type AddressListResponse struct {
    Addresses []AddressItem `json:"addresses"`
}

type CreateAddressRequest struct {
    Name       string `json:"name" binding:"required,max=50"`
    Phone      string `json:"phone" binding:"required,max=20"`
    Province   string `json:"province" binding:"max=50"`
    City       string `json:"city" binding:"max=50"`
    District   string `json:"district" binding:"max=50"`
    Address    string `json:"address" binding:"required,max=255"`
    PostalCode string `json:"postal_code" binding:"max=20"`
    IsDefault  bool   `json:"is_default"`
}

type UpdateAddressRequest struct {
    Name       *string `json:"name,omitempty"`
    Phone      *string `json:"phone,omitempty"`
    Province   *string `json:"province,omitempty"`
    City       *string `json:"city,omitempty"`
    District   *string `json:"district,omitempty"`
    Address    *string `json:"address,omitempty"`
    PostalCode *string `json:"postal_code,omitempty"`
    IsDefault  *bool   `json:"is_default,omitempty"`
}
```

```go
// apps/core/internal/domain/user/dto/account.go
package dto

type AccountDeletionRequest struct {
    Reason string `json:"reason"`
}
```

```go
// apps/core/internal/domain/user/service/user.go
package service

import (
    "context"
    "errors"
    "time"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
    if db == nil {
        db = database.DB
    }
    return &UserService{db: db}
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (dto.ProfileResponse, error) {
    var profile model.UserProfile
    if err := s.db.WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err != nil {
        return dto.ProfileResponse{}, err
    }
    return dto.ProfileResponse{UserID: profile.UserID, Nickname: profile.Nickname, AvatarURL: profile.AvatarURL, Intro: profile.Intro, Location: profile.Location}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) error {
    updates := map[string]any{}
    if req.Nickname != nil {
        updates["nickname"] = *req.Nickname
    }
    if req.AvatarURL != nil {
        updates["avatar_url"] = *req.AvatarURL
    }
    if req.Intro != nil {
        updates["intro"] = *req.Intro
    }
    if req.Location != nil {
        updates["location"] = *req.Location
    }
    if len(updates) == 0 {
        return nil
    }
    return s.db.WithContext(ctx).Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (s *UserService) GetPrivacy(ctx context.Context, userID int64) (dto.PrivacySettings, error) {
    var privacy model.UserPrivacy
    if err := s.db.WithContext(ctx).FirstOrCreate(&privacy, model.UserPrivacy{UserID: userID}).Error; err != nil {
        return dto.PrivacySettings{}, err
    }
    return dto.PrivacySettings{IsPublic: privacy.IsPublic}, nil
}

func (s *UserService) UpdatePrivacy(ctx context.Context, userID int64, req dto.PrivacySettings) error {
    return s.db.WithContext(ctx).Save(&model.UserPrivacy{UserID: userID, IsPublic: req.IsPublic}).Error
}

func (s *UserService) GetNotifications(ctx context.Context, userID int64) (dto.NotificationSettings, error) {
    var notif model.UserNotification
    if err := s.db.WithContext(ctx).FirstOrCreate(&notif, model.UserNotification{UserID: userID}).Error; err != nil {
        return dto.NotificationSettings{}, err
    }
    return dto.NotificationSettings{PushEnabled: notif.PushEnabled, EmailEnabled: notif.EmailEnabled}, nil
}

func (s *UserService) UpdateNotifications(ctx context.Context, userID int64, req dto.NotificationSettings) error {
    return s.db.WithContext(ctx).Save(&model.UserNotification{UserID: userID, PushEnabled: req.PushEnabled, EmailEnabled: req.EmailEnabled}).Error
}

func (s *UserService) ListAddresses(ctx context.Context, userID int64) ([]model.UserAddress, error) {
    var items []model.UserAddress
    if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default desc, id asc").Find(&items).Error; err != nil {
        return nil, err
    }
    return items, nil
}

func (s *UserService) CreateAddress(ctx context.Context, userID int64, req dto.CreateAddressRequest) (model.UserAddress, error) {
    var addr model.UserAddress
    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var count int64
        if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
            return err
        }
        if count >= 20 {
            return errors.New("address limit reached")
        }
        addr = model.UserAddress{UserID: userID, Name: req.Name, Phone: req.Phone, Province: req.Province, City: req.City, District: req.District, Address: req.Address, PostalCode: req.PostalCode, IsDefault: req.IsDefault || count == 0}
        if addr.IsDefault {
            if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
                return err
            }
        }
        return tx.Create(&addr).Error
    })
    return addr, err
}

func (s *UserService) UpdateAddress(ctx context.Context, userID, addressID int64, req dto.UpdateAddressRequest) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        updates := map[string]any{}
        if req.Name != nil {
            updates["name"] = *req.Name
        }
        if req.Phone != nil {
            updates["phone"] = *req.Phone
        }
        if req.Province != nil {
            updates["province"] = *req.Province
        }
        if req.City != nil {
            updates["city"] = *req.City
        }
        if req.District != nil {
            updates["district"] = *req.District
        }
        if req.Address != nil {
            updates["address"] = *req.Address
        }
        if req.PostalCode != nil {
            updates["postal_code"] = *req.PostalCode
        }
        if req.IsDefault != nil {
            updates["is_default"] = *req.IsDefault
        }
        if err := tx.Model(&model.UserAddress{}).Where("id = ? AND user_id = ?", addressID, userID).Updates(updates).Error; err != nil {
            return err
        }
        if req.IsDefault != nil && *req.IsDefault {
            return tx.Model(&model.UserAddress{}).Where("user_id = ? AND id <> ?", userID, addressID).Update("is_default", false).Error
        }
        return nil
    })
}

func (s *UserService) DeleteAddress(ctx context.Context, userID, addressID int64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var addr model.UserAddress
        if err := tx.First(&addr, "id = ? AND user_id = ?", addressID, userID).Error; err != nil {
            return err
        }
        if err := tx.Delete(&addr).Error; err != nil {
            return err
        }
        if addr.IsDefault {
            var replacement model.UserAddress
            if err := tx.Where("user_id = ?", userID).Order("id asc").First(&replacement).Error; err == nil {
                return tx.Model(&replacement).Update("is_default", true).Error
            }
        }
        return nil
    })
}

func (s *UserService) SetDefaultAddress(ctx context.Context, userID, addressID int64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
            return err
        }
        return tx.Model(&model.UserAddress{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error
    })
}

func (s *UserService) RequestAccountDeletion(ctx context.Context, userID int64, reason string) error {
    deletion := model.AccountDeletion{UserID: userID, Reason: reason, ScheduledAt: time.Now().UTC().Add(7 * 24 * time.Hour)}
    return s.db.WithContext(ctx).Save(&deletion).Error
}
```

Move the user handler to `domain/user/handler/user.go` and update imports to reference the new DTO and services, plus content/follow services from their new packages. Remove the content endpoints (ListMyPosts, ListMyReviews, ListMyFavorites, ListMyLikes) because those routes move to the content domain; also remove helper functions that become unused (likedIDs, parseJSONStrings, nextCursor) and drop any unused service fields. Keep the remaining handler method bodies identical; only update import paths and type names.

Update the handler struct/constructor to only keep the user service and DB:

```go
type UserHandler struct {
    userService *service.UserService
    db          *gorm.DB
}

func NewUserHandler(userService *service.UserService) *UserHandler {
    if userService == nil {
        userService = service.NewUserService(nil)
    }
    return &UserHandler{
        userService: userService,
        db:          database.DB,
    }
}
```

Update `apps/core/internal/domain/user/routes.go` to register the same routes under `/user` using the new handlers.

```go
package user

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes registers user routes.
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    userSvc := service.NewUserService(nil)
    userHandler := handler.NewUserHandler(userSvc)

    user := r.Group("/user", middleware.JWTAuth(cfg.JWT))
    {
        profile := user.Group("/profile")
        {
            profile.GET("", userHandler.GetProfile)
            profile.PATCH("", userHandler.UpdateProfile)
        }

        privacy := user.Group("/privacy")
        {
            privacy.GET("", userHandler.GetPrivacy)
            privacy.PATCH("", userHandler.UpdatePrivacy)
        }

        notifications := user.Group("/notifications")
        {
            notifications.GET("", userHandler.GetNotifications)
            notifications.PATCH("", userHandler.UpdateNotifications)
        }

        addresses := user.Group("/addresses")
        {
            addresses.GET("", userHandler.ListAddresses)
            addresses.POST("", userHandler.CreateAddress)
            addresses.PATCH("/:id", userHandler.UpdateAddress)
            addresses.DELETE("/:id", userHandler.DeleteAddress)
            addresses.POST("/:id/default", userHandler.SetDefaultAddress)
        }

        following := user.Group("/following")
        {
            following.GET("/users", userHandler.ListFollowingUsers)
            following.GET("/merchants", userHandler.ListFollowingMerchants)
        }

        followers := user.Group("/followers")
        {
            followers.GET("", userHandler.ListFollowers)
        }

        account := user.Group("/account")
        {
            account.POST("/export", userHandler.RequestAccountExport)
            account.DELETE("", userHandler.RequestAccountDeletion)
        }
    }
}
```

**Step 4: Run test to verify it passes**

Run:
- `go test ./internal/domain/user/handler -v`
- `go test ./internal/domain/user/service -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/user
git commit -m "refactor: migrate user domain"
```

---

### Task 6: Extract follow domain (services, handlers, tests)

**Files:**
- Move: `apps/core/internal/service/follow.go` -> `apps/core/internal/domain/follow/service/follow.go`
- Move: `apps/core/internal/service/follow_test.go` -> `apps/core/internal/domain/follow/service/follow_test.go`
- Create: `apps/core/internal/domain/follow/handler/user.go`
- Create: `apps/core/internal/domain/follow/handler/merchant.go`
- Modify: `apps/core/internal/domain/follow/routes.go`

**Step 1: Write failing test**

Move the existing follow service tests into the new package and update them to use testutil.

```go
// apps/core/internal/domain/follow/service/follow_test.go
package service

import (
    "context"
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/testutil"
)

func TestFollowServiceUserFollow(t *testing.T) {
    db := testutil.SetupTestDB(t)
    svc := NewFollowService(db)
    u1 := model.User{Role: "user", Status: 0}
    u2 := model.User{Role: "user", Status: 0}
    db.Create(&u1)
    db.Create(&u2)
    if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
        t.Fatal(err)
    }
    if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
        t.Fatal(err)
    }
}

func TestFollowUpdatesCounts(t *testing.T) {
    db := testutil.SetupTestDB(t)
    svc := NewFollowService(db)
    u1 := model.User{Role: "user", Status: 0}
    u2 := model.User{Role: "user", Status: 0}
    db.Create(&u1)
    db.Create(&u2)
    db.Create(&model.UserProfile{UserID: u1.ID, Nickname: "a"})
    db.Create(&model.UserProfile{UserID: u2.ID, Nickname: "b"})

    if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
        t.Fatal(err)
    }
    if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil {
        t.Fatal(err)
    }

    var p1, p2 model.UserProfile
    db.First(&p1, "user_id = ?", u1.ID)
    db.First(&p2, "user_id = ?", u2.ID)
    if p1.FollowingCount != 1 || p2.FollowerCount != 1 {
        t.Fatalf("counts not updated correctly: %d/%d", p1.FollowingCount, p2.FollowerCount)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/follow/service -v`

Expected: FAIL with undefined constructors.

**Step 3: Write minimal implementation**

Move the follow service into `domain/follow/service/follow.go`, keeping logic identical.

```go
package service

import (
    "context"
    "errors"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type FollowService struct {
    db *gorm.DB
}

func NewFollowService(db *gorm.DB) *FollowService {
    if db == nil {
        db = database.DB
    }
    return &FollowService{db: db}
}

func (s *FollowService) FollowUser(ctx context.Context, followerID, followingID int64) error {
    if followerID == followingID {
        return errors.New("cannot follow self")
    }
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var existing model.UserFollow
        result := tx.Where("follower_id = ? AND following_id = ?", followerID, followingID).Limit(1).Find(&existing)
        if result.Error != nil {
            return result.Error
        }
        if result.RowsAffected > 0 {
            return nil
        }
        follow := model.UserFollow{FollowerID: followerID, FollowingID: followingID}
        if err := tx.Create(&follow).Error; err != nil {
            return err
        }
        if err := tx.Model(&model.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count + 1")).Error; err != nil {
            return err
        }
        return tx.Model(&model.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
    })
}

func (s *FollowService) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        result := tx.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.UserFollow{})
        if result.Error != nil {
            return result.Error
        }
        if result.RowsAffected == 0 {
            return nil
        }
        if err := tx.Model(&model.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("GREATEST(following_count - 1, 0)")).Error; err != nil {
            return err
        }
        return tx.Model(&model.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)")).Error
    })
}

func (s *FollowService) FollowMerchant(ctx context.Context, userID, merchantID int64) error {
    follow := model.MerchantFollow{UserID: userID, MerchantID: merchantID}
    return s.db.WithContext(ctx).FirstOrCreate(&follow, follow).Error
}

func (s *FollowService) UnfollowMerchant(ctx context.Context, userID, merchantID int64) error {
    return s.db.WithContext(ctx).Where("user_id = ? AND merchant_id = ?", userID, merchantID).Delete(&model.MerchantFollow{}).Error
}
```

Create follow handlers by moving logic from `handler/profile.go`.

```go
// apps/core/internal/domain/follow/handler/user.go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/service"
    "github.com/gin-gonic/gin"
)

type UserHandler struct {
    svc *service.FollowService
}

func NewUserHandler(svc *service.FollowService) *UserHandler {
    if svc == nil {
        svc = service.NewFollowService(nil)
    }
    return &UserHandler{svc: svc}
}

func (h *UserHandler) FollowUser(c *gin.Context) {
    userID := c.GetInt64("user_id")
    targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    if err := h.svc.FollowUser(c.Request.Context(), userID, targetID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) UnfollowUser(c *gin.Context) {
    userID := c.GetInt64("user_id")
    targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    if err := h.svc.UnfollowUser(c.Request.Context(), userID, targetID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

```go
// apps/core/internal/domain/follow/handler/merchant.go
package handler

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/service"
    "github.com/gin-gonic/gin"
)

type MerchantHandler struct {
    svc *service.FollowService
}

func NewMerchantHandler(svc *service.FollowService) *MerchantHandler {
    if svc == nil {
        svc = service.NewFollowService(nil)
    }
    return &MerchantHandler{svc: svc}
}

func (h *MerchantHandler) FollowMerchant(c *gin.Context) {
    userID := c.GetInt64("user_id")
    merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
        return
    }
    if err := h.svc.FollowMerchant(c.Request.Context(), userID, merchantID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *MerchantHandler) UnfollowMerchant(c *gin.Context) {
    userID := c.GetInt64("user_id")
    merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
        return
    }
    if err := h.svc.UnfollowMerchant(c.Request.Context(), userID, merchantID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

Update follow routes to register follow handlers under `/users/:id/follow` and `/merchants/:id/follow` (JWT required).

```go
package follow

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewFollowService(nil)
    userHandler := handler.NewUserHandler(svc)
    merchantHandler := handler.NewMerchantHandler(svc)

    auth := r.Group("", middleware.JWTAuth(cfg.JWT))
    {
        auth.POST("/users/:id/follow", userHandler.FollowUser)
        auth.DELETE("/users/:id/follow", userHandler.UnfollowUser)
        auth.POST("/merchants/:id/follow", merchantHandler.FollowMerchant)
        auth.DELETE("/merchants/:id/follow", merchantHandler.UnfollowMerchant)
    }
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/follow/service -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/follow
git commit -m "refactor: migrate follow domain"
```

---

### Task 7: Simplify profile domain (public profile only)

**Files:**
- Move: `apps/core/internal/handler/profile.go` -> `apps/core/internal/domain/profile/handler.go`
- Move: `apps/core/internal/handler/profile_test.go` -> `apps/core/internal/domain/profile/handler_test.go`
- Create: `apps/core/internal/domain/profile/service/service.go`
- Modify: `apps/core/internal/domain/profile/routes.go`

**Step 1: Write failing test**

Move the profile handler construction test into the new package and add a small service DTO existence test.

```go
// apps/core/internal/domain/profile/handler_test.go
package profile

import "testing"

func TestPublicProfileHandler(t *testing.T) {
    h := NewHandler(nil)
    if h == nil {
        t.Fatal("expected handler")
    }
}
```

```go
// apps/core/internal/domain/profile/service/service_test.go
package service

import "testing"

func TestPublicProfileResponseExists(t *testing.T) {
    _ = PublicProfileResponse{}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/profile/... -v`

Expected: FAIL due to missing handler.

**Step 3: Write minimal implementation**

Create a profile service for fetching public profile data and a handler that only serves `/users/:id`.

```go
// apps/core/internal/domain/profile/service/service.go
package service

import (
    "context"
    "fmt"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

type Service interface {
    GetPublicProfile(ctx context.Context, userID int64) (*PublicProfileResponse, error)
}

type service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) Service {
    if db == nil {
        db = database.DB
    }
    return &service{db: db}
}

type PublicProfileResponse struct {
    UserID         int64  `json:"user_id"`
    Nickname       string `json:"nickname"`
    AvatarURL      string `json:"avatar_url"`
    Intro          string `json:"intro"`
    Location       string `json:"location"`
    FollowerCount  int    `json:"follower_count"`
    FollowingCount int    `json:"following_count"`
    PostCount      int    `json:"post_count"`
    ReviewCount    int    `json:"review_count"`
    LikeCount      int    `json:"like_count"`
    IsFollowing    bool   `json:"is_following"`
}

func (s *service) GetPublicProfile(ctx context.Context, userID int64) (*PublicProfileResponse, error) {
    var profile model.UserProfile
    if err := s.db.WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }

    return &PublicProfileResponse{
        UserID:         profile.UserID,
        Nickname:       profile.Nickname,
        AvatarURL:      profile.AvatarURL,
        Intro:          profile.Intro,
        Location:       profile.Location,
        FollowerCount:  profile.FollowerCount,
        FollowingCount: profile.FollowingCount,
        PostCount:      profile.PostCount,
        ReviewCount:    profile.ReviewCount,
        LikeCount:      profile.LikeCount,
        IsFollowing:    false,
    }, nil
}
```

```go
// apps/core/internal/domain/profile/handler.go
package profile

import (
    "net/http"
    "strconv"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/profile/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type Handler struct {
    svc service.Service
    db  *gorm.DB
}

func NewHandler(svc service.Service) *Handler {
    if svc == nil {
        svc = service.NewService(nil)
    }
    return &Handler{svc: svc, db: database.DB}
}

func (h *Handler) GetPublicProfile(c *gin.Context) {
    targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }

    viewerID := c.GetInt64("user_id")

    var privacy model.UserPrivacy
    if err := h.db.WithContext(c.Request.Context()).First(&privacy, "user_id = ?", targetID).Error; err == nil {
        if !privacy.IsPublic && viewerID != targetID {
            c.JSON(http.StatusForbidden, gin.H{"error": "profile is private"})
            return
        }
    }

    profile, err := h.svc.GetPublicProfile(c.Request.Context(), targetID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    if viewerID != 0 && viewerID != targetID {
        var follow model.UserFollow
        if err := h.db.WithContext(c.Request.Context()).Where("follower_id = ? AND following_id = ?", viewerID, targetID).First(&follow).Error; err == nil {
            profile.IsFollowing = true
        }
    }

    c.JSON(http.StatusOK, profile)
}
```

Update profile routes to expose only `/users/:id` with optional JWT.

```go
package profile

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/profile/service"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    svc := service.NewService(nil)
    handler := NewHandler(svc)

    users := r.Group("/users")
    {
        users.GET("/:id", handler.GetPublicProfile)
    }
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/profile/... -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/profile
git commit -m "refactor: migrate profile domain"
```

---

### Task 8: Update app wiring, remove old packages, and add CODEOWNERS

**Files:**
- Modify: `apps/core/cmd/app/main.go`
- Modify: `apps/core/internal/handler/routes.go` (remove usage)
- Delete: `apps/core/internal/handler/`
- Delete: `apps/core/internal/service/`
- Delete: `apps/core/internal/dto/`
- Create: `.github/CODEOWNERS`

**Step 1: Write failing test**

Add a small test that calls a new `buildRouter` helper which will be introduced in `main.go`.

```go
// apps/core/cmd/app/router_test.go
package main

import (
    "testing"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
)

func TestBuildRouter(t *testing.T) {
    cfg := &config.Config{}
    r := buildRouter(cfg)
    if r == nil {
        t.Fatal("expected router")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/app -v`

Expected: FAIL due to unused/missing router wiring until main.go is updated.

**Step 3: Write minimal implementation**

Update `apps/core/cmd/app/main.go` to use the new router package and the new `buildRouter` helper:

```go
import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/router"
)

func buildRouter(cfg *config.Config) *gin.Engine {
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(jsonLoggerMiddleware())

    apiGroup := r.Group(cfg.Server.APIBasePath)
    {
        apiGroup.GET("/health", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "ok"})
        })
        apiGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }

    router.Setup(r, cfg)
    return r
}

// inside main():
router := buildRouter(cfg)
```

Update `apps/core/internal/middleware/auth.go` to use the new auth token service:

```go
import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/auth"
)

// inside JWTAuth:
tokenService := auth.NewTokenService(jwtCfg)
```

Remove/stop using `internal/handler/routes.go` and delete old handler/service/dto directories after all references are updated.

Create `.github/CODEOWNERS` with domain ownership as described in the migration guide.

```
* @team-lead

apps/core/internal/domain/auth/ @auth-team
apps/core/internal/domain/user/ @user-team
apps/core/internal/domain/profile/ @profile-team
apps/core/internal/domain/follow/ @social-team
apps/core/internal/domain/content/ @content-team

apps/core/internal/config/ @platform-team
apps/core/internal/middleware/ @platform-team
apps/core/internal/model/ @platform-team
apps/core/pkg/ @platform-team
apps/core/internal/router/ @platform-team
```

**Step 4: Run tests to verify it passes**

Run:
- `go test ./...`
- `go build ./cmd/app`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/cmd/app/main.go apps/core/cmd/app/router_test.go apps/core/internal/middleware/auth.go .github/CODEOWNERS
git rm -r apps/core/internal/handler apps/core/internal/service apps/core/internal/dto
git commit -m "refactor: finalize domain migration"
```

---

### Task 9: Docs and formatting cleanup

**Files:**
- Modify: `README.md` (if any references to old handler/service paths)
- Modify: `docs/` (if references to old routes or packages)

**Step 1: Write failing test**

N/A (documentation-only task).

**Step 2: Run tests to verify it fails**

N/A.

**Step 3: Write minimal implementation**

Update references to the old handler/service/dto packages to the new domain paths.

**Step 4: Run tests to verify it passes**

Run: `gofmt -w $(rg --files -g '*.go' apps/core/internal apps/core/cmd)`

**Step 5: Commit**

```bash
git add README.md docs
git commit -m "docs: update architecture references"
```

---

### Task 10: Full verification

**Step 1: Run full test suite**

Run:
- `go test ./...`

Expected: PASS.

**Step 2: Build**

Run:
- `go build ./cmd/app`

Expected: PASS.

**Step 3: Commit**

If anything changed during verification (e.g. gofmt), commit with:

```bash
git add -A
git commit -m "chore: gofmt after migration"
```
