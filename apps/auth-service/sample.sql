-- schema.sql
-- CREATE DATABASE is typically handled by POSTGRES_DB env var or manually, 
-- but we can keep it if we want to run this manually. 
-- In docker-entrypoint-initdb.d, the DB specified in POSTGRES_DB is already created.
-- So usually we just connect to it.

CREATE TABLE IF NOT EXISTS tb_users (
    id CHAR(36) PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    nickname VARCHAR(50) DEFAULT NULL,
    avatar VARCHAR(255) DEFAULT NULL,
    bio VARCHAR(500) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP DEFAULT NULL,
    last_login_ip VARCHAR(45) DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_modtime
    BEFORE UPDATE ON tb_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 可以插入测试用户
INSERT INTO tb_users (id, username, email, password_hash, role, nickname, is_active, is_verified)
VALUES
('11111111-1111-1111-1111-111111111111', 'alice', 'alice@example.com', 'hashed_password_here', 'user', 'AliceInWonderland', TRUE, TRUE),
('11111111-1111-1111-1111-111111111112', 'weijun', 'weijun@example.com', 'hashed_password_here122', 'admin', 'WeiJunAdmin', TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;