-- +goose Up
CREATE TABLE password_reset_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT password_reset_tokens_token_hash_not_empty CHECK (length(btrim(token_hash)) > 0),
    CONSTRAINT password_reset_tokens_expires_after_created CHECK (expires_at > created_at)
);

CREATE UNIQUE INDEX uq_password_reset_tokens_token_hash ON password_reset_tokens (token_hash);
CREATE INDEX ix_password_reset_tokens_user_active ON password_reset_tokens (user_id, expires_at)
    WHERE used_at IS NULL;

-- +goose Down
DROP TABLE password_reset_tokens;
