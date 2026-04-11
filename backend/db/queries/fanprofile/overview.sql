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
FROM app.main_unlocks
WHERE user_id = $1;
