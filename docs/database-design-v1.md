# RevieU Database Design v1.0

> 基于前端需求 Matrix v2.4 (2026-02-12)
> 设计原则：遵循 3NF，允许性能冗余设计
> 数据库：PostgreSQL (推荐)

## 目录

- [1. 核心实体关系图](#1-核心实体关系图)
- [2. 表结构设计](#2-表结构设计)
- [3. 索引策略](#3-索引策略)
- [4. 冗余设计说明](#4-冗余设计说明)
- [5. 数据迁移考虑](#5-数据迁移考虑)

---

## 1. 核心实体关系图

```
Users (用户)
  ├─→ UserProfiles (用户资料)
  ├─→ Reviews (评论)
  ├─→ Posts (帖子)
  ├─→ Vouchers (券码)
  ├─→ Orders (订单)
  ├─→ UserFollowing (关注关系)
  ├─→ BrowsingHistory (浏览历史)
  └─→ Messages (消息)

Merchants (商家)
  ├─→ MerchantProfiles (商家资料)
  ├─→ Stores (店铺/门店)
  │   ├─→ StoreHours (营业时间)
  │   ├─→ StoreMenus (菜单)
  │   └─→ StoreCategories (分类关联)
  ├─→ Coupons (优惠券)
  ├─→ Packages (套餐)
  ├─→ Reviews (接收评论)
  ├─→ MerchantVerifications (认证申请)
  └─→ MarketingPosts (营销帖子)

Reviews (评论)
  ├─→ ReviewMedia (评论媒体)
  ├─→ ReviewRatings (评分明细)
  ├─→ ReviewTags (标签)
  ├─→ ReviewComments (评论回复)
  └─→ ReviewLikes (点赞)

Coupons (优惠券)
  ├─→ Vouchers (券码实例)
  └─→ Orders (订单)
```

---

## 2. 表结构设计

### 2.1 认证与用户系统 (Auth & User)

#### users (用户基础表)

```sql
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,

  -- 基础信息
  email VARCHAR(255) UNIQUE NOT NULL,
  username VARCHAR(100) UNIQUE,
  password_hash VARCHAR(255), -- 可为空(OAuth用户)

  -- 认证相关
  email_verified BOOLEAN DEFAULT FALSE,
  email_verification_code VARCHAR(10),
  email_verification_expires_at TIMESTAMP,

  -- OAuth
  google_id VARCHAR(255) UNIQUE,
  google_profile JSONB, -- 存储Google返回的原始profile

  -- 角色与状态
  role VARCHAR(20) NOT NULL DEFAULT 'customer', -- customer, merchant, admin
  status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, deleted

  -- Token管理
  refresh_token_hash VARCHAR(255),
  refresh_token_expires_at TIMESTAMP,

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_login_at TIMESTAMP,
  deleted_at TIMESTAMP, -- 软删除

  -- 索引
  INDEX idx_users_email (email),
  INDEX idx_users_google_id (google_id),
  INDEX idx_users_role (role),
  INDEX idx_users_created_at (created_at)
);
```

#### user_profiles (用户资料)

```sql
CREATE TABLE user_profiles (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- 基础资料
  display_name VARCHAR(100),
  avatar_url TEXT,
  bio TEXT,

  -- 统计数据 (冗余设计，提升性能)
  review_count INT NOT NULL DEFAULT 0,
  follower_count INT NOT NULL DEFAULT 0,
  following_count INT NOT NULL DEFAULT 0,
  like_received_count INT NOT NULL DEFAULT 0,

  -- 偏好设置
  interests JSONB, -- ["Italian", "Japanese", "Dessert"]
  notification_settings JSONB,
  privacy_settings JSONB,

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_user_profiles_user_id (user_id)
);
```

---

### 2.2 商家系统 (Merchant)

#### merchants (商家账号表)

```sql
CREATE TABLE merchants (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- 商家基础信息
  business_name VARCHAR(255) NOT NULL,
  business_type VARCHAR(50), -- restaurant, cafe, bar, etc.

  -- 认证状态
  verification_status VARCHAR(20) NOT NULL DEFAULT 'pending',
  -- pending, under_review, approved, rejected
  verified_at TIMESTAMP,

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_merchants_user_id (user_id),
  INDEX idx_merchants_verification_status (verification_status)
);
```

#### merchant_profiles (商家资料)

```sql
CREATE TABLE merchant_profiles (
  id BIGSERIAL PRIMARY KEY,
  merchant_id BIGINT UNIQUE NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

  -- 基础资料
  logo_url TEXT,
  cover_image_url TEXT,
  description TEXT,

  -- 联系方式
  phone VARCHAR(20),
  email VARCHAR(255),
  website_url TEXT,
  social_links JSONB, -- {instagram: "xxx", facebook: "xxx"}

  -- 统计数据 (冗余)
  total_reviews INT NOT NULL DEFAULT 0,
  average_rating DECIMAL(3,2) DEFAULT 0.00,
  total_followers INT NOT NULL DEFAULT 0,

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_merchant_profiles_merchant_id (merchant_id)
);
```

#### stores (店铺/门店表)

```sql
CREATE TABLE stores (
  id BIGSERIAL PRIMARY KEY,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

  -- 店铺基础信息
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE, -- URL友好标识
  description TEXT,

  -- 地理位置
  address TEXT NOT NULL,
  city VARCHAR(100),
  state VARCHAR(100),
  country VARCHAR(100) DEFAULT 'China',
  postal_code VARCHAR(20),
  latitude DECIMAL(10, 8),
  longitude DECIMAL(11, 8),

  -- 图片
  images JSONB, -- [url1, url2, url3]
  menu_images JSONB,

  -- 价格区间
  price_level INT, -- 1-4 ($ - $$$$)

  -- 状态
  status VARCHAR(20) DEFAULT 'active', -- active, inactive, closed
  is_featured BOOLEAN DEFAULT FALSE,

  -- 统计 (冗余)
  review_count INT DEFAULT 0,
  average_rating DECIMAL(3,2) DEFAULT 0.00,

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_stores_merchant_id (merchant_id),
  INDEX idx_stores_slug (slug),
  INDEX idx_stores_location (latitude, longitude),
  INDEX idx_stores_city (city),
  INDEX idx_stores_status (status),
  INDEX idx_stores_is_featured (is_featured)
);
```

#### store_hours (营业时间)

```sql
CREATE TABLE store_hours (
  id BIGSERIAL PRIMARY KEY,
  store_id BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,

  day_of_week INT NOT NULL, -- 0=Sunday, 1=Monday, ..., 6=Saturday
  open_time TIME,
  close_time TIME,
  is_closed BOOLEAN DEFAULT FALSE,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(store_id, day_of_week),
  INDEX idx_store_hours_store_id (store_id)
);
```

---

### 2.3 分类与标签 (Categories & Tags)

#### categories (分类表)

```sql
CREATE TABLE categories (
  id BIGSERIAL PRIMARY KEY,

  name VARCHAR(100) NOT NULL,
  name_zh VARCHAR(100), -- 中文名称
  slug VARCHAR(100) UNIQUE NOT NULL,

  parent_id BIGINT REFERENCES categories(id), -- 支持层级分类

  icon_url TEXT,
  display_order INT DEFAULT 0,

  is_active BOOLEAN DEFAULT TRUE,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_categories_parent_id (parent_id),
  INDEX idx_categories_slug (slug)
);
```

#### store_categories (店铺-分类关联)

```sql
CREATE TABLE store_categories (
  id BIGSERIAL PRIMARY KEY,
  store_id BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,

  is_primary BOOLEAN DEFAULT FALSE, -- 是否为主分类

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(store_id, category_id),
  INDEX idx_store_categories_store_id (store_id),
  INDEX idx_store_categories_category_id (category_id)
);
```

#### tags (标签表)

```sql
CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,

  name VARCHAR(100) NOT NULL UNIQUE,
  name_zh VARCHAR(100),
  type VARCHAR(20), -- review, post, store

  usage_count INT DEFAULT 0, -- 冗余：使用次数

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_tags_type (type),
  INDEX idx_tags_usage_count (usage_count)
);
```

---

### 2.4 评论系统 (Review System)

#### reviews (评论表)

```sql
CREATE TABLE reviews (
  id BIGSERIAL PRIMARY KEY,

  -- 关联
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  store_id BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id), -- 冗余，提升查询性能

  -- 评论内容
  content TEXT NOT NULL,

  -- 评分
  overall_rating DECIMAL(3,2) NOT NULL, -- 总体评分 1.00-5.00

  -- 子评分 (可选)
  food_rating DECIMAL(3,2),
  service_rating DECIMAL(3,2),
  atmosphere_rating DECIMAL(3,2),
  value_rating DECIMAL(3,2),

  -- 冗余数据 (提升性能，减少JOIN)
  user_display_name VARCHAR(100), -- 冗余：评论时的用户名
  user_avatar_url TEXT, -- 冗余：评论时的头像
  store_name VARCHAR(255), -- 冗余：店铺名称

  -- 统计 (冗余)
  like_count INT DEFAULT 0,
  comment_count INT DEFAULT 0,
  view_count INT DEFAULT 0,

  -- 状态
  status VARCHAR(20) DEFAULT 'published', -- draft, published, hidden, deleted
  is_featured BOOLEAN DEFAULT FALSE,

  -- AI辅助
  ai_suggestion_used BOOLEAN DEFAULT FALSE,

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  published_at TIMESTAMP,
  deleted_at TIMESTAMP,

  INDEX idx_reviews_user_id (user_id),
  INDEX idx_reviews_store_id (store_id),
  INDEX idx_reviews_merchant_id (merchant_id),
  INDEX idx_reviews_status (status),
  INDEX idx_reviews_overall_rating (overall_rating),
  INDEX idx_reviews_created_at (created_at DESC),
  INDEX idx_reviews_published_at (published_at DESC)
);
```

#### review_media (评论媒体)

```sql
CREATE TABLE review_media (
  id BIGSERIAL PRIMARY KEY,
  review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,

  media_type VARCHAR(20) NOT NULL, -- image, video
  media_url TEXT NOT NULL,
  thumbnail_url TEXT,

  file_size BIGINT, -- bytes
  mime_type VARCHAR(100),

  display_order INT DEFAULT 0,

  -- R2 存储信息
  r2_key TEXT,
  r2_bucket VARCHAR(100),

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_review_media_review_id (review_id)
);
```

#### review_tags (评论-标签关联)

```sql
CREATE TABLE review_tags (
  id BIGSERIAL PRIMARY KEY,
  review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(review_id, tag_id),
  INDEX idx_review_tags_review_id (review_id),
  INDEX idx_review_tags_tag_id (tag_id)
);
```

#### review_likes (评论点赞)

```sql
CREATE TABLE review_likes (
  id BIGSERIAL PRIMARY KEY,
  review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(review_id, user_id),
  INDEX idx_review_likes_review_id (review_id),
  INDEX idx_review_likes_user_id (user_id)
);
```

#### review_replies (商家回复评论)

```sql
CREATE TABLE review_replies (
  id BIGSERIAL PRIMARY KEY,
  review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

  content TEXT NOT NULL,

  -- 状态
  status VARCHAR(20) DEFAULT 'published', -- draft, published, deleted

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,

  INDEX idx_review_replies_review_id (review_id),
  INDEX idx_review_replies_merchant_id (merchant_id)
);
```

---

### 2.5 社区/帖子系统 (Community/Post)

#### posts (帖子表)

```sql
CREATE TABLE posts (
  id BIGSERIAL PRIMARY KEY,

  -- 关联
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  merchant_id BIGINT REFERENCES merchants(id) ON DELETE CASCADE, -- 商家营销帖
  review_id BIGINT REFERENCES reviews(id) ON DELETE SET NULL, -- 如果是评论衍生的帖子

  -- 内容
  title VARCHAR(255),
  content TEXT NOT NULL,
  media_urls JSONB, -- [url1, url2, ...]

  -- 类型
  post_type VARCHAR(20) DEFAULT 'user_post',
  -- user_post, merchant_post, review_post

  -- 统计 (冗余)
  like_count INT DEFAULT 0,
  comment_count INT DEFAULT 0,
  share_count INT DEFAULT 0,
  view_count INT DEFAULT 0,

  -- 冗余数据
  author_name VARCHAR(100),
  author_avatar_url TEXT,

  -- 状态
  status VARCHAR(20) DEFAULT 'published',
  is_pinned BOOLEAN DEFAULT FALSE,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  published_at TIMESTAMP,
  deleted_at TIMESTAMP,

  INDEX idx_posts_user_id (user_id),
  INDEX idx_posts_merchant_id (merchant_id),
  INDEX idx_posts_review_id (review_id),
  INDEX idx_posts_post_type (post_type),
  INDEX idx_posts_created_at (created_at DESC)
);
```

#### post_comments (帖子评论)

```sql
CREATE TABLE post_comments (
  id BIGSERIAL PRIMARY KEY,
  post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  parent_comment_id BIGINT REFERENCES post_comments(id), -- 支持嵌套评论

  content TEXT NOT NULL,

  like_count INT DEFAULT 0,

  -- 冗余
  user_name VARCHAR(100),
  user_avatar_url TEXT,

  status VARCHAR(20) DEFAULT 'published',

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,

  INDEX idx_post_comments_post_id (post_id),
  INDEX idx_post_comments_user_id (user_id),
  INDEX idx_post_comments_parent_comment_id (parent_comment_id)
);
```

#### post_likes (帖子点赞)

```sql
CREATE TABLE post_likes (
  id BIGSERIAL PRIMARY KEY,
  post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(post_id, user_id),
  INDEX idx_post_likes_post_id (post_id),
  INDEX idx_post_likes_user_id (user_id)
);
```

---

### 2.6 优惠券与支付系统 (Coupon & Payment)

#### coupons (优惠券表)

```sql
CREATE TABLE coupons (
  id BIGSERIAL PRIMARY KEY,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  store_id BIGINT REFERENCES stores(id) ON DELETE SET NULL,

  -- 基础信息
  title VARCHAR(255) NOT NULL,
  description TEXT,
  image_url TEXT,

  -- 优惠类型
  coupon_type VARCHAR(20) NOT NULL, -- discount, free_item, package_deal

  -- 价格
  original_price DECIMAL(10,2),
  sale_price DECIMAL(10,2), -- 为0表示免费优惠券
  discount_percentage DECIMAL(5,2), -- 折扣百分比

  -- 库存
  total_quantity INT, -- NULL = 无限制
  claimed_quantity INT DEFAULT 0,
  redeemed_quantity INT DEFAULT 0,

  -- 有效期
  valid_from TIMESTAMP,
  valid_until TIMESTAMP,

  -- 使用限制
  max_claims_per_user INT DEFAULT 1,
  min_purchase_amount DECIMAL(10,2),

  -- 使用条件
  terms_and_conditions TEXT,

  -- 状态
  status VARCHAR(20) DEFAULT 'active', -- draft, active, paused, expired, deleted

  -- 冗余
  merchant_name VARCHAR(255),

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,

  INDEX idx_coupons_merchant_id (merchant_id),
  INDEX idx_coupons_store_id (store_id),
  INDEX idx_coupons_status (status),
  INDEX idx_coupons_valid_from_until (valid_from, valid_until)
);
```

#### packages (套餐表)

```sql
CREATE TABLE packages (
  id BIGSERIAL PRIMARY KEY,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  store_id BIGINT REFERENCES stores(id) ON DELETE SET NULL,

  name VARCHAR(255) NOT NULL,
  description TEXT,
  image_url TEXT,

  -- 价格
  original_price DECIMAL(10,2) NOT NULL,
  sale_price DECIMAL(10,2) NOT NULL,
  savings DECIMAL(10,2), -- 冗余：节省金额

  -- 套餐内容
  items JSONB, -- [{name: "Item 1", quantity: 2}, ...]

  -- 有效期
  valid_from TIMESTAMP,
  valid_until TIMESTAMP,

  -- 状态
  status VARCHAR(20) DEFAULT 'active',

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_packages_merchant_id (merchant_id),
  INDEX idx_packages_store_id (store_id),
  INDEX idx_packages_status (status)
);
```

#### orders (订单表)

```sql
CREATE TABLE orders (
  id BIGSERIAL PRIMARY KEY,
  order_number VARCHAR(50) UNIQUE NOT NULL, -- 业务订单号

  -- 关联
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  coupon_id BIGINT REFERENCES coupons(id),
  package_id BIGINT REFERENCES packages(id),

  -- 订单类型
  order_type VARCHAR(20) NOT NULL, -- coupon, package

  -- 金额
  original_amount DECIMAL(10,2) NOT NULL,
  final_amount DECIMAL(10,2) NOT NULL,
  discount_amount DECIMAL(10,2) DEFAULT 0,

  -- 支付
  payment_method VARCHAR(50), -- wechat, alipay, stripe, free
  payment_status VARCHAR(20) DEFAULT 'pending',
  -- pending, processing, completed, failed, refunded
  payment_session_id VARCHAR(255), -- 第三方支付会话ID
  paid_at TIMESTAMP,

  -- 状态
  order_status VARCHAR(20) DEFAULT 'pending',
  -- pending, confirmed, completed, cancelled, refunded

  -- 冗余
  merchant_id BIGINT, -- 冗余，方便查询
  merchant_name VARCHAR(255),

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_orders_user_id (user_id),
  INDEX idx_orders_merchant_id (merchant_id),
  INDEX idx_orders_order_number (order_number),
  INDEX idx_orders_payment_status (payment_status),
  INDEX idx_orders_created_at (created_at DESC)
);
```

#### vouchers (券码表)

```sql
CREATE TABLE vouchers (
  id BIGSERIAL PRIMARY KEY,
  voucher_code VARCHAR(50) UNIQUE NOT NULL, -- 券码

  -- 关联
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  coupon_id BIGINT REFERENCES coupons(id),
  package_id BIGINT REFERENCES packages(id),
  order_id BIGINT REFERENCES orders(id),
  merchant_id BIGINT NOT NULL REFERENCES merchants(id),

  -- 券码信息
  qr_code_url TEXT, -- QR码图片URL
  barcode TEXT, -- 条形码

  -- 状态
  status VARCHAR(20) DEFAULT 'active',
  -- active, redeemed, expired, cancelled

  -- 有效期
  valid_from TIMESTAMP NOT NULL,
  valid_until TIMESTAMP NOT NULL,

  -- 核销信息
  redeemed_at TIMESTAMP,
  redeemed_by BIGINT REFERENCES merchants(id), -- 核销的商家
  redemption_notes TEXT,

  -- 冗余
  coupon_title VARCHAR(255),
  merchant_name VARCHAR(255),

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_vouchers_user_id (user_id),
  INDEX idx_vouchers_merchant_id (merchant_id),
  INDEX idx_vouchers_voucher_code (voucher_code),
  INDEX idx_vouchers_status (status),
  INDEX idx_vouchers_valid_until (valid_until)
);
```

---

### 2.7 用户行为与历史 (User Behavior)

#### user_following (用户关注)

```sql
CREATE TABLE user_following (
  id BIGSERIAL PRIMARY KEY,

  follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  followee_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- 关注用户
  merchant_id BIGINT REFERENCES merchants(id) ON DELETE CASCADE, -- 关注商家

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_user_following_follower_id (follower_id),
  INDEX idx_user_following_followee_id (followee_id),
  INDEX idx_user_following_merchant_id (merchant_id),

  -- 确保一个用户只能关注一个目标一次
  CONSTRAINT check_follow_target
    CHECK ((followee_id IS NOT NULL AND merchant_id IS NULL) OR
           (followee_id IS NULL AND merchant_id IS NOT NULL))
);
```

#### browsing_history (浏览历史)

```sql
CREATE TABLE browsing_history (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- 浏览对象
  store_id BIGINT REFERENCES stores(id) ON DELETE CASCADE,
  review_id BIGINT REFERENCES reviews(id) ON DELETE SET NULL,
  post_id BIGINT REFERENCES posts(id) ON DELETE SET NULL,

  -- 浏览元数据
  view_duration INT, -- 秒

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_browsing_history_user_id (user_id),
  INDEX idx_browsing_history_store_id (store_id),
  INDEX idx_browsing_history_created_at (created_at DESC)
);
```

#### pending_reviews (待评价提醒)

```sql
CREATE TABLE pending_reviews (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  store_id BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id),

  -- 触发来源
  trigger_source VARCHAR(50), -- coupon_redemption, visit_log, manual

  -- 提醒状态
  reminder_status VARCHAR(20) DEFAULT 'pending',
  -- pending, reminded, completed, dismissed
  reminded_at TIMESTAMP,

  -- 完成信息
  completed_at TIMESTAMP,
  review_id BIGINT REFERENCES reviews(id),

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_pending_reviews_user_id (user_id),
  INDEX idx_pending_reviews_reminder_status (reminder_status)
);
```

---

### 2.8 消息系统 (Messaging)

#### conversations (会话表)

```sql
CREATE TABLE conversations (
  id BIGSERIAL PRIMARY KEY,

  -- 会话类型
  conversation_type VARCHAR(20) NOT NULL,
  -- direct (1对1), group (群组)

  -- 参与者 (对于商家消息系统)
  merchant_id BIGINT REFERENCES merchants(id) ON DELETE CASCADE,
  customer_id BIGINT REFERENCES users(id) ON DELETE CASCADE,

  -- 会话元数据
  title VARCHAR(255), -- 群组名称或自定义标题
  avatar_url TEXT,

  -- 最后消息 (冗余，提升性能)
  last_message_content TEXT,
  last_message_at TIMESTAMP,
  last_message_sender_id BIGINT,

  -- 状态
  is_archived BOOLEAN DEFAULT FALSE,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_conversations_merchant_id (merchant_id),
  INDEX idx_conversations_customer_id (customer_id),
  INDEX idx_conversations_last_message_at (last_message_at DESC)
);
```

#### conversation_participants (会话参与者 - 用于群组)

```sql
CREATE TABLE conversation_participants (
  id BIGSERIAL PRIMARY KEY,
  conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- 参与者设置
  is_muted BOOLEAN DEFAULT FALSE,
  is_pinned BOOLEAN DEFAULT FALSE,

  -- 未读数
  unread_count INT DEFAULT 0,
  last_read_message_id BIGINT,

  joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  left_at TIMESTAMP,

  UNIQUE(conversation_id, user_id),
  INDEX idx_conversation_participants_conversation_id (conversation_id),
  INDEX idx_conversation_participants_user_id (user_id)
);
```

#### messages (消息表)

```sql
CREATE TABLE messages (
  id BIGSERIAL PRIMARY KEY,
  conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

  -- 发送者
  sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  sender_type VARCHAR(20), -- customer, merchant

  -- 消息内容
  content TEXT NOT NULL,
  message_type VARCHAR(20) DEFAULT 'text',
  -- text, image, file, system

  -- 附件
  attachments JSONB, -- [{url, type, name, size}, ...]

  -- 状态
  is_deleted BOOLEAN DEFAULT FALSE,
  deleted_at TIMESTAMP,

  -- 已读状态 (简化设计，可用单独表优化)
  read_by_ids BIGINT[], -- 已读用户ID数组

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_messages_conversation_id (conversation_id),
  INDEX idx_messages_sender_id (sender_id),
  INDEX idx_messages_created_at (created_at DESC)
);
```

---

### 2.9 商家认证系统 (Merchant Verification)

#### merchant_verifications (商家认证申请)

```sql
CREATE TABLE merchant_verifications (
  id BIGSERIAL PRIMARY KEY,
  merchant_id BIGINT UNIQUE NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

  -- 申请信息
  business_license_number VARCHAR(100),
  legal_representative VARCHAR(100),
  business_address TEXT,

  -- 文件
  documents JSONB,
  -- [{type: "license", url: "xxx", name: "xxx"}, ...]

  -- 审核状态
  verification_status VARCHAR(20) DEFAULT 'pending',
  -- pending, under_review, approved, rejected, resubmit_required

  -- 审核信息
  reviewed_by BIGINT REFERENCES users(id), -- 审核员
  reviewed_at TIMESTAMP,
  rejection_reason TEXT,
  admin_notes TEXT,

  -- 提交时间
  submitted_at TIMESTAMP,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_merchant_verifications_merchant_id (merchant_id),
  INDEX idx_merchant_verifications_status (verification_status)
);
```

---

### 2.10 营销与分析 (Marketing & Analytics)

#### marketing_posts (商家营销帖)

```sql
CREATE TABLE marketing_posts (
  id BIGSERIAL PRIMARY KEY,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  post_id BIGINT UNIQUE REFERENCES posts(id) ON DELETE CASCADE, -- 关联到posts表

  -- 营销信息
  campaign_name VARCHAR(255),
  call_to_action VARCHAR(255),
  cta_link TEXT,

  -- 标签
  tags VARCHAR(100)[],

  -- 统计
  impression_count INT DEFAULT 0,
  click_count INT DEFAULT 0,

  -- 定时发布
  scheduled_at TIMESTAMP,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_marketing_posts_merchant_id (merchant_id),
  INDEX idx_marketing_posts_scheduled_at (scheduled_at)
);
```

#### merchant_analytics (商家分析数据 - 按天聚合)

```sql
CREATE TABLE merchant_analytics (
  id BIGSERIAL PRIMARY KEY,
  merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

  date DATE NOT NULL,

  -- 流量指标
  page_views INT DEFAULT 0,
  unique_visitors INT DEFAULT 0,

  -- 互动指标
  review_count INT DEFAULT 0,
  average_rating DECIMAL(3,2),
  like_count INT DEFAULT 0,
  share_count INT DEFAULT 0,

  -- 转化指标
  coupon_views INT DEFAULT 0,
  coupon_claims INT DEFAULT 0,
  voucher_redemptions INT DEFAULT 0,

  -- 收入指标
  total_revenue DECIMAL(12,2) DEFAULT 0,
  order_count INT DEFAULT 0,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(merchant_id, date),
  INDEX idx_merchant_analytics_merchant_id (merchant_id),
  INDEX idx_merchant_analytics_date (date DESC)
);
```

#### notifications (通知表)

```sql
CREATE TABLE notifications (
  id BIGSERIAL PRIMARY KEY,

  -- 接收者
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  merchant_id BIGINT REFERENCES merchants(id) ON DELETE CASCADE,

  -- 通知类型
  notification_type VARCHAR(50) NOT NULL,
  -- new_review, new_reply, coupon_claimed, voucher_redeemed,
  -- verification_approved, payment_received, etc.

  -- 通知内容
  title VARCHAR(255) NOT NULL,
  content TEXT,

  -- 关联对象
  related_type VARCHAR(50), -- review, order, voucher, etc.
  related_id BIGINT,

  -- 动作链接
  action_url TEXT,

  -- 状态
  is_read BOOLEAN DEFAULT FALSE,
  read_at TIMESTAMP,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_notifications_user_id (user_id),
  INDEX idx_notifications_merchant_id (merchant_id),
  INDEX idx_notifications_is_read (is_read),
  INDEX idx_notifications_created_at (created_at DESC),

  -- 确保只有一个接收者
  CONSTRAINT check_notification_recipient
    CHECK ((user_id IS NOT NULL AND merchant_id IS NULL) OR
           (user_id IS NULL AND merchant_id IS NOT NULL))
);
```

---

### 2.11 管理系统 (Admin)

#### reports (举报表)

```sql
CREATE TABLE reports (
  id BIGSERIAL PRIMARY KEY,

  -- 举报人
  reporter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- 被举报对象
  reported_type VARCHAR(50) NOT NULL, -- review, post, comment, user
  reported_id BIGINT NOT NULL,

  -- 举报原因
  reason VARCHAR(50) NOT NULL,
  -- spam, inappropriate, harassment, fake, other
  details TEXT,

  -- 审核状态
  status VARCHAR(20) DEFAULT 'pending',
  -- pending, under_review, resolved, dismissed

  -- 审核信息
  reviewed_by BIGINT REFERENCES users(id),
  reviewed_at TIMESTAMP,
  resolution_notes TEXT,
  action_taken VARCHAR(100), -- deleted, warned, no_action

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_reports_reporter_id (reporter_id),
  INDEX idx_reports_reported_type_id (reported_type, reported_id),
  INDEX idx_reports_status (status)
);
```

#### admin_audit_logs (管理操作审计日志)

```sql
CREATE TABLE admin_audit_logs (
  id BIGSERIAL PRIMARY KEY,

  admin_id BIGINT NOT NULL REFERENCES users(id),

  -- 操作信息
  action VARCHAR(100) NOT NULL, -- delete_review, ban_user, approve_merchant
  target_type VARCHAR(50),
  target_id BIGINT,

  -- 操作详情
  details JSONB,
  ip_address VARCHAR(45),
  user_agent TEXT,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_admin_audit_logs_admin_id (admin_id),
  INDEX idx_admin_audit_logs_action (action),
  INDEX idx_admin_audit_logs_created_at (created_at DESC)
);
```

---

## 3. 索引策略

### 3.1 主要查询场景与索引

| 查询场景 | 涉及表 | 索引策略 |
|---------|--------|---------|
| 用户登录 | users | idx_users_email (已有) |
| 首页推荐 | stores, reviews | 复合索引 (status, is_featured, created_at) |
| 商家详情页 | stores, reviews | idx_stores_slug, idx_reviews_store_id |
| 用户评论列表 | reviews | 复合索引 (user_id, status, created_at DESC) |
| 地图搜索 | stores | 空间索引 (latitude, longitude) 或使用 PostGIS |
| 优惠券列表 | coupons | 复合索引 (status, valid_from, valid_until) |
| 我的券码 | vouchers | 复合索引 (user_id, status, valid_until) |
| 消息列表 | conversations | idx_conversations_last_message_at DESC |

### 3.2 额外推荐索引

```sql
-- 评论按评分和时间排序
CREATE INDEX idx_reviews_rating_time ON reviews(overall_rating DESC, created_at DESC)
WHERE status = 'published';

-- 优惠券有效性查询
CREATE INDEX idx_coupons_active ON coupons(status, valid_from, valid_until)
WHERE status = 'active';

-- 券码有效性查询
CREATE INDEX idx_vouchers_active ON vouchers(status, valid_until)
WHERE status = 'active';

-- 商家的活跃店铺
CREATE INDEX idx_stores_merchant_active ON stores(merchant_id, status)
WHERE status = 'active';

-- 全文搜索索引 (PostgreSQL)
CREATE INDEX idx_stores_search ON stores USING GIN(to_tsvector('english', name || ' ' || description));
CREATE INDEX idx_reviews_search ON reviews USING GIN(to_tsvector('english', content));
```

---

## 4. 冗余设计说明

### 4.1 为什么需要冗余？

1. **减少JOIN操作**: 高频查询场景（如评论列表）避免多表JOIN
2. **提升查询性能**: 关键统计数据（点赞数、评论数）冗余存储
3. **保持历史快照**: 用户名、店铺名等在数据变更后保持评论时的状态

### 4.2 冗余字段清单

| 表名 | 冗余字段 | 来源 | 更新策略 |
|------|---------|------|---------|
| reviews | user_display_name, user_avatar_url | user_profiles | 创建时写入，不随用户资料更新 |
| reviews | store_name | stores | 创建时写入，不随店铺更新 |
| reviews | merchant_id | stores -> merchants | 创建时写入，方便商家查询 |
| reviews | like_count, comment_count | 聚合计算 | 实时更新(通过trigger或应用层) |
| user_profiles | review_count, follower_count | 聚合计算 | 实时更新或定时任务 |
| stores | review_count, average_rating | reviews聚合 | 定时任务(每小时/天) |
| coupons | merchant_name | merchants | 创建时写入 |
| vouchers | coupon_title, merchant_name | coupons, merchants | 创建时写入 |
| conversations | last_message_* | messages | 新消息时更新 |
| posts | author_name, author_avatar_url | users/merchants | 创建时写入 |

### 4.3 计数器更新策略

**方案A: 数据库触发器 (推荐用于关键业务)**

```sql
-- 示例：评论点赞数自动更新
CREATE OR REPLACE FUNCTION update_review_like_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE reviews SET like_count = like_count + 1 WHERE id = NEW.review_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE reviews SET like_count = like_count - 1 WHERE id = OLD.review_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_review_like_count
AFTER INSERT OR DELETE ON review_likes
FOR EACH ROW EXECUTE FUNCTION update_review_like_count();
```

**方案B: 应用层更新 (推荐用于非关键统计)**

```javascript
// 创建点赞时
await db.transaction(async (trx) => {
  await trx('review_likes').insert({ review_id, user_id });
  await trx('reviews').where({ id: review_id }).increment('like_count', 1);
});
```

**方案C: 定时任务重算 (用于非实时统计)**

```sql
-- 每小时重算店铺平均分
UPDATE stores s
SET
  review_count = (SELECT COUNT(*) FROM reviews r WHERE r.store_id = s.id AND r.status = 'published'),
  average_rating = (SELECT AVG(overall_rating) FROM reviews r WHERE r.store_id = s.id AND r.status = 'published')
WHERE s.updated_at < NOW() - INTERVAL '1 hour';
```

---

## 5. 数据迁移考虑

### 5.1 分阶段迁移建议

**Phase 1: 核心用户与商家系统**

- users, user_profiles
- merchants, merchant_profiles, stores
- categories, tags

**Phase 2: 评论系统**

- reviews, review_media, review_tags
- review_likes, review_replies

**Phase 3: 优惠券与支付**

- coupons, packages
- orders, vouchers

**Phase 4: 社区与消息**

- posts, post_comments, post_likes
- conversations, messages

**Phase 5: 辅助功能**

- notifications, browsing_history
- merchant_analytics, reports

### 5.2 数据一致性检查

```sql
-- 检查评论计数是否一致
SELECT
  r.id,
  r.like_count AS cached_count,
  COUNT(rl.id) AS actual_count
FROM reviews r
LEFT JOIN review_likes rl ON rl.review_id = r.id
GROUP BY r.id, r.like_count
HAVING r.like_count != COUNT(rl.id);

-- 检查用户评论数是否一致
SELECT
  up.user_id,
  up.review_count AS cached_count,
  COUNT(r.id) AS actual_count
FROM user_profiles up
LEFT JOIN reviews r ON r.user_id = up.user_id AND r.status = 'published'
GROUP BY up.user_id, up.review_count
HAVING up.review_count != COUNT(r.id);
```

---

## 6. 性能优化建议

### 6.1 查询优化

1. **使用EXPLAIN ANALYZE分析慢查询**
2. **对高频查询添加物化视图**
3. **使用Redis缓存热点数据**:
   - 商家详情页
   - 用户资料
   - 优惠券列表
   - 首页推荐

### 6.2 分区策略

对于大表考虑分区：

```sql
-- 按月分区评论表 (PostgreSQL 10+)
CREATE TABLE reviews (
  -- ... 字段定义
) PARTITION BY RANGE (created_at);

CREATE TABLE reviews_2026_01 PARTITION OF reviews
  FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE reviews_2026_02 PARTITION OF reviews
  FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
```

### 6.3 读写分离

- **主库**: 处理所有写操作
- **从库**: 处理读操作（商家列表、评论查询等）
- **延迟敏感**: 用户自己的数据读主库

---

## 7. 安全性考虑

### 7.1 敏感数据

1. **密码**: 使用bcrypt/argon2加密
2. **个人信息**: 考虑字段级加密（email, phone）
3. **支付信息**: 不存储完整卡号，仅存储token
4. **软删除**: 重要数据使用deleted_at而非物理删除

### 7.2 数据权限

```sql
-- 行级安全策略示例 (PostgreSQL RLS)
ALTER TABLE reviews ENABLE ROW LEVEL SECURITY;

CREATE POLICY reviews_select_policy ON reviews
  FOR SELECT
  USING (status = 'published' OR user_id = current_user_id());
```

---

## 8. 待确认事项

### 8.1 需要前后端对齐的问题

1. **商家ID vs 店铺ID**:
   - 一个商家账号可以有多个店铺吗？
   - 评论关联到store还是merchant？

2. **venueId**:
   - 需求中提到merchantId/venueId，venue是否等同于store？

3. **Package vs Coupon**:
   - 套餐是否就是一种特殊的优惠券？
   - 是否需要单独的表？

4. **支付回调**:
   - 支付成功后的回调处理流程？
   - 需要webhook表吗？

5. **实时消息**:
   - 是否需要WebSocket支持？
   - 消息系统是否需要支持商家-用户双向？

### 8.2 技术选型建议

| 组件 | 推荐方案 | 备选方案 |
|------|---------|---------|
| 数据库 | PostgreSQL 14+ | MySQL 8+ |
| 缓存 | Redis | Memcached |
| 文件存储 | Cloudflare R2 | AWS S3, Aliyun OSS |
| 全文搜索 | PostgreSQL Full-Text | Elasticsearch, MeiliSearch |
| 消息队列 | Bull (Redis-based) | RabbitMQ, AWS SQS |
| 实时通信 | Socket.io | WebSocket, SSE |

---

## 9. 下一步行动

1. ✅ **Review数据库设计** - 前后端对齐字段定义
2. ⏳ **确认待定问题** - 需要澄清的业务逻辑
3. ⏳ **编写Migration脚本** - 生成初始化SQL
4. ⏳ **编写Seeder** - 准备测试数据
5. ⏳ **API Contract对齐** - 与前端API契约文档对比

---

**版本**: v1.0
**作者**: Backend Team
**更新日期**: 2026-02-14
