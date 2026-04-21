-- +goose Up

ALTER TABLE vouchers ADD COLUMN IF NOT EXISTS scan_token VARCHAR(128);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vouchers_scan_token ON vouchers (scan_token);

-- +goose Down

DROP INDEX IF EXISTS idx_vouchers_scan_token;
ALTER TABLE vouchers DROP COLUMN IF EXISTS scan_token;
