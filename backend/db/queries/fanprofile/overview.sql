-- name: CountCreatorFollowsByUserID :one
SELECT COUNT(*)::bigint
FROM app.creator_follows
WHERE user_id = $1;

-- name: CountPinnedShortsByUserID :one
SELECT COUNT(*)::bigint
FROM app.pinned_shorts AS pinned
JOIN app.public_shorts AS short
    ON short.id = pinned.short_id
JOIN app.media_assets AS media
    ON media.id = short.media_asset_id
JOIN app.public_creator_profiles AS profile
    ON profile.user_id = short.creator_user_id
WHERE pinned.user_id = $1;

-- name: CountUnlockedMainsByUserID :one
SELECT COUNT(*)::bigint
FROM app.main_unlocks AS unlock
JOIN app.unlockable_mains AS main
    ON main.id = unlock.main_id
JOIN app.media_assets AS main_media
    ON main_media.id = main.media_asset_id
JOIN app.public_creator_profiles AS profile
    ON profile.user_id = main.creator_user_id
JOIN LATERAL (
    SELECT short.media_asset_id
    FROM app.public_shorts AS short
    WHERE short.canonical_main_id = main.id
    ORDER BY short.published_at DESC, short.created_at DESC, short.id DESC
    LIMIT 1
) AS entry_short
    ON TRUE
JOIN app.media_assets AS entry_short_media
    ON entry_short_media.id = entry_short.media_asset_id
WHERE unlock.user_id = $1;
