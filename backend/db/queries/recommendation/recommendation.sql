-- name: InsertRecommendationEvent :one
INSERT INTO app.recommendation_events (
    viewer_user_id,
    event_kind,
    creator_user_id,
    canonical_main_id,
    short_id,
    occurred_at,
    idempotency_key
) VALUES (
    sqlc.arg(viewer_user_id),
    sqlc.arg(event_kind),
    sqlc.narg(creator_user_id),
    sqlc.narg(canonical_main_id),
    sqlc.narg(short_id),
    sqlc.arg(occurred_at),
    sqlc.arg(idempotency_key)
)
ON CONFLICT (viewer_user_id, idempotency_key) DO NOTHING
RETURNING
    id,
    viewer_user_id,
    event_kind,
    creator_user_id,
    canonical_main_id,
    short_id,
    occurred_at,
    idempotency_key,
    created_at,
    updated_at;

-- name: GetRecommendationEventByViewerAndIdempotencyKey :one
SELECT
    id,
    viewer_user_id,
    event_kind,
    creator_user_id,
    canonical_main_id,
    short_id,
    occurred_at,
    idempotency_key,
    created_at,
    updated_at
FROM app.recommendation_events
WHERE viewer_user_id = sqlc.arg(viewer_user_id)
    AND idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1;

