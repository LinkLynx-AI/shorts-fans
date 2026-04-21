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

-- name: GetInitialReviewReadyPackageByMainID :one
SELECT
    m.id,
    m.creator_user_id
FROM app.mains AS m
JOIN app.creator_capabilities AS c
    ON c.user_id = m.creator_user_id
JOIN app.media_assets AS main_asset
    ON main_asset.id = m.media_asset_id
WHERE m.id = $1
    AND c.state = 'approved'
    AND m.state = 'draft'
    AND m.price_minor > 0
    AND m.ownership_confirmed = TRUE
    AND m.consent_confirmed = TRUE
    AND main_asset.processing_state = 'ready'
    AND NOT EXISTS (
        SELECT 1
        FROM app.submission_review_intakes AS intake
        WHERE intake.canonical_main_id = m.id
            AND intake.status = 'pending_review'
    )
    AND EXISTS (
        SELECT 1
        FROM app.shorts AS s
        JOIN app.media_assets AS short_asset
            ON short_asset.id = s.media_asset_id
        WHERE s.canonical_main_id = m.id
            AND s.state = 'draft'
            AND short_asset.processing_state = 'ready'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM app.shorts AS s
        JOIN app.media_assets AS short_asset
            ON short_asset.id = s.media_asset_id
        WHERE s.canonical_main_id = m.id
            AND (
                s.state <> 'draft'
                OR short_asset.processing_state <> 'ready'
            )
    )
LIMIT 1;

-- name: GetNextSucceededInitialReviewMediaProcessingJob :one
SELECT j.*
FROM app.media_processing_jobs AS j
JOIN app.media_assets AS completed_asset
    ON completed_asset.id = j.media_asset_id
LEFT JOIN app.mains AS completed_main
    ON j.asset_role = 'main'
    AND completed_main.media_asset_id = j.media_asset_id
LEFT JOIN app.shorts AS completed_short
    ON j.asset_role = 'short'
    AND completed_short.media_asset_id = j.media_asset_id
JOIN app.mains AS m
    ON m.id = COALESCE(completed_main.id, completed_short.canonical_main_id)
JOIN app.creator_capabilities AS c
    ON c.user_id = m.creator_user_id
JOIN app.media_assets AS main_asset
    ON main_asset.id = m.media_asset_id
WHERE j.status = 'succeeded'
    AND j.last_error_code = 'review_submit_failed'
    AND completed_asset.processing_state = 'ready'
    AND j.creator_user_id = m.creator_user_id
    AND c.state = 'approved'
    AND m.state = 'draft'
    AND m.price_minor > 0
    AND m.ownership_confirmed = TRUE
    AND m.consent_confirmed = TRUE
    AND main_asset.processing_state = 'ready'
    AND NOT EXISTS (
        SELECT 1
        FROM app.submission_review_intakes AS intake
        WHERE intake.canonical_main_id = m.id
            AND intake.status = 'pending_review'
    )
    AND EXISTS (
        SELECT 1
        FROM app.shorts AS s
        JOIN app.media_assets AS short_asset
            ON short_asset.id = s.media_asset_id
        WHERE s.canonical_main_id = m.id
            AND s.state = 'draft'
            AND short_asset.processing_state = 'ready'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM app.shorts AS s
        JOIN app.media_assets AS short_asset
            ON short_asset.id = s.media_asset_id
        WHERE s.canonical_main_id = m.id
            AND (
                s.state <> 'draft'
                OR short_asset.processing_state <> 'ready'
            )
    )
ORDER BY j.completed_at ASC NULLS LAST, j.updated_at ASC, j.id ASC
LIMIT 1;

-- name: MarkMediaProcessingJobReviewSubmitFailed :one
UPDATE app.media_processing_jobs
SET
    last_error_code = 'review_submit_failed',
    last_error_message = sqlc.narg(last_error_message),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
    AND status = 'succeeded'
RETURNING *;

-- name: ClearMediaProcessingJobReviewSubmitFailure :exec
UPDATE app.media_processing_jobs
SET
    last_error_code = NULL,
    last_error_message = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
    AND status = 'succeeded'
    AND last_error_code = 'review_submit_failed';

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
