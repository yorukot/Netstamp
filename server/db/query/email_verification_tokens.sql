-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(expires_at))
RETURNING id, user_id, token_hash, expires_at, used_at, created_at;

-- name: InvalidateActiveEmailVerificationTokens :exec
UPDATE email_verification_tokens
SET used_at = sqlc.arg(used_at)
WHERE user_id = sqlc.arg(user_id)
  AND used_at IS NULL;

-- name: GetActiveEmailVerificationTokenByHash :one
SELECT id, user_id, token_hash, expires_at, used_at, created_at
FROM email_verification_tokens
WHERE token_hash = sqlc.arg(token_hash)
  AND used_at IS NULL
  AND expires_at > sqlc.arg(now_at)
LIMIT 1;

-- name: MarkEmailVerificationTokenUsed :exec
UPDATE email_verification_tokens
SET used_at = sqlc.arg(used_at)
WHERE id = sqlc.arg(id)
  AND used_at IS NULL;
