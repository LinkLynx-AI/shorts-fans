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

-- name: GetMediaProcessingJobByMediaAssetID :one
SELECT *
FROM app.media_processing_jobs
WHERE media_asset_id = $1
LIMIT 1;

-- name: ClaimMediaProcessingJobByAssetID :one
WITH candidate AS (
    SELECT id
    FROM app.media_processing_jobs
    WHERE app.media_processing_jobs.media_asset_id = $1
        AND app.media_processing_jobs.status = 'queued'
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE app.media_processing_jobs AS j
SET
    status = 'processing',
    attempt_count = j.attempt_count + 1,
    last_error_code = NULL,
    last_error_message = NULL,
    started_at = CURRENT_TIMESTAMP,
    completed_at = NULL,
    failed_at = NULL,
    updated_at = CURRENT_TIMESTAMP
FROM candidate
WHERE j.id = candidate.id
RETURNING j.*;

-- name: ClaimNextQueuedMediaProcessingJob :one
WITH candidate AS (
    SELECT id
    FROM app.media_processing_jobs
    WHERE status = 'queued'
    ORDER BY queued_at ASC, id ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE app.media_processing_jobs AS j
SET
    status = 'processing',
    attempt_count = j.attempt_count + 1,
    last_error_code = NULL,
    last_error_message = NULL,
    started_at = CURRENT_TIMESTAMP,
    completed_at = NULL,
    failed_at = NULL,
    updated_at = CURRENT_TIMESTAMP
FROM candidate
WHERE j.id = candidate.id
RETURNING j.*;

-- name: MarkMediaProcessingJobSucceeded :one
UPDATE app.media_processing_jobs
SET
    status = 'succeeded',
    last_error_code = NULL,
    last_error_message = NULL,
    completed_at = CURRENT_TIMESTAMP,
    failed_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: RequeueMediaProcessingJob :one
UPDATE app.media_processing_jobs
SET
    status = 'queued',
    last_error_code = sqlc.narg(last_error_code),
    last_error_message = sqlc.narg(last_error_message),
    queued_at = CURRENT_TIMESTAMP,
    started_at = NULL,
    completed_at = NULL,
    failed_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: MarkMediaProcessingJobFailed :one
UPDATE app.media_processing_jobs
SET
    status = 'failed',
    last_error_code = sqlc.narg(last_error_code),
    last_error_message = sqlc.narg(last_error_message),
    failed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;
