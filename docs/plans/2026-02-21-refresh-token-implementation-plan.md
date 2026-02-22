# Refresh Token Rotation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add secure refresh token rotation with `[POST] /auth/refresh` so expired access tokens can be renewed without forcing re-login.

**Architecture:** Keep access token as JWT, add database-backed refresh token persistence with hash-only storage and one-time rotation semantics. Extend auth service/handler/routes to issue both tokens on login and rotate refresh tokens on refresh. Keep errors generic and return `401` for invalid/expired/revoked refresh tokens.

**Tech Stack:** Go, Gin, GORM, SQLite (tests), PostgreSQL (runtime), golang-jwt.

---

### Task 1: Config and Token Service Primitives

**Files:**
- Modify: `apps/core/internal/config/config.go`
- Modify: `apps/core/internal/config/config_test.go`
- Modify: `apps/core/configs/config.yaml`
- Modify: `apps/core/internal/token/token.go`
- Test: `apps/core/internal/config/config_test.go`

**Step 1: Write the failing test**

```go
func TestLoad_RefreshTTL(t *testing.T) {
    // yaml contains jwt.refresh_expire_hour: 168
    // assert cfg.JWT.RefreshExpireHour == 168
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config -run TestLoad_RefreshTTL -v`
Expected: FAIL because field does not exist yet.

**Step 3: Write minimal implementation**

```go
type JWTConfig struct {
    Secret            string `yaml:"secret"`
    ExpireHour        int    `yaml:"expire_hour"`
    RefreshExpireHour int    `yaml:"refresh_expire_hour"`
}
```

Also add refresh token helpers in token service:

```go
func (s *Service) GenerateRefreshToken() (plain string, hash string, err error)
func HashToken(token string) string
```

**Step 4: Run tests**

Run: `go test ./internal/config ./internal/token -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/config/config.go apps/core/internal/config/config_test.go apps/core/configs/config.yaml apps/core/internal/token/token.go
git commit -m "feat(auth): add refresh token config and token helpers"
```

### Task 2: Persistence Model for Refresh Tokens

**Files:**
- Create: `apps/core/internal/model/refresh_token.go`
- Modify: `apps/core/internal/testutil/db.go`
- Modify: `apps/core/cmd/app/main.go`
- Test: `apps/core/internal/domain/auth/service_test.go`

**Step 1: Write failing test**

```go
func TestLoginCreatesRefreshTokenRecord(t *testing.T) {
    // login verified user
    // assert refresh token table has one active record
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/auth -run TestLoginCreatesRefreshTokenRecord -v`
Expected: FAIL due missing model/logic.

**Step 3: Write minimal implementation**

```go
type RefreshToken struct {
    ID        int64
    UserID    int64
    TokenHash string
    ExpiresAt time.Time
    RevokedAt *time.Time
    LastUsedAt *time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Add AutoMigrate registrations in app startup and test DB setup.

**Step 4: Run tests**

Run: `go test ./internal/testutil ./internal/domain/auth -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/model/refresh_token.go apps/core/internal/testutil/db.go apps/core/cmd/app/main.go apps/core/internal/domain/auth/service_test.go
git commit -m "feat(auth): add refresh token persistence model"
```

### Task 3: Auth Service Rotation Logic

**Files:**
- Modify: `apps/core/internal/domain/auth/service.go`
- Modify: `apps/core/internal/domain/auth/service_test.go`
- Modify: `apps/core/internal/domain/auth/dto.go`

**Step 1: Write failing tests**

```go
func TestLoginReturnsAccessAndRefreshTokens(t *testing.T) {}
func TestRefreshRotatesRefreshToken(t *testing.T) {}
func TestRefreshRejectsExpiredOrRevokedToken(t *testing.T) {}
```

**Step 2: Run tests to verify failures**

Run: `go test ./internal/domain/auth -run "TestLoginReturnsAccessAndRefreshTokens|TestRefreshRotatesRefreshToken|TestRefreshRejectsExpiredOrRevokedToken" -v`
Expected: FAIL due missing methods/fields.

**Step 3: Write minimal implementation**

- Extend service interface and implementation:

```go
Login(...) (LoginTokens, error)
RefreshAccessToken(ctx context.Context, refreshToken string) (LoginTokens, error)
```

- Add transaction for rotation:

```go
// validate active token by hash
// set old revoked_at + last_used_at
// insert new refresh token
// issue new access token
```

**Step 4: Run tests**

Run: `go test ./internal/domain/auth -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/auth/service.go apps/core/internal/domain/auth/service_test.go apps/core/internal/domain/auth/dto.go
git commit -m "feat(auth): implement refresh token rotation service"
```

### Task 4: API Endpoint and Routing

**Files:**
- Modify: `apps/core/internal/domain/auth/handler.go`
- Modify: `apps/core/internal/domain/auth/dto.go`
- Modify: `apps/core/internal/domain/auth/routes.go`
- Test: `apps/core/internal/router/router_test.go`

**Step 1: Write failing tests**

```go
func TestRefreshEndpoint_Success(t *testing.T) {}
func TestRefreshEndpoint_InvalidTokenReturns401(t *testing.T) {}
```

**Step 2: Run tests**

Run: `go test ./internal/router -run Refresh -v`
Expected: FAIL because route/handler absent.

**Step 3: Write minimal implementation**

- Add request/response DTOs for refresh.
- Add handler method:

```go
func (h *Handler) Refresh(c *gin.Context) {
    // bind request
    // call svc.RefreshAccessToken
    // return 200 with access+refresh token pair
}
```

- Register route: `auth.POST("/refresh", handler.Refresh)`.

**Step 4: Run tests**

Run: `go test ./internal/router ./internal/domain/auth -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add apps/core/internal/domain/auth/handler.go apps/core/internal/domain/auth/dto.go apps/core/internal/domain/auth/routes.go apps/core/internal/router/router_test.go
git commit -m "feat(auth): add refresh token API endpoint"
```

### Task 5: Docs and Final Verification

**Files:**
- Modify: `apps/core/docs/swagger.yaml`
- Modify: `apps/core/docs/swagger.json`
- Modify: `apps/core/docs/openapi.yaml`

**Step 1: Write/adjust failing docs test (if present)**

If there is an API schema test, update expected fields/endpoint first.

**Step 2: Regenerate/update API docs**

Run project doc generation command, or edit docs files to include `/auth/refresh`.

**Step 3: Run full verification**

Run: `go test ./...`
Expected: PASS.

**Step 4: Commit**

```bash
git add apps/core/docs/swagger.yaml apps/core/docs/swagger.json apps/core/docs/openapi.yaml
git commit -m "docs(api): document auth refresh endpoint"
```
