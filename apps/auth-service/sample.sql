-- schema.sql
CREATE DATABASE IF NOT EXISTS USCRE;

USE USCRE;

CREATE TABLE IF NOT EXISTS tb_users (
    id CHAR(36) PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE
);

-- 可以插入测试用户
INSERT INTO tb_users (id, username, email, password_hash, is_active, is_verified)
VALUES
('11111111-1111-1111-1111-111111111111', 'alice', 'alice@example.com', 'hashed_password_here', TRUE, TRUE),
('11111111-1111-1111-1111-111111111112', 'weijun', 'weijun@example.com', 'hashed_password_here122', TRUE, TRUE);