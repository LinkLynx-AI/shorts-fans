-- name: CreateShort :one
INSERT INTO app.shorts (
    creator_user_id,
    canonical_main_id,
    media_asset_id,
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
    sqlc.narg(caption),
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

-- name: GetShortByMediaAssetID :one
SELECT *
FROM app.shorts
WHERE media_asset_id = $1
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

-- name: UpdateShortCaption :one
UPDATE app.shorts
SET
    caption = sqlc.narg(caption),
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

-- name: CountPublicShortsByCreatorUserID :one
SELECT COUNT(*)::bigint
FROM app.public_shorts
WHERE creator_user_id = $1;

-- name: ListCreatorProfileShortGridItems :many
SELECT
    s.id,
    s.creator_user_id,
    s.canonical_main_id,
    s.media_asset_id,
    s.published_at,
    m.playback_url,
    m.duration_ms
FROM app.public_shorts AS s
JOIN app.media_assets AS m
    ON m.id = s.media_asset_id
WHERE s.creator_user_id = sqlc.arg(creator_user_id)
AND (
    sqlc.narg(cursor_published_at)::timestamptz IS NULL
    OR s.published_at < sqlc.narg(cursor_published_at)::timestamptz
    OR (
        s.published_at = sqlc.narg(cursor_published_at)::timestamptz
        AND s.id < COALESCE(sqlc.narg(cursor_short_id)::uuid, 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid)
    )
)
ORDER BY s.published_at DESC, s.id DESC
LIMIT sqlc.arg(limit_count);

-- name: GetPublicShortByID :one
SELECT *
FROM app.public_shorts
WHERE id = $1
LIMIT 1;

-- name: PutPinnedShort :exec
INSERT INTO app.pinned_shorts (
    user_id,
    short_id
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(short_id)
)
ON CONFLICT (user_id, short_id) DO NOTHING;

-- name: DeletePinnedShort :exec
DELETE FROM app.pinned_shorts
WHERE user_id = sqlc.arg(user_id)
  AND short_id = sqlc.arg(short_id);
