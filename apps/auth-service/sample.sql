-- schema.sql
CREATE DATABASE IF NOT EXISTS USCRE;

USE USCRE;

CREATE TABLE IF NOT EXISTS tb_users (
    id CHAR(36) PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at DATETIME DEFAULT NULL,
    last_login_ip VARCHAR(45) DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS tb_profiles (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL UNIQUE,
    nickname VARCHAR(50) DEFAULT NULL,
    avatar VARCHAR(255) DEFAULT NULL,
    bio VARCHAR(500) DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES tb_users(id) ON DELETE CASCADE
);

-- 可以插入测试用户
INSERT INTO tb_users (id, username, email, password_hash, role, is_active, is_verified)
VALUES
('11111111-1111-1111-1111-111111111111', 'alice', 'alice@example.com', 'hashed_password_here', 'user', TRUE, TRUE),
('11111111-1111-1111-1111-111111111112', 'weijun', 'weijun@example.com', 'hashed_password_here122', 'admin', TRUE, TRUE);

INSERT INTO tb_profiles (id, user_id, nickname)
VALUES
('22222222-2222-2222-2222-222222222221', '11111111-1111-1111-1111-111111111111', 'AliceInWonderland'),
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111112', 'WeiJunAdmin');