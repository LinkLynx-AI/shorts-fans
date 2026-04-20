-- name: GetSubmissionReviewMainByIDForUpdate :one
SELECT
    m.id,
    m.creator_user_id,
    m.media_asset_id,
    m.state,
    m.review_reason_code,
    m.post_report_state,
    m.price_minor,
    m.currency_code,
    m.ownership_confirmed,
    m.consent_confirmed,
    m.approved_for_unlock_at,
    m.created_at,
    m.updated_at,
    a.processing_state AS media_processing_state
FROM app.mains AS m
JOIN app.media_assets AS a
    ON a.id = m.media_asset_id
WHERE m.id = $1
LIMIT 1
FOR UPDATE OF m;

-- name: ListSubmissionReviewShortsByCanonicalMainIDForUpdate :many
SELECT
    s.id,
    s.creator_user_id,
    s.canonical_main_id,
    s.media_asset_id,
    s.state,
    s.review_reason_code,
    s.post_report_state,
    s.approved_for_publish_at,
    s.published_at,
    s.created_at,
    s.updated_at,
    s.caption,
    a.processing_state AS media_processing_state
FROM app.shorts AS s
JOIN app.media_assets AS a
    ON a.id = s.media_asset_id
WHERE s.canonical_main_id = $1
ORDER BY s.created_at DESC, s.id DESC
FOR UPDATE OF s;

-- name: GetPendingSubmissionReviewIntakeByCanonicalMainID :one
SELECT *
FROM app.submission_review_intakes
WHERE canonical_main_id = $1
  AND status = 'pending_review'
ORDER BY submitted_at DESC, id DESC
LIMIT 1;

-- name: GetLatestSubmissionReviewIntakeByCanonicalMainID :one
SELECT *
FROM app.submission_review_intakes
WHERE canonical_main_id = $1
ORDER BY submitted_at DESC, id DESC
LIMIT 1;

-- name: CreateSubmissionReviewIntake :one
INSERT INTO app.submission_review_intakes (
    canonical_main_id,
    creator_user_id,
    status,
    submit_kind,
    previous_intake_id,
    main_media_asset_id,
    main_price_minor,
    ownership_confirmed,
    consent_confirmed,
    submitted_at
) VALUES (
    sqlc.arg(canonical_main_id),
    sqlc.arg(creator_user_id),
    sqlc.arg(status),
    sqlc.arg(submit_kind),
    sqlc.narg(previous_intake_id),
    sqlc.arg(main_media_asset_id),
    sqlc.arg(main_price_minor),
    sqlc.arg(ownership_confirmed),
    sqlc.arg(consent_confirmed),
    sqlc.arg(submitted_at)
)
RETURNING *;

-- name: CreateSubmissionReviewIntakeShort :exec
INSERT INTO app.submission_review_intake_shorts (
    submission_review_intake_id,
    short_id,
    media_asset_id,
    caption
) VALUES (
    sqlc.arg(submission_review_intake_id),
    sqlc.arg(short_id),
    sqlc.arg(media_asset_id),
    sqlc.narg(caption)
);
