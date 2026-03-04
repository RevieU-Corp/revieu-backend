-- +goose Up

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_merchants_deleted_at ON merchants (deleted_at);

ALTER TABLE stores ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_stores_deleted_at ON stores (deleted_at);

ALTER TABLE coupons ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_coupons_deleted_at ON coupons (deleted_at);

-- +goose Down

DROP INDEX IF EXISTS idx_coupons_deleted_at;
ALTER TABLE coupons DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_stores_deleted_at;
ALTER TABLE stores DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_merchants_deleted_at;
ALTER TABLE merchants DROP COLUMN IF EXISTS deleted_at;
