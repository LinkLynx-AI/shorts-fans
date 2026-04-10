-- name: CreateMediaProcessingJob :one
INSERT INTO app.media_processing_jobs (
    creator_user_id,
    media_asset_id,
    asset_role,
    status,
    attempt_count,
    last_error_code,
    last_error_message,
    started_at,
    completed_at,
    failed_at
) VALUES (
    sqlc.arg(creator_user_id),
    sqlc.arg(media_asset_id),
    sqlc.arg(asset_role),
    sqlc.arg(status),
    sqlc.arg(attempt_count),
    sqlc.narg(last_error_code),
    sqlc.narg(last_error_message),
    sqlc.narg(started_at),
    sqlc.narg(completed_at),
    sqlc.narg(failed_at)
)
RETURNING *;
