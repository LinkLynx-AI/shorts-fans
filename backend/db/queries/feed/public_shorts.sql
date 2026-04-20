-- name: ListRecommendedPublicFeedItems :many
SELECT
    s.id,
    s.creator_user_id,
    s.canonical_main_id,
    s.media_asset_id,
    s.caption,
    s.published_at,
    short_media.duration_ms AS short_duration_ms,
    creator_profile.display_name,
    creator_profile.handle,
    creator_profile.avatar_url,
    creator_profile.bio,
    main_record.price_minor AS main_price_minor,
    main_media.duration_ms AS main_duration_ms,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE EXISTS (
            SELECT 1
            FROM app.pinned_shorts AS pinned
            WHERE pinned.user_id = sqlc.narg(viewer_user_id)::uuid
                AND pinned.short_id = s.id
        )
    END AS is_pinned,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE EXISTS (
            SELECT 1
            FROM app.main_unlocks AS main_unlock
            WHERE main_unlock.user_id = sqlc.narg(viewer_user_id)::uuid
                AND main_unlock.main_id = s.canonical_main_id
        )
    END AS is_unlocked,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE sqlc.narg(viewer_user_id)::uuid = s.creator_user_id
    END AS is_owner,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE EXISTS (
            SELECT 1
            FROM app.creator_follows AS creator_follow
            WHERE creator_follow.user_id = sqlc.narg(viewer_user_id)::uuid
                AND creator_follow.creator_user_id = s.creator_user_id
        )
    END AS is_following_creator
FROM app.public_shorts AS s
JOIN app.media_assets AS short_media
    ON short_media.id = s.media_asset_id
JOIN app.creator_profiles AS creator_profile
    ON creator_profile.user_id = s.creator_user_id
JOIN app.unlockable_mains AS main_record
    ON main_record.id = s.canonical_main_id
JOIN app.media_assets AS main_media
    ON main_media.id = main_record.media_asset_id
WHERE (
    sqlc.narg(cursor_published_at)::timestamptz IS NULL
    OR s.published_at < sqlc.narg(cursor_published_at)::timestamptz
    OR (
        s.published_at = sqlc.narg(cursor_published_at)::timestamptz
        AND s.id < COALESCE(sqlc.narg(cursor_short_id)::uuid, 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid)
    )
)
ORDER BY s.published_at DESC, s.id DESC
LIMIT sqlc.arg(limit_count);

