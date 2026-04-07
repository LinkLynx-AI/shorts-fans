-- name: ListFanProfileFollowingItems :many
SELECT
    p.user_id AS creator_user_id,
    p.display_name,
    p.handle,
    p.avatar_url,
    p.bio,
    cf.followed_at
FROM app.creator_follows AS cf
JOIN app.public_creator_profiles AS p
    ON p.user_id = cf.creator_user_id
WHERE cf.user_id = sqlc.arg(user_id)
AND (
    sqlc.narg(cursor_followed_at)::timestamptz IS NULL
    OR cf.followed_at < sqlc.narg(cursor_followed_at)::timestamptz
    OR (
        cf.followed_at = sqlc.narg(cursor_followed_at)::timestamptz
        AND cf.creator_user_id > COALESCE(
            sqlc.narg(cursor_creator_user_id)::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid
        )
    )
)
ORDER BY cf.followed_at DESC, cf.creator_user_id ASC
LIMIT sqlc.arg(limit_count);
