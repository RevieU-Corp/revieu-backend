# RevieU Backend API 全量测试案例（OpenAPI v1）

- 文档来源：`apps/core/docs/swagger.json`
- 文档用途：给协作开发者执行统一 API 回归测试
- 覆盖范围：`94` 个 API 操作（method + path）
- Base URL：`https://dev.revieu.weijun.online/api/v1`（可替换为本地）

## 通用约定

1. 需要鉴权的接口统一加：`Authorization: Bearer <ACCESS_TOKEN>`
2. 默认请求头：`Content-Type: application/json`
3. 如果返回体由 schema 自动推导，字段值为示例值，不代表线上真实数据
4. 每个接口至少验证：状态码、响应结构、关键业务字段

## 全量目录（按 Tag）

- `admin`：4 个接口
- `ai`：1 个接口
- `auth`：8 个接口
- `category`：1 个接口
- `content`：6 个接口
- `conversation`：5 个接口
- `coupon`：5 个接口
- `feed`：1 个接口
- `follow`：4 个接口
- `media`：3 个接口
- `merchant`：3 个接口
- `notification`：3 个接口
- `order`：4 个接口
- `package`：2 个接口
- `payment`：2 个接口
- `profile`：1 个接口
- `review`：5 个接口
- `store`：9 个接口
- `user`：16 个接口
- `verification`：2 个接口
- `voucher`：9 个接口

---

## Tag: admin

### TC-001 `GET /admin/merchants`

- 用例名称：List merchants for admin
- 说明：Returns a list of merchants for admin management
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/admin/merchants
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-002 `PATCH /admin/merchants/{id}`

- 用例名称：Update merchant status
- 说明：Updates a merchant's status or verification
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/admin/merchants/1
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-003 `GET /admin/reports`

- 用例名称：List reports
- 说明：Returns a list of user reports for admin review
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/admin/reports
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-004 `PATCH /admin/reports/{id}`

- 用例名称：Update report
- 说明：Updates a report status (approve/reject)
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/admin/reports/1
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: ai

### TC-005 `POST /ai/reviews/suggestions`

