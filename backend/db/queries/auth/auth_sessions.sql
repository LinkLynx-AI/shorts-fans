-- name: CreateAuthSession :one
INSERT INTO app.auth_sessions (
    user_id,
    active_mode,
    session_token_hash,
    expires_at,
    last_seen_at,
    revoked_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(active_mode),
    sqlc.arg(session_token_hash),
    sqlc.arg(expires_at),
    COALESCE(sqlc.narg(last_seen_at)::timestamptz, CURRENT_TIMESTAMP),
    sqlc.narg(revoked_at)
)
RETURNING *;

-- name: GetActiveAuthSessionByTokenHash :one
SELECT *
FROM app.auth_sessions
WHERE session_token_hash = sqlc.arg(session_token_hash)
    AND revoked_at IS NULL
    AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: GetCurrentViewerBySessionTokenHash :one
SELECT
    s.user_id,
    s.active_mode,
    EXISTS (
        SELECT 1
        FROM app.creator_capabilities AS c
        WHERE c.user_id = s.user_id
            AND c.state = 'approved'
    ) AS can_access_creator_mode
FROM app.auth_sessions AS s
WHERE s.session_token_hash = sqlc.arg(session_token_hash)
    AND s.revoked_at IS NULL
    AND s.expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: ListAuthSessionsByUserID :many
SELECT *
FROM app.auth_sessions
WHERE user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: TouchAuthSession :one
UPDATE app.auth_sessions
SET
    active_mode = sqlc.arg(active_mode),
    last_seen_at = COALESCE(sqlc.narg(last_seen_at)::timestamptz, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: TouchAuthSessionLastSeenByTokenHash :one
UPDATE app.auth_sessions
SET
    last_seen_at = COALESCE(sqlc.narg(last_seen_at)::timestamptz, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE session_token_hash = sqlc.arg(session_token_hash)
    AND revoked_at IS NULL
    AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: RevokeAuthSession :one
UPDATE app.auth_sessions
SET
    revoked_at = COALESCE(sqlc.narg(revoked_at)::timestamptz, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: RevokeActiveAuthSessionByTokenHash :one
UPDATE app.auth_sessions
SET
    revoked_at = COALESCE(sqlc.narg(revoked_at)::timestamptz, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE session_token_hash = sqlc.arg(session_token_hash)
    AND revoked_at IS NULL
    AND expires_at > CURRENT_TIMESTAMP
RETURNING *;
