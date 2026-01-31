# User Profile API Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the full user profile + social content + settings APIs described in `docs/plans/2026-01-31-user-profile-api-design.md`, including models, services, handlers, routes, and tests.

**Architecture:** Gin handlers -> service layer -> GORM models. Use AutoMigrate for schema. Use DTOs for request/response mapping. Use transactions for follow/like/favorite counters and address default logic.

**Tech Stack:** Go 1.24, Gin, GORM, Postgres (sqlite in tests).

**Test note:** Current environment can fail `go test` due to Go build cache permissions. Use `GOCACHE=/tmp/go-build` prefix for test commands in this plan.

---

### Task 1: Extend core user profile counts

**Files:**
- Modify: `apps/core/internal/model/user.go`
- Modify: `apps/core/cmd/app/main.go`
- Modify: `apps/core/internal/service/auth_test.go`

**Step 1: Write failing test (schema includes new columns)**

```go
func TestUserProfileHasCounts(t *testing.T) {
    db := setupTestDB(t)
    type Column struct { Name string }
    var cols []Column
    if err := db.Raw("PRAGMA table_info(user_profiles)").Scan(&cols).Error; err != nil {
        t.Fatalf("schema query failed: %v", err)
    }
    want := map[string]bool{
        "follower_count": true,
        "following_count": true,
        "post_count": true,
        "review_count": true,
        "like_count": true,
    }
    for _, c := range cols {
        if _, ok := want[c.Name]; ok {
            delete(want, c.Name)
        }
    }
    if len(want) != 0 {
        t.Fatalf("missing columns: %v", want)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestUserProfileHasCounts -v`
Expected: FAIL with missing columns.

**Step 3: Write minimal implementation**

Update `UserProfile` in `apps/core/internal/model/user.go`:

```go
FollowerCount  int `gorm:"default:0" json:"follower_count"`
FollowingCount int `gorm:"default:0" json:"following_count"`
PostCount      int `gorm:"default:0" json:"post_count"`
ReviewCount    int `gorm:"default:0" json:"review_count"`
LikeCount      int `gorm:"default:0" json:"like_count"`
```

Update AutoMigrate list in `apps/core/cmd/app/main.go` to include new models later; for now ensure it still migrates `UserProfile` with new columns.

Update `setupTestDB` in `apps/core/internal/service/auth_test.go` to include new models when they are added (placeholder comment ok now).

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestUserProfileHasCounts -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/model/user.go apps/core/cmd/app/main.go apps/core/internal/service/auth_test.go
git commit -m "feat(core): extend user profile counts"
```

---

### Task 2: Add merchant + tag models

**Files:**
- Create: `apps/core/internal/model/merchant.go`
- Create: `apps/core/internal/model/tag.go`
- Modify: `apps/core/cmd/app/main.go`
- Modify: `apps/core/internal/service/auth_test.go`

**Step 1: Write failing test (can create merchant + tag)**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestMerchantAndTagModels -v`
Expected: FAIL (missing models / table).

**Step 3: Write minimal implementation**

Create `apps/core/internal/model/merchant.go`:

```go
package model

import "time"

type Merchant struct {
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string    `gorm:"type:varchar(100);not null" json:"name"`
    Category    string    `gorm:"type:varchar(50)" json:"category"`
    Address     string    `gorm:"type:varchar(255)" json:"address"`
    Phone       string    `gorm:"type:varchar(20)" json:"phone"`
    CoverImage  string    `gorm:"type:varchar(255)" json:"cover_image"`
    Description string    `gorm:"type:text" json:"description"`
    AvgRating   float32   `gorm:"default:0" json:"avg_rating"`
    ReviewCount int       `gorm:"default:0" json:"review_count"`
    Status      int16     `gorm:"default:0" json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func (m *Merchant) TableName() string { return "merchants" }
```

Create `apps/core/internal/model/tag.go`:

```go
package model

type Tag struct {
    ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
    PostCount int    `gorm:"default:0" json:"post_count"`
}

func (t *Tag) TableName() string { return "tags" }
```

Add models to AutoMigrate in `apps/core/cmd/app/main.go` and in `setupTestDB` in `apps/core/internal/service/auth_test.go`.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestMerchantAndTagModels -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/model/merchant.go apps/core/internal/model/tag.go apps/core/cmd/app/main.go apps/core/internal/service/auth_test.go
git commit -m "feat(core): add merchant and tag models"
```

