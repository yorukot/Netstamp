-- +goose Up
ALTER TABLE users
    ADD COLUMN email_verified_at timestamptz;

UPDATE users
SET email_verified_at = created_at
WHERE email_verified_at IS NULL;

CREATE TABLE email_verification_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT email_verification_tokens_token_hash_not_empty CHECK (length(btrim(token_hash)) > 0),
    CONSTRAINT email_verification_tokens_expires_after_created CHECK (expires_at > created_at)
);

CREATE UNIQUE INDEX uq_email_verification_tokens_token_hash ON email_verification_tokens (token_hash);
CREATE INDEX ix_email_verification_tokens_user_active ON email_verification_tokens (user_id, expires_at)
    WHERE used_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS email_verification_tokens;

ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified_at;
