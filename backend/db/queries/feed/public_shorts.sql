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
    END AS is_owner
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
    sqlc.arg(viewer_user_id) = s.creator_user_id AS is_owner
FROM app.public_shorts AS s
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