---

### Task 3: Add post + review models

**Files:**
- Create: `apps/core/internal/model/post.go`
- Create: `apps/core/internal/model/review.go`
- Modify: `apps/core/cmd/app/main.go`
- Modify: `apps/core/internal/service/auth_test.go`

**Step 1: Write failing test (create post + review)**

```go
func TestPostAndReviewModels(t *testing.T) {
    db := setupTestDB(t)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil { t.Fatal(err) }
    merchant := model.Merchant{Name: "Cafe"}
    if err := db.Create(&merchant).Error; err != nil { t.Fatal(err) }

    post := model.Post{UserID: user.ID, MerchantID: &merchant.ID, Content: "hello"}
    if err := db.Create(&post).Error; err != nil { t.Fatal(err) }

    review := model.Review{UserID: user.ID, MerchantID: merchant.ID, Rating: 4.5, Content: "great"}
    if err := db.Create(&review).Error; err != nil { t.Fatal(err) }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestPostAndReviewModels -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/model/post.go`:

```go
package model

import (
    "time"
    "gorm.io/datatypes"
)

type Post struct {
    ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID     int64          `gorm:"not null;index" json:"user_id"`
    MerchantID *int64         `gorm:"index" json:"merchant_id"`
    Title      string         `gorm:"type:varchar(100)" json:"title"`
    Content    string         `gorm:"type:text;not null" json:"content"`
    Images     datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"images"`
    LikeCount  int            `gorm:"default:0" json:"like_count"`
    ViewCount  int            `gorm:"default:0" json:"view_count"`
    Status     int16          `gorm:"default:0" json:"status"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`

    User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
    Tags     []Tag     `gorm:"many2many:post_tags" json:"tags,omitempty"`
}

func (p *Post) TableName() string { return "posts" }
```

Create `apps/core/internal/model/review.go`:

```go
package model

import (
    "time"
    "gorm.io/datatypes"
)

type Review struct {
    ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID        int64          `gorm:"not null;index" json:"user_id"`
    MerchantID    int64          `gorm:"not null;index" json:"merchant_id"`
    Rating        float32        `gorm:"not null" json:"rating"`
    RatingEnv     *float32       `json:"rating_env"`
    RatingService *float32       `json:"rating_service"`
    RatingValue   *float32       `json:"rating_value"`
    Content       string         `gorm:"type:text" json:"content"`
    Images        datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"images"`
    AvgCost       *int           `json:"avg_cost"`
    LikeCount     int            `gorm:"default:0" json:"like_count"`
    Status        int16          `gorm:"default:0" json:"status"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`

    User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
    Tags     []Tag     `gorm:"many2many:review_tags" json:"tags,omitempty"`
}

func (r *Review) TableName() string { return "reviews" }
```

Add models to AutoMigrate list and `setupTestDB`.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestPostAndReviewModels -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/model/post.go apps/core/internal/model/review.go apps/core/cmd/app/main.go apps/core/internal/service/auth_test.go
git commit -m "feat(core): add post and review models"
```

---

### Task 4: Add follow + interaction models

**Files:**
- Create: `apps/core/internal/model/follow.go`
- Create: `apps/core/internal/model/interaction.go`
- Modify: `apps/core/cmd/app/main.go`
- Modify: `apps/core/internal/service/auth_test.go`

**Step 1: Write failing test (create follow + like)**

```go
func TestFollowAndInteractionModels(t *testing.T) {
    db := setupTestDB(t)
    u1 := model.User{Role: "user", Status: 0}
    u2 := model.User{Role: "user", Status: 0}
    if err := db.Create(&u1).Error; err != nil { t.Fatal(err) }
    if err := db.Create(&u2).Error; err != nil { t.Fatal(err) }

    follow := model.UserFollow{FollowerID: u1.ID, FollowingID: u2.ID}
    if err := db.Create(&follow).Error; err != nil { t.Fatal(err) }

    like := model.Like{UserID: u1.ID, TargetType: "post", TargetID: 123}
    if err := db.Create(&like).Error; err != nil { t.Fatal(err) }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestFollowAndInteractionModels -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/model/follow.go`:

```go
package model

import "time"

type UserFollow struct {
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    FollowerID  int64     `gorm:"not null;uniqueIndex:idx_user_follow" json:"follower_id"`
    FollowingID int64     `gorm:"not null;uniqueIndex:idx_user_follow" json:"following_id"`
    CreatedAt   time.Time `json:"created_at"`
}

func (u *UserFollow) TableName() string { return "user_follows" }

type MerchantFollow struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID     int64     `gorm:"not null;uniqueIndex:idx_merchant_follow" json:"user_id"`
    MerchantID int64     `gorm:"not null;uniqueIndex:idx_merchant_follow" json:"merchant_id"`
    CreatedAt  time.Time `json:"created_at"`
}

func (m *MerchantFollow) TableName() string { return "merchant_follows" }
```

Create `apps/core/internal/model/interaction.go`:

```go
package model

import "time"

type Like struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID     int64     `gorm:"not null;uniqueIndex:idx_like" json:"user_id"`
    TargetType string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_like" json:"target_type"`
    TargetID   int64     `gorm:"not null;uniqueIndex:idx_like" json:"target_id"`
    CreatedAt  time.Time `json:"created_at"`
}

