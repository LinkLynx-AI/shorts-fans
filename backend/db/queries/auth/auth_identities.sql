-- name: CreateAuthIdentity :one
INSERT INTO app.auth_identities (
    user_id,
    provider,
    provider_subject,
    email_normalized,
    verified_at,
    last_authenticated_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(provider),
    sqlc.arg(provider_subject),
    sqlc.narg(email_normalized),
    sqlc.narg(verified_at),
    sqlc.narg(last_authenticated_at)
)
RETURNING *;

-- name: GetAuthIdentityByProviderAndSubject :one
SELECT *
FROM app.auth_identities
WHERE provider = sqlc.arg(provider)
    AND provider_subject = sqlc.arg(provider_subject)
LIMIT 1;

-- name: GetAuthIdentityByEmailNormalized :one
SELECT *
FROM app.auth_identities
WHERE email_normalized = sqlc.arg(email_normalized)
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: ListAuthIdentitiesByUserID :many
SELECT *
FROM app.auth_identities
WHERE user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: RecordAuthIdentityAuthentication :one
UPDATE app.auth_identities
SET
    email_normalized = COALESCE(sqlc.narg(email_normalized), email_normalized),
    verified_at = COALESCE(sqlc.narg(verified_at), verified_at),
    last_authenticated_at = COALESCE(sqlc.narg(last_authenticated_at)::timestamptz, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;
