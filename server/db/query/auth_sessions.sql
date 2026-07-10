-- name: CreateAuthSession :one
INSERT INTO auth_sessions (
    user_id,
    token_hash,
    csrf_token_hash,
    created_at,
    last_used_at,
    idle_expires_at,
    absolute_expires_at
)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(csrf_token_hash),
    sqlc.arg(created_at),
    sqlc.arg(last_used_at),
    sqlc.arg(idle_expires_at),
    sqlc.arg(absolute_expires_at)
)
RETURNING id,
          user_id,
          token_hash,
          csrf_token_hash,
          created_at,
          last_used_at,
          idle_expires_at,
          absolute_expires_at,
          revoked_at,
          revoked_reason;

-- name: GetActiveAuthSessionByTokenHash :one
SELECT auth_sessions.id,
       auth_sessions.user_id,
       auth_sessions.token_hash,
       auth_sessions.csrf_token_hash,
       auth_sessions.created_at,
       auth_sessions.last_used_at,
       auth_sessions.idle_expires_at,
       auth_sessions.absolute_expires_at,
       auth_sessions.revoked_at,
       auth_sessions.revoked_reason
FROM auth_sessions
JOIN users ON users.id = auth_sessions.user_id
WHERE auth_sessions.token_hash = sqlc.arg(token_hash)
  AND auth_sessions.revoked_at IS NULL
  AND auth_sessions.idle_expires_at > sqlc.arg(now_at)
  AND auth_sessions.absolute_expires_at > sqlc.arg(now_at)
  AND users.disabled_at IS NULL;

-- name: GetActiveAuthSessionByID :one
SELECT auth_sessions.id,
       auth_sessions.user_id,
       auth_sessions.token_hash,
       auth_sessions.csrf_token_hash,
       auth_sessions.created_at,
       auth_sessions.last_used_at,
       auth_sessions.idle_expires_at,
       auth_sessions.absolute_expires_at,
       auth_sessions.revoked_at,
       auth_sessions.revoked_reason
FROM auth_sessions
JOIN users ON users.id = auth_sessions.user_id
WHERE auth_sessions.id = sqlc.arg(id)
  AND auth_sessions.revoked_at IS NULL
  AND auth_sessions.idle_expires_at > sqlc.arg(now_at)
  AND auth_sessions.absolute_expires_at > sqlc.arg(now_at)
  AND users.disabled_at IS NULL;

-- name: UpdateAuthSessionCSRFTokenHash :exec
UPDATE auth_sessions
SET csrf_token_hash = sqlc.arg(csrf_token_hash)
WHERE id = sqlc.arg(id)
  AND revoked_at IS NULL
  AND idle_expires_at > sqlc.arg(now_at)
  AND absolute_expires_at > sqlc.arg(now_at);

-- name: TouchAuthSession :exec
UPDATE auth_sessions
SET last_used_at = sqlc.arg(last_used_at),
    idle_expires_at = sqlc.arg(idle_expires_at)
WHERE id = sqlc.arg(id)
  AND revoked_at IS NULL
  AND absolute_expires_at > sqlc.arg(last_used_at);

-- name: RevokeAuthSessionByTokenHash :exec
UPDATE auth_sessions
SET revoked_at = sqlc.arg(revoked_at),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL;

-- name: RevokeAuthSessionByID :exec
UPDATE auth_sessions
SET revoked_at = sqlc.arg(revoked_at),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE id = sqlc.arg(id)
  AND revoked_at IS NULL;

-- name: RevokeAuthSessionsForUser :exec
UPDATE auth_sessions
SET revoked_at = sqlc.arg(revoked_at),
    revoked_reason = sqlc.arg(revoked_reason)
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;