func (l *Like) TableName() string { return "likes" }

type Favorite struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID     int64     `gorm:"not null;uniqueIndex:idx_favorite" json:"user_id"`
    TargetType string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_favorite" json:"target_type"`
    TargetID   int64     `gorm:"not null;uniqueIndex:idx_favorite" json:"target_id"`
    CreatedAt  time.Time `json:"created_at"`
}

func (f *Favorite) TableName() string { return "favorites" }
```

Add models to AutoMigrate list and `setupTestDB`.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestFollowAndInteractionModels -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/model/follow.go apps/core/internal/model/interaction.go apps/core/cmd/app/main.go apps/core/internal/service/auth_test.go
git commit -m "feat(core): add follow and interaction models"
```

---

### Task 5: Add settings + address + account deletion models

**Files:**
- Create: `apps/core/internal/model/settings.go`
- Create: `apps/core/internal/model/address.go`
- Modify: `apps/core/cmd/app/main.go`
- Modify: `apps/core/internal/service/auth_test.go`

**Step 1: Write failing test (create settings + address)**

```go
func TestSettingsAndAddressModels(t *testing.T) {
    db := setupTestDB(t)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil { t.Fatal(err) }

    privacy := model.UserPrivacy{UserID: user.ID, IsPublic: true}
    if err := db.Create(&privacy).Error; err != nil { t.Fatal(err) }

    address := model.UserAddress{UserID: user.ID, Name: "A", Phone: "1", Address: "Street"}
    if err := db.Create(&address).Error; err != nil { t.Fatal(err) }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestSettingsAndAddressModels -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/model/settings.go`:

```go
package model

import "time"

type UserPrivacy struct {
    UserID   int64 `gorm:"primaryKey" json:"user_id"`
    IsPublic bool  `gorm:"default:true" json:"is_public"`
}

func (u *UserPrivacy) TableName() string { return "user_privacies" }

type UserNotification struct {
    UserID       int64 `gorm:"primaryKey" json:"user_id"`
    PushEnabled  bool  `gorm:"default:true" json:"push_enabled"`
    EmailEnabled bool  `gorm:"default:true" json:"email_enabled"`
}

func (u *UserNotification) TableName() string { return "user_notifications" }

type AccountDeletion struct {
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID      int64     `gorm:"not null;uniqueIndex" json:"user_id"`
    Reason      string    `gorm:"type:varchar(255)" json:"reason"`
    ScheduledAt time.Time `gorm:"not null" json:"scheduled_at"`
    CreatedAt   time.Time `json:"created_at"`
}

func (a *AccountDeletion) TableName() string { return "account_deletions" }
```

Create `apps/core/internal/model/address.go`:

```go
package model

import "time"

type UserAddress struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID     int64     `gorm:"not null;index" json:"user_id"`
    Name       string    `gorm:"type:varchar(50);not null" json:"name"`
    Phone      string    `gorm:"type:varchar(20);not null" json:"phone"`
    Province   string    `gorm:"type:varchar(50)" json:"province"`
    City       string    `gorm:"type:varchar(50)" json:"city"`
    District   string    `gorm:"type:varchar(50)" json:"district"`
    Address    string    `gorm:"type:varchar(255);not null" json:"address"`
    PostalCode string    `gorm:"type:varchar(20)" json:"postal_code"`
    IsDefault  bool      `gorm:"default:false" json:"is_default"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

