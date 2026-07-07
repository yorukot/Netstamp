-- name: IsSystemAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM system_user_roles
    WHERE user_id = $1
      AND role = 'admin'
) AS is_system_admin;

-- name: ListSystemAdmins :many
SELECT users.id,
       users.email,
       users.display_name,
       users.email_verified_at,
       users.created_at,
       users.updated_at,
       system_user_roles.created_at AS granted_at
FROM system_user_roles
JOIN users ON users.id = system_user_roles.user_id
WHERE system_user_roles.role = 'admin'
ORDER BY system_user_roles.created_at ASC, users.email ASC;

-- name: GrantSystemAdminByEmail :one
WITH target AS (
    SELECT id,
           email,
           display_name,
           email_verified_at,
           created_at,
           updated_at
    FROM users
    WHERE email = $1
),
upserted AS (
    INSERT INTO system_user_roles (user_id, role)
    SELECT id, 'admin'
    FROM target
    ON CONFLICT (user_id) DO UPDATE
    SET role = EXCLUDED.role
    RETURNING user_id
)
SELECT target.id,
       target.email,
       target.display_name,
       target.email_verified_at,
       target.created_at,
       target.updated_at,
       system_user_roles.created_at AS granted_at
FROM target
JOIN upserted ON upserted.user_id = target.id
JOIN system_user_roles ON system_user_roles.user_id = target.id
WHERE system_user_roles.role = 'admin';

-- name: RevokeSystemAdminIfNotLast :one
WITH lock AS (
    SELECT pg_advisory_xact_lock(hashtextextended('netstamp.system_admin.manage', 0))
),
admin_count AS (
    SELECT count(*)::bigint AS total
    FROM system_user_roles, lock
    WHERE role = 'admin'
),
target AS (
    SELECT system_user_roles.user_id
    FROM system_user_roles, lock
    WHERE system_user_roles.user_id = $1
      AND system_user_roles.role = 'admin'
),
deleted AS (
    DELETE FROM system_user_roles
    WHERE system_user_roles.user_id = $1
      AND system_user_roles.role = 'admin'
      AND (SELECT total FROM admin_count) > 1
    RETURNING user_id
)
SELECT (SELECT total FROM admin_count) AS admin_count,
       EXISTS (SELECT 1 FROM target) AS target_was_admin,
       EXISTS (SELECT 1 FROM deleted) AS revoked;

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
