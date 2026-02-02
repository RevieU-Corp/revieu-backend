# Backend Architecture Migration Plan

## 1. Current State Analysis

### 1.1 Current Directory Structure

```
apps/core/internal/
├── config/          # Configuration
├── dto/             # Data Transfer Objects
├── handler/         # All handlers in one place
│   ├── auth.go      # Authentication handlers (374 lines)
│   ├── user.go      # User management handlers (733 lines)
│   ├── profile.go   # Profile handlers (313 lines)
│   ├── routes.go    # Centralized route registration
│   ├── test_handler.go
│   └── helpers.go
├── middleware/      # Middleware
├── model/           # Database models
└── service/         # All services in one place
    ├── auth.go
    ├── user.go
    ├── follow.go
    ├── content.go
    └── interaction.go
```

### 1.2 Current Problems

1. **Flat Structure**: All handlers are in a single `handler/` directory, making it hard to:
   - Identify code owners for specific business domains
   - Navigate and understand the codebase structure
   - Enable parallel development without conflicts

2. **Monolithic Handler Files**:
   - `user.go` is 733 lines, handling multiple concerns (profile, addresses, notifications, privacy, etc.)
   - `profile.go` mixes public profile and follow logic
   - `routes.go` centralizes all route registrations, creating a bottleneck

3. **No Clear Domain Boundaries**: Business domains (auth, user, profile, merchant, follow) are mixed together

4. **Scalability Issues**: As the project grows, single-directory structure will become unmanageable

### 1.3 Current Business Domains

Based on route analysis, the following domains are identified:

| Domain | Endpoints | Handler Files |
|--------|-----------|---------------|
| Auth | 6 endpoints | `auth.go` (374 lines) |
| User (Private) | 18 endpoints | `user.go` (733 lines) |
| Profile (Public) | 6 endpoints | `profile.go` (313 lines) |
| Follow | 4 endpoints | Mixed in `profile.go` & `user.go` |
| Test | 2 endpoints | `test_handler.go` (66 lines) |

**Total**: ~36 endpoints in single-layer handler structure

## 2. Target Architecture Design

### 2.1 Design Principles

1. **Domain-Driven Design (DDD)**: Organize code by business domain, not technical layer
2. **Layered Architecture**: Each domain has clear layers (Handler → Service → Repository)
3. **Modular Ownership**: Enable CODEOWNERS file for domain-level ownership
4. **Single Responsibility**: Each handler file focuses on one concern
5. **Scalability**: Easy to add new domains without modifying existing ones

### 2.2 Proposed Directory Structure

```
apps/core/internal/
├── config/              # Shared configuration
├── middleware/          # Shared middleware
├── model/              # Database models (keep as is or split)
├── pkg/                # Shared utilities (moved from internal)
│   ├── database/
│   ├── logger/
│   └── email/
│
├── domain/             # Business domains
│   ├── auth/           # Authentication domain
│   │   ├── handler.go  # Login, register, OAuth, verify
│   │   ├── service.go  # Auth business logic
│   │   ├── dto.go      # Auth-specific DTOs
│   │   └── routes.go   # Auth route registration
│   │
│   ├── user/           # User management domain
│   │   ├── handler/
│   │   │   ├── profile.go      # GET /user/profile
│   │   │   ├── privacy.go      # GET/PUT /user/privacy
│   │   │   ├── notifications.go # GET/PUT /user/notifications
│   │   │   ├── addresses.go    # CRUD /user/addresses
│   │   │   └── account.go      # Export, delete account
│   │   ├── service/
│   │   │   ├── profile.go
│   │   │   ├── privacy.go
│   │   │   ├── notification.go
│   │   │   ├── address.go
│   │   │   └── account.go
│   │   ├── dto/
│   │   │   ├── profile.go
│   │   │   ├── privacy.go
│   │   │   ├── notification.go
│   │   │   ├── address.go
│   │   │   └── account.go
│   │   └── routes.go
│   │
│   ├── profile/        # Public profile domain
│   │   ├── handler.go  # GET /users/:id, posts, reviews
│   │   ├── service.go  # Profile logic
│   │   ├── dto.go      # Profile DTOs
│   │   └── routes.go
│   │
│   ├── follow/         # Follow system domain
│   │   ├── handler/
│   │   │   ├── user.go      # Follow/unfollow users
│   │   │   └── merchant.go  # Follow/unfollow merchants
│   │   ├── service/
│   │   │   ├── user.go
│   │   │   └── merchant.go
│   │   ├── dto.go
│   │   └── routes.go
│   │
│   └── content/        # Posts, reviews, favorites, likes
│       ├── handler/
│       │   ├── post.go
│       │   ├── review.go
│       │   ├── favorite.go
│       │   └── like.go
│       ├── service/
│       │   ├── post.go
│       │   ├── review.go
│       │   ├── favorite.go
│       │   └── like.go
│       ├── dto/
│       │   ├── post.go
│       │   ├── review.go
│       │   └── favorite.go
│       └── routes.go
│
└── router/             # Central router setup
    └── router.go       # Domain route aggregation
```