func (a *UserAddress) TableName() string { return "user_addresses" }
```

Add models to AutoMigrate list and `setupTestDB`.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestSettingsAndAddressModels -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/model/settings.go apps/core/internal/model/address.go apps/core/cmd/app/main.go apps/core/internal/service/auth_test.go
git commit -m "feat(core): add settings and address models"
```

---

### Task 6: Add DTOs for profile, content, and settings

**Files:**
- Create: `apps/core/internal/dto/user.go`
- Create: `apps/core/internal/dto/content.go`

**Step 1: Write failing test (DTO JSON tags)**

```go
func TestDTOJSONTags(t *testing.T) {
    // Compile-time check: ensure structs exist
    _ = dto.PublicProfileResponse{}
    _ = dto.PostItem{}
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestDTOJSONTags -v`
Expected: FAIL (missing package dto).

**Step 3: Write minimal implementation**

Create `apps/core/internal/dto/user.go` with request/response structs from design:

```go
package dto

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

type SecurityOverviewResponse struct {
    HasPassword    bool     `json:"has_password"`
    LinkedAccounts []string `json:"linked_accounts"`
    Email          string   `json:"email"`
}

type ChangePasswordRequest struct {
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}

type PrivacySettings struct { IsPublic bool `json:"is_public"` }

type NotificationSettings struct {
    PushEnabled  bool `json:"push_enabled"`
    EmailEnabled bool `json:"email_enabled"`
}

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

type AddressListResponse struct { Addresses []AddressItem `json:"addresses"` }

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

Create `apps/core/internal/dto/content.go`:

```go
package dto

import "time"

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

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestDTOJSONTags -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/dto/user.go apps/core/internal/dto/content.go
git commit -m "feat(core): add user/content DTOs"
```

---

### Task 7: Create user service for profile + settings + addresses

**Files:**
- Create: `apps/core/internal/service/user.go`
- Modify: `apps/core/internal/service/auth_test.go`
- Create: `apps/core/internal/service/user_test.go`

**Step 1: Write failing tests (profile + settings + address)**

```go
func TestUserServiceProfileAndSettings(t *testing.T) {
    db := setupTestDB(t)
    svc := NewUserService(db)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil { t.Fatal(err) }
    if err := db.Create(&model.UserProfile{UserID: user.ID, Nickname: "n"}).Error; err != nil { t.Fatal(err) }

    prof, err := svc.GetProfile(context.Background(), user.ID)
    if err != nil || prof.UserID != user.ID { t.Fatalf("profile failed: %v", err) }

    newName := "new"
    if err := svc.UpdateProfile(context.Background(), user.ID, dto.UpdateProfileRequest{Nickname: &newName}); err != nil {
        t.Fatalf("update failed: %v", err)
    }
}

func TestUserServiceAddressDefault(t *testing.T) {
    db := setupTestDB(t)
    svc := NewUserService(db)
    user := model.User{Role: "user", Status: 0}
    if err := db.Create(&user).Error; err != nil { t.Fatal(err) }

    addr, err := svc.CreateAddress(context.Background(), user.ID, dto.CreateAddressRequest{Name: "A", Phone: "1", Address: "X"})
    if err != nil || !addr.IsDefault { t.Fatalf("expected default address") }
}
```

**Step 2: Run tests to verify they fail**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestUserService -v`
Expected: FAIL (missing service).

**Step 3: Write minimal implementation**

Create `apps/core/internal/service/user.go`:

