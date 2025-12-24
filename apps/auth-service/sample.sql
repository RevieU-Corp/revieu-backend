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
('11111111-1111-1111-1111-111111111112', 'weijun', 'weijun@example.com', 'hashed_password_here122', 'admin', TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tb_profiles (id, user_id, nickname)
VALUES
('22222222-2222-2222-2222-222222222221', '11111111-1111-1111-1111-111111111111', 'AliceInWonderland'),
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111112', 'WeiJunAdmin')
ON CONFLICT (id) DO NOTHING;