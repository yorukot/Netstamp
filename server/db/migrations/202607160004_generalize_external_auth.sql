-- +goose Up
ALTER TABLE user_identities
    DROP CONSTRAINT uq_user_identities_issuer_subject,
    DROP CONSTRAINT uq_user_identities_user_issuer,
    ADD COLUMN username text,
    ADD COLUMN avatar_url text,
    ADD CONSTRAINT user_identities_username_not_empty CHECK (username IS NULL OR length(btrim(username)) > 0),
    ADD CONSTRAINT user_identities_avatar_url_not_empty CHECK (avatar_url IS NULL OR length(btrim(avatar_url)) > 0),
    ADD CONSTRAINT uq_user_identities_provider_issuer_subject UNIQUE (provider, issuer, subject),
    ADD CONSTRAINT uq_user_identities_user_provider_issuer UNIQUE (user_id, provider, issuer);

ALTER TABLE auth_sessions
    DROP CONSTRAINT auth_sessions_authentication_method_valid,
    ADD COLUMN sudo_eligible boolean NOT NULL DEFAULT false,
    ADD CONSTRAINT auth_sessions_authentication_method_valid CHECK (authentication_method IN ('password', 'google', 'github', 'oidc'));

UPDATE auth_sessions
SET sudo_eligible = (authentication_method = 'password');

ALTER TABLE oidc_auth_flows RENAME TO external_auth_flows;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT uq_oidc_auth_flows_state_hash TO uq_external_auth_flows_state_hash;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_state_hash_not_empty TO external_auth_flows_state_hash_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_browser_token_hash_not_empty TO external_auth_flows_browser_token_hash_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_nonce_not_empty TO external_auth_flows_nonce_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_pkce_not_empty TO external_auth_flows_pkce_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_intent_valid TO external_auth_flows_intent_valid;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_session_required TO external_auth_flows_session_required;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_return_to_relative TO external_auth_flows_return_to_relative;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_expires_after_created TO external_auth_flows_expires_after_created;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT oidc_auth_flows_used_after_created TO external_auth_flows_used_after_created;
ALTER INDEX ix_oidc_auth_flows_expires_at RENAME TO ix_external_auth_flows_expires_at;

ALTER TABLE external_auth_flows
    ADD COLUMN provider text NOT NULL DEFAULT 'oidc';

ALTER TABLE external_auth_flows
    ALTER COLUMN provider DROP DEFAULT,
    ADD CONSTRAINT external_auth_flows_provider_valid CHECK (provider IN ('google', 'github', 'oidc'));

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM auth_sessions WHERE authentication_method NOT IN ('password', 'oidc'))
       OR EXISTS (SELECT 1 FROM user_identities WHERE provider <> 'oidc')
       OR EXISTS (SELECT 1 FROM external_auth_flows WHERE provider <> 'oidc') THEN
        RAISE EXCEPTION 'cannot roll back while Google or GitHub authentication data exists';
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE external_auth_flows DROP CONSTRAINT external_auth_flows_provider_valid;
ALTER TABLE external_auth_flows DROP COLUMN provider;
ALTER INDEX ix_external_auth_flows_expires_at RENAME TO ix_oidc_auth_flows_expires_at;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT uq_external_auth_flows_state_hash TO uq_oidc_auth_flows_state_hash;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_state_hash_not_empty TO oidc_auth_flows_state_hash_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_browser_token_hash_not_empty TO oidc_auth_flows_browser_token_hash_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_nonce_not_empty TO oidc_auth_flows_nonce_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_pkce_not_empty TO oidc_auth_flows_pkce_not_empty;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_intent_valid TO oidc_auth_flows_intent_valid;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_session_required TO oidc_auth_flows_session_required;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_return_to_relative TO oidc_auth_flows_return_to_relative;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_expires_after_created TO oidc_auth_flows_expires_after_created;
ALTER TABLE external_auth_flows
    RENAME CONSTRAINT external_auth_flows_used_after_created TO oidc_auth_flows_used_after_created;
ALTER TABLE external_auth_flows RENAME TO oidc_auth_flows;

ALTER TABLE auth_sessions
    DROP CONSTRAINT auth_sessions_authentication_method_valid,
    DROP COLUMN sudo_eligible,
    ADD CONSTRAINT auth_sessions_authentication_method_valid CHECK (authentication_method IN ('password', 'oidc'));

ALTER TABLE user_identities
    DROP CONSTRAINT uq_user_identities_provider_issuer_subject,
    DROP CONSTRAINT uq_user_identities_user_provider_issuer,
    DROP CONSTRAINT user_identities_username_not_empty,
    DROP CONSTRAINT user_identities_avatar_url_not_empty,
    DROP COLUMN username,
    DROP COLUMN avatar_url,
    ADD CONSTRAINT uq_user_identities_issuer_subject UNIQUE (issuer, subject),
    ADD CONSTRAINT uq_user_identities_user_issuer UNIQUE (user_id, issuer);
