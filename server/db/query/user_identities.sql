-- name: CreateUserIdentity :one
INSERT INTO user_identities (
    user_id, provider, issuer, subject, email, email_verified, display_name, created_at, last_login_at
)
VALUES (
    sqlc.arg(user_id), sqlc.arg(provider), sqlc.arg(issuer), sqlc.arg(subject),
    sqlc.narg(email), sqlc.arg(email_verified), sqlc.narg(display_name),
    sqlc.arg(created_at), sqlc.narg(last_login_at)
)
RETURNING *;

-- name: GetUserIdentityByIssuerSubject :one
SELECT *
FROM user_identities
WHERE issuer = sqlc.arg(issuer)
  AND subject = sqlc.arg(subject);

-- name: GetUserIdentityByIDForUser :one
SELECT *
FROM user_identities
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id);

-- name: ListUserIdentities :many
SELECT *
FROM user_identities
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at ASC;

-- name: TouchUserIdentityLogin :one
UPDATE user_identities
SET email = sqlc.narg(email),
    email_verified = sqlc.arg(email_verified),
    display_name = sqlc.narg(display_name),
    last_login_at = sqlc.arg(last_login_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUserIdentityForUser :one
DELETE FROM user_identities
WHERE user_identities.id = sqlc.arg(id)
  AND user_identities.user_id = sqlc.arg(user_id)
  AND (
      EXISTS (SELECT 1 FROM password_credentials WHERE password_credentials.user_id = user_identities.user_id)
      OR EXISTS (
          SELECT 1 FROM user_identities AS other
          WHERE other.user_id = user_identities.user_id AND other.id <> user_identities.id
      )
  )
RETURNING user_identities.id;

-- name: CountUserAuthenticationMethods :one
SELECT EXISTS (
           SELECT 1 FROM password_credentials WHERE password_credentials.user_id = sqlc.arg(target_user_id)
       ) AS has_password,
       count(user_identities.id)::bigint AS identity_count
FROM users
LEFT JOIN user_identities ON user_identities.user_id = users.id
WHERE users.id = sqlc.arg(target_user_id)
GROUP BY users.id;

-- name: CreateOIDCUser :one
WITH created_user AS (
    INSERT INTO users (email, display_name, email_verified_at)
    VALUES (sqlc.arg(email), sqlc.arg(display_name), sqlc.arg(email_verified_at))
    RETURNING *
), created_identity AS (
    INSERT INTO user_identities (
        user_id, provider, issuer, subject, email, email_verified, display_name, created_at, last_login_at
    )
    SELECT id, 'oidc', sqlc.arg(issuer), sqlc.arg(subject), sqlc.arg(email), true,
           sqlc.arg(display_name), sqlc.arg(created_at), sqlc.arg(created_at)
    FROM created_user
    RETURNING *
)
SELECT created_user.id,
       created_user.email,
       created_user.display_name,
       created_user.email_verified_at,
       created_user.disabled_at,
       created_user.created_at,
       created_user.updated_at,
       created_identity.id AS identity_id
FROM created_user
CROSS JOIN created_identity;

-- name: CreateOIDCAuthFlow :one
INSERT INTO oidc_auth_flows (
    state_hash, browser_token_hash, nonce, pkce_verifier, intent, session_id, return_to, created_at, expires_at
)
VALUES (
    sqlc.arg(state_hash), sqlc.arg(browser_token_hash), sqlc.arg(nonce), sqlc.arg(pkce_verifier),
    sqlc.arg(intent), sqlc.narg(session_id), sqlc.arg(return_to), sqlc.arg(created_at), sqlc.arg(expires_at)
)
RETURNING *;

-- name: ConsumeOIDCAuthFlow :one
UPDATE oidc_auth_flows
SET used_at = sqlc.arg(used_at)
WHERE state_hash = sqlc.arg(state_hash)
  AND browser_token_hash = sqlc.arg(browser_token_hash)
  AND used_at IS NULL
  AND expires_at > sqlc.arg(used_at)
RETURNING *;

-- name: DeleteExpiredOIDCAuthFlows :exec
DELETE FROM oidc_auth_flows
WHERE expires_at <= sqlc.arg(now_at)
   OR used_at IS NOT NULL;
