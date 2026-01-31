# User Profile API Design

## Overview

用户 Profile 系统 API 设计，支持类似小红书+大众点评风格的社交内容+点评应用。

**本次设计范围**：
- 核心数据模型：Merchant、Post、Review、Tag、关注/点赞/收藏
- Profile API：公开主页、关注系统、内容列表、收藏/点赞、账号设置

**后续设计**：
- E-Coupon 优惠券系统
- Task 任务系统（写评价任务、发帖任务、消费返券等）

---

## 数据模型

### Merchant 商家表

```go
type Merchant struct {
    ID          int64     `gorm:"primaryKey;autoIncrement"`
    Name        string    `gorm:"type:varchar(100);not null"`        // 商家名称
    Category    string    `gorm:"type:varchar(50)"`                  // 分类：餐饮、美容、娱乐等
    Address     string    `gorm:"type:varchar(255)"`                 // 地址
    Phone       string    `gorm:"type:varchar(20)"`                  // 联系电话
    CoverImage  string    `gorm:"type:varchar(255)"`                 // 封面图
    Description string    `gorm:"type:text"`                         // 商家简介
    AvgRating   float32   `gorm:"default:0"`                         // 平均评分（冗余字段）
    ReviewCount int       `gorm:"default:0"`                         // 评价数量（冗余字段）
    Status      int16     `gorm:"default:0"`                         // 0:正常 1:关闭 2:审核中
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Post 帖子表

```go
type Post struct {
    ID         int64     `gorm:"primaryKey;autoIncrement"`
    UserID     int64     `gorm:"not null;index"`                    // 发布者
    MerchantID *int64    `gorm:"index"`                             // 关联商家（可选）
    Title      string    `gorm:"type:varchar(100)"`                 // 标题（可选）
    Content    string    `gorm:"type:text;not null"`                // 正文
    Images     string    `gorm:"type:jsonb;default:'[]'"`           // 图片URL数组
    LikeCount  int       `gorm:"default:0"`                         // 点赞数（冗余）
    ViewCount  int       `gorm:"default:0"`                         // 浏览数（冗余）
    Status     int16     `gorm:"default:0"`                         // 0:正常 1:隐藏 2:审核中
    CreatedAt  time.Time
    UpdatedAt  time.Time

    // Relations
    User     *User     `gorm:"foreignKey:UserID"`
    Merchant *Merchant `gorm:"foreignKey:MerchantID"`
    Tags     []Tag     `gorm:"many2many:post_tags"`
}
```

### Tag 标签表

```go
type Tag struct {
    ID        int64  `gorm:"primaryKey;autoIncrement"`
    Name      string `gorm:"type:varchar(50);uniqueIndex;not null"` // 标签名，如 #美食
    PostCount int    `gorm:"default:0"`                             // 使用次数（冗余）
}
```

### Review 商家评价表

```go
type Review struct {
    ID            int64     `gorm:"primaryKey;autoIncrement"`
    UserID        int64     `gorm:"not null;index"`                   // 评价者
    MerchantID    int64     `gorm:"not null;index"`                   // 被评价商家
    Rating        float32   `gorm:"not null"`                         // 总评分 1.0-5.0
    RatingEnv     *float32  `gorm:""`                                 // 环境评分（可选）
    RatingService *float32  `gorm:""`                                 // 服务评分（可选）
    RatingValue   *float32  `gorm:""`                                 // 性价比评分（可选）
    Content       string    `gorm:"type:text"`                        // 评价内容
    Images        string    `gorm:"type:jsonb;default:'[]'"`          // 图片URL数组
    AvgCost       *int      `gorm:""`                                 // 人均消费（可选）
    LikeCount     int       `gorm:"default:0"`                        // 点赞数（冗余）
    Status        int16     `gorm:"default:0"`                        // 0:正常 1:隐藏 2:审核中
    CreatedAt     time.Time
    UpdatedAt     time.Time

    // Relations
    User     *User     `gorm:"foreignKey:UserID"`
    Merchant *Merchant `gorm:"foreignKey:MerchantID"`
    Tags     []Tag     `gorm:"many2many:review_tags"`
}
```

### 关注系统

```go
// UserFollow 用户关注表
type UserFollow struct {
    ID          int64     `gorm:"primaryKey;autoIncrement"`
    FollowerID  int64     `gorm:"not null;uniqueIndex:idx_user_follow"`  // 关注者
    FollowingID int64     `gorm:"not null;uniqueIndex:idx_user_follow"`  // 被关注者
    CreatedAt   time.Time
}

