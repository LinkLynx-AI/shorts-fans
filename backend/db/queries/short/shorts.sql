-- name: CreateShort :one
INSERT INTO app.shorts (
    creator_user_id,
    canonical_main_id,
    media_asset_id,
    title,
    caption,
    state,
    review_reason_code,
    post_report_state,
    approved_for_publish_at,
    published_at
) VALUES (
    sqlc.arg(creator_user_id),
    sqlc.arg(canonical_main_id),
    sqlc.arg(media_asset_id),
    sqlc.arg(title),
    sqlc.arg(caption),
    sqlc.arg(state),
    sqlc.narg(review_reason_code),
    sqlc.narg(post_report_state),
    sqlc.narg(approved_for_publish_at),
    sqlc.narg(published_at)
)
RETURNING *;

-- name: GetShortByID :one
SELECT *
FROM app.shorts
WHERE id = $1
LIMIT 1;

-- name: ListShortsByCreatorUserID :many
SELECT *
FROM app.shorts
WHERE creator_user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: UpdateShortState :one
UPDATE app.shorts
SET
    state = sqlc.arg(state),
    review_reason_code = sqlc.narg(review_reason_code),
    post_report_state = sqlc.narg(post_report_state),
    approved_for_publish_at = sqlc.narg(approved_for_publish_at),
    published_at = sqlc.narg(published_at),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: PublishShort :one
UPDATE app.shorts
SET
    state = 'approved_for_publish',
    approved_for_publish_at = COALESCE(approved_for_publish_at, CURRENT_TIMESTAMP),
    published_at = COALESCE(published_at, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListPublicShortsByCreatorUserID :many
SELECT *
FROM app.public_shorts
WHERE creator_user_id = $1
ORDER BY published_at DESC, created_at DESC, id DESC;

-- name: GetPublicShortByID :one
SELECT *
FROM app.public_shorts
WHERE id = $1
LIMIT 1;
