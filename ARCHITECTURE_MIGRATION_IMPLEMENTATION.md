# Architecture Migration Implementation Guide

## Overview

本文档是《Architecture Migration Plan》的详细实现指南，面向高级程序员，提供可执行的迁移步骤、代码示例和最佳实践。

**目标读者**: Senior Backend Engineers  
**预计阅读时间**: 30 分钟  
**前置要求**: 已阅读 `ARCHITECTURE_MIGRATION_PLAN.md`

---

## Table of Contents

1. [Pre-Migration Setup](#1-pre-migration-setup)
2. [Phase 1: Auth Domain Migration](#2-phase-1-auth-domain-migration)
3. [Phase 2: Content Domain Migration](#3-phase-2-content-domain-migration)
4. [Phase 3: User Domain Refactor](#4-phase-3-user-domain-refactor)
5. [Phase 4: Follow Domain Extraction](#5-phase-4-follow-domain-extraction)
6. [Phase 5: Profile Domain Simplification](#6-phase-5-profile-domain-simplification)
7. [Phase 6: Cleanup & Finalization](#7-phase-6-cleanup--finalization)
8. [Testing Strategy](#8-testing-strategy)
9. [Common Pitfalls & Solutions](#9-common-pitfalls--solutions)
10. [Post-Migration Checklist](#10-post-migration-checklist)

---

## 1. Pre-Migration Setup

### 1.1 创建 Feature Branch

```bash
# 从最新的 main 分支创建
git checkout main
git pull origin main
git checkout -b refactor/domain-driven-architecture
```

### 1.2 创建基础目录结构

```bash
cd apps/core/internal

# 创建 domain 目录结构
mkdir -p domain/{auth,user,profile,follow,content}/{handler,service,dto}
mkdir -p router

# 移动 shared packages
cd ../../..
mkdir -p apps/core/pkg/{database,logger,email}

# 移动现有 pkg 文件
mv apps/core/pkg/database/*.go apps/core/pkg/database/ 2>/dev/null || true
mv apps/core/pkg/logger/*.go apps/core/pkg/logger/ 2>/dev/null || true
mv apps/core/pkg/email/*.go apps/core/pkg/email/ 2>/dev/null || true
```

### 1.3 更新 go.mod 路径别名（可选）

如果项目规模大，考虑添加 import 别名：

```go
// go.mod 添加 replace 或直接使用完整路径
// 本指南使用完整路径，保持简单
```

### 1.4 创建共享接口

在 `apps/core/internal/domain/common.go` 创建跨领域共享的接口：

```go
package domain

import "context"

// Transactional 支持事务的接口
type Transactional interface {
    WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CursorPagination 游标分页请求
type CursorPagination struct {
    Cursor *int64
    Limit  int
}

// CursorPaginationResponse 游标分页响应
type CursorPaginationResponse struct {
    Total  int   `json:"total"`
    Cursor *int64 `json:"cursor,omitempty"`
}
```

---

## 2. Phase 1: Auth Domain Migration

**目标**: 迁移 `handler/auth.go` → `domain/auth/`

### 2.1 分析现有代码

当前 `auth.go` 包含：
- `AuthHandler` struct
- `NewAuthHandler` constructor
- 6 个 handlers: Register, Login, GoogleLogin, GoogleCallback, VerifyEmail, Me
- 2 个 request structs: RegisterRequest, LoginRequest

### 2.2 创建 Auth DTO

**File**: `apps/core/internal/domain/auth/dto.go`

```go
package auth

import "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"

// RegisterRequest 注册请求
type RegisterRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
    Message string `json:"message"`
    UserID  int64  `json:"user_id"`
}

// LoginResponse 登录响应
type LoginResponse struct {
    Token string `json:"token"`
    Type  string `json:"type"`
}

// UserInfoResponse 当前用户信息
type UserInfoResponse struct {
    UserID  interface{} `json:"user_id"`
    Email   interface{} `json:"email"`
    Role    interface{} `json:"role"`
    Message string      `json:"message"`
}

// ToUserResponse 将 model.User 转换为响应
func ToUserResponse(user *model.User) UserInfoResponse {
    return UserInfoResponse{
        UserID:  user.ID,
        Email:   "", // 从 user.Auths 获取
        Role:    user.Role,
        Message: "Token is valid!",
    }
}
```

### 2.3 创建 Auth Service

**File**: `apps/core/internal/domain/auth/service.go`

```go
package auth

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
    "gorm.io/gorm"
)

// Service 认证服务接口
type Service interface {
    Register(ctx context.Context, username, email, password, baseURL string) (*model.User, error)
    Login(ctx context.Context, email, password, ipAddress string) (string, error)
    LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, picture string) (string, error)
    VerifyEmail(ctx context.Context, token string) error
}

// service 认证服务实现
type service struct {
    db     *gorm.DB
    jwtCfg config.JWTConfig
    smtpCfg config.SMTPConfig
}

// NewService 创建认证服务
func NewService(db *gorm.DB, jwtCfg config.JWTConfig, smtpCfg config.SMTPConfig) Service {
    if db == nil {
        db = database.DB
    }
    return &service{
        db:      db,
        jwtCfg:  jwtCfg,
        smtpCfg: smtpCfg,
    }
}

// Register 用户注册
func (s *service) Register(ctx context.Context, username, email, password, baseURL string) (*model.User, error) {
    // 复制原有 service/auth.go 的 Register 逻辑
    // 注意：需要重构原有逻辑，移除 logger 依赖，通过返回值传递错误
    
    var user model.User
    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. 检查邮箱是否已存在
        var existing model.UserAuth
        if err := tx.Where("identity_type = ? AND identifier = ?", "email", email).First(&existing).Error; err == nil {
            return fmt.Errorf("email already registered")
        }
        
        // 2. 创建用户
        user = model.User{
            Role:   "user",
            Status: 0,
        }
        if err := tx.Create(&user).Error; err != nil {
            return fmt.Errorf("failed to create user: %w", err)
        }
        
        // 3. 创建认证记录
        auth := model.UserAuth{
            UserID:       user.ID,
            IdentityType: "email",
            Identifier:   email,
        }
        if err := auth.SetPassword(password); err != nil {
            return fmt.Errorf("failed to hash password: %w", err)
        }
        if err := tx.Create(&auth).Error; err != nil {
            return fmt.Errorf("failed to create auth record: %w", err)
        }
        
        // 4. 创建用户画像
        profile := model.UserProfile{
            UserID:   user.ID,
            Nickname: username,
        }
        if err := tx.Create(&profile).Error; err != nil {
            return fmt.Errorf("failed to create profile: %w", err)
        }
        
        // 5. 创建验证令牌
        // TODO: 实现邮件验证逻辑
        _ = baseURL // 用于构建验证链接
        
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

// Login 用户登录
func (s *service) Login(ctx context.Context, email, password, ipAddress string) (string, error) {
    // 复制原有登录逻辑
    var auth model.UserAuth
    if err := s.db.WithContext(ctx).Where("identity_type = ? AND identifier = ?", "email", email).First(&auth).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return "", fmt.Errorf("invalid email or password")
        }
        return "", fmt.Errorf("database error: %w", err)
    }
    
    if !auth.CheckPassword(password) {
        return "", fmt.Errorf("invalid email or password")
    }
    
    // TODO: 生成 JWT token
    _ = ipAddress
    return "jwt_token_placeholder", nil
}

// LoginOrRegisterOAuthUser OAuth 登录/注册
func (s *service) LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, picture string) (string, error) {
    // 复制原有 OAuth 逻辑
    // 简化版：查找或创建用户，返回 token
    _ = picture
    
    var auth model.UserAuth
    err := s.db.WithContext(ctx).Where("identity_type = ? AND identifier = ?", provider, email).First(&auth).Error
    
    if err == gorm.ErrRecordNotFound {
        // 创建新用户
        user := model.User{
            Role:   "user",
            Status: 0,
        }
        if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
            return "", fmt.Errorf("failed to create user: %w", err)
        }
        
        auth = model.UserAuth{
            UserID:       user.ID,
            IdentityType: provider,
            Identifier:   email,
        }
        if err := s.db.WithContext(ctx).Create(&auth).Error; err != nil {
            return "", fmt.Errorf("failed to create auth: %w", err)
        }
        
        profile := model.UserProfile{
            UserID:   user.ID,
            Nickname: name,
        }
        if err := s.db.WithContext(ctx).Create(&profile).Error; err != nil {
            return "", fmt.Errorf("failed to create profile: %w", err)
        }
    } else if err != nil {
        return "", fmt.Errorf("database error: %w", err)
    }
    
    // TODO: 生成 JWT
    return "jwt_token_placeholder", nil
}

// VerifyEmail 验证邮箱
func (s *service) VerifyEmail(ctx context.Context, token string) error {
    // 复制原有验证逻辑
    var verification model.EmailVerification
    if err := s.db.WithContext(ctx).Where("token = ?", token).First(&verification).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return fmt.Errorf("invalid or expired token")
        }
        return fmt.Errorf("database error: %w", err)
    }
    
    if verification.IsExpired() {
        return fmt.Errorf("token expired")
    }
    
    // 更新用户状态为已验证
    // TODO: 实现用户状态更新
    
    return nil
}

// ExchangeGoogleToken 交换 Google Token（辅助函数）
func (s *service) ExchangeGoogleToken(ctx context.Context, code, redirectURI string, oauthCfg config.OAuthConfig) (*http.Response, error) {
    return http.PostForm("https://oauth2.googleapis.com/token", url.Values{
        "code":          {code},
        "client_id":     {oauthCfg.Google.ClientID},
        "client_secret": {oauthCfg.Google.ClientSecret},
        "redirect_uri":  {redirectURI},
        "grant_type":    {"authorization_code"},
    })
}
```

### 2.4 创建 Auth Handler

**File**: `apps/core/internal/domain/auth/handler.go`

```go
package auth

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/logger"
    "github.com/gin-gonic/gin"
)

// Handler 认证处理器
type Handler struct {
    svc         Service
    oauthCfg    config.OAuthConfig
    frontendURL string
    apiBasePath string
}

// NewHandler 创建认证处理器
func NewHandler(svc Service, oauthCfg config.OAuthConfig, frontendURL, apiBasePath string) *Handler {
    return &Handler{
        svc:         svc,
        oauthCfg:    oauthCfg,
        frontendURL: frontendURL,
        apiBasePath: apiBasePath,
    }
}

// Register 用户注册
// @Summary Register a new user
// @Description Register a new user with username, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register Request"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 构建 baseURL
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
        Message: "User created successfully. Please check your email for verification link.",
        UserID:  user.ID,
    })
}

// Login 用户登录
// @Summary Login user
// @Description Login with email and password to get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
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
    
    c.JSON(http.StatusOK, LoginResponse{
        Token: token,
        Type:  "Bearer",
    })
}

// GoogleLogin Google OAuth 跳转
// @Summary Redirect to Google OAuth
// @Description Redirects user to Google OAuth authorization page
// @Tags auth
// @Success 302 "Redirect to Google OAuth"
// @Router /auth/login/google [get]
func (h *Handler) GoogleLogin(c *gin.Context) {
    clientID := h.oauthCfg.Google.ClientID
    if clientID == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Google OAuth not configured"})
        return
    }
    
    frontendURL := h.getFrontendURL(c)
    redirectURI := h.buildRedirectURI(c)
    state := c.Query("state")
    if state == "" {
        state = frontendURL
    }
    
    authURL := fmt.Sprintf(
        "https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&state=%s",
        clientID,
        redirectURI,
        "openid email profile",
        state,
    )
    
    c.Redirect(http.StatusFound, authURL)
}

// GoogleCallback Google OAuth 回调
// @Summary Handle Google OAuth callback
// @Description Handles Google OAuth callback, creates/logs in user, redirects to frontend with token
// @Tags auth
// @Param code query string true "Authorization code from Google"
// @Success 302 "Redirect to frontend with token"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/callback/google [get]
func (h *Handler) GoogleCallback(c *gin.Context) {
    code := c.Query("code")
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
        return
    }
    
    state := c.Query("state")
    frontendURL := h.getFrontendURL(c)
    if state != "" {
        frontendURL = state
    }
    
    redirectURI := h.buildRedirectURI(c)
    
    // 调用 service 交换 token
    tokenResp, err := h.svc.(*service).ExchangeGoogleToken(c.Request.Context(), code, redirectURI, h.oauthCfg)
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
        c.JSON(http.StatusBadRequest, gin.H{"error": tokenData.Error})
        return
    }
    
    // 获取用户信息
    userInfo, err := h.getGoogleUserInfo(tokenData.AccessToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // 登录或注册用户
    token, err := h.svc.LoginOrRegisterOAuthUser(c.Request.Context(), userInfo.Email, userInfo.Name, "google", userInfo.Picture)
    if err != nil {
        logger.Error(c.Request.Context(), "Failed to login/register OAuth user",
            "error", err.Error(),
            "event", "google_oauth_login_failed",
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process login"})
        return
    }
    
    redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, token)
    c.Redirect(http.StatusFound, redirectURL)
}

// getGoogleUserInfo 获取 Google 用户信息
func (h *Handler) getGoogleUserInfo(accessToken string) (*struct {
    ID      string `json:"id"`
    Email   string `json:"email"`
    Name    string `json:"name"`
    Picture string `json:"picture"`
}, error) {
    resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read user info: %w", err)
    }
    
    var userInfo struct {
        ID      string `json:"id"`
        Email   string `json:"email"`
        Name    string `json:"name"`
        Picture string `json:"picture"`
    }
    if err := json.Unmarshal(body, &userInfo); err != nil {
        return nil, fmt.Errorf("failed to decode user info: %w", err)
    }
    
    return &userInfo, nil
}

// VerifyEmail 验证邮箱
// @Summary Verify user email
// @Description Verify user email using the token sent to their email
// @Tags auth
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/verify [get]
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
    
    frontendURL := h.getFrontendURL(c)
    c.Redirect(http.StatusFound, frontendURL+"/auth/verified")
}

// Me 获取当前用户信息
// @Summary Get current user info
// @Description Get the current authenticated user's information (protected route)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserInfoResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
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

// Helper 方法

func (h *Handler) getFrontendURL(c *gin.Context) string {
    if h.frontendURL != "" {
        return h.frontendURL
    }
    
    if referer := c.GetHeader("Referer"); referer != "" {
        return referer
    }
    if origin := c.GetHeader("Origin"); origin != "" {
        return origin
    }
    
    return "http://localhost:3000"
}

func (h *Handler) buildRedirectURI(c *gin.Context) string {
    scheme := "http"
    if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
        scheme = "https"
    } else if c.Request.TLS != nil {
        scheme = "https"
    }
    return fmt.Sprintf("%s://%s%s/auth/callback/google", scheme, c.Request.Host, h.apiBasePath)
}
```

### 2.5 创建 Auth Routes

**File**: `apps/core/internal/domain/auth/routes.go`

```go
package auth

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes 注册认证路由
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    // 初始化服务
    svc := NewService(nil, cfg.JWT, cfg.SMTP)
    
    // 初始化处理器
    handler := NewHandler(svc, cfg.OAuth, cfg.FrontendURL, cfg.Server.APIBasePath)
    
    // 认证路由组
    auth := r.Group("/auth")
    {
        auth.POST("/register", handler.Register)
        auth.POST("/login", handler.Login)
        auth.GET("/login/google", handler.GoogleLogin)
        auth.GET("/callback/google", handler.GoogleCallback)
        auth.GET("/verify", handler.VerifyEmail)
        
        // 受保护的路由
        auth.GET("/me", middleware.JWTAuth(cfg.JWT), handler.Me)
    }
}
```

### 2.6 更新主路由

**File**: `apps/core/internal/router/router.go`

```go
package router

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/auth"
    "github.com/gin-gonic/gin"
)

// Setup 设置所有路由
func Setup(router *gin.Engine, cfg *config.Config) {
    api := router.Group(cfg.Server.APIBasePath)
    
    // 注册 auth 域路由
    auth.RegisterRoutes(api, cfg)
    
    // TODO: 后续添加其他域
    // user.RegisterRoutes(api, cfg)
    // profile.RegisterRoutes(api, cfg)
    // follow.RegisterRoutes(api, cfg)
    // content.RegisterRoutes(api, cfg)
}
```

### 2.7 更新 main.go

**File**: `apps/core/cmd/app/main.go`

```go
// 找到路由注册部分，替换为：
// 旧代码：
// handler.RegisterRoutes(router, cfg)

// 新代码：
router.Setup(router, cfg)
```

### 2.8 编译和测试

```bash
cd apps/core

# 编译
go build ./cmd/app

# 运行测试
go test ./domain/auth/...

# 如果测试不存在，先复制原有测试
cp internal/handler/auth_test.go internal/domain/auth/handler_test.go
# 然后修改 import 路径
```

---

## 3. Phase 2: Content Domain Migration

**目标**: 从 `service/content.go` 和分散的 handler 逻辑中提取 content 域

### 3.1 识别 Content 相关代码

当前分散在：
- `service/content.go`: Posts, Reviews, Favorites, Likes 的业务逻辑
- `user.go`: ListMyPosts, ListMyReviews, ListMyFavorites, ListMyLikes
- `profile.go`: ListUserPosts, ListUserReviews

### 3.2 Content DTO 定义

**File**: `apps/core/internal/domain/content/dto/post.go`

```go
package dto

import "time"

// PostItem 帖子列表项
type PostItem struct {
    ID        int64         `json:"id"`
    Title     string        `json:"title"`
    Content   string        `json:"content"`
    Images    []string      `json:"images"`
    LikeCount int           `json:"like_count"`
    ViewCount int           `json:"view_count"`
    IsLiked   bool          `json:"is_liked"`
    Merchant  *MerchantBrief `json:"merchant,omitempty"`
    Tags      []string      `json:"tags"`
    CreatedAt time.Time     `json:"created_at"`
}

// PostListResponse 帖子列表响应
type PostListResponse struct {
    Posts  []PostItem `json:"posts"`
    Total  int        `json:"total"`
    Cursor *int64     `json:"cursor,omitempty"`
}

// MerchantBrief 商家简要信息
type MerchantBrief struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Category string `json:"category"`
}
```

**File**: `apps/core/internal/domain/content/dto/review.go`

```go
package dto

import "time"

// ReviewItem 评论列表项
type ReviewItem struct {
    ID            int64         `json:"id"`
    Rating        int           `json:"rating"`
    RatingEnv     int           `json:"rating_env"`
    RatingService int           `json:"rating_service"`
    RatingValue   int           `json:"rating_value"`
    Content       string        `json:"content"`
    Images        []string      `json:"images"`
    AvgCost       int           `json:"avg_cost"`
    LikeCount     int           `json:"like_count"`
    IsLiked       bool          `json:"is_liked"`
    Merchant      MerchantBrief `json:"merchant"`
    Tags          []string      `json:"tags"`
    CreatedAt     time.Time     `json:"created_at"`
}

// ReviewListResponse 评论列表响应
type ReviewListResponse struct {
    Reviews []ReviewItem `json:"reviews"`
    Total   int          `json:"total"`
    Cursor  *int64       `json:"cursor,omitempty"`
}
```

**File**: `apps/core/internal/domain/content/dto/favorite.go`

```go
package dto

import "time"

// FavoriteItem 收藏项
type FavoriteItem struct {
    ID         int64      `json:"id"`
    TargetType string     `json:"target_type"`
    TargetID   int64      `json:"target_id"`
    Post       *PostItem  `json:"post,omitempty"`
    Review     *ReviewItem `json:"review,omitempty"`
    Merchant   *MerchantBrief `json:"merchant,omitempty"`
    CreatedAt  time.Time  `json:"created_at"`
}

// FavoriteListResponse 收藏列表响应
type FavoriteListResponse struct {
    Items  []FavoriteItem `json:"items"`
    Total  int            `json:"total"`
    Cursor *int64         `json:"cursor,omitempty"`
}
```

### 3.3 Content Service 接口和实现

**File**: `apps/core/internal/domain/content/service/post.go`

```go
package service

import (
    "context"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// PostService 帖子服务接口
type PostService interface {
    ListUserPosts(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Post, int64, error)
    ListMyPosts(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Post, int64, error)
}

// postService 帖子服务实现
type postService struct {
    db *gorm.DB
}

// NewPostService 创建帖子服务
func NewPostService(db *gorm.DB) PostService {
    if db == nil {
        db = database.DB
    }
    return &postService{db: db}
}

// ListUserPosts 获取用户的帖子列表（公开）
func (s *postService) ListUserPosts(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Post, int64, error) {
    return s.listPosts(ctx, userID, cursor, limit)
}

// ListMyPosts 获取当前用户的帖子列表
func (s *postService) ListMyPosts(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Post, int64, error) {
    return s.listPosts(ctx, userID, cursor, limit)
}

// listPosts 内部实现
func (s *postService) listPosts(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Post, int64, error) {
    var posts []model.Post
    var total int64
    
    query := s.db.WithContext(ctx).Where("user_id = ?", userID)
    
    if cursor != nil {
        query = query.Where("id < ?", *cursor)
    }
    
    if err := query.Order("id desc").Limit(limit).Find(&posts).Error; err != nil {
        return nil, 0, err
    }
    
    if err := s.db.WithContext(ctx).Model(&model.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    return posts, total, nil
}
```

**File**: `apps/core/internal/domain/content/service/review.go`

```go
package service

import (
    "context"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// ReviewService 评论服务接口
type ReviewService interface {
    ListUserReviews(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Review, int64, error)
    ListMyReviews(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Review, int64, error)
}

// reviewService 评论服务实现
type reviewService struct {
    db *gorm.DB
}

// NewReviewService 创建评论服务
func NewReviewService(db *gorm.DB) ReviewService {
    if db == nil {
        db = database.DB
    }
    return &reviewService{db: db}
}

// ListUserReviews 获取用户的评论列表（公开）
func (s *reviewService) ListUserReviews(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Review, int64, error) {
    return s.listReviews(ctx, userID, cursor, limit)
}

// ListMyReviews 获取当前用户的评论列表
func (s *reviewService) ListMyReviews(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Review, int64, error) {
    return s.listReviews(ctx, userID, cursor, limit)
}

func (s *reviewService) listReviews(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Review, int64, error) {
    var reviews []model.Review
    var total int64
    
    query := s.db.WithContext(ctx).Where("user_id = ?", userID)
    
    if cursor != nil {
        query = query.Where("id < ?", *cursor)
    }
    
    if err := query.Order("id desc").Limit(limit).Find(&reviews).Error; err != nil {
        return nil, 0, err
    }
    
    if err := s.db.WithContext(ctx).Model(&model.Review{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    return reviews, total, nil
}
```

**File**: `apps/core/internal/domain/content/service/favorite.go`

```go
package service

import (
    "context"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// FavoriteService 收藏服务接口
type FavoriteService interface {
    ListFavorites(ctx context.Context, userID int64, targetType string, cursor *int64, limit int) ([]model.Favorite, int64, error)
}

// favoriteService 收藏服务实现
type favoriteService struct {
    db *gorm.DB
}

// NewFavoriteService 创建收藏服务
func NewFavoriteService(db *gorm.DB) FavoriteService {
    if db == nil {
        db = database.DB
    }
    return &favoriteService{db: db}
}

// ListFavorites 获取收藏列表
func (s *favoriteService) ListFavorites(ctx context.Context, userID int64, targetType string, cursor *int64, limit int) ([]model.Favorite, int64, error) {
    var favorites []model.Favorite
    var total int64
    
    query := s.db.WithContext(ctx).Where("user_id = ?", userID)
    
    if targetType != "" {
        query = query.Where("target_type = ?", targetType)
    }
    
    if cursor != nil {
        query = query.Where("id < ?", *cursor)
    }
    
    if err := query.Order("id desc").Limit(limit).Find(&favorites).Error; err != nil {
        return nil, 0, err
    }
    
    countQuery := s.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ?", userID)
    if targetType != "" {
        countQuery = countQuery.Where("target_type = ?", targetType)
    }
    if err := countQuery.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    return favorites, total, nil
}
```

**File**: `apps/core/internal/domain/content/service/like.go`

```go
package service

import (
    "context"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// LikeService 点赞服务接口
type LikeService interface {
    ListLikes(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Like, int64, error)
}

// likeService 点赞服务实现
type likeService struct {
    db *gorm.DB
}

// NewLikeService 创建点赞服务
func NewLikeService(db *gorm.DB) LikeService {
    if db == nil {
        db = database.DB
    }
    return &likeService{db: db}
}

// ListLikes 获取点赞列表
func (s *likeService) ListLikes(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Like, int64, error) {
    var likes []model.Like
    var total int64
    
    query := s.db.WithContext(ctx).Where("user_id = ?", userID)
    
    if cursor != nil {
        query = query.Where("id < ?", *cursor)
    }
    
    if err := query.Order("id desc").Limit(limit).Find(&likes).Error; err != nil {
        return nil, 0, err
    }
    
    if err := s.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    return likes, total, nil
}
```

### 3.4 Content Handlers

**File**: `apps/core/internal/domain/content/handler/post.go`

```go
package handler

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// PostHandler 帖子处理器
type PostHandler struct {
    postSvc service.PostService
    db      *gorm.DB
}

// NewPostHandler 创建帖子处理器
func NewPostHandler(postSvc service.PostService) *PostHandler {
    if postSvc == nil {
        postSvc = service.NewPostService(nil)
    }
    return &PostHandler{
        postSvc: postSvc,
        db:      database.DB,
    }
}

// ListUserPosts 获取用户的帖子列表（公开）
// @Summary List user's posts
// @Description Returns a user's posts
// @Tags content
// @Produce json
// @Param id path int true "User ID"
// @Param cursor query int false "Cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.PostListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/posts [get]
func (h *PostHandler) ListUserPosts(c *gin.Context) {
    targetID, err := parseIDParam(c, "id")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    
    cursor, limit := parseCursorLimit(c)
    posts, total, err := h.postSvc.ListUserPosts(c.Request.Context(), targetID, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    liked := h.getLikedIDs(c, "post")
    items := make([]dto.PostItem, 0, len(posts))
    for _, post := range posts {
        var merchant *dto.MerchantBrief
        if post.MerchantID != nil {
            merchant = h.loadMerchantBrief(*post.MerchantID)
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
    
    c.JSON(http.StatusOK, dto.PostListResponse{
        Posts:  items,
        Total:  int(total),
        Cursor: nextCursor(posts),
    })
}

// ListMyPosts 获取当前用户的帖子列表
// @Summary List my posts
// @Description Returns posts created by the authenticated user
// @Tags content
// @Produce json
// @Param cursor query int false "Cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.PostListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/posts [get]
func (h *PostHandler) ListMyPosts(c *gin.Context) {
    userID := c.GetInt64("user_id")
    cursor, limit := parseCursorLimit(c)
    
    posts, total, err := h.postSvc.ListMyPosts(c.Request.Context(), userID, cursor, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    liked := h.getLikedIDs(c, "post")
    items := make([]dto.PostItem, 0, len(posts))
    for _, post := range posts {
        var merchant *dto.MerchantBrief
        if post.MerchantID != nil {
            merchant = h.loadMerchantBrief(*post.MerchantID)
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
    
    c.JSON(http.StatusOK, dto.PostListResponse{
        Posts:  items,
        Total:  int(total),
        Cursor: nextCursor(posts),
    })
}

// Helper 方法

func (h *PostHandler) getLikedIDs(c *gin.Context, targetType string) map[int64]bool {
    userID := c.GetInt64("user_id")
    if userID == 0 {
        return map[int64]bool{}
    }
    
    var likes []model.Like
    if err := h.db.WithContext(c.Request.Context()).
        Where("user_id = ? AND target_type = ?", userID, targetType).
        Find(&likes).Error; err != nil {
        return map[int64]bool{}
    }
    
    result := make(map[int64]bool, len(likes))
    for _, like := range likes {
        result[like.TargetID] = true
    }
    return result
}

func (h *PostHandler) loadMerchantBrief(id int64) *dto.MerchantBrief {
    var merchant model.Merchant
    if err := h.db.First(&merchant, id).Error; err != nil {
        return nil
    }
    return &dto.MerchantBrief{
        ID:       merchant.ID,
        Name:     merchant.Name,
        Category: merchant.Category,
    }
}

// parseIDParam 从 URL 参数解析 ID
func parseIDParam(c *gin.Context, name string) (int64, error) {
    return strconv.ParseInt(c.Param(name), 10, 64)
}

// parseCursorLimit 解析分页参数
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

// parseJSONStrings 解析 JSON 字符串数组
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

// nextCursor 获取下一页游标
func nextCursor[T any](items []T) *int64 {
    if len(items) == 0 {
        return nil
    }
    // 这里应该根据实际模型获取最后一个 ID
    // 简化处理，返回 nil
    return nil
}
```

### 3.5 Content Routes

**File**: `apps/core/internal/domain/content/routes.go`

```go
package content

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/content/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes 注册内容路由
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    // 初始化服务
    postSvc := service.NewPostService(nil)
    reviewSvc := service.NewReviewService(nil)
    favSvc := service.NewFavoriteService(nil)
    likeSvc := service.NewLikeService(nil)
    
    // 初始化处理器
    postHandler := handler.NewPostHandler(postSvc)
    reviewHandler := handler.NewReviewHandler(reviewSvc)
    favHandler := handler.NewFavoriteHandler(favSvc)
    likeHandler := handler.NewLikeHandler(likeSvc)
    
    // 公开路由
    users := r.Group("/users")
    {
        users.GET("/:id/posts", postHandler.ListUserPosts)
        users.GET("/:id/reviews", reviewHandler.ListUserReviews)
    }
    
    // 需要认证的路由
    user := r.Group("/user", middleware.JWTAuth(cfg.JWT))
    {
        user.GET("/posts", postHandler.ListMyPosts)
        user.GET("/reviews", reviewHandler.ListMyReviews)
        user.GET("/favorites", favHandler.ListMyFavorites)
        user.GET("/likes", likeHandler.ListMyLikes)
    }
}
```

---

## 4. Phase 3: User Domain Refactor

**目标**: 拆分 `user.go` (733 行) 为多个专注的 handler

### 4.1 User Domain 结构

```
apps/core/internal/domain/user/
├── handler/
│   ├── profile.go      # GET/PUT /user/profile
│   ├── privacy.go      # GET/PUT /user/privacy
│   ├── notification.go # GET/PUT /user/notifications
│   ├── address.go      # CRUD /user/addresses
│   ├── following.go    # 关注/粉丝列表
│   └── account.go      # 导出/删除账号
├── service/
│   ├── profile.go
│   ├── privacy.go
│   ├── notification.go
│   ├── address.go
│   ├── following.go
│   └── account.go
├── dto/
│   ├── profile.go
│   ├── privacy.go
│   ├── notification.go
│   ├── address.go
│   ├── following.go
│   └── account.go
└── routes.go
```

### 4.2 User DTO 示例

**File**: `apps/core/internal/domain/user/dto/address.go`

```go
package dto

// AddressItem 地址项
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

// AddressListResponse 地址列表响应
type AddressListResponse struct {
    Addresses []AddressItem `json:"addresses"`
}

// CreateAddressRequest 创建地址请求
type CreateAddressRequest struct {
    Name       string `json:"name" binding:"required"`
    Phone      string `json:"phone" binding:"required"`
    Province   string `json:"province" binding:"required"`
    City       string `json:"city" binding:"required"`
    District   string `json:"district"`
    Address    string `json:"address" binding:"required"`
    PostalCode string `json:"postal_code"`
    IsDefault  bool   `json:"is_default"`
}

// UpdateAddressRequest 更新地址请求
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

### 4.3 Address Handler

**File**: `apps/core/internal/domain/user/handler/address.go`

```go
package handler

import (
    "net/http"
    "strconv"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/service"
    "github.com/gin-gonic/gin"
)

// AddressHandler 地址处理器
type AddressHandler struct {
    svc service.AddressService
}

// NewAddressHandler 创建地址处理器
func NewAddressHandler(svc service.AddressService) *AddressHandler {
    return &AddressHandler{svc: svc}
}

// ListAddresses 获取地址列表
// @Summary List addresses
// @Description Returns the authenticated user's saved addresses
// @Tags user
// @Produce json
// @Success 200 {object} dto.AddressListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/addresses [get]
func (h *AddressHandler) ListAddresses(c *gin.Context) {
    userID := c.GetInt64("user_id")
    items, err := h.svc.ListAddresses(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    resp := dto.AddressListResponse{Addresses: make([]dto.AddressItem, 0, len(items))}
    for _, item := range items {
        resp.Addresses = append(resp.Addresses, dto.AddressItem{
            ID:         item.ID,
            Name:       item.Name,
            Phone:      item.Phone,
            Province:   item.Province,
            City:       item.City,
            District:   item.District,
            Address:    item.Address,
            PostalCode: item.PostalCode,
            IsDefault:  item.IsDefault,
        })
    }
    c.JSON(http.StatusOK, resp)
}

// CreateAddress 创建地址
// @Summary Create address
// @Description Adds a new address for the authenticated user
// @Tags user
// @Accept json
// @Produce json
// @Param request body dto.CreateAddressRequest true "Create Address Request"
// @Success 201 {object} dto.AddressItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /user/addresses [post]
func (h *AddressHandler) CreateAddress(c *gin.Context) {
    userID := c.GetInt64("user_id")
    var req dto.CreateAddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    addr, err := h.svc.CreateAddress(c.Request.Context(), userID, req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, dto.AddressItem{
        ID:         addr.ID,
        Name:       addr.Name,
        Phone:      addr.Phone,
        Province:   addr.Province,
        City:       addr.City,
        District:   addr.District,
        Address:    addr.Address,
        PostalCode: addr.PostalCode,
        IsDefault:  addr.IsDefault,
    })
}

// UpdateAddress 更新地址
// @Summary Update address
// @Description Updates an existing address
// @Tags user
// @Accept json
// @Produce json
// @Param id path int true "Address ID"
// @Param request body dto.UpdateAddressRequest true "Update Address Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/addresses/{id} [patch]
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
    userID := c.GetInt64("user_id")
    addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
        return
    }
    
    var req dto.UpdateAddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.svc.UpdateAddress(c.Request.Context(), userID, addressID, req); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeleteAddress 删除地址
// @Summary Delete address
// @Description Deletes an address
// @Tags user
// @Produce json
// @Param id path int true "Address ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/addresses/{id} [delete]
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
    userID := c.GetInt64("user_id")
    addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
        return
    }
    
    if err := h.svc.DeleteAddress(c.Request.Context(), userID, addressID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SetDefaultAddress 设置默认地址
// @Summary Set default address
// @Description Sets an address as default
// @Tags user
// @Produce json
// @Param id path int true "Address ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/addresses/{id}/default [post]
func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
    userID := c.GetInt64("user_id")
    addressID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
        return
    }
    
    if err := h.svc.SetDefaultAddress(c.Request.Context(), userID, addressID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

### 4.4 User Routes

**File**: `apps/core/internal/domain/user/routes.go`

```go
package user

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/user/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes 注册用户路由
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    // 初始化服务
    profileSvc := service.NewProfileService(nil)
    privacySvc := service.NewPrivacyService(nil)
    notifSvc := service.NewNotificationService(nil)
    addrSvc := service.NewAddressService(nil)
    followingSvc := service.NewFollowingService(nil)
    accountSvc := service.NewAccountService(nil)
    
    // 初始化处理器
    profileHandler := handler.NewProfileHandler(profileSvc)
    privacyHandler := handler.NewPrivacyHandler(privacySvc)
    notifHandler := handler.NewNotificationHandler(notifSvc)
    addrHandler := handler.NewAddressHandler(addrSvc)
    followingHandler := handler.NewFollowingHandler(followingSvc)
    accountHandler := handler.NewAccountHandler(accountSvc)
    
    // 需要认证的路由组
    user := r.Group("/user", middleware.JWTAuth(cfg.JWT))
    {
        // Profile
        profile := user.Group("/profile")
        {
            profile.GET("", profileHandler.GetProfile)
            profile.PATCH("", profileHandler.UpdateProfile)
        }
        
        // Privacy
        privacy := user.Group("/privacy")
        {
            privacy.GET("", privacyHandler.GetPrivacy)
            privacy.PATCH("", privacyHandler.UpdatePrivacy)
        }
        
        // Notifications
        notifications := user.Group("/notifications")
        {
            notifications.GET("", notifHandler.GetNotifications)
            notifications.PATCH("", notifHandler.UpdateNotifications)
        }
        
        // Addresses
        addresses := user.Group("/addresses")
        {
            addresses.GET("", addrHandler.ListAddresses)
            addresses.POST("", addrHandler.CreateAddress)
            addresses.PATCH("/:id", addrHandler.UpdateAddress)
            addresses.DELETE("/:id", addrHandler.DeleteAddress)
            addresses.POST("/:id/default", addrHandler.SetDefaultAddress)
        }
        
        // Following
        following := user.Group("/following")
        {
            following.GET("/users", followingHandler.ListFollowingUsers)
            following.GET("/merchants", followingHandler.ListFollowingMerchants)
        }
        followers := user.Group("/followers")
        {
            followers.GET("", followingHandler.ListFollowers)
        }
        
        // Account
        account := user.Group("/account")
        {
            account.POST("/export", accountHandler.RequestAccountExport)
            account.DELETE("", accountHandler.RequestAccountDeletion)
        }
    }
}
```

---

## 5. Phase 4: Follow Domain Extraction

**目标**: 从 `profile.go` 和 `user.go` 提取关注/粉丝逻辑

### 5.1 Follow Service

**File**: `apps/core/internal/domain/follow/service/user.go`

```go
package service

import (
    "context"
    "fmt"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// UserService 用户关注服务接口
type UserService interface {
    FollowUser(ctx context.Context, followerID, followingID int64) error
    UnfollowUser(ctx context.Context, followerID, followingID int64) error
}

// userService 用户关注服务实现
type userService struct {
    db *gorm.DB
}

// NewUserService 创建用户关注服务
func NewUserService(db *gorm.DB) UserService {
    if db == nil {
        db = database.DB
    }
    return &userService{db: db}
}

// FollowUser 关注用户
func (s *userService) FollowUser(ctx context.Context, followerID, followingID int64) error {
    if followerID == followingID {
        return fmt.Errorf("cannot follow yourself")
    }
    
    // 检查是否已关注
    var existing model.UserFollow
    if err := s.db.WithContext(ctx).
        Where("follower_id = ? AND following_id = ?", followerID, followingID).
        First(&existing).Error; err == nil {
        return fmt.Errorf("already following")
    }
    
    // 创建关注记录
    follow := model.UserFollow{
        FollowerID:  followerID,
        FollowingID: followingID,
    }
    
    return s.db.WithContext(ctx).Create(&follow).Error
}

// UnfollowUser 取消关注用户
func (s *userService) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
    return s.db.WithContext(ctx).
        Where("follower_id = ? AND following_id = ?", followerID, followingID).
        Delete(&model.UserFollow{}).Error
}
```

**File**: `apps/core/internal/domain/follow/service/merchant.go`

```go
package service

import (
    "context"
    "fmt"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// MerchantService 商家关注服务接口
type MerchantService interface {
    FollowMerchant(ctx context.Context, userID, merchantID int64) error
    UnfollowMerchant(ctx context.Context, userID, merchantID int64) error
}

// merchantService 商家关注服务实现
type merchantService struct {
    db *gorm.DB
}

// NewMerchantService 创建商家关注服务
func NewMerchantService(db *gorm.DB) MerchantService {
    if db == nil {
        db = database.DB
    }
    return &merchantService{db: db}
}

// FollowMerchant 关注商家
func (s *merchantService) FollowMerchant(ctx context.Context, userID, merchantID int64) error {
    // 检查是否已关注
    var existing model.MerchantFollow
    if err := s.db.WithContext(ctx).
        Where("user_id = ? AND merchant_id = ?", userID, merchantID).
        First(&existing).Error; err == nil {
        return fmt.Errorf("already following")
    }
    
    follow := model.MerchantFollow{
        UserID:     userID,
        MerchantID: merchantID,
    }
    
    return s.db.WithContext(ctx).Create(&follow).Error
}

// UnfollowMerchant 取消关注商家
func (s *merchantService) UnfollowMerchant(ctx context.Context, userID, merchantID int64) error {
    return s.db.WithContext(ctx).
        Where("user_id = ? AND merchant_id = ?", userID, merchantID).
        Delete(&model.MerchantFollow{}).Error
}
```

**File**: `apps/core/internal/domain/follow/service/list.go`

```go
package service

import (
    "context"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// ListService 列表服务接口
type ListService interface {
    ListFollowingUsers(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.UserFollow, error)
    ListFollowingMerchants(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.MerchantFollow, error)
    ListFollowers(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.UserFollow, error)
}

// listService 列表服务实现
type listService struct {
    db *gorm.DB
}

// NewListService 创建列表服务
func NewListService(db *gorm.DB) ListService {
    if db == nil {
        db = database.DB
    }
    return &listService{db: db}
}

// ListFollowingUsers 获取关注用户列表
func (s *listService) ListFollowingUsers(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.UserFollow, error) {
    query := s.db.WithContext(ctx).Where("follower_id = ?", userID).Order("id desc")
    
    if cursor != nil {
        query = query.Where("id < ?", *cursor)
    }
    
    var follows []model.UserFollow
    if err := query.Limit(limit).Find(&follows).Error; err != nil {
        return nil, err
    }
    
    return follows, nil
}

// ListFollowingMerchants 获取关注商家列表
func (s *listService) ListFollowingMerchants(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.MerchantFollow, error) {
    query := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id desc")
    
    if cursor != nil {
        query = query.Where("id < ?", *cursor)
    }
    
    var follows []model.MerchantFollow
    if err := query.Limit(limit).Find(&follows).Error; err != nil {
        return nil, err
    }
    
    return follows, nil
}

// ListFollowers 获取粉丝列表
func (s *listService) ListFollowers(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.UserFollow, error) {
    query := s.db.WithContext(ctx).Where("following_id = ?", userID).Order("id desc")
    
    if cursor != nil {
        query = query.Where("id < ?", *cursor)
    }
    
    var follows []model.UserFollow
    if err := query.Limit(limit).Find(&follows).Error; err != nil {
        return nil, err
    }
    
    return follows, nil
}
```

### 5.2 Follow Handlers

**File**: `apps/core/internal/domain/follow/handler/user.go`

```go
package handler

import (
    "net/http"
    "strconv"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/service"
    "github.com/gin-gonic/gin"
)

// UserHandler 用户关注处理器
type UserHandler struct {
    svc service.UserService
}

// NewUserHandler 创建用户关注处理器
func NewUserHandler(svc service.UserService) *UserHandler {
    return &UserHandler{svc: svc}
}

// FollowUser 关注用户
// @Summary Follow user
// @Description Follow a user
// @Tags follow
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /users/{id}/follow [post]
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

// UnfollowUser 取消关注用户
// @Summary Unfollow user
// @Description Unfollow a user
// @Tags follow
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/follow [delete]
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

**File**: `apps/core/internal/domain/follow/handler/merchant.go`

```go
package handler

import (
    "net/http"
    "strconv"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/service"
    "github.com/gin-gonic/gin"
)

// MerchantHandler 商家关注处理器
type MerchantHandler struct {
    svc service.MerchantService
}

// NewMerchantHandler 创建商家关注处理器
func NewMerchantHandler(svc service.MerchantService) *MerchantHandler {
    return &MerchantHandler{svc: svc}
}

// FollowMerchant 关注商家
// @Summary Follow merchant
// @Description Follow a merchant
// @Tags follow
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /merchants/{id}/follow [post]
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

// UnfollowMerchant 取消关注商家
// @Summary Unfollow merchant
// @Description Unfollow a merchant
// @Tags follow
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /merchants/{id}/follow [delete]
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

### 5.3 Follow Routes

**File**: `apps/core/internal/domain/follow/routes.go`

```go
package follow

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/handler"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/follow/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes 注册关注路由
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    // 初始化服务
    userSvc := service.NewUserService(nil)
    merchantSvc := service.NewMerchantService(nil)
    listSvc := service.NewListService(nil)
    
    // 初始化处理器
    userHandler := handler.NewUserHandler(userSvc)
    merchantHandler := handler.NewMerchantHandler(merchantSvc)
    
    // 需要认证的路由
    auth := r.Group("", middleware.JWTAuth(cfg.JWT))
    {
        // 用户关注
        auth.POST("/users/:id/follow", userHandler.FollowUser)
        auth.DELETE("/users/:id/follow", userHandler.UnfollowUser)
        
        // 商家关注
        auth.POST("/merchants/:id/follow", merchantHandler.FollowMerchant)
        auth.DELETE("/merchants/:id/follow", merchantHandler.UnfollowMerchant)
    }
    
    // 关注列表路由在 user domain 中处理，因为返回的是用户自己的数据
    _ = listSvc // 在 following handler 中使用
}
```

---

## 6. Phase 5: Profile Domain Simplification

**目标**: 将 `profile.go` 简化为只处理公开资料查看

### 6.1 Profile Handler

**File**: `apps/core/internal/domain/profile/handler.go`

```go
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

// Handler 资料处理器
type Handler struct {
    svc service.Service
    db  *gorm.DB
}

// NewHandler 创建资料处理器
func NewHandler(svc service.Service) *Handler {
    return &Handler{
        svc: svc,
        db:  database.DB,
    }
}

// GetPublicProfile 获取公开资料
// @Summary Get public user profile
// @Description Returns a user's public profile
// @Tags profile
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} service.PublicProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (h *Handler) GetPublicProfile(c *gin.Context) {
    targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    
    viewerID := c.GetInt64("user_id")
    
    // 检查隐私设置
    var privacy model.UserPrivacy
    if err := h.db.WithContext(c.Request.Context()).First(&privacy, "user_id = ?", targetID).Error; err == nil {
        if !privacy.IsPublic && viewerID != targetID {
            c.JSON(http.StatusForbidden, gin.H{"error": "profile is private"})
            return
        }
    }
    
    // 获取资料
    profile, err := h.svc.GetPublicProfile(c.Request.Context(), targetID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    
    // 检查是否已关注
    if viewerID != 0 && viewerID != targetID {
        var follow model.UserFollow
        if err := h.db.WithContext(c.Request.Context()).
            Where("follower_id = ? AND following_id = ?", viewerID, targetID).
            First(&follow).Error; err == nil {
        }
    }
    
    c.JSON(http.StatusOK, profile)
}
```

### 6.2 Profile Service

**File**: `apps/core/internal/domain/profile/service.go`

```go
package service

import (
    "context"
    "fmt"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "github.com/RevieU-Corp/revieu-backend/apps/core/pkg/database"
    "gorm.io/gorm"
)

// Service 资料服务接口
type Service interface {
    GetPublicProfile(ctx context.Context, userID int64) (*PublicProfileResponse, error)
}

// service 资料服务实现
type service struct {
    db *gorm.DB
}

// NewService 创建资料服务
func NewService(db *gorm.DB) Service {
    if db == nil {
        db = database.DB
    }
    return &service{db: db}
}

// PublicProfileResponse 公开资料响应
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

// GetPublicProfile 获取公开资料
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
        IsFollowing:    false, // 由调用方设置
    }, nil
}
```

### 6.3 Profile Routes

**File**: `apps/core/internal/domain/profile/routes.go`

```go
package profile

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/domain/profile/service"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterRoutes 注册资料路由
func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    // 初始化服务
    svc := service.NewService(nil)
    
    // 初始化处理器
    handler := NewHandler(svc)
    
    // 公开路由（可选认证）
    users := r.Group("/users")
    {
        users.GET("/:id", middleware.OptionalJWTAuth(cfg.JWT), handler.GetPublicProfile)
    }
}
```

---

## 7. Phase 6: Cleanup & Finalization

### 7.1 更新主路由

**File**: `apps/core/internal/router/router.go` (Final)

```go
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

// Setup 设置所有路由
func Setup(router *gin.Engine, cfg *config.Config) {
    api := router.Group(cfg.Server.APIBasePath)
    
    // 注册所有域路由
    auth.RegisterRoutes(api, cfg)
    user.RegisterRoutes(api, cfg)
    profile.RegisterRoutes(api, cfg)
    follow.RegisterRoutes(api, cfg)
    content.RegisterRoutes(api, cfg)
    
    // 测试路由（可选保留或删除）
    // test.RegisterRoutes(api, cfg)
}
```

### 7.2 删除旧文件

```bash
cd apps/core/internal

# 删除旧目录
rm -rf handler/
rm -rf service/
rm -rf dto/

# 验证没有遗漏的引用
cd ../..
go build ./cmd/app
```

### 7.3 更新 main.go (Final)

**File**: `apps/core/cmd/app/main.go`

```go
// ... 其他代码 ...

import (
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/router"
    // ... 其他 imports ...
)

func main() {
    // ... 初始化和配置 ...
    
    // 设置路由
    router.Setup(r, cfg)
    
    // ... 启动服务 ...
}
```

### 7.4 创建 CODEOWNERS

**File**: `.github/CODEOWNERS`

```
# Global fallback
* @team-lead

# Domain ownership
apps/core/internal/domain/auth/ @auth-team
apps/core/internal/domain/user/ @user-team
apps/core/internal/domain/profile/ @profile-team
apps/core/internal/domain/follow/ @social-team
apps/core/internal/domain/content/ @content-team

# Infrastructure
apps/core/internal/config/ @platform-team
apps/core/internal/middleware/ @platform-team
apps/core/internal/model/ @platform-team
apps/core/pkg/ @platform-team
apps/core/internal/router/ @platform-team
```

---

## 8. Testing Strategy

### 8.1 单元测试迁移

每个新文件应该有对应的测试文件：

```
domain/auth/
├── handler.go
├── handler_test.go
├── service.go
├── service_test.go
└── dto.go
```

### 8.2 测试模板

**File**: `apps/core/internal/domain/auth/handler_test.go`

```go
package auth

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/config"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockService 模拟认证服务
type MockService struct {
    mock.Mock
}

func (m *MockService) Register(ctx context.Context, username, email, password, baseURL string) (*model.User, error) {
    args := m.Called(ctx, username, email, password, baseURL)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockService) Login(ctx context.Context, email, password, ipAddress string) (string, error) {
    args := m.Called(ctx, email, password, ipAddress)
    return args.String(0), args.Error(1)
}

func (m *MockService) LoginOrRegisterOAuthUser(ctx context.Context, email, name, provider, picture string) (string, error) {
    args := m.Called(ctx, email, name, provider, picture)
    return args.String(0), args.Error(1)
}

func (m *MockService) VerifyEmail(ctx context.Context, token string) error {
    args := m.Called(ctx, token)
    return args.Error(0)
}

func TestHandler_Register(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    mockSvc := new(MockService)
    handler := NewHandler(mockSvc, config.OAuthConfig{}, "", "")
    
    router := gin.New()
    router.POST("/auth/register", handler.Register)
    
    t.Run("success", func(t *testing.T) {
        mockSvc.On("Register", mock.Anything, "testuser", "test@example.com", "password123", mock.Anything).
            Return(&model.User{ID: 1}, nil).Once()
        
        body := RegisterRequest{
            Username: "testuser",
            Email:    "test@example.com",
            Password: "password123",
        }
        jsonBody, _ := json.Marshal(body)
        
        req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var resp RegisterResponse
        err := json.Unmarshal(w.Body.Bytes(), &resp)
        assert.NoError(t, err)
        assert.Equal(t, int64(1), resp.UserID)
    })
    
    t.Run("invalid request", func(t *testing.T) {
        body := `{"username": "test"}` // missing email and password
        req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusBadRequest, w.Code)
    })
}
```

### 8.3 集成测试

创建端到端测试：

**File**: `apps/core/tests/integration/auth_test.go`

```go
package integration

import (
    "testing"
    
    "github.com/stretchr/testify/suite"
)

type AuthSuite struct {
    suite.Suite
    // 测试服务器、数据库连接等
}

func (s *AuthSuite) SetupSuite() {
    // 设置测试环境
}

func (s *AuthSuite) TearDownSuite() {
    // 清理
}

func (s *AuthSuite) TestRegisterAndLogin() {
    // 注册
    // 登录
    // 验证
}

func TestAuthSuite(t *testing.T) {
    suite.Run(t, new(AuthSuite))
}
```

### 8.4 运行所有测试

```bash
cd apps/core

# 运行所有测试
go test ./...

# 运行特定域的测试
go test ./domain/auth/...
go test ./domain/user/...

# 运行集成测试
go test ./tests/integration/...

# 带覆盖率
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 9. Common Pitfalls & Solutions

### 9.1 Import 循环依赖

**问题**: Handler → Service → Model, 但 Model 需要引用其他域的 DTO

**解决方案**:
1. DTO 和 Model 分离，Model 不依赖 DTO
2. 使用接口解耦
3. 必要时提取 common types

```go
// 在 domain/common.go 定义共享类型
package domain

type UserBrief struct {
    ID       int64  `json:"id"`
    Nickname string `json:"nickname"`
    Avatar   string `json:"avatar"`
}
```

### 9.2 事务管理

**问题**: 跨域操作需要事务

**解决方案**:
1. 使用 Unit of Work 模式
2. 在应用层（handler）管理事务
3. 使用事件驱动（异步）

```go
// 在 handler 中管理事务
func (h *Handler) ComplexOperation(c *gin.Context) {
    err := h.db.Transaction(func(tx *gorm.DB) error {
        // 使用事务的 service 实例
        svc1 := service.NewSomethingService(tx)
        svc2 := service.NewOtherService(tx)
        
        if err := svc1.DoSomething(); err != nil {
            return err
        }
        if err := svc2.DoOther(); err != nil {
            return err
        }
        return nil
    })
}
```

### 9.3 共享 Helper 函数

**问题**: 多个 handler 需要 parseCursorLimit 等工具函数

**解决方案**:
1. 创建 `internal/common/http.go`
2. 使用 middleware 注入
3. 每个 handler 独立定义（保持简单）

```go
package common

import (
    "strconv"
    "github.com/gin-gonic/gin"
)

// ParsePagination 解析分页参数
func ParsePagination(c *gin.Context) (*int64, int) {
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
```

### 9.4 Swagger 文档更新

**问题**: 迁移后 Swagger 注解可能不匹配

**解决方案**:
1. 确保每个 handler 都有完整的 Swagger 注解
2. 使用 swag 工具重新生成

```bash
cd apps/core

# 重新生成 Swagger 文档
swag init -g cmd/app/main.go -o docs/

# 或者使用 go generate（如果配置了）
go generate ./...
```

### 9.5 测试数据准备

**问题**: 测试需要数据库状态

**解决方案**:
1. 使用 testify/mock 模拟 service
2. 使用 testcontainers 进行集成测试
3. 使用 sqlmock

```bash
# 安装 sqlmock
go get github.com/DATA-DOG/go-sqlmock
```

---

## 10. Post-Migration Checklist

### 10.1 代码检查

- [ ] 所有旧文件已删除
- [ ] 所有新文件已创建
- [ ] 没有编译错误
- [ ] 没有循环依赖
- [ ] Swagger 文档已更新
- [ ] 代码格式化（gofmt）

### 10.2 测试检查

- [ ] 单元测试通过率 > 80%
- [ ] 集成测试通过
- [ ] 手动测试关键路径
- [ ] 性能测试无回归

### 10.3 文档检查

- [ ] README 已更新
- [ ] API 文档已更新
- [ ] CODEOWNERS 已创建
- [ ] CHANGELOG 已记录

### 10.4 部署检查

- [ ] CI/CD 配置更新
- [ ] 环境变量检查
- [ ] 数据库迁移脚本（如有）
- [ ] 回滚计划准备

---

## Appendix A: Quick Reference

### A.1 Directory Structure (Final)

```
apps/core/
├── cmd/app/
│   └── main.go
├── internal/
│   ├── config/
│   ├── middleware/
│   ├── model/
│   ├── domain/
│   │   ├── auth/
│   │   ├── user/
│   │   ├── profile/
│   │   ├── follow/
│   │   └── content/
│   └── router/
├── pkg/
│   ├── database/
│   ├── logger/
│   └── email/
├── docs/
└── tests/
    └── integration/
```

### A.2 Migration Commands Summary

```bash
# 1. Setup
git checkout -b refactor/domain-driven-architecture
mkdir -p apps/core/internal/domain/{auth,user,profile,follow,content}/{handler,service,dto}
mkdir -p apps/core/internal/router

# 2. Migrate auth domain
# Copy and adapt code from handler/auth.go to domain/auth/

# 3. Migrate content domain
# Copy from service/content.go to domain/content/service/

# 4. Migrate user domain
# Split user.go into multiple files in domain/user/

# 5. Migrate follow domain
# Extract from profile.go and user.go

# 6. Cleanup
rm -rf apps/core/internal/{handler,service,dto}

# 7. Verify
go build ./apps/core/cmd/app
go test ./apps/core/...
```

### A.3 Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Domain per directory | Clear ownership, parallel development |
| Service interfaces | Testability, dependency injection |
| DTO per domain | Type safety, decoupling |
| Central router | Single entry point, clear dependencies |
| Interface-based services | Mocking for tests |

---

**End of Implementation Guide**

For questions or issues, refer to the main migration plan or contact the architecture team.
