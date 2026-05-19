-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING id, email, display_name, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, display_name, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, display_name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = $2
WHERE id = $1
RETURNING id, email, password_hash, display_name, created_at, updated_at;

-- name: UpdateUserEmail :one
UPDATE users
SET email = $2
WHERE id = $1
RETURNING id, email, password_hash, display_name, created_at, updated_at;

-- name: UpdateUserPasswordHash :one
UPDATE users
SET password_hash = $2
WHERE id = $1
RETURNING id, email, password_hash, display_name, created_at, updated_at;
