# Refresh Token Rotation Design

Date: 2026-02-21

## Context
Current auth uses a single JWT access token and returns 401 when expired. There is no refresh token mechanism. The issue requests silent refresh via a refresh token and a new endpoint `[POST] /auth/refresh`.

## Goals
- Add refresh token support with server-side persistence.
- Enable silent access token renewal when access token expires.
- Improve security via refresh token rotation.

## Non-Goals
- Frontend `apiClient` changes (only backend design here).
- OAuth provider refresh handling (this is backend refresh for our own JWTs).

## Proposed Approach (Recommended)
**Refresh token rotation with DB persistence.**

### Components
- `refresh_tokens` table to store hashed refresh tokens.
- Auth service methods:
  - `IssueTokens` for login/register to generate access + refresh.
  - `RefreshAccessToken` to validate and rotate refresh tokens.
- New route `POST /auth/refresh`.

### Data Model (refresh_tokens)
Fields (minimum):
- `id` (PK)
- `user_id`
- `token_hash`
- `expires_at`
- `revoked_at` (nullable)
- `last_used_at` (nullable)
- `created_at` / `updated_at`

### Token Strategy
- Access token: JWT with short TTL.
- Refresh token: high-entropy random string.
- Store only `token_hash` (e.g., SHA-256) in DB.

## Data Flow
### Login/Register
1. Generate access token.
2. Generate refresh token.
3. Store refresh token hash + metadata.
4. Return both tokens to client.

### Refresh (POST /auth/refresh)
1. Client posts refresh token.
2. Server hashes token and finds matching record.
3. Validate not expired or revoked.
4. Rotate: revoke old token and issue a new refresh token record.
5. Return new access token + new refresh token.

## Error Handling & Security
- Invalid/expired/revoked refresh token => `401` with a generic error.
- Refresh token never stored in plaintext.
- Require HTTPS at deployment level.
- Default: allow multiple active refresh tokens per user.
  - Optional future toggle: single-device by revoking all other tokens on login.

## Testing
- Unit tests:
  - Login issues refresh token record.
  - Refresh rotates token (old revoked, new valid).
  - Expired or revoked token returns 401.
- Optional integration:
  - Login -> expire access token -> refresh -> access protected endpoint.

## Acceptance Criteria Mapping
- Valid refresh token silently refreshes access token.
- Refresh endpoint validates token server-side.
- Access token rotation occurs without user disruption.
