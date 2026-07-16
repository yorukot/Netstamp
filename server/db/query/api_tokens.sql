-- name: CreateAPIToken :one
INSERT INTO api_tokens (user_id, name, token_hash, token_hint, scopes, created_at, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(name), sqlc.arg(token_hash), sqlc.arg(token_hint), sqlc.arg(scopes), sqlc.arg(created_at), sqlc.arg(expires_at))
RETURNING *;

-- name: CountActiveAPITokensForUser :one
SELECT count(*)
FROM api_tokens
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(now_at);

-- name: LockUserForAPITokenCreate :one
SELECT id
FROM users
WHERE id = sqlc.arg(user_id)
FOR UPDATE;

-- name: ListAPITokensForUser :many
SELECT *
FROM api_tokens
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: GetActiveAPITokenByHash :one
SELECT api_tokens.*
FROM api_tokens
JOIN users ON users.id = api_tokens.user_id
WHERE api_tokens.token_hash = sqlc.arg(token_hash)
  AND api_tokens.revoked_at IS NULL
  AND api_tokens.expires_at > sqlc.arg(now_at)
  AND users.disabled_at IS NULL;

-- name: TouchAPIToken :exec
UPDATE api_tokens
SET last_used_at = sqlc.arg(last_used_at)
WHERE id = sqlc.arg(id)
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(last_used_at)
  AND (last_used_at IS NULL OR last_used_at < sqlc.arg(touch_before));

-- name: RevokeAPITokenForUser :one
UPDATE api_tokens
SET revoked_at = sqlc.arg(revoked_at), revoked_reason = sqlc.arg(revoked_reason)
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
RETURNING id;

-- name: RevokeAPITokensForUser :exec
UPDATE api_tokens
SET revoked_at = sqlc.arg(revoked_at), revoked_reason = sqlc.arg(revoked_reason)
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;