-- name: UpsertRecommendationViewerShortFeatures :execrows
INSERT INTO app.recommendation_viewer_short_features (
    viewer_user_id,
    short_id,
    creator_user_id,
    canonical_main_id,
    impression_count,
    last_impression_at,
    view_start_count,
    last_view_start_at,
    view_completion_count,
    last_view_completion_at,
    rewatch_loop_count,
    last_rewatch_loop_at,
    main_click_count,
    last_main_click_at,
    unlock_conversion_count,
    last_unlock_conversion_at
) VALUES (
    sqlc.arg(viewer_user_id),
    sqlc.arg(short_id),
    sqlc.arg(creator_user_id),
    sqlc.arg(canonical_main_id),
    sqlc.arg(impression_count),
    sqlc.narg(last_impression_at),
    sqlc.arg(view_start_count),
    sqlc.narg(last_view_start_at),
    sqlc.arg(view_completion_count),
    sqlc.narg(last_view_completion_at),
    sqlc.arg(rewatch_loop_count),
    sqlc.narg(last_rewatch_loop_at),
    sqlc.arg(main_click_count),
    sqlc.narg(last_main_click_at),
    sqlc.arg(unlock_conversion_count),
    sqlc.narg(last_unlock_conversion_at)
)
ON CONFLICT (viewer_user_id, short_id) DO UPDATE
SET
    impression_count = app.recommendation_viewer_short_features.impression_count + EXCLUDED.impression_count,
    last_impression_at = COALESCE(
        GREATEST(app.recommendation_viewer_short_features.last_impression_at, EXCLUDED.last_impression_at),
        app.recommendation_viewer_short_features.last_impression_at,
        EXCLUDED.last_impression_at
    ),
    view_start_count = app.recommendation_viewer_short_features.view_start_count + EXCLUDED.view_start_count,
    last_view_start_at = COALESCE(
        GREATEST(app.recommendation_viewer_short_features.last_view_start_at, EXCLUDED.last_view_start_at),
        app.recommendation_viewer_short_features.last_view_start_at,
        EXCLUDED.last_view_start_at
    ),
    view_completion_count = app.recommendation_viewer_short_features.view_completion_count + EXCLUDED.view_completion_count,
    last_view_completion_at = COALESCE(
        GREATEST(app.recommendation_viewer_short_features.last_view_completion_at, EXCLUDED.last_view_completion_at),
        app.recommendation_viewer_short_features.last_view_completion_at,
        EXCLUDED.last_view_completion_at
    ),
    rewatch_loop_count = app.recommendation_viewer_short_features.rewatch_loop_count + EXCLUDED.rewatch_loop_count,
    last_rewatch_loop_at = COALESCE(
        GREATEST(app.recommendation_viewer_short_features.last_rewatch_loop_at, EXCLUDED.last_rewatch_loop_at),
        app.recommendation_viewer_short_features.last_rewatch_loop_at,
        EXCLUDED.last_rewatch_loop_at
    ),
    main_click_count = app.recommendation_viewer_short_features.main_click_count + EXCLUDED.main_click_count,
    last_main_click_at = COALESCE(
        GREATEST(app.recommendation_viewer_short_features.last_main_click_at, EXCLUDED.last_main_click_at),
        app.recommendation_viewer_short_features.last_main_click_at,
        EXCLUDED.last_main_click_at
    ),
    unlock_conversion_count = app.recommendation_viewer_short_features.unlock_conversion_count + EXCLUDED.unlock_conversion_count,
    last_unlock_conversion_at = COALESCE(
        GREATEST(app.recommendation_viewer_short_features.last_unlock_conversion_at, EXCLUDED.last_unlock_conversion_at),
        app.recommendation_viewer_short_features.last_unlock_conversion_at,
        EXCLUDED.last_unlock_conversion_at
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE app.recommendation_viewer_short_features.creator_user_id = EXCLUDED.creator_user_id
    AND app.recommendation_viewer_short_features.canonical_main_id = EXCLUDED.canonical_main_id;

-- name: UpsertRecommendationViewerCreatorFeatures :exec
INSERT INTO app.recommendation_viewer_creator_features (
    viewer_user_id,
    creator_user_id,
    impression_count,
    last_impression_at,
    view_start_count,
    last_view_start_at,
    view_completion_count,
    last_view_completion_at,
    rewatch_loop_count,
    last_rewatch_loop_at,
    profile_click_count,
    last_profile_click_at,
    main_click_count,
    last_main_click_at,
    unlock_conversion_count,
    last_unlock_conversion_at
) VALUES (
    sqlc.arg(viewer_user_id),
    sqlc.arg(creator_user_id),
    sqlc.arg(impression_count),
    sqlc.narg(last_impression_at),
    sqlc.arg(view_start_count),
    sqlc.narg(last_view_start_at),
    sqlc.arg(view_completion_count),
    sqlc.narg(last_view_completion_at),
    sqlc.arg(rewatch_loop_count),
    sqlc.narg(last_rewatch_loop_at),
    sqlc.arg(profile_click_count),
    sqlc.narg(last_profile_click_at),
    sqlc.arg(main_click_count),
    sqlc.narg(last_main_click_at),
    sqlc.arg(unlock_conversion_count),
    sqlc.narg(last_unlock_conversion_at)
)
ON CONFLICT (viewer_user_id, creator_user_id) DO UPDATE
SET
    impression_count = app.recommendation_viewer_creator_features.impression_count + EXCLUDED.impression_count,
    last_impression_at = COALESCE(
        GREATEST(app.recommendation_viewer_creator_features.last_impression_at, EXCLUDED.last_impression_at),
        app.recommendation_viewer_creator_features.last_impression_at,
        EXCLUDED.last_impression_at
    ),
    view_start_count = app.recommendation_viewer_creator_features.view_start_count + EXCLUDED.view_start_count,
    last_view_start_at = COALESCE(
        GREATEST(app.recommendation_viewer_creator_features.last_view_start_at, EXCLUDED.last_view_start_at),
        app.recommendation_viewer_creator_features.last_view_start_at,
        EXCLUDED.last_view_start_at
    ),
    view_completion_count = app.recommendation_viewer_creator_features.view_completion_count + EXCLUDED.view_completion_count,
    last_view_completion_at = COALESCE(
        GREATEST(app.recommendation_viewer_creator_features.last_view_completion_at, EXCLUDED.last_view_completion_at),
        app.recommendation_viewer_creator_features.last_view_completion_at,
        EXCLUDED.last_view_completion_at
    ),
    rewatch_loop_count = app.recommendation_viewer_creator_features.rewatch_loop_count + EXCLUDED.rewatch_loop_count,
    last_rewatch_loop_at = COALESCE(
        GREATEST(app.recommendation_viewer_creator_features.last_rewatch_loop_at, EXCLUDED.last_rewatch_loop_at),
        app.recommendation_viewer_creator_features.last_rewatch_loop_at,
        EXCLUDED.last_rewatch_loop_at
    ),
    profile_click_count = app.recommendation_viewer_creator_features.profile_click_count + EXCLUDED.profile_click_count,
    last_profile_click_at = COALESCE(
        GREATEST(app.recommendation_viewer_creator_features.last_profile_click_at, EXCLUDED.last_profile_click_at),
        app.recommendation_viewer_creator_features.last_profile_click_at,
        EXCLUDED.last_profile_click_at
    ),
    main_click_count = app.recommendation_viewer_creator_features.main_click_count + EXCLUDED.main_click_count,
    last_main_click_at = COALESCE(
        GREATEST(app.recommendation_viewer_creator_features.last_main_click_at, EXCLUDED.last_main_click_at),
        app.recommendation_viewer_creator_features.last_main_click_at,
        EXCLUDED.last_main_click_at
    ),
    unlock_conversion_count = app.recommendation_viewer_creator_features.unlock_conversion_count + EXCLUDED.unlock_conversion_count,
    last_unlock_conversion_at = COALESCE(
        GREATEST(app.recommendation_viewer_creator_features.last_unlock_conversion_at, EXCLUDED.last_unlock_conversion_at),
        app.recommendation_viewer_creator_features.last_unlock_conversion_at,
        EXCLUDED.last_unlock_conversion_at
    ),
    updated_at = CURRENT_TIMESTAMP;

-- name: UpsertRecommendationViewerMainFeatures :execrows
INSERT INTO app.recommendation_viewer_main_features (
    viewer_user_id,
    canonical_main_id,
    creator_user_id,
    impression_count,
    last_impression_at,
    view_start_count,
    last_view_start_at,
    view_completion_count,
    last_view_completion_at,
    rewatch_loop_count,
    last_rewatch_loop_at,
    main_click_count,
    last_main_click_at,
    unlock_conversion_count,
    last_unlock_conversion_at
) VALUES (
    sqlc.arg(viewer_user_id),
    sqlc.arg(canonical_main_id),
    sqlc.arg(creator_user_id),
    sqlc.arg(impression_count),
    sqlc.narg(last_impression_at),
    sqlc.arg(view_start_count),
    sqlc.narg(last_view_start_at),
    sqlc.arg(view_completion_count),
    sqlc.narg(last_view_completion_at),
    sqlc.arg(rewatch_loop_count),
    sqlc.narg(last_rewatch_loop_at),
    sqlc.arg(main_click_count),
    sqlc.narg(last_main_click_at),
    sqlc.arg(unlock_conversion_count),
    sqlc.narg(last_unlock_conversion_at)
)
ON CONFLICT (viewer_user_id, canonical_main_id) DO UPDATE
SET
    impression_count = app.recommendation_viewer_main_features.impression_count + EXCLUDED.impression_count,
    last_impression_at = COALESCE(
        GREATEST(app.recommendation_viewer_main_features.last_impression_at, EXCLUDED.last_impression_at),
        app.recommendation_viewer_main_features.last_impression_at,
        EXCLUDED.last_impression_at
    ),
    view_start_count = app.recommendation_viewer_main_features.view_start_count + EXCLUDED.view_start_count,
    last_view_start_at = COALESCE(
        GREATEST(app.recommendation_viewer_main_features.last_view_start_at, EXCLUDED.last_view_start_at),
        app.recommendation_viewer_main_features.last_view_start_at,
        EXCLUDED.last_view_start_at
    ),
    view_completion_count = app.recommendation_viewer_main_features.view_completion_count + EXCLUDED.view_completion_count,
    last_view_completion_at = COALESCE(
        GREATEST(app.recommendation_viewer_main_features.last_view_completion_at, EXCLUDED.last_view_completion_at),
        app.recommendation_viewer_main_features.last_view_completion_at,
        EXCLUDED.last_view_completion_at
    ),
    rewatch_loop_count = app.recommendation_viewer_main_features.rewatch_loop_count + EXCLUDED.rewatch_loop_count,
    last_rewatch_loop_at = COALESCE(
        GREATEST(app.recommendation_viewer_main_features.last_rewatch_loop_at, EXCLUDED.last_rewatch_loop_at),
        app.recommendation_viewer_main_features.last_rewatch_loop_at,
        EXCLUDED.last_rewatch_loop_at
    ),
    main_click_count = app.recommendation_viewer_main_features.main_click_count + EXCLUDED.main_click_count,
    last_main_click_at = COALESCE(
        GREATEST(app.recommendation_viewer_main_features.last_main_click_at, EXCLUDED.last_main_click_at),
        app.recommendation_viewer_main_features.last_main_click_at,
        EXCLUDED.last_main_click_at
    ),
    unlock_conversion_count = app.recommendation_viewer_main_features.unlock_conversion_count + EXCLUDED.unlock_conversion_count,
    last_unlock_conversion_at = COALESCE(
        GREATEST(app.recommendation_viewer_main_features.last_unlock_conversion_at, EXCLUDED.last_unlock_conversion_at),
        app.recommendation_viewer_main_features.last_unlock_conversion_at,
        EXCLUDED.last_unlock_conversion_at
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE app.recommendation_viewer_main_features.creator_user_id = EXCLUDED.creator_user_id;

-- name: UpsertRecommendationShortGlobalFeatures :execrows
INSERT INTO app.recommendation_short_global_features (
    short_id,
    creator_user_id,
    canonical_main_id,
    impression_count,
    last_impression_at,
    view_start_count,
    last_view_start_at,
    view_completion_count,
    last_view_completion_at,
    rewatch_loop_count,
    last_rewatch_loop_at,
    main_click_count,
    last_main_click_at,
    unlock_conversion_count,
    last_unlock_conversion_at
) VALUES (
    sqlc.arg(short_id),
    sqlc.arg(creator_user_id),
    sqlc.arg(canonical_main_id),
    sqlc.arg(impression_count),
    sqlc.narg(last_impression_at),
    sqlc.arg(view_start_count),
    sqlc.narg(last_view_start_at),
    sqlc.arg(view_completion_count),
    sqlc.narg(last_view_completion_at),
    sqlc.arg(rewatch_loop_count),
    sqlc.narg(last_rewatch_loop_at),
    sqlc.arg(main_click_count),
    sqlc.narg(last_main_click_at),
    sqlc.arg(unlock_conversion_count),
    sqlc.narg(last_unlock_conversion_at)
)
ON CONFLICT (short_id) DO UPDATE
SET
    impression_count = app.recommendation_short_global_features.impression_count + EXCLUDED.impression_count,
    last_impression_at = COALESCE(
        GREATEST(app.recommendation_short_global_features.last_impression_at, EXCLUDED.last_impression_at),
        app.recommendation_short_global_features.last_impression_at,
        EXCLUDED.last_impression_at
    ),
    view_start_count = app.recommendation_short_global_features.view_start_count + EXCLUDED.view_start_count,
    last_view_start_at = COALESCE(
        GREATEST(app.recommendation_short_global_features.last_view_start_at, EXCLUDED.last_view_start_at),
        app.recommendation_short_global_features.last_view_start_at,
        EXCLUDED.last_view_start_at
    ),
    view_completion_count = app.recommendation_short_global_features.view_completion_count + EXCLUDED.view_completion_count,
    last_view_completion_at = COALESCE(
        GREATEST(app.recommendation_short_global_features.last_view_completion_at, EXCLUDED.last_view_completion_at),
        app.recommendation_short_global_features.last_view_completion_at,
        EXCLUDED.last_view_completion_at
    ),
    rewatch_loop_count = app.recommendation_short_global_features.rewatch_loop_count + EXCLUDED.rewatch_loop_count,
    last_rewatch_loop_at = COALESCE(
        GREATEST(app.recommendation_short_global_features.last_rewatch_loop_at, EXCLUDED.last_rewatch_loop_at),
        app.recommendation_short_global_features.last_rewatch_loop_at,
        EXCLUDED.last_rewatch_loop_at
    ),
    main_click_count = app.recommendation_short_global_features.main_click_count + EXCLUDED.main_click_count,
    last_main_click_at = COALESCE(
        GREATEST(app.recommendation_short_global_features.last_main_click_at, EXCLUDED.last_main_click_at),
        app.recommendation_short_global_features.last_main_click_at,
        EXCLUDED.last_main_click_at
    ),
    unlock_conversion_count = app.recommendation_short_global_features.unlock_conversion_count + EXCLUDED.unlock_conversion_count,
    last_unlock_conversion_at = COALESCE(
        GREATEST(app.recommendation_short_global_features.last_unlock_conversion_at, EXCLUDED.last_unlock_conversion_at),
        app.recommendation_short_global_features.last_unlock_conversion_at,
        EXCLUDED.last_unlock_conversion_at
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE app.recommendation_short_global_features.creator_user_id = EXCLUDED.creator_user_id
    AND app.recommendation_short_global_features.canonical_main_id = EXCLUDED.canonical_main_id;

-- name: ListRecommendationViewerShortFeaturesByViewerAndShortIDs :many
SELECT *
FROM app.recommendation_viewer_short_features
WHERE viewer_user_id = sqlc.arg(viewer_user_id)
    AND short_id = ANY(sqlc.arg(short_ids)::uuid[])
ORDER BY short_id ASC;

-- name: ListRecommendationViewerCreatorFeaturesByViewerAndCreatorIDs :many
SELECT *
FROM app.recommendation_viewer_creator_features
WHERE viewer_user_id = sqlc.arg(viewer_user_id)
    AND creator_user_id = ANY(sqlc.arg(creator_user_ids)::uuid[])
ORDER BY creator_user_id ASC;

-- name: ListRecommendationViewerMainFeaturesByViewerAndMainIDs :many
SELECT *
FROM app.recommendation_viewer_main_features
WHERE viewer_user_id = sqlc.arg(viewer_user_id)
    AND canonical_main_id = ANY(sqlc.arg(canonical_main_ids)::uuid[])
ORDER BY canonical_main_id ASC;

-- name: ListRecommendationShortGlobalFeaturesByShortIDs :many
SELECT *
FROM app.recommendation_short_global_features
WHERE short_id = ANY(sqlc.arg(short_ids)::uuid[])
ORDER BY short_id ASC;

-- name: ListRecommendationPinnedShortIDsByViewerAndShortIDs :many
SELECT short_id
FROM app.pinned_shorts
WHERE user_id = sqlc.arg(viewer_user_id)
    AND short_id = ANY(sqlc.arg(short_ids)::uuid[])
ORDER BY short_id ASC;

-- name: ListRecommendationFollowedCreatorIDsByViewerAndCreatorIDs :many
SELECT creator_user_id
FROM app.creator_follows
WHERE user_id = sqlc.arg(viewer_user_id)
    AND creator_user_id = ANY(sqlc.arg(creator_user_ids)::uuid[])
ORDER BY creator_user_id ASC;

-- name: ListRecommendationUnlockedMainIDsByViewerAndMainIDs :many
SELECT main_id
FROM app.main_unlocks
WHERE user_id = sqlc.arg(viewer_user_id)
    AND main_id = ANY(sqlc.arg(canonical_main_ids)::uuid[])
ORDER BY main_id ASC;
