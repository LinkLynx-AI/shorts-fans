-- name: CreateAuthLoginChallenge :one
INSERT INTO app.auth_login_challenges (
    provider,
    provider_subject,
    email_normalized,
    challenge_token_hash,
    purpose,
    expires_at,
    consumed_at,
    attempt_count
) VALUES (
    sqlc.arg(provider),
    sqlc.arg(provider_subject),
    sqlc.narg(email_normalized),
    sqlc.arg(challenge_token_hash),
    sqlc.arg(purpose),
    sqlc.arg(expires_at),
    sqlc.narg(consumed_at),
    sqlc.arg(attempt_count)
)
RETURNING *;

-- name: GetLatestPendingAuthLoginChallengeByProviderAndSubject :one
SELECT *
FROM app.auth_login_challenges
WHERE provider = sqlc.arg(provider)
    AND provider_subject = sqlc.arg(provider_subject)
    AND consumed_at IS NULL
    AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: IncrementAuthLoginChallengeAttemptCount :one
UPDATE app.auth_login_challenges
SET
    attempt_count = attempt_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ConsumeAuthLoginChallenge :one
UPDATE app.auth_login_challenges
SET
    consumed_at = COALESCE(sqlc.narg(consumed_at)::timestamptz, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;
