-- +goose Up
CREATE TABLE auth_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL,
    csrf_token_hash bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz NOT NULL DEFAULT now(),
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz NULL,
    revoked_reason text NULL,
    CONSTRAINT auth_sessions_token_hash_not_empty CHECK (length(token_hash) > 0),
    CONSTRAINT auth_sessions_csrf_token_hash_not_empty CHECK (length(csrf_token_hash) > 0),
    CONSTRAINT auth_sessions_idle_expires_after_created CHECK (idle_expires_at > created_at),
    CONSTRAINT auth_sessions_absolute_expires_after_created CHECK (absolute_expires_at > created_at),
    CONSTRAINT auth_sessions_revoked_reason_present CHECK (revoked_at IS NULL OR length(btrim(coalesce(revoked_reason, ''))) > 0)
);

CREATE UNIQUE INDEX uq_auth_sessions_token_hash ON auth_sessions (token_hash);
CREATE INDEX ix_auth_sessions_user_id ON auth_sessions (user_id);
CREATE INDEX ix_auth_sessions_active_expiry
    ON auth_sessions (idle_expires_at, absolute_expires_at)
    WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;
