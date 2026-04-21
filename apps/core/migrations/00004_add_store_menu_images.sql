-- +goose Up

ALTER TABLE stores
ADD COLUMN IF NOT EXISTS menu_images JSONB NOT NULL DEFAULT '[]'::jsonb;

-- +goose Down

ALTER TABLE stores
DROP COLUMN IF EXISTS menu_images;