-- name: ListFollowingPublicFeedItems :many
WITH ranking_context AS (
    SELECT COALESCE(
        sqlc.narg(ranking_reference_at)::timestamptz,
        DATE_TRUNC('hour', CURRENT_TIMESTAMP)
    ) AS reference_at
),
ranking_weights AS (
    SELECT
        2000000::bigint AS unlocked_main_penalty,
        900000::bigint AS seen_completion_penalty,
        180000::bigint AS rewatch_penalty,
        350000::bigint AS main_unlock_weight,
        90000::bigint AS main_click_weight,
        40000::bigint AS main_completion_weight,
        20000::bigint AS main_rewatch_weight,
        180000::bigint AS creator_unlock_weight,
        50000::bigint AS creator_main_click_weight,
        25000::bigint AS creator_profile_click_weight,
        12000::bigint AS creator_completion_weight,
        6000::bigint AS creator_rewatch_weight,
        7000::bigint AS global_unlock_weight,
        1800::bigint AS global_main_click_weight,
        900::bigint AS global_completion_weight,
        450::bigint AS global_rewatch_weight,
        120::bigint AS freshness_per_hour_weight,
        60000::bigint AS same_creator_penalty,
        120000::bigint AS same_main_penalty
),
candidate_following AS (
    SELECT
        s.id,
        s.creator_user_id,
        s.canonical_main_id,
        s.media_asset_id,
        s.caption,
        s.published_at,
        short_media.duration_ms AS short_duration_ms,
        creator_profile.display_name,
        creator_profile.handle,
        creator_profile.avatar_url,
        creator_profile.bio,
        main_record.price_minor AS main_price_minor,
        main_media.duration_ms AS main_duration_ms,
        pinned.short_id IS NOT NULL AS is_pinned,
        unlocked.main_id IS NOT NULL AS is_unlocked,
        sqlc.arg(viewer_user_id) = s.creator_user_id AS is_owner,
        TRUE AS is_following_creator,
        COALESCE(viewer_short.view_completion_count, 0) AS viewer_short_view_completion_count,
        COALESCE(viewer_short.rewatch_loop_count, 0) AS viewer_short_rewatch_loop_count,
        COALESCE(viewer_creator.unlock_conversion_count, 0) AS viewer_creator_unlock_conversion_count,
        COALESCE(viewer_creator.main_click_count, 0) AS viewer_creator_main_click_count,
        COALESCE(viewer_creator.profile_click_count, 0) AS viewer_creator_profile_click_count,
        COALESCE(viewer_creator.view_completion_count, 0) AS viewer_creator_view_completion_count,
        COALESCE(viewer_creator.rewatch_loop_count, 0) AS viewer_creator_rewatch_loop_count,
        COALESCE(viewer_main.unlock_conversion_count, 0) AS viewer_main_unlock_conversion_count,
        COALESCE(viewer_main.main_click_count, 0) AS viewer_main_main_click_count,
        COALESCE(viewer_main.view_completion_count, 0) AS viewer_main_view_completion_count,
        COALESCE(viewer_main.rewatch_loop_count, 0) AS viewer_main_view_rewatch_loop_count,
        COALESCE(short_global.unlock_conversion_count, 0) AS short_global_unlock_conversion_count,
        COALESCE(short_global.main_click_count, 0) AS short_global_main_click_count,
        COALESCE(short_global.view_completion_count, 0) AS short_global_view_completion_count,
        COALESCE(short_global.rewatch_loop_count, 0) AS short_global_rewatch_loop_count
    FROM app.public_shorts AS s
    CROSS JOIN ranking_context
    JOIN app.media_assets AS short_media
        ON short_media.id = s.media_asset_id
    JOIN app.creator_profiles AS creator_profile
        ON creator_profile.user_id = s.creator_user_id
    JOIN app.unlockable_mains AS main_record
        ON main_record.id = s.canonical_main_id
    JOIN app.media_assets AS main_media
        ON main_media.id = main_record.media_asset_id
    JOIN app.creator_follows AS followed_creator
        ON followed_creator.user_id = sqlc.arg(viewer_user_id)
        AND followed_creator.creator_user_id = s.creator_user_id
    LEFT JOIN app.pinned_shorts AS pinned
        ON pinned.user_id = sqlc.arg(viewer_user_id)
        AND pinned.short_id = s.id
    LEFT JOIN app.main_unlocks AS unlocked
        ON unlocked.user_id = sqlc.arg(viewer_user_id)
        AND unlocked.main_id = s.canonical_main_id
    LEFT JOIN app.recommendation_viewer_short_features AS viewer_short
        ON viewer_short.viewer_user_id = sqlc.arg(viewer_user_id)
        AND viewer_short.short_id = s.id
    LEFT JOIN app.recommendation_viewer_creator_features AS viewer_creator
        ON viewer_creator.viewer_user_id = sqlc.arg(viewer_user_id)
        AND viewer_creator.creator_user_id = s.creator_user_id
    LEFT JOIN app.recommendation_viewer_main_features AS viewer_main
        ON viewer_main.viewer_user_id = sqlc.arg(viewer_user_id)
        AND viewer_main.canonical_main_id = s.canonical_main_id
    LEFT JOIN app.recommendation_short_global_features AS short_global
        ON short_global.short_id = s.id
    WHERE s.published_at <= ranking_context.reference_at
),
scored_following AS (
    SELECT
        candidate_following.*,
        (
            CASE
                WHEN candidate_following.is_unlocked THEN -ranking_weights.unlocked_main_penalty
                ELSE 0
            END
            - (
                LEAST(candidate_following.viewer_short_view_completion_count, 1)
                * ranking_weights.seen_completion_penalty
            )
            - (
                LEAST(candidate_following.viewer_short_rewatch_loop_count, 1)
                * ranking_weights.rewatch_penalty
            )
            + (
                LEAST(candidate_following.viewer_main_unlock_conversion_count, 1)
                * ranking_weights.main_unlock_weight
            )
            + (
                LEAST(candidate_following.viewer_main_main_click_count, 3)
                * ranking_weights.main_click_weight
            )
            + (
                LEAST(candidate_following.viewer_main_view_completion_count, 1)
                * ranking_weights.main_completion_weight
            )
            + (
                LEAST(candidate_following.viewer_main_view_rewatch_loop_count, 3)
                * ranking_weights.main_rewatch_weight
            )
            + (
                LEAST(candidate_following.viewer_creator_unlock_conversion_count, 1)
                * ranking_weights.creator_unlock_weight
            )
            + (
                LEAST(candidate_following.viewer_creator_main_click_count, 3)
                * ranking_weights.creator_main_click_weight
            )
            + (
                LEAST(candidate_following.viewer_creator_profile_click_count, 3)
                * ranking_weights.creator_profile_click_weight
            )
            + (
                LEAST(candidate_following.viewer_creator_view_completion_count, 3)
                * ranking_weights.creator_completion_weight
            )
            + (
                LEAST(candidate_following.viewer_creator_rewatch_loop_count, 3)
                * ranking_weights.creator_rewatch_weight
            )
            + (
                LEAST(candidate_following.short_global_unlock_conversion_count, 10)
                * ranking_weights.global_unlock_weight
            )
            + (
                LEAST(candidate_following.short_global_main_click_count, 10)
                * ranking_weights.global_main_click_weight
            )
            + (
                LEAST(candidate_following.short_global_view_completion_count, 10)
                * ranking_weights.global_completion_weight
            )
            + (
                LEAST(candidate_following.short_global_rewatch_loop_count, 10)
                * ranking_weights.global_rewatch_weight
            )
            + (
                GREATEST(
                    0::bigint,
                    72::bigint
                    - GREATEST(
                        0::bigint,
                        FLOOR(EXTRACT(EPOCH FROM (ranking_context.reference_at - candidate_following.published_at)) / 3600)::bigint
                    )
                )
                * ranking_weights.freshness_per_hour_weight
            )
        ) AS base_rank_score
    FROM candidate_following
    CROSS JOIN ranking_weights
    CROSS JOIN ranking_context
),
diversified_following AS (
    SELECT
        scored_following.*,
        ROW_NUMBER() OVER (
            PARTITION BY scored_following.creator_user_id
            ORDER BY scored_following.base_rank_score DESC, scored_following.published_at DESC, scored_following.id DESC
        ) AS creator_rank_position,
        ROW_NUMBER() OVER (
            PARTITION BY scored_following.canonical_main_id
            ORDER BY scored_following.base_rank_score DESC, scored_following.published_at DESC, scored_following.id DESC
        ) AS main_rank_position
    FROM scored_following
),
ordered_following AS (
    SELECT
        diversified_following.id,
        diversified_following.creator_user_id,
        diversified_following.canonical_main_id,
        diversified_following.media_asset_id,
        diversified_following.caption,
        diversified_following.published_at,
        diversified_following.short_duration_ms,
        diversified_following.display_name,
        diversified_following.handle,
        diversified_following.avatar_url,
        diversified_following.bio,
        diversified_following.main_price_minor,
        diversified_following.main_duration_ms,
        diversified_following.is_pinned,
        diversified_following.is_unlocked,
        diversified_following.is_owner,
        diversified_following.is_following_creator,
        (
            diversified_following.base_rank_score
            - ((diversified_following.creator_rank_position - 1) * ranking_weights.same_creator_penalty)
            - ((diversified_following.main_rank_position - 1) * ranking_weights.same_main_penalty)
        )::bigint AS rank_score
    FROM diversified_following
    CROSS JOIN ranking_weights
)
SELECT
    ordered_following.id,
    ordered_following.creator_user_id,
    ordered_following.canonical_main_id,
    ordered_following.media_asset_id,
    ordered_following.caption,
    ordered_following.published_at,
    ordered_following.short_duration_ms,
    ordered_following.display_name,
    ordered_following.handle,
    ordered_following.avatar_url,
    ordered_following.bio,
    ordered_following.main_price_minor,
    ordered_following.main_duration_ms,
    ordered_following.is_pinned,
    ordered_following.is_unlocked,
    ordered_following.is_owner,
    ordered_following.is_following_creator,
    ordered_following.rank_score
