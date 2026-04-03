-- name: CreateMediaAsset :one
INSERT INTO app.media_assets (
    creator_user_id,
    processing_state,
    storage_provider,
    storage_bucket,
    storage_key,
    playback_url,
    mime_type,
    duration_ms,
    external_upload_ref
) VALUES (
    sqlc.arg(creator_user_id),
    sqlc.arg(processing_state),
    sqlc.arg(storage_provider),
    sqlc.arg(storage_bucket),
    sqlc.arg(storage_key),
    sqlc.narg(playback_url),
    sqlc.arg(mime_type),
    sqlc.narg(duration_ms),
    sqlc.narg(external_upload_ref)
)
RETURNING *;

-- name: GetMediaAssetByID :one
SELECT *
FROM app.media_assets
WHERE id = $1
LIMIT 1;

-- name: ListMediaAssetsByCreatorUserID :many
SELECT *
FROM app.media_assets
WHERE creator_user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: UpdateMediaAssetProcessingState :one
UPDATE app.media_assets
SET
    processing_state = sqlc.arg(processing_state),
    playback_url = sqlc.narg(playback_url),
    duration_ms = sqlc.narg(duration_ms),
    external_upload_ref = sqlc.narg(external_upload_ref),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;