### 2.3 Domain Breakdown

| Domain | Responsibilities | Ownership |
|--------|------------------|-----------|
| `auth` | Login, register, OAuth, email verification, password reset | Auth Team |
| `user` | Private user management: profile, settings, addresses, account | User Team |
| `profile` | Public profile viewing: user profiles, public posts/reviews | Profile Team |
| `follow` | Follow/unfollow users and merchants, follower lists | Social Team |
| `content` | Posts, reviews, favorites, likes (current in service/content.go) | Content Team |

## 3. Migration Strategy

### 3.1 Phase 1: Foundation (Week 1)

**Goal**: Set up new structure and migrate simple domains

**Tasks**:
1. Create new directory structure
2. Move shared packages to `pkg/`
3. Migrate `auth` domain (simplest, 1 handler file)
4. Set up `router/` for domain aggregation
5. Update imports and ensure tests pass

**File Changes**:
```
# New files to create
apps/core/internal/domain/auth/
├── handler.go      # Move from handler/auth.go
├── service.go      # Move from service/auth.go
├── dto.go          # Extract from internal/dto/
└── routes.go       # Domain-specific routes

apps/core/internal/router/
└── router.go       # Aggregates all domain routes
```

### 3.2 Phase 2: Content Domain (Week 1-2)

**Goal**: Migrate content-related functionality

**Current State**:
- `service/content.go` handles posts, reviews, favorites, likes
- Mixed with `user.go` handlers for `/user/posts`, `/user/reviews`, etc.

**Migration**:
```
apps/core/internal/domain/content/
├── handler/
│   ├── post.go     # GET /posts, GET /users/:id/posts, etc.
│   ├── review.go   # GET /reviews, GET /users/:id/reviews
│   ├── favorite.go # GET/POST /favorites
│   └── like.go     # POST/DELETE /likes
├── service/
│   ├── post.go     # From service/content.go
│   ├── review.go   # From service/content.go
│   ├── favorite.go # From service/content.go
│   └── like.go     # From service/content.go
├── dto/
│   ├── post.go     # From internal/dto/content.go
│   ├── review.go
│   └── favorite.go
└── routes.go
```

**Note**: Move `/user/posts`, `/user/reviews` handlers from `user.go` to `content/handler/post.go` and `content/handler/review.go`

### 3.3 Phase 3: User Domain Refactor (Week 2)

**Goal**: Split monolithic `user.go` into focused handlers