// MerchantFollow 商家关注表
type MerchantFollow struct {
    ID         int64     `gorm:"primaryKey;autoIncrement"`
    UserID     int64     `gorm:"not null;uniqueIndex:idx_merchant_follow"`
    MerchantID int64     `gorm:"not null;uniqueIndex:idx_merchant_follow"`
    CreatedAt  time.Time
}
```

### 互动系统

```go
// Like 点赞表（通用）
type Like struct {
    ID         int64     `gorm:"primaryKey;autoIncrement"`
    UserID     int64     `gorm:"not null;uniqueIndex:idx_like"`
    TargetType string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_like"` // "post" 或 "review"
    TargetID   int64     `gorm:"not null;uniqueIndex:idx_like"`
    CreatedAt  time.Time
}

// Favorite 收藏表（通用）
type Favorite struct {
    ID         int64     `gorm:"primaryKey;autoIncrement"`
    UserID     int64     `gorm:"not null;uniqueIndex:idx_favorite"`
    TargetType string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_favorite"` // "post", "review", "merchant"
    TargetID   int64     `gorm:"not null;uniqueIndex:idx_favorite"`
    CreatedAt  time.Time
}
```

### UserProfile 扩展

在现有 `UserProfile` 基础上添加统计字段：

```go
type UserProfile struct {
    UserID         int64  `gorm:"primaryKey"`
    Nickname       string `gorm:"type:varchar(50);not null"`
    AvatarURL      string `gorm:"type:varchar(255)"`
    Intro          string `gorm:"type:varchar(255)"`
    Location       string `gorm:"type:varchar(100)"`

    // 统计字段（事件驱动同步更新）
    FollowerCount  int    `gorm:"default:0"`
    FollowingCount int    `gorm:"default:0"`
    PostCount      int    `gorm:"default:0"`
    ReviewCount    int    `gorm:"default:0"`
    LikeCount      int    `gorm:"default:0"`  // 获赞数
}
```

### 账号设置相关

```go
// UserPrivacy 隐私设置
type UserPrivacy struct {
    UserID   int64 `gorm:"primaryKey"`
    IsPublic bool  `gorm:"default:true"` // 公开/私密账号
}

// UserNotification 通知设置
type UserNotification struct {
    UserID       int64 `gorm:"primaryKey"`
    PushEnabled  bool  `gorm:"default:true"`
    EmailEnabled bool  `gorm:"default:true"`
}

