# RevieU Database Design v2.0 (Full Edition)

> 基于前端需求 Matrix v2.4 + 现有后端代码审查
> 设计原则：3NF + 性能冗余 | 保留现有代码优秀模式 | 全模块覆盖
> 数据库：PostgreSQL 14+ with PostGIS
> ORM：GORM (Go)

## 设计决策摘要

| 决策点 | 结论 | 理由 |
|--------|------|------|
| 认证架构 | 沿用 `user_auths` 多 provider 模式 | 现有设计优于 v1 单表方案 |
| 点赞/收藏 | 沿用多态 `likes` + `favorites` | 1 张表覆盖多 target_type |
| 关注关系 | 沿用 `user_follows` + `merchant_follows` 分表 | 比 CHECK constraint 更清晰 |
| 商家 vs 门店 | 拆分: `merchants`(账号) + `stores`(门店) | 支持多门店/地理查询/营业时间 |
| 评论图片 | `media_uploads` 通用表 + `review_media` 关联表 | 兼顾 R2 追踪和评论关联 |
| venue_id 兼容 | `stores` 替代原 `venues`，review 中用 `store_id` | 统一命名，migration 时处理 |

---

## 0. 总表清单

### 保留不变 (11 张)

| 表名 | 说明 | 来源 |
|------|------|------|
| `users` | 核心用户表 | 现有 |
| `user_auths` | 多 provider 认证 | 现有 |
| `user_profiles` | 用户资料 | 现有 (增补字段) |
| `email_verifications` | 邮箱验证 | 现有 |
| `user_addresses` | 收货地址 | 现有 |
| `user_privacies` | 隐私设置 | 现有 |
| `user_notifications` | 通知偏好 | 现有 |
| `account_deletions` | 注销请求 | 现有 |
| `user_follows` | 用户互关 | 现有 |
| `merchant_follows` | 关注商家 | 现有 (改关联到 stores) |
| `media_uploads` | 通用媒体上传 | 现有 (增补字段) |

### 重构 (6 张)

| 表名 | 变化 | 来源 |
|------|------|------|
| `merchants` | 从"实体"改为"账号"，拆出 stores | 现有重构 |
| `reviews` | venue_id → store_id，增加 comment_count | 现有重构 |
| `posts` | 增加 comment_count, share_count, post_type | 现有重构 |
| `tags` | 增加 type, review_count 字段 | 现有重构 |
| `likes` | 不变 | 现有 |
| `favorites` | 不变 | 现有 |

### 新增 (24 张)

| 表名 | 模块 | 覆盖需求 |
|------|------|---------|
| `stores` | 商家 | DIS-001~010 |
| `store_hours` | 商家 | DIS-003 (Open Now) |
| `categories` | 发现 | DIS-004 |
| `store_categories` | 发现 | DIS-004 |
| `review_media` | 评论 | REV-002/003 |
| `review_tags` | 评论 | REV-001 (已有但未迁移) |
| `review_comments` | 评论 | REV-007, MRC-002 |
| `post_comments` | 社区 | COM-004 |
| `post_tags` | 社区 | (已有但确认保留) |
| `coupons` | 优惠券 | CPV-001~003 |
| `packages` | 套餐 | MRC-004 |
| `orders` | 订单 | CPV-004/005 |
| `vouchers` | 券码 | CPV-006~008 |
| `payments` | 支付 | CPV-003/004 |
| `conversations` | 消息 | MSG-001 |
| `conversation_participants` | 消息 | MSG-001/003 |
| `messages` | 消息 | MSG-002 |
| `merchant_verifications` | 认证 | VER-001/002 |
| `marketing_posts` | 营销 | MRC-005 |
| `merchant_analytics` | 分析 | MRC-007 |
| `notifications` | 通知 | MRC-009 |
| `reports` | 管理 | ADM-002 |
| `admin_audit_logs` | 管理 | ADM-001 |
| `browsing_history` | 用户行为 | PRF-006 |

**总计：41 张表**

---

## 1. 用户与认证系统

> 覆盖: AUT-001 ~ AUT-013

### users — 保留不变

```sql
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    role       VARCHAR(20) NOT NULL DEFAULT 'user',  -- 'user', 'merchant', 'admin'
    status     SMALLINT    NOT NULL DEFAULT 0,        -- 0:active 1:banned 2:pending
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

> **变更**: `role` 增加 `'merchant'` 值。现有 merchant 用户需要通过 merchants 表关联，
> 但 role 字段也应标记，方便路由守卫快速判断。

### user_auths — 保留不变

```sql
CREATE TABLE user_auths (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    identity_type VARCHAR(20) NOT NULL,  -- 'email', 'google', 'apple', 'wechat'
    identifier    VARCHAR(255) NOT NULL, -- email 或 open_id
    credential    VARCHAR(255),          -- 密码 hash 或 access_token
    last_login_at TIMESTAMPTZ,

    UNIQUE (identity_type, identifier)
);

CREATE INDEX idx_user_auths_user_id ON user_auths(user_id);
CREATE INDEX idx_user_auths_email ON user_auths(identifier)
    WHERE identity_type = 'email';
```

### user_profiles — 保留，增补字段

```sql
CREATE TABLE user_profiles (
    user_id        BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    nickname       VARCHAR(50)  NOT NULL,
    avatar_url     VARCHAR(255),
    intro          VARCHAR(255),
    location       VARCHAR(100),

    -- 冗余统计
    follower_count  INT NOT NULL DEFAULT 0,
    following_count INT NOT NULL DEFAULT 0,
    post_count      INT NOT NULL DEFAULT 0,
    review_count    INT NOT NULL DEFAULT 0,
    like_count      INT NOT NULL DEFAULT 0,

    -- v2 新增
    coupon_count    INT NOT NULL DEFAULT 0  -- 持有券码数
);
```

### email_verifications — 保留不变

```sql
CREATE TABLE email_verifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email      VARCHAR(255) NOT NULL,
    token      VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verifications_token ON email_verifications(token);