FROM ordered_following
ORDER BY ordered_following.rank_score DESC, ordered_following.published_at DESC, ordered_following.id DESC
;

-- name: ListFeedItemsByShortIDs :many
WITH snapshot_short_ids AS (
    SELECT
        ordered_short_ids.short_id,
        ordered_short_ids.ordinality
    FROM unnest(sqlc.arg(short_ids)::uuid[]) WITH ORDINALITY AS ordered_short_ids(short_id, ordinality)
)
SELECT
    s.id,
    s.creator_user_id,
    s.canonical_main_id,
    s.media_asset_id,
    s.caption,
    s.published_at,
    short_media.duration_ms AS short_duration_ms,
    creator_profile.display_name,
    creator_profile.handle,
    creator_profile.avatar_url,
    creator_profile.bio,
    main_record.price_minor AS main_price_minor,
    main_media.duration_ms AS main_duration_ms,
    EXISTS (
        SELECT 1
        FROM app.pinned_shorts AS pinned
        WHERE pinned.user_id = sqlc.arg(viewer_user_id)
            AND pinned.short_id = s.id
    ) AS is_pinned,
    EXISTS (
        SELECT 1
        FROM app.main_unlocks AS main_unlock
        WHERE main_unlock.user_id = sqlc.arg(viewer_user_id)
            AND main_unlock.main_id = s.canonical_main_id
    ) AS is_unlocked,
    sqlc.arg(viewer_user_id) = s.creator_user_id AS is_owner,
    EXISTS (
        SELECT 1
        FROM app.creator_follows AS creator_follow
        WHERE creator_follow.user_id = sqlc.arg(viewer_user_id)
            AND creator_follow.creator_user_id = s.creator_user_id
    ) AS is_following_creator
