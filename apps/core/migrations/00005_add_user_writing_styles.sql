-- +goose Up

CREATE TABLE IF NOT EXISTS user_writing_styles (
    user_id                       BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    features                      JSONB NOT NULL DEFAULT '{}'::jsonb,
    samples                       JSONB NOT NULL DEFAULT '[]'::jsonb,
    reviews_since_last_derive     INTEGER NOT NULL DEFAULT 0,
    derived_from_review_count     INTEGER NOT NULL DEFAULT 0,
    last_derived_at               TIMESTAMPTZ,
    last_derive_started_at        TIMESTAMPTZ,
    updated_at                    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS user_writing_styles;
