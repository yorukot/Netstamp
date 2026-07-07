-- name: IsSystemAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM system_user_roles
    WHERE user_id = $1
      AND role = 'admin'
) AS is_system_admin;

-- name: GrantFirstSystemAdminIfNone :one
WITH lock AS (
    SELECT pg_advisory_xact_lock(hashtextextended('netstamp.system_admin.bootstrap', 0))
),
inserted AS (
    INSERT INTO system_user_roles (user_id, role)
    SELECT sqlc.arg(user_id), 'admin'
    FROM lock
    WHERE NOT EXISTS (
        SELECT 1
        FROM system_user_roles
        WHERE role = 'admin'
    )
    ON CONFLICT (user_id) DO NOTHING
    RETURNING user_id
)
SELECT EXISTS (SELECT 1 FROM inserted) AS granted;

-- name: ListSystemSettings :many
SELECT key,
       value,
       encrypted_value,
       encrypted_value_nonce,
       secret,
       updated_by_user_id,
       created_at,
       updated_at
FROM system_settings
ORDER BY key ASC;

-- name: UpsertSystemSetting :one
INSERT INTO system_settings (
    key,
    value,
    encrypted_value,
    encrypted_value_nonce,
    secret,
    updated_by_user_id
) VALUES (
    sqlc.arg(key),
    sqlc.narg(value),
    sqlc.narg(encrypted_value),
    sqlc.narg(encrypted_value_nonce),
    sqlc.arg(secret),
    sqlc.narg(updated_by_user_id)
)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    encrypted_value = EXCLUDED.encrypted_value,
    encrypted_value_nonce = EXCLUDED.encrypted_value_nonce,
    secret = EXCLUDED.secret,
    updated_by_user_id = EXCLUDED.updated_by_user_id
RETURNING key,
          value,
          encrypted_value,
          encrypted_value_nonce,
          secret,
          updated_by_user_id,
          created_at,
          updated_at;

-- name: CreateSystemSettingAuditEvent :exec
INSERT INTO system_setting_audit_events (key, action, updated_by_user_id)
VALUES ($1, $2, sqlc.narg(updated_by_user_id));