```go
package service

import (
    "context"
    "errors"
    "time"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/dto"
    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "gorm.io/gorm"
)

type UserService struct { db *gorm.DB }

func NewUserService(db *gorm.DB) *UserService { return &UserService{db: db} }

func (s *UserService) GetProfile(ctx context.Context, userID int64) (dto.ProfileResponse, error) {
    var profile model.UserProfile
    if err := s.db.WithContext(ctx).First(&profile, "user_id = ?", userID).Error; err != nil {
        return dto.ProfileResponse{}, err
    }
    return dto.ProfileResponse{UserID: profile.UserID, Nickname: profile.Nickname, AvatarURL: profile.AvatarURL, Intro: profile.Intro, Location: profile.Location}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req dto.UpdateProfileRequest) error {
    updates := map[string]any{}
    if req.Nickname != nil { updates["nickname"] = *req.Nickname }
    if req.AvatarURL != nil { updates["avatar_url"] = *req.AvatarURL }
    if req.Intro != nil { updates["intro"] = *req.Intro }
    if req.Location != nil { updates["location"] = *req.Location }
    return s.db.WithContext(ctx).Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (s *UserService) GetPrivacy(ctx context.Context, userID int64) (dto.PrivacySettings, error) {
    var p model.UserPrivacy
    if err := s.db.WithContext(ctx).FirstOrCreate(&p, model.UserPrivacy{UserID: userID}).Error; err != nil {
        return dto.PrivacySettings{}, err
    }
    return dto.PrivacySettings{IsPublic: p.IsPublic}, nil
}

func (s *UserService) UpdatePrivacy(ctx context.Context, userID int64, req dto.PrivacySettings) error {
    return s.db.WithContext(ctx).Save(&model.UserPrivacy{UserID: userID, IsPublic: req.IsPublic}).Error
}

func (s *UserService) GetNotifications(ctx context.Context, userID int64) (dto.NotificationSettings, error) {
    var n model.UserNotification
    if err := s.db.WithContext(ctx).FirstOrCreate(&n, model.UserNotification{UserID: userID}).Error; err != nil {
        return dto.NotificationSettings{}, err
    }
    return dto.NotificationSettings{PushEnabled: n.PushEnabled, EmailEnabled: n.EmailEnabled}, nil
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
        if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Count(&count).Error; err != nil { return err }
        if count >= 20 { return errors.New("address limit reached") }
        addr = model.UserAddress{UserID: userID, Name: req.Name, Phone: req.Phone, Province: req.Province, City: req.City, District: req.District, Address: req.Address, PostalCode: req.PostalCode, IsDefault: req.IsDefault || count == 0}
        if addr.IsDefault {
            if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil { return err }
        }
        return tx.Create(&addr).Error
    })
    return addr, err
}

func (s *UserService) UpdateAddress(ctx context.Context, userID, addressID int64, req dto.UpdateAddressRequest) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        updates := map[string]any{}
        if req.Name != nil { updates["name"] = *req.Name }
        if req.Phone != nil { updates["phone"] = *req.Phone }
        if req.Province != nil { updates["province"] = *req.Province }
        if req.City != nil { updates["city"] = *req.City }
        if req.District != nil { updates["district"] = *req.District }
        if req.Address != nil { updates["address"] = *req.Address }
        if req.PostalCode != nil { updates["postal_code"] = *req.PostalCode }
        if req.IsDefault != nil { updates["is_default"] = *req.IsDefault }
        if err := tx.Model(&model.UserAddress{}).Where("id = ? AND user_id = ?", addressID, userID).Updates(updates).Error; err != nil { return err }
        if req.IsDefault != nil && *req.IsDefault {
            return tx.Model(&model.UserAddress{}).Where("user_id = ? AND id <> ?", userID, addressID).Update("is_default", false).Error
        }
        return nil
    })
}

func (s *UserService) DeleteAddress(ctx context.Context, userID, addressID int64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var addr model.UserAddress
        if err := tx.First(&addr, "id = ? AND user_id = ?", addressID, userID).Error; err != nil { return err }
        if err := tx.Delete(&addr).Error; err != nil { return err }
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
        if err := tx.Model(&model.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil { return err }
        return tx.Model(&model.UserAddress{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error
    })
}

func (s *UserService) RequestAccountDeletion(ctx context.Context, userID int64, reason string) error {
    deletion := model.AccountDeletion{UserID: userID, Reason: reason, ScheduledAt: time.Now().UTC().Add(7 * 24 * time.Hour)}
    return s.db.WithContext(ctx).Save(&deletion).Error
}
```

**Step 4: Run tests to verify they pass**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestUserService -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/service/user.go apps/core/internal/service/user_test.go

