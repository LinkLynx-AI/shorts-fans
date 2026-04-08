-- name: CreateCreatorProfile :one
INSERT INTO app.creator_profiles (
    user_id,
    display_name,
    handle,
    avatar_url,
    bio,
    published_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.narg(display_name),
    sqlc.narg(handle),
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
    handle = sqlc.narg(handle),
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

-- name: GetPublicCreatorProfileByHandle :one
SELECT *
FROM app.public_creator_profiles
WHERE handle = sqlc.arg(handle)
LIMIT 1;

-- name: CountCreatorFollowersByCreatorUserID :one
SELECT COUNT(*)::bigint
FROM app.creator_follows
WHERE creator_user_id = $1;

-- name: PutCreatorFollow :exec
INSERT INTO app.creator_follows (
    user_id,
    creator_user_id
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(creator_user_id)
)
ON CONFLICT (user_id, creator_user_id) DO NOTHING;

-- name: DeleteCreatorFollow :exec
DELETE FROM app.creator_follows
WHERE user_id = sqlc.arg(user_id)
AND creator_user_id = sqlc.arg(creator_user_id);

-- name: GetViewerCreatorFollowState :one
SELECT EXISTS (
    SELECT 1
    FROM app.creator_follows
    WHERE user_id = sqlc.arg(user_id)
    AND creator_user_id = sqlc.arg(creator_user_id)
) AS is_following;

-- name: ListRecentPublicCreatorProfiles :many
SELECT *
FROM app.public_creator_profiles
WHERE (
    sqlc.narg(cursor_published_at)::timestamptz IS NULL
    OR published_at < sqlc.narg(cursor_published_at)::timestamptz
    OR (
        published_at = sqlc.narg(cursor_published_at)::timestamptz
        AND handle > COALESCE(sqlc.narg(cursor_handle)::text, '')
    )
)
ORDER BY published_at DESC, handle ASC
LIMIT sqlc.arg(limit_count);

-- name: SearchPublicCreatorProfiles :many
SELECT *
FROM app.public_creator_profiles
WHERE (
    display_name ILIKE '%' || sqlc.arg(display_name_query) || '%' ESCAPE '\'
    OR (
        sqlc.arg(handle_prefix_query) <> ''
        AND handle LIKE sqlc.arg(handle_prefix_query) || '%' ESCAPE '\'
    )
)
AND (
    sqlc.narg(cursor_published_at)::timestamptz IS NULL
    OR published_at < sqlc.narg(cursor_published_at)::timestamptz
    OR (
        published_at = sqlc.narg(cursor_published_at)::timestamptz
        AND handle > COALESCE(sqlc.narg(cursor_handle)::text, '')
    )
)
ORDER BY published_at DESC, handle ASC
LIMIT sqlc.arg(limit_count);
