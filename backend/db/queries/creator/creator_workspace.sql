-- name: GetCreatorWorkspaceOverviewMetrics :one
SELECT
    COALESCE(
        SUM(
            CASE
                WHEN mu.user_id IS NULL THEN 0
                ELSE COALESCE(m.price_minor, 0)
            END
        ),
        0
    )::bigint AS gross_unlock_revenue_jpy,
    COUNT(mu.main_id)::bigint AS unlock_count,
    COUNT(DISTINCT mu.user_id)::bigint AS unique_purchaser_count
FROM app.mains AS m
LEFT JOIN app.main_unlocks AS mu
    ON mu.main_id = m.id
WHERE m.creator_user_id = sqlc.arg(owner_user_id);

-- name: GetCreatorWorkspaceRevisionRequestedSummary :one
SELECT
    (
        SELECT COUNT(*)::bigint
        FROM app.mains AS m
        WHERE m.creator_user_id = sqlc.arg(owner_user_id)
            AND m.state = 'revision_requested'
    ) AS main_count,
    (
        SELECT COUNT(*)::bigint
        FROM app.shorts AS s
        WHERE s.creator_user_id = sqlc.arg(owner_user_id)
            AND s.state = 'revision_requested'
    ) AS short_count;

-- name: ListCreatorWorkspacePreviewMainsByCreatorUserID :many
SELECT
    id,
    creator_user_id,
    media_asset_id,
    state,
    review_reason_code,
    post_report_state,
    COALESCE(price_minor, 0)::bigint AS price_minor,
    COALESCE(currency_code, '')::text AS currency_code,
    ownership_confirmed,
    consent_confirmed,
    approved_for_unlock_at,
    created_at,
    updated_at
FROM app.mains
WHERE creator_user_id = sqlc.arg(owner_user_id)
    AND media_asset_id IS NOT NULL
    AND price_minor IS NOT NULL
    AND currency_code IS NOT NULL
ORDER BY created_at DESC, id DESC;

-- name: UpdateCreatorWorkspaceMainPrice :one
UPDATE app.mains
SET
    price_minor = sqlc.arg(price_minor),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(main_id)
    AND creator_user_id = sqlc.arg(owner_user_id)
    AND media_asset_id IS NOT NULL
    AND price_minor IS NOT NULL
    AND currency_code = 'JPY'
RETURNING
    id,
    COALESCE(price_minor, 0)::bigint AS price_minor,
    COALESCE(currency_code, '')::text AS currency_code;

-- name: ListCreatorWorkspaceTopMainCandidatesByCreatorUserID :many
SELECT
    m.id,
    m.media_asset_id,
    COALESCE(m.price_minor, 0)::bigint AS price_minor,
    COALESCE(m.currency_code, '')::text AS currency_code,
    m.created_at,
    COUNT(mu.user_id)::bigint AS unlock_count
FROM app.mains AS m
LEFT JOIN app.main_unlocks AS mu
    ON mu.main_id = m.id
WHERE m.creator_user_id = sqlc.arg(owner_user_id)
    AND m.media_asset_id IS NOT NULL
    AND m.price_minor IS NOT NULL
    AND m.currency_code IS NOT NULL
GROUP BY
    m.id,
    m.media_asset_id,
    m.price_minor,
    m.currency_code,
    m.created_at
ORDER BY unlock_count DESC, m.created_at DESC, m.id DESC;

-- name: ListCreatorWorkspaceTopShortCandidatesByCreatorUserID :many
SELECT
    s.id,
    s.canonical_main_id,
    s.media_asset_id,
    s.created_at,
    COUNT(mu.user_id)::bigint AS attributed_unlock_count
FROM app.shorts AS s
LEFT JOIN app.main_unlocks AS mu
    ON mu.main_id = s.canonical_main_id
WHERE s.creator_user_id = sqlc.arg(owner_user_id)
    AND s.media_asset_id IS NOT NULL
GROUP BY
    s.id,
    s.canonical_main_id,
    s.media_asset_id,
    s.created_at
ORDER BY attributed_unlock_count DESC, s.created_at DESC, s.id DESC;
