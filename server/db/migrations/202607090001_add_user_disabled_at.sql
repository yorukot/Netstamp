-- +goose Up
ALTER TABLE users
    ADD COLUMN disabled_at timestamptz,
    ADD CONSTRAINT users_disabled_at_after_created_at CHECK (disabled_at IS NULL OR disabled_at >= created_at);

CREATE INDEX idx_users_disabled_at
    ON users (disabled_at)
    WHERE disabled_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_disabled_at;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_disabled_at_after_created_at,
    DROP COLUMN IF EXISTS disabled_at;