CREATE INDEX idx_email_verifications_user ON email_verifications(user_id);
```

### user_addresses — 保留不变

```sql
CREATE TABLE user_addresses (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(50)  NOT NULL,
    phone       VARCHAR(20)  NOT NULL,
    province    VARCHAR(50),
    city        VARCHAR(50),
    district    VARCHAR(50),
    address     VARCHAR(255) NOT NULL,
    postal_code VARCHAR(20),
    is_default  BOOLEAN      DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
```

### user_privacies — 保留不变

```sql
CREATE TABLE user_privacies (
    user_id  BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT TRUE
);
```

### user_notifications — 保留不变

```sql
CREATE TABLE user_notifications (
    user_id       BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    push_enabled  BOOLEAN DEFAULT TRUE,
    email_enabled BOOLEAN DEFAULT TRUE
);
```

### account_deletions — 保留不变

```sql
CREATE TABLE account_deletions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    reason       VARCHAR(255),
    scheduled_at TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

---

## 2. 商家与门店系统

> 覆盖: DIS-001 ~ DIS-010, MRC-006

### merchants — 重构为"商家账号"

```sql
CREATE TABLE merchants (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- 商家基础信息
    business_name  VARCHAR(100) NOT NULL,
    business_type  VARCHAR(50),    -- restaurant, cafe, bar, salon, etc.
    logo_url       VARCHAR(255),
    description    TEXT,

    -- 联系方式
    contact_phone  VARCHAR(20),
    contact_email  VARCHAR(255),
    website_url    VARCHAR(512),
    social_links   JSONB,          -- {"instagram":"...", "facebook":"..."}

    -- 认证状态
    verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified',
    -- unverified, pending, under_review, approved, rejected
    verified_at TIMESTAMPTZ,

    -- 冗余统计 (品牌级汇总)
    total_stores    INT     NOT NULL DEFAULT 0,
    total_reviews   INT     NOT NULL DEFAULT 0,
    avg_rating      DECIMAL(3,2) DEFAULT 0.00,
    follower_count  INT     NOT NULL DEFAULT 0,

    status     SMALLINT    NOT NULL DEFAULT 0,  -- 0:active 1:suspended 2:closed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchants_user_id ON merchants(user_id);
CREATE INDEX idx_merchants_verification_status ON merchants(verification_status);
CREATE INDEX idx_merchants_status ON merchants(status);
```

### stores — 新增 (门店/地点)

```sql
CREATE TABLE stores (
    id          BIGSERIAL PRIMARY KEY,
    merchant_id BIGINT       NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

    -- 基础信息
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) UNIQUE,
    description TEXT,

    -- 地理位置 (PostGIS)
    address     VARCHAR(255) NOT NULL,
    city        VARCHAR(100),
    state       VARCHAR(100),
    country     VARCHAR(100) DEFAULT 'US',
    postal_code VARCHAR(20),
    location    GEOGRAPHY(POINT, 4326),  -- PostGIS 地理点

    -- 联系
    phone       VARCHAR(20),

    -- 图片
    cover_image_url VARCHAR(255),
    images          JSONB DEFAULT '[]',   -- 店铺相册
    menu_images     JSONB DEFAULT '[]',   -- 菜单图片

    -- 价格区间
    price_level SMALLINT,  -- 1~4 ($~$$$$)

    -- 冗余统计
    review_count INT        NOT NULL DEFAULT 0,
    avg_rating   DECIMAL(3,2) DEFAULT 0.00,

    -- 状态
    status      SMALLINT    NOT NULL DEFAULT 0,  -- 0:active 1:inactive 2:closed
    is_featured BOOLEAN     DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stores_merchant_id ON stores(merchant_id);
CREATE INDEX idx_stores_slug ON stores(slug);
CREATE INDEX idx_stores_location ON stores USING GIST(location);
CREATE INDEX idx_stores_city ON stores(city);
CREATE INDEX idx_stores_status ON stores(status) WHERE status = 0;
CREATE INDEX idx_stores_featured ON stores(is_featured) WHERE is_featured = TRUE;
```

### store_hours — 新增 (营业时间)

```sql
CREATE TABLE store_hours (
    id          BIGSERIAL PRIMARY KEY,
    store_id    BIGINT  NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL,  -- 0=Sun, 1=Mon, ..., 6=Sat
    open_time   TIME,
    close_time  TIME,
    is_closed   BOOLEAN DEFAULT FALSE,

    UNIQUE (store_id, day_of_week)
);

CREATE INDEX idx_store_hours_store_id ON store_hours(store_id);
```

---

## 3. 分类与标签

> 覆盖: DIS-003, DIS-004

### categories — 新增

```sql
CREATE TABLE categories (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(50)  NOT NULL,
    slug       VARCHAR(50)  UNIQUE NOT NULL,
    parent_id  INT REFERENCES categories(id) ON DELETE SET NULL,
    icon_url   VARCHAR(255),
    sort_order INT DEFAULT 0,
    is_active  BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);
```

### store_categories — 新增

```sql
CREATE TABLE store_categories (
    store_id    BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id INT    NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    is_primary  BOOLEAN DEFAULT FALSE,

    PRIMARY KEY (store_id, category_id)
);

CREATE INDEX idx_store_categories_category_id ON store_categories(category_id);
```

### tags — 重构 (增加 type 和 review_count)

```sql
CREATE TABLE tags (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(50) NOT NULL UNIQUE,
    type         VARCHAR(20) DEFAULT 'general',  -- general, review, post, store
    post_count   INT NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tags_type ON tags(type);
```

---

## 4. 评论系统

> 覆盖: REV-001 ~ REV-008

### reviews — 重构

```sql
CREATE TABLE reviews (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id    BIGINT  NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    merchant_id BIGINT  NOT NULL REFERENCES merchants(id),  -- 冗余

    -- 评分
    rating         DECIMAL(3,2) NOT NULL,  -- 总分 1.00~5.00
    rating_food    DECIMAL(3,2),           -- 菜品
    rating_env     DECIMAL(3,2),           -- 环境
    rating_service DECIMAL(3,2),           -- 服务
    rating_value   DECIMAL(3,2),           -- 性价比

    -- 内容
    content    TEXT NOT NULL,
    visit_date DATE,
    avg_cost   INT,               -- 人均消费 (分/cents)

    -- 冗余统计
    like_count    INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,

    -- 冗余快照 (创建时写入，不跟随源表更新)
    user_nickname  VARCHAR(50),
    user_avatar    VARCHAR(255),
    store_name     VARCHAR(255),

    -- 状态
    status     SMALLINT NOT NULL DEFAULT 0,   -- 0:published 1:hidden 2:deleted
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_store_id ON reviews(store_id, created_at DESC);
CREATE INDEX idx_reviews_merchant_id ON reviews(merchant_id);
CREATE INDEX idx_reviews_status ON reviews(status) WHERE status = 0;
CREATE INDEX idx_reviews_rating ON reviews(rating DESC, created_at DESC)
    WHERE status = 0;
```

### review_media — 新增 (评论媒体关联)

```sql
CREATE TABLE review_media (
    id         BIGSERIAL PRIMARY KEY,
    review_id  BIGINT      NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    upload_id  BIGINT      REFERENCES media_uploads(id),  -- 关联通用上传记录

    media_type VARCHAR(10) NOT NULL,  -- 'image', 'video'
    url        VARCHAR(512) NOT NULL,
    sort_order INT DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_media_review_id ON review_media(review_id);
```

### review_tags — 新增 (评论-标签关联)

```sql
CREATE TABLE review_tags (
    review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    tag_id    BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,

    PRIMARY KEY (review_id, tag_id)
);

CREATE INDEX idx_review_tags_tag_id ON review_tags(tag_id);
```

### review_comments — 重构 (支持商家回复 + 嵌套)

```sql
CREATE TABLE review_comments (
    id                BIGSERIAL PRIMARY KEY,
    review_id         BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_comment_id BIGINT REFERENCES review_comments(id) ON DELETE CASCADE,

    content    TEXT    NOT NULL,
    is_merchant_reply BOOLEAN DEFAULT FALSE, -- 标记是否为商家回复

    -- 冗余快照
    user_nickname VARCHAR(50),
    user_avatar   VARCHAR(255),

    status     SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_comments_review_id ON review_comments(review_id);
CREATE INDEX idx_review_comments_user_id ON review_comments(user_id);
CREATE INDEX idx_review_comments_parent ON review_comments(parent_comment_id);
```

---

## 5. 社区/帖子系统

> 覆盖: COM-001 ~ COM-005, MRC-005

### posts — 重构

```sql
CREATE TABLE posts (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_id *BIGINT REFERENCES merchants(id) ON DELETE SET NULL,  -- 商家营销帖
    review_id   *BIGINT REFERENCES reviews(id) ON DELETE SET NULL,    -- 评论衍生帖

    post_type   VARCHAR(20) NOT NULL DEFAULT 'user',  -- 'user', 'merchant', 'review'
    title       VARCHAR(100),
    content     TEXT NOT NULL,
    images      JSONB DEFAULT '[]',

    -- 冗余统计
    like_count    INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    share_count   INT NOT NULL DEFAULT 0,
    view_count    INT NOT NULL DEFAULT 0,

    -- 冗余快照
    author_name   VARCHAR(100),
    author_avatar VARCHAR(255),

    status     SMALLINT NOT NULL DEFAULT 0,  -- 0:published 1:hidden 2:deleted
    is_pinned  BOOLEAN  DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_merchant_id ON posts(merchant_id);
CREATE INDEX idx_posts_post_type ON posts(post_type);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC) WHERE status = 0;
```

> 注: SQL 中 `*BIGINT` 表示 nullable，GORM 中对应 `*int64`。

### post_tags — 保留不变

```sql
CREATE TABLE post_tags (
    post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id  BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,

    PRIMARY KEY (post_id, tag_id)
);
```

### post_comments — 新增

```sql
CREATE TABLE post_comments (
    id                BIGSERIAL PRIMARY KEY,
    post_id           BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_comment_id BIGINT REFERENCES post_comments(id) ON DELETE CASCADE,

    content TEXT NOT NULL,

    like_count INT NOT NULL DEFAULT 0,

    -- 冗余快照
    user_nickname VARCHAR(50),
    user_avatar   VARCHAR(255),

    status     SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_post_comments_post_id ON post_comments(post_id);
CREATE INDEX idx_post_comments_user_id ON post_comments(user_id);
CREATE INDEX idx_post_comments_parent ON post_comments(parent_comment_id);
```

---

## 6. 互动系统

> 覆盖: REV-007, COM-003, COM-005, PRF-006

### likes — 保留不变 (多态)

```sql
CREATE TABLE likes (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL,
    target_type VARCHAR(20) NOT NULL,  -- 'post', 'review', 'comment'
    target_id   BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, target_type, target_id)
);
```

### favorites — 保留不变 (多态)

```sql
CREATE TABLE favorites (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL,
    target_type VARCHAR(20) NOT NULL,  -- 'post', 'review', 'store'
    target_id   BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, target_type, target_id)
);
```

### user_follows — 保留不变

```sql
CREATE TABLE user_follows (
    id           BIGSERIAL PRIMARY KEY,
    follower_id  BIGINT NOT NULL,
    following_id BIGINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (follower_id, following_id)
);
```

### merchant_follows — 保留 (改名说明)

```sql
-- 注: 虽然叫 merchant_follows, 实际关注的是品牌(merchant)
-- 如果需要关注具体门店，可增加 store_follows 表，MVP 先不加
CREATE TABLE merchant_follows (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, merchant_id)
);
```

### browsing_history — 新增

```sql
CREATE TABLE browsing_history (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(20) NOT NULL,  -- 'store', 'review', 'post'
    target_id   BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 只保留最近 N 条，可通过定时任务清理
CREATE INDEX idx_browsing_history_user ON browsing_history(user_id, created_at DESC);
```

---

## 7. 优惠券、订单与券码系统

> 覆盖: CPV-001 ~ CPV-009, MRC-003, MRC-004

### coupons — 重构 (补齐字段)

```sql
CREATE TABLE coupons (
    id          BIGSERIAL PRIMARY KEY,
    merchant_id BIGINT       NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    store_id    BIGINT       REFERENCES stores(id) ON DELETE SET NULL,  -- NULL = 品牌通用

    title       VARCHAR(100) NOT NULL,
    description TEXT,
    image_url   VARCHAR(255),

    -- 优惠类型
    coupon_type VARCHAR(20) NOT NULL,  -- 'free', 'discount', 'package_deal'

    -- 价格
    original_price      DECIMAL(10,2),
    sale_price          DECIMAL(10,2) DEFAULT 0,  -- 0 = 免费
    discount_percentage DECIMAL(5,2),

    -- 库存
    total_quantity  INT,  -- NULL = 无限
    claimed_count   INT NOT NULL DEFAULT 0,
    redeemed_count  INT NOT NULL DEFAULT 0,

    -- 有效期
    valid_from  TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,

    -- 使用限制
    max_per_user INT DEFAULT 1,
    terms        TEXT,

    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    -- draft, active, paused, expired, deleted

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coupons_merchant_id ON coupons(merchant_id);
CREATE INDEX idx_coupons_store_id ON coupons(store_id);
CREATE INDEX idx_coupons_status ON coupons(status);
CREATE INDEX idx_coupons_valid ON coupons(valid_from, valid_until)
    WHERE status = 'active';
```

### packages — 新增

```sql
CREATE TABLE packages (
    id          BIGSERIAL PRIMARY KEY,
    merchant_id BIGINT       NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    store_id    BIGINT       REFERENCES stores(id) ON DELETE SET NULL,

    name        VARCHAR(255) NOT NULL,
    description TEXT,
    image_url   VARCHAR(255),

    original_price DECIMAL(10,2) NOT NULL,
    sale_price     DECIMAL(10,2) NOT NULL,

    -- 套餐内容
    items JSONB NOT NULL DEFAULT '[]',
    -- [{"name":"Item A","qty":2,"price":10.00}, ...]

    valid_from  TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,

    status VARCHAR(20) NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_packages_merchant_id ON packages(merchant_id);
CREATE INDEX idx_packages_status ON packages(status);
```

### orders — 新增

```sql
CREATE TABLE orders (
    id           BIGSERIAL PRIMARY KEY,
    order_no     VARCHAR(50) UNIQUE NOT NULL,  -- 业务订单号 (e.g. "ORD-20260214-XXXXX")

    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_id  BIGINT      NOT NULL REFERENCES merchants(id),

    -- 购买对象 (二选一)
    coupon_id    BIGINT      REFERENCES coupons(id),
    package_id   BIGINT      REFERENCES packages(id),
    order_type   VARCHAR(20) NOT NULL,  -- 'coupon', 'package'

    -- 金额
    original_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    final_amount    DECIMAL(10,2) NOT NULL,

    -- 订单状态
    order_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending, confirmed, completed, cancelled, refunded

    -- 冗余快照
    item_title   VARCHAR(255),  -- 购买时的商品标题
    merchant_name VARCHAR(100),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_merchant_id ON orders(merchant_id);
CREATE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_status ON orders(order_status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
```

### payments — 重构 (关联 order)

```sql
CREATE TABLE payments (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT      NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users(id),

    amount     DECIMAL(10,2) NOT NULL,
    currency   VARCHAR(10) NOT NULL DEFAULT 'USD',

    -- 支付渠道
    payment_method VARCHAR(50),  -- 'stripe', 'wechat', 'alipay', 'free'
    payment_session_id VARCHAR(255),  -- 第三方支付会话 ID

    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending, processing, completed, failed, refunded

    paid_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
```

### vouchers — 重构 (补齐核销链路)

```sql
CREATE TABLE vouchers (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50) UNIQUE NOT NULL,

    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coupon_id   BIGINT REFERENCES coupons(id),
    package_id  BIGINT REFERENCES packages(id),
    order_id    BIGINT REFERENCES orders(id),
    merchant_id BIGINT NOT NULL REFERENCES merchants(id),

    -- QR
    qr_code     VARCHAR(512),  -- QR 码内容 / URL

    -- 有效期
    valid_from  TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,

    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    -- active, redeemed, expired, cancelled

    -- 核销信息
    redeemed_at    TIMESTAMPTZ,
    redeemed_by    BIGINT REFERENCES users(id),  -- 核销操作人 (商家员工)
    redemption_note TEXT,

    -- 冗余快照
    coupon_title  VARCHAR(100),
    merchant_name VARCHAR(100),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vouchers_user_id ON vouchers(user_id);
CREATE INDEX idx_vouchers_merchant_id ON vouchers(merchant_id);
CREATE INDEX idx_vouchers_status ON vouchers(status);
CREATE INDEX idx_vouchers_code ON vouchers(code);
CREATE INDEX idx_vouchers_valid ON vouchers(valid_until)
    WHERE status = 'active';
```

---

## 8. 消息系统

> 覆盖: MSG-001 ~ MSG-005

### conversations — 新增

```sql
CREATE TABLE conversations (
    id                BIGSERIAL PRIMARY KEY,
    conversation_type VARCHAR(20) NOT NULL DEFAULT 'direct',  -- 'direct', 'group'

    -- group 元信息
    title      VARCHAR(255),
    avatar_url VARCHAR(255),

    -- 冗余: 最后一条消息 (提升列表查询性能)
    last_message_content   TEXT,
    last_message_at        TIMESTAMPTZ,
    last_message_sender_id BIGINT,

    is_archived BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_last_msg ON conversations(last_message_at DESC);
```

### conversation_participants — 新增

```sql
CREATE TABLE conversation_participants (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 角色 (可区分商家身份)
    role VARCHAR(20) DEFAULT 'member',  -- 'member', 'merchant', 'admin'

    -- 设置
    is_muted  BOOLEAN DEFAULT FALSE,
    is_pinned BOOLEAN DEFAULT FALSE,

    -- 未读计数
    unread_count       INT    DEFAULT 0,
    last_read_message_id BIGINT,

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at   TIMESTAMPTZ,

    UNIQUE (conversation_id, user_id)
);

CREATE INDEX idx_conv_participants_user ON conversation_participants(user_id);
CREATE INDEX idx_conv_participants_conv ON conversation_participants(conversation_id);
```

### messages — 新增

```sql
CREATE TABLE messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    content      TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text',  -- 'text', 'image', 'file', 'system'
    attachments  JSONB DEFAULT '[]',
    -- [{"url":"...","type":"image","name":"...","size":1024}]

    is_deleted BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conv ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
```

---

## 9. 商家认证

> 覆盖: VER-001, VER-002

### merchant_verifications — 新增

```sql
CREATE TABLE merchant_verifications (
    id          BIGSERIAL PRIMARY KEY,
    merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

    -- 申请资料
    business_license_no VARCHAR(100),
    legal_representative VARCHAR(100),
    business_address     TEXT,
    documents            JSONB DEFAULT '[]',
    -- [{"type":"license","url":"...","name":"..."}]

    -- 审核状态
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending, under_review, approved, rejected, resubmit_required

    -- 审核信息
    reviewed_by      BIGINT REFERENCES users(id),
    reviewed_at      TIMESTAMPTZ,
    rejection_reason TEXT,
    admin_notes      TEXT,

    submitted_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchant_verifications_merchant ON merchant_verifications(merchant_id);
CREATE INDEX idx_merchant_verifications_status ON merchant_verifications(status);
```

---

## 10. 营销与数据分析

> 覆盖: MRC-005, MRC-007, MRC-008, MRC-009

### marketing_posts — 新增

```sql
CREATE TABLE marketing_posts (
    id          BIGSERIAL PRIMARY KEY,
    merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    post_id     BIGINT UNIQUE REFERENCES posts(id) ON DELETE CASCADE,  -- 关联 posts 表

    campaign_name    VARCHAR(255),
    call_to_action   VARCHAR(255),
    cta_link         VARCHAR(512),

    -- 统计
    impression_count INT NOT NULL DEFAULT 0,
    click_count      INT NOT NULL DEFAULT 0,

    -- 定时发布
    scheduled_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_marketing_posts_merchant ON marketing_posts(merchant_id);
```

### merchant_analytics — 新增 (按天聚合)

```sql
CREATE TABLE merchant_analytics (
    id          BIGSERIAL PRIMARY KEY,
    merchant_id BIGINT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    date        DATE   NOT NULL,

    -- 流量
    page_views      INT DEFAULT 0,
    unique_visitors INT DEFAULT 0,

    -- 互动
    new_reviews   INT DEFAULT 0,
    avg_rating    DECIMAL(3,2),
    new_likes     INT DEFAULT 0,
    new_shares    INT DEFAULT 0,

    -- 转化
    coupon_views       INT DEFAULT 0,
    coupon_claims      INT DEFAULT 0,
    voucher_redemptions INT DEFAULT 0,

    -- 收入
    revenue     DECIMAL(12,2) DEFAULT 0,
    order_count INT DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (merchant_id, date)
);

CREATE INDEX idx_merchant_analytics_date ON merchant_analytics(merchant_id, date DESC);
```

### notifications — 新增 (业务通知)

```sql
CREATE TABLE notifications (
    id        BIGSERIAL PRIMARY KEY,
    user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 通知类型
    type VARCHAR(50) NOT NULL,
    -- new_review, new_reply, coupon_claimed, voucher_redeemed,
    -- verification_update, payment_received, new_follower, system

    title   VARCHAR(255) NOT NULL,
    content TEXT,

    -- 关联对象 (多态)
    related_type VARCHAR(50),  -- 'review', 'order', 'voucher', 'merchant', etc.
    related_id   BIGINT,

    action_url TEXT,

    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);
```

---

## 11. 管理系统

> 覆盖: ADM-001, ADM-002

### reports — 新增

```sql
CREATE TABLE reports (
    id          BIGSERIAL PRIMARY KEY,
    reporter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 被举报对象 (多态)
    target_type VARCHAR(50) NOT NULL,  -- 'review', 'post', 'comment', 'user', 'store'
    target_id   BIGINT      NOT NULL,

    reason  VARCHAR(50) NOT NULL,  -- 'spam', 'inappropriate', 'fake', 'harassment', 'other'
    details TEXT,

    -- 审核
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending, under_review, resolved, dismissed

    reviewed_by      BIGINT REFERENCES users(id),
    reviewed_at      TIMESTAMPTZ,
    resolution_notes TEXT,
    action_taken     VARCHAR(50),  -- 'deleted', 'warned', 'banned', 'no_action'

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_reporter ON reports(reporter_id);
CREATE INDEX idx_reports_target ON reports(target_type, target_id);
CREATE INDEX idx_reports_status ON reports(status);
```

### admin_audit_logs — 新增

```sql
CREATE TABLE admin_audit_logs (
    id       BIGSERIAL PRIMARY KEY,
    admin_id BIGINT NOT NULL REFERENCES users(id),

    action      VARCHAR(100) NOT NULL,  -- 'delete_review', 'ban_user', 'approve_merchant'
    target_type VARCHAR(50),
    target_id   BIGINT,
    details     JSONB,

    ip_address VARCHAR(45),
    user_agent TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_admin ON admin_audit_logs(admin_id);
CREATE INDEX idx_audit_logs_action ON admin_audit_logs(action);
CREATE INDEX idx_audit_logs_created ON admin_audit_logs(created_at DESC);
```

---

## 12. 媒体上传

> 覆盖: REV-002, REV-003

### media_uploads — 保留，增补字段

```sql
CREATE TABLE media_uploads (
    id         BIGSERIAL PRIMARY KEY,
    uuid       VARCHAR(36)  NOT NULL UNIQUE,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_key VARCHAR(512) NOT NULL,
    file_url   VARCHAR(512),

    -- v2 新增
    file_size  BIGINT,        -- bytes
    mime_type  VARCHAR(100),
    r2_bucket  VARCHAR(100),

    status     VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'completed', 'failed'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_uploads_user_id ON media_uploads(user_id);
CREATE INDEX idx_media_uploads_uuid ON media_uploads(uuid);
```

---

## 13. 冗余设计总结

### 13.1 计数器字段

| 表 | 冗余字段 | 来源 | 更新策略 |
|----|---------|------|---------|
| `user_profiles` | follower_count, following_count, post_count, review_count, like_count, coupon_count | 各表聚合 | 应用层事务内 +1/-1 |
| `merchants` | total_stores, total_reviews, avg_rating, follower_count | stores, reviews, merchant_follows | 应用层 + 定时校准 |
| `stores` | review_count, avg_rating | reviews | 定时任务 (每小时) |
| `reviews` | like_count, comment_count | likes, review_comments | 应用层事务内 +1/-1 |
| `posts` | like_count, comment_count, share_count, view_count | likes, post_comments | 应用层 + Redis 缓冲 |
| `tags` | post_count, review_count | post_tags, review_tags | 应用层 +1/-1 |
| `coupons` | claimed_count, redeemed_count | vouchers | 应用层事务内 |

### 13.2 快照字段

| 表 | 冗余字段 | 来源 | 更新策略 |
|----|---------|------|---------|
| `reviews` | user_nickname, user_avatar, store_name | user_profiles, stores | 创建时写入，不更新 |
| `review_comments` | user_nickname, user_avatar | user_profiles | 创建时写入 |
| `posts` | author_name, author_avatar | user_profiles / merchants | 创建时写入 |
| `post_comments` | user_nickname, user_avatar | user_profiles | 创建时写入 |
| `orders` | item_title, merchant_name | coupons/packages, merchants | 创建时写入 |
| `vouchers` | coupon_title, merchant_name | coupons, merchants | 创建时写入 |
| `conversations` | last_message_content, last_message_at, last_message_sender_id | messages | 新消息时更新 |

### 13.3 计数器校准 SQL

```sql
-- 定时任务: 校准 store 评分和评论数
UPDATE stores s SET
    review_count = sub.cnt,
    avg_rating = sub.avg
FROM (
    SELECT store_id, COUNT(*) AS cnt, AVG(rating)::DECIMAL(3,2) AS avg
    FROM reviews WHERE status = 0
    GROUP BY store_id
) sub
WHERE s.id = sub.store_id
  AND (s.review_count != sub.cnt OR s.avg_rating != sub.avg);

-- 定时任务: 校准 merchant 汇总
UPDATE merchants m SET
    total_reviews = sub.cnt,
    avg_rating = sub.avg,
    total_stores = (SELECT COUNT(*) FROM stores WHERE merchant_id = m.id AND status = 0)
FROM (
    SELECT merchant_id, COUNT(*) AS cnt, AVG(rating)::DECIMAL(3,2) AS avg
    FROM reviews WHERE status = 0
    GROUP BY merchant_id
) sub
WHERE m.id = sub.merchant_id;

-- 定时任务: 校准 user_profiles.review_count
UPDATE user_profiles up SET
    review_count = (SELECT COUNT(*) FROM reviews WHERE user_id = up.user_id AND status = 0);
```

---

## 14. 从现有库迁移策略

### 14.1 需要做的数据迁移

| 步骤 | 操作 | 风险 | 回滚方案 |
|------|------|------|---------|
| 1 | `users.role` 增加 `'merchant'` 枚举值 | 低 | 无需回滚 |
| 2 | 创建 `stores` 表，从 `merchants` 迁移门店级字段 | 中 | 保留原字段 |
| 3 | `reviews.venue_id` → `reviews.store_id` (重命名 + 外键更新) | 高 | 备份原表 |
| 4 | 创建 `categories`, `store_categories` | 低 | 新表可直接 DROP |
| 5 | 创建 `store_hours` | 低 | 新表可直接 DROP |
| 6 | `merchants` 增加 `user_id`, `verification_status` 等字段 | 中 | ALTER ADD 可逆 |
| 7 | 创建 `review_media`，从 `reviews.images` JSONB 迁移 | 中 | 保留 images 字段 |
| 8 | `review_comments` 增加 `parent_comment_id`, `is_merchant_reply` | 低 | ALTER ADD 可逆 |
| 9 | `posts` 增加 `post_type`, `comment_count`, `share_count` | 低 | ALTER ADD 可逆 |
| 10 | `tags` 增加 `type`, `review_count` | 低 | ALTER ADD 可逆 |
| 11 | `coupons` 补齐字段 | 低 | ALTER ADD 可逆 |
| 12 | 创建 `orders`, `packages` | 低 | 新表可直接 DROP |
| 13 | `payments` 增加 `order_id`, `payment_method`, `payment_session_id` | 中 | ALTER ADD 可逆 |
| 14 | `vouchers` 增加核销字段 | 低 | ALTER ADD 可逆 |
| 15 | `media_uploads` 增加 `file_size`, `mime_type`, `r2_bucket` | 低 | ALTER ADD 可逆 |
| 16 | 创建消息系统三张表 | 低 | 新表可直接 DROP |
| 17 | 创建 `merchant_verifications` | 低 | 新表可直接 DROP |
| 18 | 创建营销/分析/通知/管理四张表 | 低 | 新表可直接 DROP |
| 19 | 创建 `browsing_history` | 低 | 新表可直接 DROP |
| 20 | 将 `coupons`, `vouchers`, `payments`, `review_comments` 加入 AutoMigrate | 低 | 移除即可 |

### 14.2 迁移顺序建议

```
Phase 1 (Week 1): 商家/门店体系
  ├─ merchants 增加 user_id 等字段
  ├─ 创建 stores (从 merchants 迁移门店数据)
  ├─ 创建 categories + store_categories + store_hours
  └─ reviews.venue_id → store_id

Phase 2 (Week 2): 评论 + 社区补全
  ├─ 创建 review_media (迁移 reviews.images)
  ├─ review_comments 增加嵌套 + 商家回复
  ├─ 创建 post_comments
  └─ posts, tags 增补字段

Phase 3 (Week 3): 优惠券/订单链路
  ├─ coupons 补齐字段
  ├─ 创建 orders, packages
  ├─ payments 增加 order_id 关联
  ├─ vouchers 补齐核销字段
  └─ AutoMigrate 添加上述表

Phase 4 (Week 4): 消息 + 认证 + 管理
  ├─ 创建消息系统三表
  ├─ 创建 merchant_verifications
  ├─ 创建 notifications, reports, admin_audit_logs
  ├─ 创建 marketing_posts, merchant_analytics
  └─ 创建 browsing_history
```

### 14.3 关键迁移 SQL 示例

```sql
-- Step 2: merchants → stores 数据迁移
-- 为每个现有 merchant 创建一条 store 记录
INSERT INTO stores (merchant_id, name, description, address, phone, cover_image_url, avg_rating, review_count, status, created_at, updated_at)
SELECT id, name, description, address, phone, cover_image, avg_rating, review_count, status, created_at, updated_at
FROM merchants;

-- Step 3: reviews.venue_id → store_id
-- 先建立 venue_id → store_id 的映射 (venue_id 原本就是 merchant.id)
ALTER TABLE reviews ADD COLUMN store_id BIGINT;
UPDATE reviews r SET store_id = s.id
FROM stores s WHERE s.merchant_id = r.venue_id;
ALTER TABLE reviews ALTER COLUMN store_id SET NOT NULL;
ALTER TABLE reviews ADD CONSTRAINT fk_reviews_store FOREIGN KEY (store_id) REFERENCES stores(id);
-- 确认无误后可以 DROP venue_id

-- Step 7: reviews.images JSONB → review_media
INSERT INTO review_media (review_id, media_type, url, sort_order, created_at)
SELECT
    r.id,
    'image',
    img.value::text,
    img.ordinality - 1,
    r.created_at
FROM reviews r,
LATERAL jsonb_array_elements_text(r.images::jsonb) WITH ORDINALITY AS img(value, ordinality)
WHERE r.images IS NOT NULL AND r.images != '[]';
```

---

## 15. AutoMigrate 完整列表 (目标态)

```go
// cmd/app/main.go
if err := database.DB.AutoMigrate(
    // === 用户系统 (现有) ===
    &model.User{},
    &model.UserAuth{},
    &model.UserProfile{},
    &model.EmailVerification{},
    &model.UserAddress{},
    &model.UserPrivacy{},
    &model.UserNotification{},
    &model.AccountDeletion{},

    // === 商家与门店 ===
    &model.Merchant{},          // 重构
    &model.Store{},             // 新增
    &model.StoreHour{},         // 新增

    // === 分类与标签 ===
    &model.Category{},          // 新增
    &model.Tag{},               // 重构

    // === 内容 ===
    &model.Review{},            // 重构
    &model.ReviewMedia{},       // 新增
    &model.ReviewComment{},     // 重构
    &model.Post{},              // 重构
    &model.PostComment{},       // 新增

    // === 互动 ===
    &model.Like{},
    &model.Favorite{},
    &model.UserFollow{},
    &model.MerchantFollow{},

    // === 优惠券与订单 ===
    &model.Coupon{},            // 重构
    &model.Package{},           // 新增
    &model.Order{},             // 新增
    &model.Payment{},           // 重构
    &model.Voucher{},           // 重构

    // === 消息 ===
    &model.Conversation{},      // 新增
    &model.ConversationParticipant{}, // 新增
    &model.Message{},           // 新增

    // === 认证 ===
    &model.MerchantVerification{}, // 新增

    // === 营销与分析 ===
    &model.MarketingPost{},     // 新增
    &model.MerchantAnalytics{}, // 新增
    &model.Notification{},      // 新增

    // === 管理 ===
    &model.Report{},            // 新增
    &model.AdminAuditLog{},     // 新增

    // === 用户行为 ===
    &model.MediaUpload{},       // 重构
    &model.BrowsingHistory{},   // 新增
); err != nil {
    logger.Error(ctx, "Failed to migrate database", "error", err.Error())
    os.Exit(1)
}
```

---

## 16. 需求 ID 覆盖度矩阵

| 需求 ID | 涉及的表 | 覆盖状态 |
|---------|---------|---------|
| AUT-001~007 | users, user_auths, email_verifications | ✅ 已覆盖 |
| AUT-008 | (前端) | ✅ 无需后端表 |
| AUT-009~010 | email_verifications | ✅ 已覆盖 |
| AUT-011 | user_auths (refresh token) | ✅ credential 字段可存 |
| AUT-012~013 | (前端) | ✅ 无需后端表 |
| DIS-001 | stores, categories, store_categories | ✅ |
| DIS-002~003 | stores, categories, store_hours | ✅ |
| DIS-004 | categories, store_categories | ✅ |
| DIS-005 | stores (全文索引) | ✅ |
| DIS-006~007 | stores (PostGIS location) | ✅ |
| DIS-008 | stores, merchants, reviews | ✅ |
| DIS-009 | reviews, review_media, review_comments | ✅ |
| DIS-010 | stores.menu_images | ✅ |
| REV-001 | reviews, tags, review_tags | ✅ |
| REV-002 | review_media, media_uploads | ✅ |
| REV-003 | media_uploads | ✅ |
| REV-004 | reviews | ✅ |
| REV-005 | (前端 localStorage) | ✅ 无需后端表 |
| REV-006 | (前端 AI) | ✅ 无需后端表 |
| REV-007 | likes, review_comments | ✅ |
| REV-008 | (前端) | ✅ 无需后端表 |
| COM-001 | reviews, posts | ✅ |
| COM-002 | posts, post_comments | ✅ |
| COM-003 | likes | ✅ |
| COM-004 | post_comments | ✅ |
| COM-005 | posts.share_count | ✅ |
| CPV-001 | coupons, vouchers | ✅ |
| CPV-002 | coupons, vouchers | ✅ |
| CPV-003 | payments | ✅ |
| CPV-004 | orders, payments | ✅ |
| CPV-005 | orders | ✅ |
| CPV-006 | vouchers | ✅ |
| CPV-007 | vouchers | ✅ |
| CPV-008 | vouchers | ✅ |
| CPV-009 | vouchers (核销字段) | ✅ |
| PRF-001 | user_profiles, reviews | ✅ |
| PRF-002 | reviews | ✅ |
| PRF-003 | user_profiles, user_privacies, user_notifications | ✅ |
| PRF-004 | (前端 AI) | ✅ 无需后端表 |
| PRF-005 | (可通过 orders/vouchers 查询) | ✅ |
| PRF-006 | merchant_follows, browsing_history, vouchers | ✅ |
| MRC-001 | merchants, merchant_analytics | ✅ |
| MRC-002 | review_comments (is_merchant_reply) | ✅ |
| MRC-003 | coupons | ✅ |
| MRC-004 | packages | ✅ |
| MRC-005 | marketing_posts, posts | ✅ |
| MRC-006 | stores | ✅ |
| MRC-007 | merchant_analytics | ✅ |
| MRC-008 | (可复用 marketing_posts 或后续独立表) | ⚠️ 暂用 marketing_posts |
| MRC-009 | notifications | ✅ |
| MSG-001 | conversations, conversation_participants | ✅ |
| MSG-002 | messages | ✅ |
| MSG-003 | conversation_participants (mute/pin) | ✅ |
| MSG-004 | messages (全文搜索) | ✅ |
| MSG-005 | conversations (group), conversations (delete) | ✅ |
| VER-001 | merchant_verifications | ✅ |
| VER-002 | merchants, users | ✅ |
| VER-003 | (前端) | ✅ 无需后端表 |
| ADM-001 | admin_audit_logs, merchants, stores | ✅ |
| ADM-002 | reports | ✅ |

**覆盖率: 100% (所有 Backend-Required 需求均有对应表结构)**

---

**版本**: v2.0
**更新日期**: 2026-02-14
**状态**: 待确认
