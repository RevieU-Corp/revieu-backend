# 用户认证 API 文档

## 概述
- 提供用户认证相关接口：注册、邮箱验证、登录、获取用户信息。
- 返回统一 JSON 格式：`code`（0 表示成功）、`message`、`data`（成功时为对象/数组，失败时通常为 `null` 或省略）。

## 认证
- 使用 JWT（Bearer token）
- token 有效期为 1 小时

## 统一响应格式（JSON）
成功响应示例：
```json
{
    "code": 0,
    "message": "success message",
    "data": { /* response data */ }
}
```

错误响应示例（常见样式）：
```json
{
    "code": 1,
    "message": "error message",
    "data": null
}
```

> 说明：代码实现中 `app/auth/routes.py` 使用统一的 `code`/`message`/`data` 格式。HTTP 状态码与 `code` 字段配合使用以表达语义（例如 201 创建成功、400 参数错误、401 认证失败、404 未找到）。

## 常见 HTTP 错误码
- `400`: Bad Request — 参数校验失败或请求无效
- `401`: Unauthorized — 认证失败、token 无效或过期
- `404`: Not Found — 资源（如用户）不存在
- `500`: Internal Server Error — 服务内部错误


## 接口列表

### 1. 用户注册 — `POST` `/api/v1/auth/register`

#### 请求头
        Content-Type: application/json

#### 请求体
```json
{
    "username": "alice",
    "email": "alice@example.com",
    "password": "your_password"
}
```

#### 成功响应 (201)
```json
{
    "code": 0,
    "message": "user created, verification email sent",
    "data": {
        "id": "uuid-string",
        "email": "alice@example.com",
        "username": "alice",
        "created_at": "2025-10-27T12:00:00+00:00"
    }
}
```

#### 错误响应 (示例)
- `400`: 用户已存在或参数错误
```json
{
    "code": 1,
    "message": "user already exists",
    "data": null
}
```

### 2. 邮箱验证 — `GET` `/api/v1/auth/verify/<token>`

#### 参数
- `token`: 通过邮件发送的验证 token

#### 成功响应 (200)
```json
{
    "code": 0,
    "message": "email verified successfully",
    "data": null
}
```

#### 其它响应示例
- `400`: token 无效或过期
```json
{
    "code": 1,
    "message": "invalid or expired token",
    "data": null
}
```
- `404`: 用户不存在
```json
{
    "code": 2,
    "message": "user not found",
    "data": null
}
```
- `200`: 已验证过（实现中返回 code=3 与 200 状态）
```json
{
    "code": 3,
    "message": "already verified",
    "data": null
}
```

### 3. 用户登录 — `POST` `/api/v1/auth/login`

#### 请求头
        Content-Type: application/json

#### 请求体
```json
{
    "email": "alice@example.com",
    "password": "your_password"
}
```

#### 成功响应 (200)
```json
{
    "code": 0,
    "message": "login successful",
    "data": { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
}
```

#### 错误响应 (401)
```json
{
    "code": 1,
    "message": "Invalid credentials, incorrect email or password",
    "data": null
}
```

### 4. 获取用户信息 — `GET` `/api/v1/auth/profile`

#### 请求头
        Authorization: Bearer <token>

#### 成功响应 (200)
```json
{
    "code": 0,
    "message": "User profile fetched successfully",
    "data": {
        "username": "alice",
        "email": "alice@example.com",
        "created_at": "2025-10-27T12:00:00+00:00"
    }
}
```

#### 错误响应（示例）
- `401`: token 缺失/无效/过期
```json
{
    "code": 1,
    "message": "Missing token",
    "data": null
}
```
或
```json
{
    "code": 1,
    "message": "Invalid token",
    "data": null
}
```
- `401`: token 过期
```json
{
    "code": 1,
    "message": "Token expired",
    "data": null
}
```
- `404`: 用户不存在
```json
{
    "code": 2,
    "message": "User not found",
    "data": null
}
```

## 第三方 OAuth 登录（Google / GitHub）接口列表

该项目在 `app/auth/routes.py` 中同时实现了 GitHub 与 Google 的 OAuth 登录流程。路由基准路径为 `/api/v1/auth/`（视 Blueprint 注册前缀而定）。下面列出常用端点及其行为：

### 5. GitHub 登录

- 1) 跳转到 GitHub 授权页 — `GET` `/api/v1/auth/github/login`
    - 功能：将用户重定向到 GitHub 的 OAuth 授权页面。无需请求体或特殊 header。

