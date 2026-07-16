-- +goose Up
CREATE TABLE password_credentials (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT password_credentials_hash_not_empty CHECK (length(btrim(password_hash)) > 0)
);

CREATE TRIGGER set_password_credentials_updated_at
    BEFORE UPDATE ON password_credentials
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

INSERT INTO password_credentials (user_id, password_hash, created_at, updated_at)
SELECT id, password_hash, created_at, updated_at
FROM users;

CREATE TABLE user_identities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider text NOT NULL,
    issuer text NOT NULL,
    subject text NOT NULL,
    email citext,
    email_verified boolean NOT NULL DEFAULT false,
    display_name text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz,
    CONSTRAINT user_identities_provider_not_empty CHECK (length(btrim(provider)) > 0),
    CONSTRAINT user_identities_issuer_not_empty CHECK (length(btrim(issuer)) > 0),
    CONSTRAINT user_identities_subject_not_empty CHECK (length(btrim(subject)) > 0),
    CONSTRAINT user_identities_email_not_empty CHECK (email IS NULL OR length(btrim(email::text)) > 0),
    CONSTRAINT user_identities_display_name_not_empty CHECK (display_name IS NULL OR length(btrim(display_name)) > 0),
    CONSTRAINT user_identities_last_login_after_created CHECK (last_login_at IS NULL OR last_login_at >= created_at),
    CONSTRAINT uq_user_identities_issuer_subject UNIQUE (issuer, subject),
    CONSTRAINT uq_user_identities_user_issuer UNIQUE (user_id, issuer)
);

CREATE INDEX ix_user_identities_user_id ON user_identities (user_id);

CREATE TRIGGER set_user_identities_updated_at
    BEFORE UPDATE ON user_identities
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

ALTER TABLE auth_sessions
    ADD COLUMN authenticated_at timestamptz,
    ADD COLUMN authentication_method text,
    ADD COLUMN identity_id uuid REFERENCES user_identities(id) ON DELETE SET NULL;

UPDATE auth_sessions
SET authenticated_at = created_at,
    authentication_method = 'password';

ALTER TABLE auth_sessions
    ALTER COLUMN authenticated_at SET NOT NULL,
    ALTER COLUMN authentication_method SET NOT NULL,
    ADD CONSTRAINT auth_sessions_authentication_method_valid CHECK (authentication_method IN ('password', 'oidc')),
    ADD CONSTRAINT auth_sessions_authenticated_after_created CHECK (authenticated_at >= created_at);

CREATE INDEX ix_auth_sessions_identity_id ON auth_sessions (identity_id) WHERE identity_id IS NOT NULL;

CREATE TABLE oidc_auth_flows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    state_hash bytea NOT NULL,
    browser_token_hash bytea NOT NULL,
    nonce text NOT NULL,
    pkce_verifier text NOT NULL,
    intent text NOT NULL,
    session_id uuid REFERENCES auth_sessions(id) ON DELETE CASCADE,
    return_to text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    CONSTRAINT uq_oidc_auth_flows_state_hash UNIQUE (state_hash),
    CONSTRAINT oidc_auth_flows_state_hash_not_empty CHECK (length(state_hash) > 0),
    CONSTRAINT oidc_auth_flows_browser_token_hash_not_empty CHECK (length(browser_token_hash) > 0),
    CONSTRAINT oidc_auth_flows_nonce_not_empty CHECK (length(nonce) > 0),
    CONSTRAINT oidc_auth_flows_pkce_not_empty CHECK (length(pkce_verifier) > 0),
    CONSTRAINT oidc_auth_flows_intent_valid CHECK (intent IN ('login', 'sudo', 'link')),
    CONSTRAINT oidc_auth_flows_session_required CHECK ((intent = 'login' AND session_id IS NULL) OR (intent IN ('sudo', 'link') AND session_id IS NOT NULL)),
    CONSTRAINT oidc_auth_flows_return_to_relative CHECK (
        return_to LIKE '/%'
        AND return_to NOT LIKE '//%'
        AND position(chr(92) IN return_to) = 0
        AND position(chr(10) IN return_to) = 0
        AND position(chr(13) IN return_to) = 0
    ),
    CONSTRAINT oidc_auth_flows_expires_after_created CHECK (expires_at > created_at),
    CONSTRAINT oidc_auth_flows_used_after_created CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX ix_oidc_auth_flows_expires_at ON oidc_auth_flows (expires_at);

ALTER TABLE users DROP CONSTRAINT users_password_hash_not_empty;
ALTER TABLE users DROP COLUMN password_hash;

-- +goose Down
ALTER TABLE users ADD COLUMN password_hash text;

UPDATE users
SET password_hash = password_credentials.password_hash
FROM password_credentials
WHERE password_credentials.user_id = users.id;

-- Refuse to discard passwordless accounts during rollback.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE password_hash IS NULL) THEN
        RAISE EXCEPTION 'cannot roll back while passwordless users exist';
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE users
    ALTER COLUMN password_hash SET NOT NULL,
    ADD CONSTRAINT users_password_hash_not_empty CHECK (length(btrim(password_hash)) > 0);

DROP TABLE oidc_auth_flows;
DROP INDEX ix_auth_sessions_identity_id;
ALTER TABLE auth_sessions
    DROP CONSTRAINT auth_sessions_authenticated_after_created,
    DROP CONSTRAINT auth_sessions_authentication_method_valid,
    DROP COLUMN identity_id,
    DROP COLUMN authentication_method,
    DROP COLUMN authenticated_at;
DROP TABLE user_identities;
DROP TABLE password_credentials;