FROM snapshot_short_ids
JOIN app.public_shorts AS s
    ON s.id = snapshot_short_ids.short_id
JOIN app.media_assets AS short_media
    ON short_media.id = s.media_asset_id
JOIN app.creator_profiles AS creator_profile
    ON creator_profile.user_id = s.creator_user_id
JOIN app.unlockable_mains AS main_record
    ON main_record.id = s.canonical_main_id
JOIN app.media_assets AS main_media
    ON main_media.id = main_record.media_asset_id
ORDER BY snapshot_short_ids.ordinality ASC;

-- name: GetPublicShortDetailItem :one
SELECT
    s.id,
    s.creator_user_id,
    s.canonical_main_id,
    s.media_asset_id,
    s.caption,
    s.published_at,
    short_media.duration_ms AS short_duration_ms,
    creator_profile.display_name,
    creator_profile.handle,
    creator_profile.avatar_url,
    creator_profile.bio,
    main_record.price_minor AS main_price_minor,
    main_media.duration_ms AS main_duration_ms,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE EXISTS (
            SELECT 1
            FROM app.pinned_shorts AS pinned
            WHERE pinned.user_id = sqlc.narg(viewer_user_id)::uuid
                AND pinned.short_id = s.id
        )
    END AS is_pinned,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE EXISTS (
            SELECT 1
            FROM app.main_unlocks AS main_unlock
            WHERE main_unlock.user_id = sqlc.narg(viewer_user_id)::uuid
                AND main_unlock.main_id = s.canonical_main_id
        )
    END AS is_unlocked,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE sqlc.narg(viewer_user_id)::uuid = s.creator_user_id
    END AS is_owner,
    CASE
        WHEN sqlc.narg(viewer_user_id)::uuid IS NULL THEN FALSE
        ELSE EXISTS (
            SELECT 1
            FROM app.creator_follows AS creator_follow
            WHERE creator_follow.user_id = sqlc.narg(viewer_user_id)::uuid
                AND creator_follow.creator_user_id = s.creator_user_id
        )
    END AS is_following_creator
FROM app.public_shorts AS s
JOIN app.media_assets AS short_media
    ON short_media.id = s.media_asset_id
JOIN app.creator_profiles AS creator_profile
    ON creator_profile.user_id = s.creator_user_id
JOIN app.unlockable_mains AS main_record
    ON main_record.id = s.canonical_main_id
JOIN app.media_assets AS main_media
    ON main_media.id = main_record.media_asset_id
WHERE s.id = sqlc.arg(short_id)
LIMIT 1;
