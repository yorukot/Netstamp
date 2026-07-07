-- +goose Up
CREATE TYPE system_role AS ENUM ('admin');

CREATE TABLE system_user_roles (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    role system_role NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_system_user_roles_role ON system_user_roles (role);

INSERT INTO system_user_roles (user_id, role)
SELECT id, 'admin'::system_role
FROM users
ORDER BY created_at ASC, id ASC
LIMIT 1;

CREATE TABLE system_settings (
    key text PRIMARY KEY,
    value jsonb,
    encrypted_value bytea,
    encrypted_value_nonce bytea,
    secret boolean NOT NULL DEFAULT false,
    updated_by_user_id uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT system_settings_key_not_empty CHECK (length(btrim(key)) > 0),
    CONSTRAINT system_settings_public_or_secret_value CHECK (
        (secret = false AND value IS NOT NULL AND encrypted_value IS NULL AND encrypted_value_nonce IS NULL)
        OR
        (secret = true AND value IS NULL AND encrypted_value IS NOT NULL AND encrypted_value_nonce IS NOT NULL)
    )
);

CREATE TRIGGER set_system_settings_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE system_setting_audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    key text NOT NULL,
    action text NOT NULL,
    updated_by_user_id uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT system_setting_audit_key_not_empty CHECK (length(btrim(key)) > 0),
    CONSTRAINT system_setting_audit_action_not_empty CHECK (length(btrim(action)) > 0)
);

CREATE INDEX idx_system_setting_audit_events_created
    ON system_setting_audit_events (created_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS system_setting_audit_events;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS system_user_roles;
DROP TYPE IF EXISTS system_role;