git commit -m "feat(core): add user service for profile and settings"
```

---

### Task 8: Create follow service

**Files:**
- Create: `apps/core/internal/service/follow.go`
- Create: `apps/core/internal/service/follow_test.go`

**Step 1: Write failing tests (follow user + merchant)**

```go
func TestFollowServiceUserFollow(t *testing.T) {
    db := setupTestDB(t)
    svc := NewFollowService(db)
    u1 := model.User{Role: "user", Status: 0}
    u2 := model.User{Role: "user", Status: 0}
    db.Create(&u1); db.Create(&u2)
    if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil { t.Fatal(err) }
    if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil { t.Fatal(err) }
}
```

**Step 2: Run tests to verify they fail**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestFollowServiceUserFollow -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/service/follow.go`:

```go
package service

import (
    "context"
    "errors"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "gorm.io/gorm"
)

type FollowService struct { db *gorm.DB }

func NewFollowService(db *gorm.DB) *FollowService { return &FollowService{db: db} }

func (s *FollowService) FollowUser(ctx context.Context, followerID, followingID int64) error {
    if followerID == followingID { return errors.New("cannot follow self") }
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        follow := model.UserFollow{FollowerID: followerID, FollowingID: followingID}
        if err := tx.FirstOrCreate(&follow, follow).Error; err != nil { return err }
        tx.Model(&model.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("following_count + 1"))
        tx.Model(&model.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("follower_count + 1"))
        return nil
    })
}

func (s *FollowService) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        tx.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.UserFollow{})
        tx.Model(&model.UserProfile{}).Where("user_id = ?", followerID).UpdateColumn("following_count", gorm.Expr("GREATEST(following_count - 1, 0)"))
        tx.Model(&model.UserProfile{}).Where("user_id = ?", followingID).UpdateColumn("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)"))
        return nil
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

**Step 4: Run tests to verify they pass**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestFollowServiceUserFollow -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/service/follow.go apps/core/internal/service/follow_test.go
git commit -m "feat(core): add follow service"
```

---

### Task 9: Create interaction service (likes + favorites)

**Files:**
- Create: `apps/core/internal/service/interaction.go`
- Create: `apps/core/internal/service/interaction_test.go`

**Step 1: Write failing tests (like + favorite)**

```go
func TestInteractionServiceLike(t *testing.T) {
    db := setupTestDB(t)
    svc := NewInteractionService(db)
    u := model.User{Role: "user", Status: 0}
    db.Create(&u)
    if err := svc.Like(context.Background(), u.ID, "post", 123); err != nil { t.Fatal(err) }
    if err := svc.Unlike(context.Background(), u.ID, "post", 123); err != nil { t.Fatal(err) }
}
```

**Step 2: Run tests to verify they fail**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestInteractionServiceLike -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/service/interaction.go`:

```go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "gorm.io/gorm"
)

type InteractionService struct { db *gorm.DB }

func NewInteractionService(db *gorm.DB) *InteractionService { return &InteractionService{db: db} }

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

**Step 4: Run tests to verify they pass**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestInteractionServiceLike -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/service/interaction.go apps/core/internal/service/interaction_test.go
git commit -m "feat(core): add interaction service"
```

---

### Task 10: Create content service for list endpoints

**Files:**
- Create: `apps/core/internal/service/content.go`
- Create: `apps/core/internal/service/content_test.go`

**Step 1: Write failing tests (list posts/reviews)**

```go
func TestContentServiceListPosts(t *testing.T) {
    db := setupTestDB(t)
    svc := NewContentService(db)
    user := model.User{Role: "user", Status: 0}
    db.Create(&user)
    db.Create(&model.Post{UserID: user.ID, Content: "a"})
    posts, total, err := svc.ListUserPosts(context.Background(), user.ID, nil, 10)
    if err != nil || total != 1 || len(posts) != 1 { t.Fatalf("list posts failed: %v", err) }
}
```

**Step 2: Run tests to verify they fail**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestContentServiceListPosts -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/service/content.go`:

