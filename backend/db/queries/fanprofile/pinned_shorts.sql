-- name: ListFanProfilePinnedShortItems :many
SELECT
    pinned.short_id,
    pinned.pinned_at,
    short.creator_user_id,
    short.canonical_main_id,
    short.media_asset_id,
    short.caption,
    media.duration_ms,
    profile.display_name,
    profile.handle,
    profile.avatar_url,
    profile.bio
FROM app.pinned_shorts AS pinned
JOIN app.public_shorts AS short
    ON short.id = pinned.short_id
JOIN app.media_assets AS media
    ON media.id = short.media_asset_id
JOIN app.public_creator_profiles AS profile
    ON profile.user_id = short.creator_user_id
WHERE pinned.user_id = sqlc.arg(user_id)
AND (
    sqlc.narg(cursor_pinned_at)::timestamptz IS NULL
    OR pinned.pinned_at < sqlc.narg(cursor_pinned_at)::timestamptz
    OR (
        pinned.pinned_at = sqlc.narg(cursor_pinned_at)::timestamptz
        AND pinned.short_id < COALESCE(
            sqlc.narg(cursor_short_id)::uuid,
            'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid
        )
    )
)
ORDER BY pinned.pinned_at DESC, pinned.short_id DESC
LIMIT sqlc.arg(limit_count);
