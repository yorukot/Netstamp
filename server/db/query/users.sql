-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name, email_verified_at)
VALUES ($1, $2, $3, sqlc.narg(email_verified_at))
RETURNING id, email, password_hash, display_name, email_verified_at, false::boolean AS is_system_admin, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id,
       email,
       password_hash,
       display_name,
       email_verified_at,
       EXISTS (
           SELECT 1
           FROM system_user_roles
           WHERE system_user_roles.user_id = users.id
             AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       created_at,
       updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id,
       email,
       password_hash,
       display_name,
       email_verified_at,
       EXISTS (
           SELECT 1
           FROM system_user_roles
           WHERE system_user_roles.user_id = users.id
             AND system_user_roles.role = 'admin'
       ) AS is_system_admin,
       created_at,
       updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = $2
WHERE id = $1
RETURNING id,
          email,
          password_hash,
          display_name,
          email_verified_at,
          EXISTS (
              SELECT 1
              FROM system_user_roles
              WHERE system_user_roles.user_id = users.id
                AND system_user_roles.role = 'admin'
          ) AS is_system_admin,
          created_at,
          updated_at;

-- name: UpdateUserEmail :one
UPDATE users
SET email = $2
WHERE id = $1
RETURNING id,
          email,
          password_hash,
          display_name,
          email_verified_at,
          EXISTS (
              SELECT 1
              FROM system_user_roles
              WHERE system_user_roles.user_id = users.id
                AND system_user_roles.role = 'admin'
          ) AS is_system_admin,
          created_at,
          updated_at;

-- name: UpdateUserPasswordHash :one
UPDATE users
SET password_hash = $2
WHERE id = $1
RETURNING id,
          email,
          password_hash,
          display_name,
          email_verified_at,
          EXISTS (
              SELECT 1
              FROM system_user_roles
              WHERE system_user_roles.user_id = users.id
                AND system_user_roles.role = 'admin'
          ) AS is_system_admin,
          created_at,
          updated_at;

-- name: MarkUserEmailVerified :one
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, sqlc.arg(verified_at))
WHERE id = sqlc.arg(id)
RETURNING id,
          email,
          password_hash,
          display_name,
          email_verified_at,
          EXISTS (
              SELECT 1
              FROM system_user_roles
              WHERE system_user_roles.user_id = users.id
                AND system_user_roles.role = 'admin'
          ) AS is_system_admin,
          created_at,
          updated_at;
