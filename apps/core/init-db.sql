-- =============================================
-- RevieU Database Initialization Script
-- Platform: PostgreSQL 14+ with PostGIS
-- Author: Wayne & Gemini
-- =============================================

-- 1. 初始化设置 & 扩展
-- ---------------------------------------------
-- 启用 PostGIS 扩展 (必须先安装 postgis)
CREATE EXTENSION IF NOT EXISTS postgis;

-- 清理旧表 (开发阶段专用，生产环境请慎用 CASCADE)
DROP TABLE IF EXISTS review_votes CASCADE;
DROP TABLE IF EXISTS review_comments CASCADE;
DROP TABLE IF EXISTS review_media CASCADE;
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS venue_categories CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS venues CASCADE;
DROP TABLE IF EXISTS email_verifications CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;
DROP TABLE IF EXISTS user_auths CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- =============================================
-- A. 用户系统 (Identity & Authentication)
-- =============================================

-- 1. 核心用户表 (只存不可变/系统级信息)
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- 'user', 'admin'
    status SMALLINT NOT NULL DEFAULT 0,       -- 0: active, 1: banned, 2: pending
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. 用户认证表 (支持多重登录方式：Email/Google/WeChat)
CREATE TABLE user_auths (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    identity_type VARCHAR(20) NOT NULL, -- 'email', 'google', 'apple'
    identifier VARCHAR(255) NOT NULL,   -- email地址 或 open_id/sub
    credential VARCHAR(255),            -- 密码hash 或 access_token (可选)
    last_login_at TIMESTAMPTZ,
    
    -- 保证同一类型的登录方式唯一 (比如一个邮箱只能注册一次)
    UNIQUE (identity_type, identifier)
);

-- 3. 用户画像表 (业务展示信息，解耦)
CREATE TABLE user_profiles (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    nickname VARCHAR(50) NOT NULL,
    avatar_url VARCHAR(255),
    intro VARCHAR(255), -- 一句话简介
    location VARCHAR(100)
);

-- 4. 邮箱验证表 (Email Verification)
CREATE TABLE email_verifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verifications_token ON email_verifications(token);
CREATE INDEX idx_email_verifications_user ON email_verifications(user_id);

-- =============================================
-- B. 商户/地点系统 (Venues & Taxonomy)
-- =============================================

-- 4. 商户表 (核心实体)
CREATE TABLE venues (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    address VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    website VARCHAR(255),
    
    -- [PostGIS] 地理位置: 使用 WGS 84 (SRID 4326)
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    
    -- 商家认领 (可选)
    owner_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    
    -- [反范式设计] 性能冗余字段 (由触发器或应用层异步更新)
    avg_rating DECIMAL(3, 2) DEFAULT 0.00, -- 缓存平均分 (e.g. 4.50)
    review_count INT DEFAULT 0,            -- 缓存评论数
    cover_image_url VARCHAR(255),          -- 缓存封面图
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. 分类表 (支持层级, e.g. Food -> Asian -> Chinese)
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    parent_id INT REFERENCES categories(id) ON DELETE SET NULL,
    icon_url VARCHAR(255)
);

-- 6. 商户-分类关联表 (多对多)
CREATE TABLE venue_categories (
    venue_id BIGINT REFERENCES venues(id) ON DELETE CASCADE,
    category_id INT REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (venue_id, category_id)
);

-- =============================================
-- C. 点评系统 (Reviews & Interactions)
-- =============================================

-- 7. 评论主表
CREATE TABLE reviews (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    venue_id BIGINT NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    
    rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    content TEXT, -- 允许长文本
    visit_date DATE NOT NULL, -- 用户实际去的时间
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 约束: 一个用户对一个商户原则上只有一条主评
    UNIQUE (user_id, venue_id) 
);

-- 8. 评论多媒体表 (图床链接)
CREATE TABLE review_media (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    media_type VARCHAR(10) NOT NULL CHECK (media_type IN ('image', 'video')),
    url VARCHAR(512) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. 评论回复表 (互动层)
CREATE TABLE review_comments (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- 回复者
    
    reply_to_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL, -- 被回复者(可选)
    content TEXT NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 10. 评论点赞/投票表 (质量层)
CREATE TABLE review_votes (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_type VARCHAR(20) NOT NULL, -- 'useful', 'funny', 'cool'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 防止重复刷票
    UNIQUE (review_id, user_id, vote_type)
);

-- =============================================
-- D. 索引优化 (Indexes)
-- =============================================

-- Users: 登录查找
CREATE INDEX idx_users_email ON user_auths(identifier) WHERE identity_type = 'email';
CREATE INDEX idx_users_role ON users(role);

-- Venues: PostGIS 核心索引 (GiST R-Tree)
CREATE INDEX idx_venues_location ON venues USING GIST (location);
-- Venues: 名称搜索 (简单B-tree，如需模糊搜索建议加 pg_trgm)
CREATE INDEX idx_venues_name ON venues(name);

-- Reviews: 
-- 1. 餐厅详情页加载评论: WHERE venue_id = ? ORDER BY created_at DESC
CREATE INDEX idx_reviews_venue_date ON reviews(venue_id, created_at DESC);
-- 2. 用户个人页加载评论: WHERE user_id = ?
CREATE INDEX idx_reviews_user ON reviews(user_id);

-- Foreign Keys (PG不会自动给FK建索引，手动加上以优化 JOIN)
CREATE INDEX idx_review_media_rid ON review_media(review_id);
CREATE INDEX idx_review_comments_rid ON review_comments(review_id);
CREATE INDEX idx_venue_cats_vid ON venue_categories(venue_id);
CREATE INDEX idx_venue_cats_cid ON venue_categories(category_id);

-- =============================================
-- E. 测试数据种子 (Seed Data Example)
-- =============================================
-- 仅作为示例，演示如何插入 PostGIS 数据
/*
INSERT INTO venues (name, address, location, created_at)
VALUES (
    'USC Village', 
    '3015 S Hoover St, Los Angeles, CA 90007',
    ST_SetSRID(ST_MakePoint(-118.2851, 34.0224), 4326)::geography, 
    NOW()
);
*/