// UserAddress 收货地址
type UserAddress struct {
    ID         int64     `gorm:"primaryKey;autoIncrement"`
    UserID     int64     `gorm:"not null;index"`
    Name       string    `gorm:"type:varchar(50);not null"`   // 收件人
    Phone      string    `gorm:"type:varchar(20);not null"`   // 电话
    Province   string    `gorm:"type:varchar(50)"`            // 省
    City       string    `gorm:"type:varchar(50)"`            // 市
    District   string    `gorm:"type:varchar(50)"`            // 区
    Address    string    `gorm:"type:varchar(255);not null"`  // 详细地址
    PostalCode string    `gorm:"type:varchar(20)"`            // 邮编
    IsDefault  bool      `gorm:"default:false"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// AccountDeletion 账号删除请求
type AccountDeletion struct {
    ID          int64     `gorm:"primaryKey;autoIncrement"`
    UserID      int64     `gorm:"not null;uniqueIndex"`
    Reason      string    `gorm:"type:varchar(255)"`
    ScheduledAt time.Time `gorm:"not null"`  // 计划删除时间（7天后）
    CreatedAt   time.Time
}
```

---

## API 路由设计

### /user - 当前用户（全部需JWT）

```
/user
├── /profile
│   ├── GET    /                        # 获取自己的资料
│   └── PATCH  /                        # 更新资料
├── /security
│   ├── GET    /                        # 获取安全概览
│   ├── POST   /password                # 修改密码
│   ├── POST   /link/:provider          # 绑定第三方账号
│   └── DELETE /link/:provider          # 解绑第三方账号
├── /privacy
│   ├── GET    /                        # 获取隐私设置
│   └── PATCH  /                        # 更新隐私设置
├── /notifications
│   ├── GET    /                        # 获取通知设置
│   └── PATCH  /                        # 更新通知设置
├── /addresses
│   ├── GET    /                        # 获取地址列表
│   ├── POST   /                        # 添加地址
│   ├── PATCH  /:id                     # 更新地址
│   ├── DELETE /:id                     # 删除地址
│   └── POST   /:id/default             # 设为默认
├── /account
│   ├── POST   /export                  # 请求数据导出
│   └── DELETE /                        # 删除账号
├── /posts
│   └── GET    /                        # 我的帖子列表
├── /reviews
│   └── GET    /                        # 我的评价列表
├── /favorites
│   └── GET    /                        # 我的收藏列表
├── /likes
│   └── GET    /                        # 我的点赞列表
├── /following
│   ├── GET    /users                   # 关注的用户
│   └── GET    /merchants               # 关注的商家
└── /followers
    └── GET    /                        # 我的粉丝
```

### /users/:id - 查看其他用户（公开）

```
/users/:id
├── GET    /                            # 用户公开主页
├── GET    /posts                       # 用户的帖子
├── GET    /reviews                     # 用户的评价
├── POST   /follow                      # 关注（需JWT）
└── DELETE /follow                      # 取关（需JWT）
```

### /merchants/:id - 商家关注

```
/merchants/:id
├── POST   /follow                      # 关注商家（需JWT）
└── DELETE /follow                      # 取关商家（需JWT）
```

---

## Request/Response 结构

### 公开主页

```go
// GET /users/:id
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
```

### 关注系统

```go
// GET /user/following/users
type FollowingUsersResponse struct {
    Users []UserBrief `json:"users"`
    Total int         `json:"total"`
}

type UserBrief struct {
    UserID    int64  `json:"user_id"`
    Nickname  string `json:"nickname"`
    AvatarURL string `json:"avatar_url"`
    Intro     string `json:"intro"`
}

// GET /user/followers
type FollowersResponse struct {
    Users []UserBrief `json:"users"`
    Total int         `json:"total"`
}
```

### 帖子列表

```go
// GET /user/posts 或 /users/:id/posts
type PostListResponse struct {
    Posts  []PostItem `json:"posts"`
    Total  int        `json:"total"`
    Cursor *int64     `json:"cursor,omitempty"`
}

type PostItem struct {
    ID         int64          `json:"id"`
    Title      string         `json:"title"`
    Content    string         `json:"content"`
    Images     []string       `json:"images"`
    LikeCount  int            `json:"like_count"`
    ViewCount  int            `json:"view_count"`
    IsLiked    bool           `json:"is_liked"`
    Merchant   *MerchantBrief `json:"merchant,omitempty"`
    Tags       []string       `json:"tags"`
    CreatedAt  time.Time      `json:"created_at"`
}

type MerchantBrief struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Category string `json:"category"`
}
```

### 评价列表

```go
// GET /user/reviews 或 /users/:id/reviews
type ReviewListResponse struct {
    Reviews []ReviewItem `json:"reviews"`
    Total   int          `json:"total"`
    Cursor  *int64       `json:"cursor,omitempty"`
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
```

### 收藏列表

```go
// GET /user/favorites?type=post|review|merchant
type FavoriteListResponse struct {
    Items  []FavoriteItem `json:"items"`
    Total  int            `json:"total"`
    Cursor *int64         `json:"cursor,omitempty"`
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
```

### 账号设置

```go
// GET /user/security
type SecurityOverviewResponse struct {
    HasPassword    bool     `json:"has_password"`
    LinkedAccounts []string `json:"linked_accounts"`
    Email          string   `json:"email"`  // 脱敏显示
}

// POST /user/security/password
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}

// GET/PATCH /user/privacy
type PrivacySettings struct {
    IsPublic bool `json:"is_public"`
}

// GET/PATCH /user/notifications
type NotificationSettings struct {
    PushEnabled  bool `json:"push_enabled"`
    EmailEnabled bool `json:"email_enabled"`
}
```

### 地址管理

```go
// GET /user/addresses
type AddressListResponse struct {
    Addresses []AddressItem `json:"addresses"`
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

// POST /user/addresses
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

// PATCH /user/addresses/:id
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

---

## 业务逻辑 & 边界情况

### 关注系统

| 场景 | 处理方式 |
|------|----------|
| 关注自己 | 拒绝，返回错误 |
| 重复关注 | 幂等处理，返回成功 |
| 取关未关注的用户 | 幂等处理，返回成功 |
| 关注私密账号 | 允许关注（后续可扩展为需要审批） |

### 安全设置

| 场景 | 处理方式 |
|------|----------|
| 修改密码 - 无现有密码 | OAuth用户首次设置，`old_password` 不要求 |
| 修改密码 - 有现有密码 | 必须验证 `old_password` |
| 解绑OAuth - 最后一个登录方式 | 拒绝："必须保留至少一个登录方式" |
| 解绑OAuth - 只剩邮箱但无密码 | 拒绝："请先设置密码" |

### 地址管理

| 场景 | 处理方式 |
|------|----------|
| 添加第一个地址 | 自动设为默认 |
| 删除默认地址 | 自动将最早的地址设为默认 |
| 设为默认 | 取消其他地址的默认标记 |
| 地址数量限制 | 最多20个，超出返回错误 |

### 账号删除流程

```
用户请求删除
    │
    ▼
验证身份（密码/OAuth）
    │
    ▼
创建 AccountDeletion 记录
设置 scheduled_at = now + 7天
    │
    ▼
发送确认邮件
    │
    ▼
设置用户状态为 "pending_deletion"
（可登录，但显示警告）
    │
    ▼
7天内可取消
    │
    ▼
定时任务执行软删除
```

### 数据导出流程

```
用户请求导出
    │
    ▼
创建异步任务
    │
    ▼
收集用户数据（资料、地址、帖子、评价等）
    │
    ▼
生成 JSON/ZIP 文件
    │
    ▼
上传到临时存储（7天过期）
    │
    ▼
发送下载链接到用户邮箱
```

---

## 文件结构

```
apps/core/internal/
├── model/
│   ├── user.go              # 现有，扩展 UserProfile
│   ├── merchant.go          # 新增：商家模型
│   ├── post.go              # 新增：帖子模型
│   ├── review.go            # 新增：评价模型
│   ├── tag.go               # 新增：标签模型
│   ├── follow.go            # 新增：关注模型
│   ├── interaction.go       # 新增：点赞、收藏模型
│   ├── address.go           # 新增：地址模型
│   └── settings.go          # 新增：隐私、通知设置模型
├── handler/
│   ├── user.go              # 新增：/user 路由处理
│   ├── profile.go           # 新增：公开主页处理
│   └── routes.go            # 更新：注册新路由
├── service/
│   ├── user.go              # 新增：用户服务
│   ├── follow.go            # 新增：关注服务
│   └── interaction.go       # 新增：点赞收藏服务
└── dto/
    ├── user.go              # 新增：用户相关 DTO
    └── content.go           # 新增：内容相关 DTO
```

---

## 数据库新增表

```
merchants              # 商家表
posts                  # 帖子表
reviews                # 评价表
tags                   # 标签表
post_tags              # 帖子-标签关联
review_tags            # 评价-标签关联
user_follows           # 用户关注
merchant_follows       # 商家关注
likes                  # 点赞
favorites              # 收藏
user_addresses         # 收货地址
user_privacies         # 隐私设置
user_notifications     # 通知设置
account_deletions      # 账号删除请求
```

### UserProfile 表更新

新增字段：
- `follower_count`
- `following_count`
- `post_count`
- `review_count`
- `like_count`

---

## Implementation Notes (2026-01-31)

- `posts.images` and `reviews.images` are stored as JSON text (`jsonb` column) without a custom JSON type.
- Cursor pagination is wired in handlers but currently returns `null` for next cursor.
- Account export endpoint returns `202 Accepted` with a placeholder response; async pipeline still pending.
- Follow counts are updated idempotently (repeat follow does not double-count).
