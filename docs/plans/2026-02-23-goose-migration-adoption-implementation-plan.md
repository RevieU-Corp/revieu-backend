# Goose Migration Adoption Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace startup-time AutoMigrate as the production schema mechanism with explicit Goose SQL migrations and operational commands.

**Architecture:** Keep application DB connection unchanged, but gate AutoMigrate behind a config flag that defaults to disabled. Introduce a baseline Goose migration SQL file as the canonical schema for new environments, and add Makefile commands for migration lifecycle (`up/down/status/create`).

**Tech Stack:** Go, GORM, PostgreSQL, Goose CLI, Makefile

---

### Task 1: Add failing tests for AutoMigrate gating and config parsing

**Files:**
- Create: `apps/core/cmd/app/migrate_test.go`
- Modify: `apps/core/internal/config/config_test.go`

**Step 1: Write failing tests**
- Add tests asserting:
  - migration function does nothing when disabled.
  - migration function creates schema when enabled.
  - `database.auto_migrate` defaults to false when absent.
  - `database.auto_migrate` parses true when provided.

**Step 2: Run tests to verify failure**
Run:
- `cd apps/core && GOCACHE=/tmp/go-build go test ./cmd/app ./internal/config`

Expected:
- build/test failure because migration helper and config field do not exist yet.

### Task 2: Implement AutoMigrate gating in app startup

**Files:**
- Create: `apps/core/cmd/app/migrate.go`
- Modify: `apps/core/cmd/app/main.go`
- Modify: `apps/core/internal/config/config.go`
- Modify: `apps/core/configs/config.yaml`

**Step 1: Implement minimal code**
- Add `DatabaseConfig.AutoMigrate bool`.
- Extract migration model list into helper function.
- Replace direct `AutoMigrate(...)` in `main.go` with helper call based on config flag.
- Set `database.auto_migrate: false` in default config.

**Step 2: Run tests to verify pass**
Run:
- `cd apps/core && GOCACHE=/tmp/go-build go test ./cmd/app ./internal/config`

Expected:
- all tests pass.

### Task 3: Add Goose migration directory and baseline SQL migration

**Files:**
- Create: `apps/core/migrations/00001_init_schema.sql`

**Step 1: Write baseline migration**
- Add Goose `Up` section creating all required tables/indexes (without destructive `DROP TABLE`).
- Add Goose `Down` section dropping created objects in reverse dependency order.
- Ensure schema includes current app-required tables such as `refresh_tokens` and model-compatible `user_addresses` columns.

**Step 2: Validate SQL structure quickly**
Run:
- `rg -n "\+goose Up|\+goose Down|CREATE TABLE refresh_tokens|CREATE TABLE user_addresses" apps/core/migrations/00001_init_schema.sql`

Expected:
- required markers and key tables present.

### Task 4: Add migration operational commands and docs

**Files:**
- Modify: `apps/core/Makefile`
- Modify: `apps/core/README.md`
- Modify: `README.md`

**Step 1: Implement commands**
- Add `migrate-up`, `migrate-down`, `migrate-status`, `migrate-create` make targets using Goose CLI.
- Add DSN and migrations-dir variables.

**Step 2: Document workflow**
- Document local migration flow and deployment recommendation: run migrations before app rollout.

**Step 3: Verify docs and Makefile references**
Run:
- `cd apps/core && make help | rg -n "migrate"`
- `rg -n "migrate-up|goose|AutoMigrate|database.auto_migrate" README.md apps/core/README.md apps/core/Makefile`

Expected:
- migration commands visible and documented.

### Task 5: Full verification before completion

**Files:**
- N/A

**Step 1: Run focused tests**
Run:
- `cd apps/core && GOCACHE=/tmp/go-build go test ./cmd/app ./internal/config ./internal/testutil`

**Step 2: Run repository status check**
Run:
- `git status --short`

Expected:
- only intended files changed and tests passing.