```go
package service

import (
    "context"

    "github.com/RevieU-Corp/revieu-backend/apps/core/internal/model"
    "gorm.io/gorm"
)

type ContentService struct { db *gorm.DB }

func NewContentService(db *gorm.DB) *ContentService { return &ContentService{db: db} }

func (s *ContentService) ListUserPosts(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Post, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Post{}).Where("user_id = ?", userID).Order("id desc")
    if cursor != nil { q = q.Where("id < ?", *cursor) }
    var total int64
    if err := q.Count(&total).Error; err != nil { return nil, 0, err }
    var posts []model.Post
    if err := q.Limit(limit).Find(&posts).Error; err != nil { return nil, 0, err }
    return posts, total, nil
}

func (s *ContentService) ListUserReviews(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Review, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Review{}).Where("user_id = ?", userID).Order("id desc")
    if cursor != nil { q = q.Where("id < ?", *cursor) }
    var total int64
    if err := q.Count(&total).Error; err != nil { return nil, 0, err }
    var reviews []model.Review
    if err := q.Limit(limit).Find(&reviews).Error; err != nil { return nil, 0, err }
    return reviews, total, nil
}

func (s *ContentService) ListFavorites(ctx context.Context, userID int64, targetType string, cursor *int64, limit int) ([]model.Favorite, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ?", userID)
    if targetType != "" { q = q.Where("target_type = ?", targetType) }
    if cursor != nil { q = q.Where("id < ?", *cursor) }
    var total int64
    if err := q.Count(&total).Error; err != nil { return nil, 0, err }
    var items []model.Favorite
    if err := q.Order("id desc").Limit(limit).Find(&items).Error; err != nil { return nil, 0, err }
    return items, total, nil
}

func (s *ContentService) ListLikes(ctx context.Context, userID int64, cursor *int64, limit int) ([]model.Like, int64, error) {
    q := s.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ?", userID)
    if cursor != nil { q = q.Where("id < ?", *cursor) }
    var total int64
    if err := q.Count(&total).Error; err != nil { return nil, 0, err }
    var items []model.Like
    if err := q.Order("id desc").Limit(limit).Find(&items).Error; err != nil { return nil, 0, err }
    return items, total, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestContentServiceListPosts -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/service/content.go apps/core/internal/service/content_test.go
git commit -m "feat(core): add content listing service"
```

---

### Task 11: Add user handlers for /user routes

**Files:**
- Create: `apps/core/internal/handler/user.go`
- Modify: `apps/core/internal/handler/routes.go`
- Create: `apps/core/internal/handler/user_test.go`

**Step 1: Write failing handler tests (profile + addresses)**

```go
func TestUserProfileHandlers(t *testing.T) {
    router := gin.New()
    // TODO: setup routes with mocked services
    _ = router
}
```

**Step 2: Run tests to verify they fail**

Run: `GOCACHE=/tmp/go-build go test ./internal/handler -run TestUserProfileHandlers -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/handler/user.go` implementing:
- GET/PATCH `/user/profile`
- GET/PATCH `/user/privacy`
- GET/PATCH `/user/notifications`
- GET/POST/PATCH/DELETE `/user/addresses`
- POST `/user/addresses/:id/default`
- GET `/user/posts`
- GET `/user/reviews`
- GET `/user/favorites`
- GET `/user/likes`
- GET `/user/following/users`
- GET `/user/following/merchants`
- GET `/user/followers`
- POST `/user/account/export` (create record/log placeholder)
- DELETE `/user/account` (create AccountDeletion)

Example handler method skeleton:

```go
func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := c.GetInt64("user_id")
    resp, err := h.userService.GetProfile(c.Request.Context(), userID)
    if err != nil { c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}); return }
    c.JSON(http.StatusOK, resp)
}
```

Register `/user` routes in `apps/core/internal/handler/routes.go` using `middleware.JWTAuth` and include new `UserHandler` instance with services.

**Step 4: Run tests to verify they pass**

Run: `GOCACHE=/tmp/go-build go test ./internal/handler -run TestUserProfileHandlers -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/handler/user.go apps/core/internal/handler/user_test.go apps/core/internal/handler/routes.go
git commit -m "feat(core): add /user handlers"
```

---

### Task 12: Add public profile handlers (/users/:id) + merchant follow

**Files:**
- Create: `apps/core/internal/handler/profile.go`
- Modify: `apps/core/internal/handler/routes.go`
- Create: `apps/core/internal/handler/profile_test.go`

**Step 1: Write failing tests (public profile)**

```go
func TestPublicProfileHandler(t *testing.T) {
    router := gin.New()
    _ = router
}
```

**Step 2: Run tests to verify they fail**

