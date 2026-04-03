-- name: ListRecommendedFeedItems :many
SELECT
    s.id AS short_id,
    s.canonical_main_id,
    s.creator_user_id,
    s.title AS short_title,
    s.caption AS short_caption,
    s.media_asset_id AS short_media_asset_id,
    s.published_at AS short_published_at,
    short_media.playback_url AS short_playback_url,
    short_media.duration_ms AS short_duration_ms,
    c.display_name AS creator_display_name,
    c.handle AS creator_handle,
    c.avatar_url AS creator_avatar_url,
    c.bio AS creator_bio,
    m.id AS main_id,
    m.price_minor AS main_price_minor,
    main_media.duration_ms AS main_duration_ms
FROM app.public_shorts AS s
JOIN app.media_assets AS short_media
    ON short_media.id = s.media_asset_id
    AND short_media.creator_user_id = s.creator_user_id
    AND short_media.processing_state = 'ready'
JOIN app.public_creator_profiles AS c
    ON c.user_id = s.creator_user_id
JOIN app.unlockable_mains AS m
    ON m.id = s.canonical_main_id
    AND m.creator_user_id = s.creator_user_id
JOIN app.media_assets AS main_media
    ON main_media.id = m.media_asset_id
    AND main_media.creator_user_id = m.creator_user_id
    AND main_media.processing_state = 'ready'
WHERE (
    sqlc.narg(cursor_published_at)::timestamptz IS NULL
    OR s.published_at < sqlc.narg(cursor_published_at)::timestamptz
    OR (
        s.published_at = sqlc.narg(cursor_published_at)::timestamptz
        AND s.id < sqlc.narg(cursor_short_id)::uuid
    )
)
    AND m.price_minor IS NOT NULL
    AND m.currency_code = 'JPY'
ORDER BY s.published_at DESC, s.id DESC
LIMIT sqlc.arg(page_limit);
