-- name: CreateUser :one
WITH created AS (
    INSERT INTO users (email, display_name, email_verified_at)
    VALUES (sqlc.arg(email), sqlc.arg(display_name), sqlc.narg(email_verified_at))
    RETURNING *
), credential AS (
    INSERT INTO password_credentials (user_id, password_hash)
    SELECT id, sqlc.arg(password_hash)
    FROM created
    RETURNING user_id, password_hash
)
SELECT created.id,
       created.email,
       credential.password_hash,
       true::boolean AS has_password,
       created.display_name,
       created.email_verified_at,
       created.disabled_at,
       false::boolean AS is_system_admin,
       created.created_at,
       created.updated_at
FROM created
JOIN credential ON credential.user_id = created.id;

-- name: GetUserByEmail :one
SELECT users.id,
       users.email,
       COALESCE(password_credentials.password_hash, '')::text AS password_hash,
       (password_credentials.user_id IS NOT NULL)::boolean AS has_password,
       users.display_name,
       users.email_verified_at,
       users.disabled_at,
       EXISTS (
           SELECT 1
           FROM system_user_roles
           WHERE system_user_roles.user_id = users.id
             AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       users.created_at,
       users.updated_at
FROM users
LEFT JOIN password_credentials ON password_credentials.user_id = users.id
WHERE users.email = $1;

-- name: GetUserByID :one
SELECT users.id,
       users.email,
       COALESCE(password_credentials.password_hash, '')::text AS password_hash,
       (password_credentials.user_id IS NOT NULL)::boolean AS has_password,
       users.display_name,
       users.email_verified_at,
       users.disabled_at,
       EXISTS (
           SELECT 1
           FROM system_user_roles
           WHERE system_user_roles.user_id = users.id
             AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       users.created_at,
       users.updated_at
FROM users
LEFT JOIN password_credentials ON password_credentials.user_id = users.id
WHERE users.id = $1;

-- name: UpdateUserDisplayName :one
WITH updated AS (
    UPDATE users
    SET display_name = $2
    WHERE id = $1
    RETURNING *
)
SELECT updated.id,
       updated.email,
       COALESCE(password_credentials.password_hash, '')::text AS password_hash,
       (password_credentials.user_id IS NOT NULL)::boolean AS has_password,
       updated.display_name,
       updated.email_verified_at,
       updated.disabled_at,
       EXISTS (
           SELECT 1 FROM system_user_roles
           WHERE system_user_roles.user_id = updated.id AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       updated.created_at,
       updated.updated_at
FROM updated
LEFT JOIN password_credentials ON password_credentials.user_id = updated.id;

-- name: UpdateUserEmail :one
WITH updated AS (
    UPDATE users
    SET email = $2
    WHERE id = $1
    RETURNING *
)
SELECT updated.id,
       updated.email,
       COALESCE(password_credentials.password_hash, '')::text AS password_hash,
       (password_credentials.user_id IS NOT NULL)::boolean AS has_password,
       updated.display_name,
       updated.email_verified_at,
       updated.disabled_at,
       EXISTS (
           SELECT 1 FROM system_user_roles
           WHERE system_user_roles.user_id = updated.id AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       updated.created_at,
       updated.updated_at
FROM updated
LEFT JOIN password_credentials ON password_credentials.user_id = updated.id;

-- name: UpdateUserPasswordHash :one
WITH credential AS (
    INSERT INTO password_credentials (user_id, password_hash)
    VALUES ($1, $2)
    ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash
    RETURNING user_id, password_hash
)
SELECT users.id,
       users.email,
       credential.password_hash,
       true::boolean AS has_password,
       users.display_name,
       users.email_verified_at,
       users.disabled_at,
       EXISTS (
           SELECT 1 FROM system_user_roles
           WHERE system_user_roles.user_id = users.id AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       users.created_at,
       users.updated_at
FROM users
JOIN credential ON credential.user_id = users.id;

-- name: DeleteUserPasswordCredential :one
WITH deleted AS (
    DELETE FROM password_credentials
    WHERE password_credentials.user_id = sqlc.arg(id)
      AND EXISTS (SELECT 1 FROM user_identities WHERE user_identities.user_id = password_credentials.user_id)
    RETURNING user_id
)
SELECT users.id,
       users.email,
       ''::text AS password_hash,
       false::boolean AS has_password,
       users.display_name,
       users.email_verified_at,
       users.disabled_at,
       EXISTS (
           SELECT 1 FROM system_user_roles
           WHERE system_user_roles.user_id = users.id AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       users.created_at,
       users.updated_at
FROM users
JOIN deleted ON deleted.user_id = users.id;

-- name: MarkUserEmailVerified :one
UPDATE users
SET email_verified_at = COALESCE(users.email_verified_at, $2)
WHERE users.id = $1
RETURNING users.id,
       users.email,
       COALESCE((SELECT password_hash FROM password_credentials WHERE user_id = users.id), '')::text AS password_hash,
       EXISTS(SELECT 1 FROM password_credentials WHERE user_id = users.id)::boolean AS has_password,
       users.display_name,
       users.email_verified_at,
       users.disabled_at,
       EXISTS (
           SELECT 1 FROM system_user_roles
           WHERE system_user_roles.user_id = users.id AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       users.created_at,
       users.updated_at;

-- name: DisableUser :one
UPDATE users
SET disabled_at = COALESCE(users.disabled_at, $1)
WHERE users.id = $2
RETURNING users.id,
       users.email,
       COALESCE((SELECT password_hash FROM password_credentials WHERE user_id = users.id), '')::text AS password_hash,
       EXISTS(SELECT 1 FROM password_credentials WHERE user_id = users.id)::boolean AS has_password,
       users.display_name,
       users.email_verified_at,
       users.disabled_at,
       EXISTS (
           SELECT 1 FROM system_user_roles
           WHERE system_user_roles.user_id = users.id AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       users.created_at,
       users.updated_at;