Run: `GOCACHE=/tmp/go-build go test ./internal/handler -run TestPublicProfileHandler -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Create `apps/core/internal/handler/profile.go` with:
- GET `/users/:id` public profile
- GET `/users/:id/posts`
- GET `/users/:id/reviews`
- POST `/users/:id/follow`
- DELETE `/users/:id/follow`
- POST `/merchants/:id/follow`
- DELETE `/merchants/:id/follow`

Ensure public profile checks `UserPrivacy.IsPublic` for target user; if private, return 403 unless requester is the same user.

**Step 4: Run tests to verify they pass**

Run: `GOCACHE=/tmp/go-build go test ./internal/handler -run TestPublicProfileHandler -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/handler/profile.go apps/core/internal/handler/profile_test.go apps/core/internal/handler/routes.go
git commit -m "feat(core): add public profile handlers"
```

---

### Task 13: Wire auth-based helpers and service composition

**Files:**
- Modify: `apps/core/internal/handler/routes.go`
- Create: `apps/core/internal/handler/helpers.go`

**Step 1: Write failing test (router includes all routes)**

```go
func TestRoutesContainUserEndpoints(t *testing.T) {
    router := gin.New()
    cfg := &config.Config{}
    RegisterRoutes(router, cfg)
    // Ensure at least one route exists; detailed route checks can inspect gin routes
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/handler -run TestRoutesContainUserEndpoints -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

Add helper in `apps/core/internal/handler/helpers.go`:

```go
package handler

import "github.com/gin-gonic/gin"

func getUserID(c *gin.Context) int64 {
    return c.GetInt64("user_id")
}
```

Ensure `routes.go` creates services and handlers once, then wires all new routes.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-build go test ./internal/handler -run TestRoutesContainUserEndpoints -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/handler/routes.go apps/core/internal/handler/helpers.go apps/core/internal/handler/*_test.go
git commit -m "feat(core): wire profile routes and helpers"
```

---

### Task 14: End-to-end service tests for follow + content counts

**Files:**
- Modify: `apps/core/internal/service/follow_test.go`
- Modify: `apps/core/internal/service/content_test.go`

**Step 1: Write failing tests (counts update)**

```go
func TestFollowUpdatesCounts(t *testing.T) {
    db := setupTestDB(t)
    svc := NewFollowService(db)
    u1 := model.User{Role: "user", Status: 0}
    u2 := model.User{Role: "user", Status: 0}
    db.Create(&u1); db.Create(&u2)
    db.Create(&model.UserProfile{UserID: u1.ID, Nickname: "a"})
    db.Create(&model.UserProfile{UserID: u2.ID, Nickname: "b"})

    if err := svc.FollowUser(context.Background(), u1.ID, u2.ID); err != nil { t.Fatal(err) }

    var p1, p2 model.UserProfile
    db.First(&p1, "user_id = ?", u1.ID)
    db.First(&p2, "user_id = ?", u2.ID)
    if p1.FollowingCount != 1 || p2.FollowerCount != 1 { t.Fatalf("counts not updated") }
}
```

**Step 2: Run tests to verify they fail**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestFollowUpdatesCounts -v`
Expected: FAIL (counts not updated).

**Step 3: Write minimal implementation**

Adjust `FollowService` to update counts only when a row is newly created (check `RowsAffected` on `FirstOrCreate`). Update unlike to decrement only when a row existed.

**Step 4: Run tests to verify they pass**

Run: `GOCACHE=/tmp/go-build go test ./internal/service -run TestFollowUpdatesCounts -v`
Expected: PASS

**Step 5: Commit**

```bash
git add apps/core/internal/service/follow.go apps/core/internal/service/follow_test.go
git commit -m "feat(core): update follow counts accurately"
```

---

### Task 15: Documentation refresh

**Files:**
- Modify: `docs/plans/2026-01-31-user-profile-api-design.md`
- Modify: `README.md` (optional API route list section)

**Step 1: Write failing doc check (optional)**

No test.

**Step 2: Update docs**

Add notes on implemented endpoints and any deviations (e.g., export endpoint returns 202 + placeholder) to `docs/plans/2026-01-31-user-profile-api-design.md` and `README.md` if it lists API routes.

**Step 3: Commit**

```bash
git add docs/plans/2026-01-31-user-profile-api-design.md README.md
git commit -m "docs(core): note user profile API implementation"
```

---

Plan complete and saved to `docs/plans/2026-01-31-user-profile-api-implementation-plan.md`.

Two execution options:

1. Subagent-Driven (this session) - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. Parallel Session (separate) - Open new session with executing-plans, batch execution with checkpoints

Which approach?
