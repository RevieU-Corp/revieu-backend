# OpenAPI V1 Missing Endpoints Design

**Context**
OpenAPI v1 defines several endpoints not yet implemented in this repo: feed, merchants, reviews, coupons, vouchers, payments, media, and AI suggestions. Existing auth/user/profile behavior must remain unchanged; only missing routes should be added.

## Architecture
- Add new domain packages per OpenAPI tags: `feed`, `merchant`, `review`, `coupon`, `voucher`, `payment`, `media`, `ai`.
- Each domain follows existing layering: `routes.go` → `handler/*.go` → `service/*.go`, using GORM for persistence.
- Reuse existing models where possible (`model.Merchant`, `model.Review`, `model.Like`, `model.Tag`), and add new models for coupons, vouchers, payments, media uploads, and review comments.
- Keep existing auth/user/profile routes untouched; only register new routes in `router.Setup`.

## Data Flow
- Public reads: `/feed/home`, `/merchants`, `/merchants/{id}`, `/merchants/{id}/reviews`.
- Auth-required writes: `/reviews` POST, `/reviews/{id}/like`, `/reviews/{id}/comments`, voucher status/use/share, coupon redeem/payment initiation, payment create, media upload/analysis, AI suggestions.
- Handlers validate input, map DTO ↔ model, call service; services use GORM with `database.DB` and explicit transactions for state changes.
- Feed is assembled from latest reviews + featured merchants; no dedicated feed table.

## Error Handling
- 400 for invalid input or bad IDs, 401 for unauthenticated requests, 404 for missing records, 500 for unexpected failures.
- Action endpoints return 200 with empty body (or minimal status JSON if needed).
- Review likes and comments validate target existence; voucher status transitions validate current state.

## Testing
- Add HTTP-level tests for new routes using Gin + `testutil.SetupTestDB`.
- Seed minimal data per test, verify status codes and key response fields.
- Keep tests deterministic by using placeholder outputs for AI/media/payment services.

## Schema/Migrations
- Update models and `internal/testutil/db.go` AutoMigrate list.
- Extend `init-db.sql` with new tables and indexes for coupons, vouchers, payments, media uploads, and review comments.

