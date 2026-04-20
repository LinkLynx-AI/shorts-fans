-- name: ListAdminSubmissionReviewQueue :many
SELECT
    i.id AS intake_id,
    i.submit_kind,
    i.submitted_at,
    i.canonical_main_id,
    i.creator_user_id,
    u.display_name,
    u.handle,
    u.avatar_url,
    COALESCE(p.bio, '') AS creator_bio,
    m.state AS main_state,
    m.state = 'pending_review' AS main_decision_required,
    COUNT(intake_short.short_id)::bigint AS short_count,
    COUNT(*) FILTER (WHERE s.state = 'pending_review')::bigint AS pending_short_count
FROM app.submission_review_intakes AS i
JOIN app.user_profiles AS u
    ON u.user_id = i.creator_user_id
LEFT JOIN app.creator_profiles AS p
    ON p.user_id = i.creator_user_id
JOIN app.mains AS m
    ON m.id = i.canonical_main_id
LEFT JOIN app.submission_review_intake_shorts AS intake_short
    ON intake_short.submission_review_intake_id = i.id
LEFT JOIN app.shorts AS s
    ON s.id = intake_short.short_id
WHERE i.status = 'pending_review'
GROUP BY
    i.id,
    i.submit_kind,
    i.submitted_at,
    i.canonical_main_id,
    i.creator_user_id,
    u.display_name,
    u.handle,
    u.avatar_url,
    p.bio,
    m.state
ORDER BY i.submitted_at DESC, i.id DESC;

-- name: GetAdminSubmissionReviewCaseSummaryByIntakeID :one
SELECT
    i.id AS intake_id,
    i.status,
    i.submit_kind,
    i.previous_intake_id,
    i.canonical_main_id,
    i.creator_user_id,
    i.main_media_asset_id,
    i.main_price_minor,
    i.ownership_confirmed,
    i.consent_confirmed,
    i.submitted_at,
    u.display_name,
    u.handle,
    u.avatar_url,
    COALESCE(p.bio, '') AS creator_bio
FROM app.submission_review_intakes AS i
JOIN app.user_profiles AS u
    ON u.user_id = i.creator_user_id
LEFT JOIN app.creator_profiles AS p
    ON p.user_id = i.creator_user_id
WHERE i.id = $1
LIMIT 1;

-- name: GetAdminSubmissionReviewMainByIntakeID :one
SELECT
    i.canonical_main_id AS main_id,
    i.main_media_asset_id AS media_asset_id,
    m.state,
    m.review_reason_code,
    m.review_note,
    m.review_decision_source,
    m.review_decisioned_at,
    i.main_price_minor AS price_minor,
    m.currency_code,
    a.duration_ms,
    a.processing_state AS media_processing_state,
    d.target_state AS intake_target_state,
    d.reason_code AS intake_reason_code,
    d.review_note AS intake_review_note,
    d.decision_source AS intake_decision_source,
    d.decisioned_at AS intake_decisioned_at
FROM app.submission_review_intakes AS i
JOIN app.mains AS m
    ON m.id = i.canonical_main_id
JOIN app.media_assets AS a
    ON a.id = i.main_media_asset_id
LEFT JOIN app.submission_review_main_decisions AS d
    ON d.submission_review_intake_id = i.id
    AND d.main_id = m.id
WHERE i.id = $1
LIMIT 1;

-- name: ListAdminSubmissionReviewShortsByIntakeID :many
SELECT
    intake_short.short_id,
    intake_short.media_asset_id AS media_asset_id,
    intake_short.caption AS intake_caption,
    s.state,
    s.review_reason_code,
    s.review_note,
    s.review_decision_source,
    s.review_decisioned_at,
    a.duration_ms,
    a.processing_state AS media_processing_state,
    d.target_state AS intake_target_state,
    d.reason_code AS intake_reason_code,
    d.review_note AS intake_review_note,
    d.decision_source AS intake_decision_source,
    d.decisioned_at AS intake_decisioned_at
FROM app.submission_review_intake_shorts AS intake_short
JOIN app.shorts AS s
    ON s.id = intake_short.short_id
JOIN app.media_assets AS a
    ON a.id = intake_short.media_asset_id
LEFT JOIN app.submission_review_short_decisions AS d
    ON d.submission_review_intake_id = intake_short.submission_review_intake_id
    AND d.short_id = intake_short.short_id
WHERE intake_short.submission_review_intake_id = $1
ORDER BY intake_short.created_at DESC, intake_short.short_id DESC;