- 2) 授权回调 — `GET` `/api/v1/auth/github/callback?code=<code>`
    - 功能：从查询参数读取 `code`，用该 code 向 GitHub 换取 access token，然后使用该 token 获取用户信息与邮箱列表。若找到经过验证的 primary 邮箱，会使用该邮箱登录或注册（通过服务层 `AuthService.login_or_register_github_user`），并返回应用内 JWT token。

#### 成功响应示例 (200)：
```json
{
    "code": 0,
    "message": "GitHub login successful",
    "data": { "email": "alice@example.com", "name": "alice", "token": "<jwt>" }
}
```

#### 错误响应：
- `400` 缺少 code：
```json
{
    "code": 1,
    "message": "Missing code",
    "data": null
}
```
- `400` 未能获取 access token：
```json
{
    "code": 1,
    "message": "Failed to get access token",
    "data": { /* token endpoint 返回的原始响应 */ }
}
```
- `400` 无经过验证的邮箱：
```json
{
    "code": 1,
    "message": "No verified email found in GitHub account",
    "data": null
}
```

### 6. Google 登录

- 1) 跳转到 Google 授权页 — `GET` `/api/v1/auth/google/login`
    - 功能：重定向到 Google 的 OAuth 授权页面，使用 `scope=open id email profile` 等参数。

- 2) 授权回调 — `GET` `/api/v1/auth/google/callback?code=<code>`
    - 功能：用 `code` 交换 access token（向 `https://oauth2.googleapis.com/token`），然后用 access token 请求用户信息（`/oauth2/v3/userinfo`）。使用返回的 email/name 调用 `AuthService.login_or_register_google_user`，并返回应用内 JWT token。

#### 成功响应示例 (200)：
```json
{
    "code": 0,
    "message": "Google login successful",
    "data": { "email": "alice@example.com", "name": "Alice", "token": "<jwt>" }
}
```

#### 错误响应：
- `400` 缺少 code：
```json
{
    "code": 1,
    "message": "Missing code",
    "data": null
}
```
- `400` 未能获取 access token：
```json
{
    "code": 1,
    "message": "Failed to get access token",
    "data": { /* token endpoint 返回的原始响应 */ }
}
```

> 备注：两端 OAuth 回调处都会将第三方获得的 email/name 传入服务层，服务层负责注册/登录逻辑并返回应用内的 JWT token。回调会将最终结果包装为统一响应格式 {code, message, data}。

## 实现细节

### 安全特性
- JWT token 有效期为 1 小时
- 使用 `werkzeug.security` 进行密码哈希
- 邮箱验证 token 有效期为 15 分钟
- 所有密码相关操作使用安全哈希存储

### 数据验证
- 用户名无唯一性验证
- 邮箱唯一性检查
- 邮箱验证流程（发送邮件 + 回调验证）

### 邮件服务
- 支持通过 SMTP 发送验证邮件
- 可配置的邮件服务器设置（见环境变量配置）

### 环境变量配置
- `JWT_SECRET_KEY`: JWT 签名密钥
- `MYSQL_ROOT_PASSWORD`: MySQL 根密码
- `MYSQL_DATABASE`: MySQL 数据库名称
- `PORT`: 应用监听端口
- `ADDRESS`: 应用监听地址
- `LOG_DIR`: 日志文件目录
- `FLASK_ENV`: Flask 环境（默认：development）

- `MAIL_SERVER`: SMTP 服务器地址（默认：smtp.gmail.com）
- `MAIL_PORT`: SMTP 端口（默认：465）
- `MAIL_USE_TLS`: 是否使用 TLS（默认：False）
- `MAIL_USE_SSL`: 是否使用 SSL（默认：True）
- `MAIL_USERNAME`: SMTP 用户名
- `MAIL_PASSWORD`: SMTP 密码
- `MAIL_SENDER_NAME`: 发件人名称（默认：USCRE）
- `MAIL_SENDER_EMAIL`: 发件人邮箱（默认：与 MAIL_USERNAME 相同）

- `GOOGLE_CLIENT_ID`: Google 客户端 ID
- `GOOGLE_CLIENT_SECRET`: Google 客户端密钥
- `GOOGLE_REDIRECT_URI`: Google 授权回调 URI

- `GITHUB_CLIENT_ID`: GitHub 客户端 ID
- `GITHUB_CLIENT_SECRET`: GitHub 客户端密钥
- `GITHUB_REDIRECT_URI`: GitHub 授权回调 URI