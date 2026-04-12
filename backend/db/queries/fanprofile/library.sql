-- name: ListFanProfileLibraryItems :many
SELECT
    unlock.main_id,
    unlock.purchased_at,
    unlock.created_at AS unlock_created_at,
    main.creator_user_id,
    main_media.duration_ms AS main_duration_ms,
    profile.display_name,
    profile.handle,
    profile.avatar_url,
    profile.bio,
    entry_short.id AS entry_short_id,
    entry_short.canonical_main_id AS entry_short_canonical_main_id,
    entry_short.caption AS entry_short_caption,
    entry_short.media_asset_id AS entry_short_media_asset_id,
    entry_short_media.duration_ms AS entry_short_duration_ms
FROM app.main_unlocks AS unlock
JOIN app.unlockable_mains AS main
    ON main.id = unlock.main_id
JOIN app.media_assets AS main_media
    ON main_media.id = main.media_asset_id
JOIN app.public_creator_profiles AS profile
    ON profile.user_id = main.creator_user_id
JOIN LATERAL (
    SELECT
        short.id,
        short.canonical_main_id,
        short.caption,
        short.media_asset_id
    FROM app.public_shorts AS short
    WHERE short.canonical_main_id = main.id
    ORDER BY short.published_at DESC, short.created_at DESC, short.id DESC
    LIMIT 1
) AS entry_short
    ON TRUE
JOIN app.media_assets AS entry_short_media
    ON entry_short_media.id = entry_short.media_asset_id
WHERE unlock.user_id = sqlc.arg(user_id)
AND (
    sqlc.narg(cursor_purchased_at)::timestamptz IS NULL
    OR unlock.purchased_at < sqlc.narg(cursor_purchased_at)::timestamptz
    OR (
        unlock.purchased_at = sqlc.narg(cursor_purchased_at)::timestamptz
        AND (
            unlock.created_at < COALESCE(
                sqlc.narg(cursor_unlock_created_at)::timestamptz,
                'infinity'::timestamptz
            )
            OR (
                unlock.created_at = COALESCE(
                    sqlc.narg(cursor_unlock_created_at)::timestamptz,
                    'infinity'::timestamptz
                )
                AND unlock.main_id < COALESCE(
                    sqlc.narg(cursor_main_id)::uuid,
                    'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid
                )
            )
        )
    )
)
ORDER BY unlock.purchased_at DESC, unlock.created_at DESC, unlock.main_id DESC
LIMIT sqlc.arg(limit_count);
