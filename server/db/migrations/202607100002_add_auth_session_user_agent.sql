-- +goose Up
ALTER TABLE auth_sessions
    ADD COLUMN user_agent text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE auth_sessions
    DROP COLUMN IF EXISTS user_agent;