- 用例名称：Get review suggestions
- 说明：Generates AI suggestions for improving a review
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/ai/reviews/suggestions
Content-Type: application/json
```

Body 示例：

```json
{
  "businessCategory": "string",
  "currentText": "string",
  "merchantName": "string",
  "overallRating": 1.1
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "suggestions": [
    "string"
  ]
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: auth

### TC-006 `GET /auth/callback/google`

- 用例名称：Handle Google OAuth callback
- 说明：Handles Google OAuth callback, creates/logs in user, redirects to frontend with token
- 鉴权：不需要

**请求示例**

```http
GET /api/v1/auth/callback/google?code=sample-code
```

Query 参数：
- `code` (string): `sample-code`

**期望响应**

- HTTP `302` (Redirect to frontend with token)

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-007 `POST /auth/forgot-password`

- 用例名称：Send password reset email
- 说明：Sends a password reset email if the account exists
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/auth/forgot-password
Content-Type: application/json
```

Body 示例：

```json
{
  "email": "string"
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-008 `POST /auth/login`

- 用例名称：Login user
- 说明：Login with email and password to get JWT token
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/auth/login
Content-Type: application/json
```

Body 示例：

```json
{
  "email": "string",
  "password": "string"
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "access_token": "string",
  "refresh_token": "string",
  "token": "string",
  "type": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-009 `GET /auth/login/google`

- 用例名称：Redirect to Google OAuth
- 说明：Redirects user to Google OAuth authorization page
- 鉴权：不需要

**请求示例**

```http
GET /api/v1/auth/login/google
```

**期望响应**

- HTTP `302` (Redirect to Google OAuth)

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-010 `GET /auth/me`

- 用例名称：Get current user info
- 说明：Get the current authenticated user's information (protected route)
- 鉴权：需要 Bearer Token
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/auth/me
Authorization: Bearer <ACCESS_TOKEN>
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "email": "string",
  "message": "string",
  "role": "string",
  "user_id": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-011 `POST /auth/refresh`

- 用例名称：Refresh access token
- 说明：Rotates refresh token and returns a new access token pair
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/auth/refresh
Content-Type: application/json
```

Body 示例：

```json
{
  "refresh_token": "string"
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "access_token": "string",
  "refresh_token": "string",
  "type": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-012 `POST /auth/register`

- 用例名称：Register a new user
- 说明：Register a new user with username, email and password
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/auth/register
Content-Type: application/json
```

Body 示例：

```json
{
  "email": "string",
  "password": "string",
  "username": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "message": "string",
  "user_id": 1
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-013 `GET /auth/verify`

- 用例名称：Verify user email
- 说明：Verify user email using the token sent to their email
- 鉴权：不需要

**请求示例**

```http
GET /api/v1/auth/verify?token=sample-token
```

Query 参数：
- `token` (string): `sample-token`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: category

### TC-014 `GET /categories`

- 用例名称：List categories
- 说明：Returns a list of all categories
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/categories
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: content

### TC-015 `GET /user/favorites`

- 用例名称：List my favorites
- 说明：Returns favorites for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/favorites?type=sample&cursor=1&limit=1
```

Query 参数：
- `type` (string): `sample`
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "cursor": 1,
  "items": [
    {
      "created_at": "string",
      "id": 1,
      "merchant": {
        "category": "...",
        "id": "...",
        "name": "..."
      },
      "post": {
        "content": "...",
        "created_at": "...",
        "id": "...",
        "images": "...",
        "is_liked": "...",
        "like_count": "...",
        "merchant": "...",
        "tags": "...",
        "title": "...",
        "view_count": "..."
      },
      "review": {
        "avg_cost": "...",
        "content": "...",
        "created_at": "...",
        "id": "...",
        "images": "...",
        "is_liked": "...",
        "like_count": "...",
        "merchant": "...",
        "rating": "...",
        "rating_env": "...",
        "rating_service": "...",
        "rating_value": "...",
        "tags": "..."
      },
      "target_id": 1,
      "target_type": "string"
    }
  ],
  "total": 1
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-016 `GET /user/likes`

- 用例名称：List my likes
- 说明：Returns likes for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/likes?cursor=1&limit=1
```

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-017 `GET /user/posts`

- 用例名称：List my posts
- 说明：Returns posts created by the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/posts?cursor=1&limit=1
```

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "cursor": 1,
  "posts": [
    {
      "content": "string",
      "created_at": "string",
      "id": 1,
      "images": [
        "string"
      ],
      "is_liked": true,
      "like_count": 1,
      "merchant": {
        "category": "...",
        "id": "...",
        "name": "..."
      },
      "tags": [
        "string"
      ],
      "title": "string",
      "view_count": 1
    }
  ],
  "total": 1
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-018 `GET /user/reviews`

- 用例名称：List my reviews
- 说明：Returns reviews created by the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/reviews?cursor=1&limit=1
```

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "cursor": 1,
  "reviews": [
    {
      "avg_cost": 1,
      "content": "string",
      "created_at": "string",
      "id": 1,
      "images": [
        "string"
      ],
      "is_liked": true,
      "like_count": 1,
      "merchant": {
        "category": "...",
        "id": "...",
        "name": "..."
      },
      "rating": 1.1,
      "rating_env": 1.1,
      "rating_service": 1.1,
      "rating_value": 1.1,
      "tags": [
        "string"
      ]
    }
  ],
  "total": 1
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-019 `GET /users/{id}/posts`

- 用例名称：List user's posts
- 说明：Returns a user's posts
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/users/1/posts?cursor=1&limit=1
```

路径参数：
- `id` (integer): `1`

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "cursor": 1,
  "posts": [
    {
      "content": "string",
      "created_at": "string",
      "id": 1,
      "images": [
        "string"
      ],
      "is_liked": true,
      "like_count": 1,
      "merchant": {
        "category": "...",
        "id": "...",
        "name": "..."
      },
      "tags": [
        "string"
      ],
      "title": "string",
      "view_count": 1
    }
  ],
  "total": 1
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-020 `GET /users/{id}/reviews`

- 用例名称：List user's reviews
- 说明：Returns a user's reviews
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/users/1/reviews?cursor=1&limit=1
```

路径参数：
- `id` (integer): `1`

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "cursor": 1,
  "reviews": [
    {
      "avg_cost": 1,
      "content": "string",
      "created_at": "string",
      "id": 1,
      "images": [
        "string"
      ],
      "is_liked": true,
      "like_count": 1,
      "merchant": {
        "category": "...",
        "id": "...",
        "name": "..."
      },
      "rating": 1.1,
      "rating_env": 1.1,
      "rating_service": 1.1,
      "rating_value": 1.1,
      "tags": [
        "string"
      ]
    }
  ],
  "total": 1
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: conversation

### TC-021 `GET /conversations`

- 用例名称：List conversations
- 说明：Returns conversations for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/conversations
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-022 `POST /conversations`

- 用例名称：Create conversation
- 说明：Creates a new conversation
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/conversations
Content-Type: application/json
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-023 `GET /conversations/{id}/messages`

- 用例名称：Get conversation messages
- 说明：Returns messages for a conversation
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/conversations/1/messages
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-024 `POST /conversations/{id}/messages`

- 用例名称：Send message
- 说明：Sends a message in a conversation
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/conversations/1/messages
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-025 `PATCH /conversations/{id}/settings`

- 用例名称：Update conversation settings
- 说明：Updates settings for a conversation (e.g. mute)
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/conversations/1/settings
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: coupon

### TC-026 `POST /coupons/{id}/payment/initiate`

- 用例名称：Initiate coupon payment
- 说明：Initiates payment flow for a coupon
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/coupons/1/payment/initiate
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

Body 示例：

```json
{
  "userId": "string"
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-027 `POST /coupons/{id}/redeem`

- 用例名称：Redeem coupon
- 说明：Redeems a coupon for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/coupons/1/redeem
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-028 `POST /coupons/{id}/validate`

- 用例名称：Validate coupon
- 说明：Validates a coupon by ID
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/coupons/1/validate
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

Body 示例：

```json
{
  "quantity": 1
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-029 `POST /merchant/stores/{id}/coupons`

- 用例名称：Create store coupon
- 说明：Creates a store-scoped coupon for an owned published store
- 鉴权：需要 Bearer Token
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/merchant/stores/1/coupons
Authorization: Bearer <ACCESS_TOKEN>
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

Body 示例：

```json
{
  "description": "string",
  "max_per_user": 1,
  "price": 1.1,
  "status": "string",
  "terms": "string",
  "title": "string",
  "total_quantity": 1,
  "type": "string",
  "valid_from": "string",
  "valid_until": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-030 `GET /stores/{id}/coupons`

- 用例名称：List store coupons
- 说明：Lists published active coupons under a store
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/stores/1/coupons
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: feed

### TC-031 `GET /feed/home`

- 用例名称：Get home feed
- 说明：Returns the home feed items
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/feed/home
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: follow

### TC-032 `POST /merchants/{id}/follow`

- 用例名称：Follow merchant
- 说明：Follow a merchant
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/merchants/1/follow
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-033 `DELETE /merchants/{id}/follow`

- 用例名称：Unfollow merchant
- 说明：Unfollow a merchant
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
DELETE /api/v1/merchants/1/follow
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-034 `POST /users/{id}/follow`

- 用例名称：Follow user
- 说明：Follow a user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/users/1/follow
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-035 `DELETE /users/{id}/follow`

- 用例名称：Unfollow user
- 说明：Unfollow a user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
DELETE /api/v1/users/1/follow
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: media

### TC-036 `POST /media/presigned-urls`

- 用例名称：Create presigned URLs for media upload
- 说明：Generates presigned URLs for uploading files directly to R2 storage
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/media/presigned-urls
Content-Type: application/json
```

Body 示例：

```json
{
  "files": [
    {
      "content_type": "string",
      "filename": "string"
    }
  ]
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "uploads": [
    {
      "expires_at": "string",
      "file_url": "string",
      "id": "string",
      "upload_url": "string"
    }
  ]
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-037 `POST /media/uploads`

- 用例名称：Create media upload
- 说明：Creates a media upload and returns upload URLs
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/media/uploads
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-038 `POST /media/{id}/analysis`

- 用例名称：Analyze media upload
- 说明：Triggers analysis for a media upload
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/media/1/analysis
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: merchant

### TC-039 `GET /merchants`

- 用例名称：List merchants
- 说明：Returns merchants, optionally filtered by category
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/merchants?category=sample
```

Query 参数：
- `category` (string): `sample`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-040 `GET /merchants/{id}`

- 用例名称：Get merchant detail
- 说明：Returns a merchant by ID
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/merchants/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "businessName": "string",
  "category": "string",
  "coverImage": "string",
  "distance": "string",
  "id": "string",
  "name": "string",
  "rating": 1.1,
  "reviewCount": 1,
  "tags": [
    "string"
  ]
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-041 `GET /merchants/{id}/reviews`

- 用例名称：List merchant reviews
- 说明：Returns reviews for a merchant
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/merchants/1/reviews
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: notification

### TC-042 `GET /notifications`

- 用例名称：List notifications
- 说明：Returns notifications for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/notifications
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-043 `POST /notifications/read-all`

- 用例名称：Mark all notifications as read
- 说明：Marks all notifications as read for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/notifications/read-all
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-044 `PATCH /notifications/{id}/read`

- 用例名称：Mark notification as read
- 说明：Marks a single notification as read
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/notifications/1/read
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: order

### TC-045 `GET /orders`

- 用例名称：List orders
- 说明：Returns orders for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/orders
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-046 `POST /orders`

- 用例名称：Create order
- 说明：Creates a new order for the authenticated user
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/orders
Content-Type: application/json
```

Body 示例：

```json
{
  "coupon_id": 1,
  "quantity": 1
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-047 `GET /orders/{id}`

- 用例名称：Get order detail
- 说明：Returns an order by ID
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/orders/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-048 `POST /orders/{id}/pay`

- 用例名称：Pay order
- 说明：Simulates payment success for an order and issues vouchers
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/orders/1/pay
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: package

### TC-049 `GET /packages`

- 用例名称：List packages
- 说明：Returns a list of available packages
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/packages
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-050 `GET /packages/{id}`

- 用例名称：Get package detail
- 说明：Returns a package by ID
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/packages/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: payment

### TC-051 `POST /payments`

- 用例名称：Create payment
- 说明：Creates a payment record
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/payments
Content-Type: application/json
```

Body 示例：

```json
{
  "amount": 1.1,
  "currency": "string",
  "status": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-052 `GET /payments/{id}`

- 用例名称：Get payment detail
- 说明：Returns a payment by ID
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/payments/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: profile

### TC-053 `GET /users/{id}`

- 用例名称：Get public user profile
- 说明：Returns a user's public profile
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/users/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "avatar_url": "string",
  "follower_count": 1,
  "following_count": 1,
  "intro": "string",
  "is_following": true,
  "like_count": 1,
  "location": "string",
  "nickname": "string",
  "post_count": 1,
  "review_count": 1,
  "user_id": 1
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: review

### TC-054 `GET /reviews`

- 用例名称：List my reviews
- 说明：Returns reviews created by the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/reviews
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-055 `POST /reviews`

- 用例名称：Create a review
- 说明：Creates a new review for the authenticated user
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/reviews
Content-Type: application/json
```

Body 示例：

```json
{
  "businessImage": "string",
  "businessName": "string",
  "createdAt": "string",
  "id": "string",
  "images": [
    "string"
  ],
  "likeCount": 1,
  "location": "string",
  "merchantId": "string",
  "rating": 1.1,
  "storeId": "string",
  "tags": [
    "string"
  ],
  "text": "string",
  "userId": "string",
  "venueId": "string",
  "visitDate": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "businessImage": "string",
  "businessName": "string",
  "createdAt": "string",
  "id": "string",
  "images": [
    "string"
  ],
  "likeCount": 1,
  "location": "string",
  "merchantId": "string",
  "rating": 1.1,
  "storeId": "string",
  "tags": [
    "string"
  ],
  "text": "string",
  "userId": "string",
  "venueId": "string",
  "visitDate": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

- HTTP `422` (Unprocessable Entity)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-056 `GET /reviews/{id}`

- 用例名称：Get review detail
- 说明：Returns a single review by ID
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/reviews/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "businessImage": "string",
  "businessName": "string",
  "createdAt": "string",
  "id": "string",
  "images": [
    "string"
  ],
  "likeCount": 1,
  "location": "string",
  "merchantId": "string",
  "rating": 1.1,
  "storeId": "string",
  "tags": [
    "string"
  ],
  "text": "string",
  "userId": "string",
  "venueId": "string",
  "visitDate": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-057 `POST /reviews/{id}/comments`

- 用例名称：Add a review comment
- 说明：Adds a comment to a review
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/reviews/1/comments
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

Body 示例：

```json
{
  "text": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-058 `POST /reviews/{id}/like`

- 用例名称：Like a review
- 说明：Likes a review for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/reviews/1/like
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: store

### TC-059 `GET /merchant/stores`

- 用例名称：List current merchant stores
- 说明：Returns stores owned by the authenticated merchant user
- 鉴权：需要 Bearer Token
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/merchant/stores
Authorization: Bearer <ACCESS_TOKEN>
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-060 `POST /merchant/stores`

- 用例名称：Create a store
- 说明：Creates a new store for the authenticated merchant
- 鉴权：需要 Bearer Token
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/merchant/stores
Authorization: Bearer <ACCESS_TOKEN>
Content-Type: application/json
```

Body 示例：

```json
{
  "address": "string",
  "category_ids": [
    1
  ],
  "city": "string",
  "country": "string",
  "cover_image_url": "string",
  "description": "string",
  "hours": [
    {
      "close_time": "string",
      "day_of_week": 1,
      "is_closed": true,
      "open_time": "string"
    }
  ],
  "images": [
    "string"
  ],
  "latitude": 1.1,
  "longitude": 1.1,
  "name": "string",
  "phone": "string",
  "state": "string",
  "website": "string",
  "zip_code": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-061 `PATCH /merchant/stores/{id}`

- 用例名称：Update a store
- 说明：Updates a store for the authenticated merchant
- 鉴权：需要 Bearer Token
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/merchant/stores/1
Authorization: Bearer <ACCESS_TOKEN>
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

Body 示例：

```json
{
  "address": "string",
  "category_ids": [
    1
  ],
  "city": "string",
  "country": "string",
  "cover_image_url": "string",
  "description": "string",
  "hours": [
    {
      "close_time": "string",
      "day_of_week": 1,
      "is_closed": true,
      "open_time": "string"
    }
  ],
  "images": [
    "string"
  ],
  "latitude": 1.1,
  "longitude": 1.1,
  "name": "string",
  "phone": "string",
  "state": "string",
  "website": "string",
  "zip_code": "string"
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-062 `POST /merchant/stores/{id}/activate`

- 用例名称：Activate a store
- 说明：Marks a merchant-owned store as published
- 鉴权：需要 Bearer Token
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/merchant/stores/1/activate
Authorization: Bearer <ACCESS_TOKEN>
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-063 `POST /merchant/stores/{id}/deactivate`

- 用例名称：Deactivate a store
- 说明：Marks a merchant-owned store as hidden
- 鉴权：需要 Bearer Token
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/merchant/stores/1/deactivate
Authorization: Bearer <ACCESS_TOKEN>
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-064 `GET /stores`

- 用例名称：List stores
- 说明：Returns a list of stores
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/stores?category=sample&lat=1&lng=1&rating=1&radius_km=1&cursor=1&limit=1
```

Query 参数：
- `category` (string): `sample`
- `lat` (number): `1`
- `lng` (number): `1`
- `rating` (number): `1`
- `radius_km` (number): `1`
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-065 `GET /stores/{id}`

- 用例名称：Get store detail
- 说明：Returns a store by ID
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/stores/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-066 `GET /stores/{id}/hours`

- 用例名称：Get store hours
- 说明：Returns operating hours for a store
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/stores/1/hours
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-067 `GET /stores/{id}/reviews`

- 用例名称：Get store reviews
- 说明：Returns reviews for a store
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/stores/1/reviews?cursor=1&limit=1
```

路径参数：
- `id` (integer): `1`

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: user

### TC-068 `DELETE /user/account`

- 用例名称：Request account deletion
- 说明：Schedules account deletion (cooling period)
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
DELETE /api/v1/user/account
Content-Type: application/json
```

Body 示例：

```json
{
  "key": "string"
}
```

**期望响应**

- HTTP `202` (Accepted)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-069 `POST /user/account/export`

- 用例名称：Request account export
- 说明：Queues a user data export
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/user/account/export
```

**期望响应**

- HTTP `202` (Accepted)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-070 `GET /user/addresses`

- 用例名称：List addresses
- 说明：Returns the authenticated user's saved addresses
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/addresses
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "addresses": [
    {
      "address": "string",
      "city": "string",
      "district": "string",
      "id": 1,
      "is_default": true,
      "name": "string",
      "phone": "string",
      "postal_code": "string",
      "province": "string"
    }
  ]
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-071 `POST /user/addresses`

- 用例名称：Create address
- 说明：Adds a new address for the authenticated user
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/user/addresses
Content-Type: application/json
```

Body 示例：

```json
{
  "address": "string",
  "city": "string",
  "district": "string",
  "is_default": true,
  "name": "string",
  "phone": "string",
  "postal_code": "string",
  "province": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "address": "string",
  "city": "string",
  "district": "string",
  "id": 1,
  "is_default": true,
  "name": "string",
  "phone": "string",
  "postal_code": "string",
  "province": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-072 `PATCH /user/addresses/{id}`

- 用例名称：Update address
- 说明：Updates an existing address
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/user/addresses/1
Content-Type: application/json
```

路径参数：
- `id` (integer): `1`

Body 示例：

```json
{
  "address": "string",
  "city": "string",
  "district": "string",
  "is_default": true,
  "name": "string",
  "phone": "string",
  "postal_code": "string",
  "province": "string"
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-073 `DELETE /user/addresses/{id}`

- 用例名称：Delete address
- 说明：Deletes an address
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
DELETE /api/v1/user/addresses/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-074 `POST /user/addresses/{id}/default`

- 用例名称：Set default address
- 说明：Sets an address as default
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/user/addresses/1/default
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-075 `GET /user/followers`

- 用例名称：List followers
- 说明：Returns followers of the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/followers?cursor=1&limit=1
```

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-076 `GET /user/following/merchants`

- 用例名称：List following merchants
- 说明：Returns merchants the authenticated user follows
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/following/merchants?cursor=1&limit=1
```

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-077 `GET /user/following/users`

- 用例名称：List following users
- 说明：Returns users the authenticated user follows
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/following/users?cursor=1&limit=1
```

Query 参数：
- `cursor` (integer): `1`
- `limit` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-078 `GET /user/notifications`

- 用例名称：Get notification settings
- 说明：Returns the authenticated user's notification settings
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/notifications
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "email_enabled": true,
  "push_enabled": true
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-079 `PATCH /user/notifications`

- 用例名称：Update notification settings
- 说明：Updates the authenticated user's notification settings
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/user/notifications
Content-Type: application/json
```

Body 示例：

```json
{
  "email_enabled": true,
  "push_enabled": true
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-080 `GET /user/privacy`

- 用例名称：Get privacy settings
- 说明：Returns the authenticated user's privacy settings
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/privacy
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "is_public": true
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-081 `PATCH /user/privacy`

- 用例名称：Update privacy settings
- 说明：Updates the authenticated user's privacy settings
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/user/privacy
Content-Type: application/json
```

Body 示例：

```json
{
  "is_public": true
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-082 `GET /user/profile`

- 用例名称：Get current user profile
- 说明：Returns the authenticated user's profile
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/user/profile
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "avatar_url": "string",
  "intro": "string",
  "location": "string",
  "nickname": "string",
  "user_id": 1
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-083 `PATCH /user/profile`

- 用例名称：Update current user profile
- 说明：Updates nickname, avatar, intro, or location
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/user/profile
Content-Type: application/json
```

Body 示例：

```json
{
  "avatar_url": "string",
  "intro": "string",
  "location": "string",
  "nickname": "string"
}
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: verification

### TC-084 `GET /merchant/verification`

- 用例名称：Get verification status
- 说明：Returns the verification status for the authenticated merchant
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/merchant/verification
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-085 `POST /merchant/verification`

- 用例名称：Submit merchant verification
- 说明：Submits verification documents for the authenticated merchant
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/merchant/verification
Content-Type: application/json
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

## Tag: voucher

### TC-086 `POST /merchant/vouchers/{id}/redeem`

- 用例名称：Redeem voucher by merchant owner
- 说明：Redeems a voucher for merchant-owned store operations
- 鉴权：需要 Bearer Token
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/merchant/vouchers/1/redeem
Authorization: Bearer <ACCESS_TOKEN>
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `403` (Forbidden)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema
- 未携带/无效 token 时应返回 401 或 403

---

### TC-087 `GET /vouchers`

- 用例名称：List vouchers
- 说明：Returns vouchers for the authenticated user
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/vouchers
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `500` (Internal Server Error)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-088 `POST /vouchers`

- 用例名称：Create voucher
- 说明：Creates a voucher for the authenticated user
- 鉴权：不需要
- 请求 Content-Type：`application/json`
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/vouchers
Content-Type: application/json
```

Body 示例：

```json
{
  "code": "string",
  "couponId": "string",
  "userId": "string"
}
```

**期望响应**

- HTTP `201` (Created)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-089 `GET /vouchers/code/{code}`

- 用例名称：Get voucher by code
- 说明：Returns a voucher by code
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/vouchers/code/sample-code
```

路径参数：
- `code` (string): `sample-code`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-090 `POST /vouchers/share/email`

- 用例名称：Share voucher via email
- 说明：Sends voucher share email
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/vouchers/share/email
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-091 `POST /vouchers/share/sms`

- 用例名称：Share voucher via SMS
- 说明：Sends voucher share SMS
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
POST /api/v1/vouchers/share/sms
```

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-092 `GET /vouchers/{id}`

- 用例名称：Get voucher detail
- 说明：Returns a voucher by ID
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
GET /api/v1/vouchers/1
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "value"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

- HTTP `404` (Not Found)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-093 `PATCH /vouchers/{id}/status`

- 用例名称：Update voucher status
- 说明：Updates voucher status to used
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/vouchers/1/status
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

### TC-094 `PATCH /vouchers/{id}/use`

- 用例名称：Use voucher
- 说明：Marks a voucher as used
- 鉴权：不需要
- 响应 Content-Type：`application/json`

**请求示例**

```http
PATCH /api/v1/vouchers/1/use
```

路径参数：
- `id` (integer): `1`

**期望响应**

- HTTP `200` (OK)

```json
{
  "key": "string"
}
```

- HTTP `400` (Bad Request)

```json
{
  "key": "string"
}
```

- HTTP `401` (Unauthorized)

```json
{
  "key": "string"
}
```

断言重点：
- 状态码与 swagger 定义一致
- 返回 JSON 字段类型符合 schema

---

