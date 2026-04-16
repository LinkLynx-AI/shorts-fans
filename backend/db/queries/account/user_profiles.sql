-- name: CreateUserProfile :one
INSERT INTO app.user_profiles (
    user_id,
    display_name,
    handle,
    avatar_url
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(display_name),
    sqlc.arg(handle),
    sqlc.narg(avatar_url)
)
RETURNING *;

-- name: GetUserProfileByUserID :one
SELECT *
FROM app.user_profiles
WHERE user_id = $1
LIMIT 1;

-- name: GetUserProfileByHandle :one
SELECT *
FROM app.user_profiles
WHERE handle = sqlc.arg(handle)
LIMIT 1;

-- name: UpdateUserProfile :one
UPDATE app.user_profiles
SET
    display_name = sqlc.arg(display_name),
    handle = sqlc.arg(handle),
    avatar_url = sqlc.narg(avatar_url),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
RETURNING *;
