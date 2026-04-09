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
