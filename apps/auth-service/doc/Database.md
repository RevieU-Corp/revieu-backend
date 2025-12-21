# USCRE 数据库设计文档

## 1. 数据库概述

**数据库名称**：`USCRE`
**用途**：用户注册、登录、密码找回、账号安全管理等核心功能。
**字符集**：`utf8mb4`，支持多语言字符
**Collation**：`utf8mb4_unicode_ci`

---

## 2. 表结构

### 2.1 用户表 `tb_users`

| 字段名                     | 类型           | 主键 | 唯一 | 默认值                                           | 允许空 | 说明              |
| ----------------------- | ------------ | -- | -- | --------------------------------------------- | --- | --------------- |
| `id`                  | CHAR(36)     | ✅  | ✅  | 无                                             | ❌   | 用户全局唯一标识（UUID）  |
| `username`              | VARCHAR(50)  | ❌  | ❌  | 无                                             | ❌   | 用户名，可用于登录       |
| `email`                 | VARCHAR(120) | ❌  | ✅  | 无                                             | ❌   | 邮箱，可用于登录/找回密码   |
| `password_hash`         | VARCHAR(255) | ❌  | ❌  | 无                                             | ❌   | 哈希后的密码，不存明文     |
| `created_at`            | DATETIME     | ❌  | ❌  | CURRENT_TIMESTAMP                             | ❌   | 用户注册时间          |
| `updated_at`            | DATETIME     | ❌  | ❌  | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | ❌   | 最后更新时间          |
| `is_active`             | BOOLEAN      | ❌  | ❌  | TRUE                                          | ❌   | 账户是否启用          |
| `is_verified`           | BOOLEAN      | ❌  | ❌  | FALSE                                         | ❌   | 邮箱是否验证          |

**说明**：

* `id` 是全局唯一标识，用户注册时由后端生成。
* `password_hash` 必须使用安全哈希算法（如 PBKDF2、bcrypt、argon2 等）。
* `is_active` 用于软删除用户或禁用账户。
* `is_verified` 标记用户邮箱是否已通过验证。
---

## 3. 示例 SQL 脚本

```sql
-- schema.sql
CREATE DATABASE IF NOT EXISTS USCRE;

USE USCRE;

CREATE TABLE IF NOT EXISTS tb_users (
    id CHAR(36) PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(128) NOT NULL,
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
```

---

## 4. 设计规范与建议

1. **UUID 主键**：保证分布式唯一性，便于跨服务引用。
2. **密码安全**：禁止存储明文密码，必须哈希并加盐。
3. **时间字段**：统一使用 UTC 时间或数据库默认时区，方便日志统计和跨时区处理。
4. **安全字段**：`is_active`、`is_verified`、`failed_login_attempts`、`lock_until` 统一管理用户状态和登录安全。
5. **扩展性**：可以在 `tb_users` 之外增加用户配置表、角色权限表等，保持主表简洁。

---