**Current**: `user.go` (733 lines) handles:
- Profile (GET/PUT /user/profile)
- Privacy (GET/PUT /user/privacy)
- Notifications (GET/PUT /user/notifications)
- Addresses (CRUD /user/addresses)
- Content lists (GET /user/posts, /user/reviews, /user/favorites, /user/likes)
- Following (GET /user/following/*, /user/followers)
- Account (POST /user/account/export, DELETE /user/account)

**Target Structure**:
```
apps/core/internal/domain/user/
├── handler/
│   ├── profile.go        # 80 lines
│   ├── privacy.go        # 50 lines
│   ├── notification.go   # 50 lines
│   ├── address.go        # 150 lines
│   ├── following.go      # 120 lines (moved from content domain)
│   └── account.go        # 40 lines
├── service/
│   ├── profile.go
│   ├── privacy.go
│   ├── notification.go
│   ├── address.go
│   └── account.go
├── dto/
│   ├── profile.go
│   ├── privacy.go
│   ├── notification.go
│   ├── address.go
│   └── account.go
└── routes.go
```

### 3.4 Phase 4: Follow Domain (Week 2)

**Goal**: Extract follow logic from profile and user handlers

**Current State**:
- `profile.go`: FollowUser, UnfollowUser, FollowMerchant, UnfollowMerchant
- `user.go`: ListFollowingUsers, ListFollowingMerchants, ListFollowers

**Target**:
```
apps/core/internal/domain/follow/
├── handler/
│   ├── user.go      # Follow/unfollow users
│   ├── merchant.go  # Follow/unfollow merchants
│   └── list.go      # List following/followers
├── service/
│   ├── user.go      # From service/follow.go
│   └── merchant.go
├── dto.go
└── routes.go
```

### 3.5 Phase 5: Profile Domain (Week 2-3)

**Goal**: Simplify profile handler to only public profile viewing

**Current**: `profile.go` (313 lines) handles:
- GetPublicProfile
- ListUserPosts (delegate to content service)
- ListUserReviews (delegate to content service)
- Follow operations (move to follow domain)

**Target**:
```
apps/core/internal/domain/profile/
├── handler.go  # GetPublicProfile only
├── service.go  # Profile viewing logic
├── dto.go
└── routes.go
```

### 3.6 Phase 6: Cleanup (Week 3)

**Goal**: Remove old files and finalize

**Tasks**:
1. Delete old directories:
   - `apps/core/internal/handler/`
   - `apps/core/internal/service/`
   - `apps/core/internal/dto/`
2. Update all imports
3. Run full test suite
4. Update documentation
5. Create CODEOWNERS file

## 4. Implementation Details

### 4.1 Handler Structure Pattern

Each handler file should follow this pattern:

```go
package user

import (
    "github.com/gin-gonic/gin"
    "apps/core/internal/domain/user/service"
)

type AddressHandler struct {
    svc service.AddressService
}

func NewAddressHandler(svc service.AddressService) *AddressHandler {
    return &AddressHandler{svc: svc}
}

// ListAddresses godoc
// @Summary List addresses
// ...
func (h *AddressHandler) ListAddresses(c *gin.Context) {
    // Implementation
}
```

### 4.2 Route Registration Pattern

Domain-level `routes.go`:

```go
package user

import (
    "github.com/gin-gonic/gin"
    "apps/core/internal/domain/user/handler"
    "apps/core/internal/domain/user/service"
    "apps/core/internal/middleware"
    "apps/core/internal/config"
)

func RegisterRoutes(r *gin.RouterGroup, cfg *config.Config) {
    // Initialize services
    addrSvc := service.NewAddressService()
    profileSvc := service.NewProfileService()
    
    // Initialize handlers
    addrHandler := handler.NewAddressHandler(addrSvc)
    profileHandler := handler.NewProfileHandler(profileSvc)
    
    // Register routes
    user := r.Group("/user", middleware.JWTAuth(cfg.JWT))
    {
        user.GET("/profile", profileHandler.GetProfile)
        user.PATCH("/profile", profileHandler.UpdateProfile)
        
        addresses := user.Group("/addresses")
        {
            addresses.GET("", addrHandler.ListAddresses)
            addresses.POST("", addrHandler.CreateAddress)
        }
    }
}
```

### 4.3 Central Router

```go
package router

import (
    "github.com/gin-gonic/gin"
    "apps/core/internal/config"
    auth "apps/core/internal/domain/auth"
    user "apps/core/internal/domain/user"
    profile "apps/core/internal/domain/profile"
    follow "apps/core/internal/domain/follow"
    content "apps/core/internal/domain/content"
)

func Setup(router *gin.Engine, cfg *config.Config) {
    api := router.Group(cfg.Server.APIBasePath)
    
    // Register domain routes
    auth.RegisterRoutes(api, cfg)
    user.RegisterRoutes(api, cfg)
    profile.RegisterRoutes(api, cfg)
    follow.RegisterRoutes(api, cfg)
    content.RegisterRoutes(api, cfg)
}
```

## 5. CODEOWNERS Example

After migration, create `.github/CODEOWNERS`:

```
# Global owners
* @team-lead

# Domain-specific owners
apps/core/internal/domain/auth/ @auth-team
apps/core/internal/domain/user/ @user-team
apps/core/internal/domain/profile/ @profile-team
apps/core/internal/domain/follow/ @social-team
apps/core/internal/domain/content/ @content-team

# Shared infrastructure
apps/core/internal/config/ @platform-team
apps/core/internal/middleware/ @platform-team
apps/core/pkg/ @platform-team
```

## 6. Migration Timeline

| Phase | Duration | Domains | Effort |
|-------|----------|---------|--------|
| 1 | Week 1 | auth | Low |
| 2 | Week 1-2 | content | Medium |
| 3 | Week 2 | user | High (split 733 lines) |
| 4 | Week 2 | follow | Low |
| 5 | Week 2-3 | profile | Low |
| 6 | Week 3 | cleanup | Low |

**Total Duration**: 3 weeks
**Recommended Team**: 2-3 developers working in parallel on different domains

## 7. Risk Mitigation

1. **Test Coverage**: Ensure all existing tests pass before each phase
2. **Feature Freezes**: Freeze features in domains being migrated
3. **Incremental Deployment**: Deploy domain by domain, not all at once
4. **Rollback Plan**: Keep old files until phase 6, use git for rollback
5. **Documentation**: Update API docs (Swagger) continuously

## 8. Benefits After Migration

1. **Clear Domain Boundaries**: Easy to understand and navigate
2. **Code Ownership**: Clear responsibility with CODEOWNERS
3. **Parallel Development**: Teams can work on different domains
4. **Scalability**: Easy to add new domains (e.g., `notification`, `search`)
5. **Testability**: Smaller, focused files are easier to test
6. **Maintainability**: Single responsibility per file

## 9. Appendix: Current Endpoint Mapping

| Endpoint | Current Handler | Target Domain | Target Handler |
|----------|----------------|---------------|----------------|
| POST /auth/register | auth.go | auth | handler.go |
| POST /auth/login | auth.go | auth | handler.go |
| GET /auth/login/google | auth.go | auth | handler.go |
| GET /auth/callback/google | auth.go | auth | handler.go |
| GET /auth/verify | auth.go | auth | handler.go |
| GET /auth/me | auth.go | auth | handler.go |
| GET /user/profile | user.go | user | handler/profile.go |
| PATCH /user/profile | user.go | user | handler/profile.go |
| GET /user/privacy | user.go | user | handler/privacy.go |
| PATCH /user/privacy | user.go | user | handler/privacy.go |
| GET /user/notifications | user.go | user | handler/notification.go |
| PATCH /user/notifications | user.go | user | handler/notification.go |
| GET /user/addresses | user.go | user | handler/address.go |
| POST /user/addresses | user.go | user | handler/address.go |
| PATCH /user/addresses/:id | user.go | user | handler/address.go |
| DELETE /user/addresses/:id | user.go | user | handler/address.go |
| POST /user/addresses/:id/default | user.go | user | handler/address.go |
| GET /user/posts | user.go | content | handler/post.go |
| GET /user/reviews | user.go | content | handler/review.go |
| GET /user/favorites | user.go | content | handler/favorite.go |
| GET /user/likes | user.go | content | handler/like.go |
| GET /user/following/users | user.go | follow | handler/list.go |
| GET /user/following/merchants | user.go | follow | handler/list.go |
| GET /user/followers | user.go | follow | handler/list.go |
| POST /user/account/export | user.go | user | handler/account.go |
| DELETE /user/account | user.go | user | handler/account.go |
| GET /users/:id | profile.go | profile | handler.go |
| GET /users/:id/posts | profile.go | content | handler/post.go |
| GET /users/:id/reviews | profile.go | content | handler/review.go |
| POST /users/:id/follow | profile.go | follow | handler/user.go |
| DELETE /users/:id/follow | profile.go | follow | handler/user.go |
| POST /merchants/:id/follow | profile.go | follow | handler/merchant.go |
| DELETE /merchants/:id/follow | profile.go | follow | handler/merchant.go |
| GET /test | test_handler.go | test | handler.go |
| POST /test | test_handler.go | test | handler.go |
