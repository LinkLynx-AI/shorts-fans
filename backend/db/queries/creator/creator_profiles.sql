-- name: CreateCreatorProfile :one
INSERT INTO app.creator_profiles (
    user_id,
    display_name,
    avatar_url,
    bio,
    published_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.narg(display_name),
    sqlc.narg(avatar_url),
    sqlc.arg(bio),
    sqlc.narg(published_at)
)
RETURNING *;

-- name: GetCreatorProfileByUserID :one
SELECT *
FROM app.creator_profiles
WHERE user_id = $1
LIMIT 1;

-- name: UpdateCreatorProfile :one
UPDATE app.creator_profiles
SET
    display_name = sqlc.narg(display_name),
    avatar_url = sqlc.narg(avatar_url),
    bio = sqlc.arg(bio),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
RETURNING *;

-- name: PublishCreatorProfile :one
UPDATE app.creator_profiles
SET
    published_at = COALESCE(published_at, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING *;

-- name: GetPublicCreatorProfileByUserID :one
SELECT *
FROM app.public_creator_profiles
WHERE user_id = $1
LIMIT 1;
