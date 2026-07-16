-- +goose Up
CREATE TABLE api_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    token_hash bytea NOT NULL,
    token_hint text NOT NULL,
    scopes text[] NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revoked_reason text,
    CONSTRAINT api_tokens_name_valid CHECK (length(btrim(name)) BETWEEN 1 AND 100),
    CONSTRAINT api_tokens_token_hash_not_empty CHECK (length(token_hash) > 0),
    CONSTRAINT api_tokens_token_hint_valid CHECK (length(token_hint) = 8),
    CONSTRAINT api_tokens_scopes_not_empty CHECK (cardinality(scopes) > 0),
    CONSTRAINT api_tokens_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT api_tokens_last_used_after_created CHECK (last_used_at IS NULL OR last_used_at >= created_at),
    CONSTRAINT api_tokens_revoked_reason_present CHECK (revoked_at IS NULL OR length(btrim(coalesce(revoked_reason, ''))) > 0)
);

CREATE UNIQUE INDEX uq_api_tokens_token_hash ON api_tokens (token_hash);
CREATE INDEX ix_api_tokens_user_id ON api_tokens (user_id);
CREATE INDEX ix_api_tokens_active_expiry ON api_tokens (expires_at) WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS api_tokens;
