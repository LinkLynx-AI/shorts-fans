-- name: GetPublicShortForComments :one
SELECT id
FROM app.public_shorts
WHERE public_shorts.id = sqlc.arg(short_id);

-- name: ListShortCommentsFirstPage :many
WITH public_short AS (
    SELECT public_shorts.id
    FROM app.public_shorts
    WHERE public_shorts.id = sqlc.arg(short_id)
),
comment_page AS (
    SELECT
        short_comment.id,
        short_comment.short_id,
        short_comment.author_user_id,
        short_comment.body,
        short_comment.created_at
    FROM app.short_comments AS short_comment
    WHERE short_comment.short_id = sqlc.arg(short_id)
    ORDER BY short_comment.created_at DESC, short_comment.id DESC
    LIMIT sqlc.arg(limit_count)
)
SELECT
    public_short.id AS public_short_id,
    comment_page.id,
    comment_page.short_id,
    comment_page.author_user_id,
    comment_page.body,
    comment_page.created_at,
    user_profile.display_name,
    user_profile.handle,
    user_profile.avatar_url
FROM public_short
LEFT JOIN comment_page
    ON TRUE
LEFT JOIN app.user_profiles AS user_profile
    ON user_profile.user_id = comment_page.author_user_id
ORDER BY comment_page.created_at DESC NULLS LAST, comment_page.id DESC NULLS LAST;

-- name: ListShortCommentsAfterCursor :many
WITH public_short AS (
    SELECT public_shorts.id
    FROM app.public_shorts
    WHERE public_shorts.id = sqlc.arg(short_id)
),
comment_page AS (
    SELECT
        short_comment.id,
        short_comment.short_id,
        short_comment.author_user_id,
        short_comment.body,
        short_comment.created_at
    FROM app.short_comments AS short_comment
    WHERE short_comment.short_id = sqlc.arg(short_id)
        AND (short_comment.created_at, short_comment.id) < (
            sqlc.arg(cursor_created_at)::timestamptz,
            sqlc.arg(cursor_comment_id)::uuid
        )
    ORDER BY short_comment.created_at DESC, short_comment.id DESC
    LIMIT sqlc.arg(limit_count)
)
SELECT
    public_short.id AS public_short_id,
    comment_page.id,
    comment_page.short_id,
    comment_page.author_user_id,
    comment_page.body,
    comment_page.created_at,
    user_profile.display_name,
    user_profile.handle,
    user_profile.avatar_url
FROM public_short
LEFT JOIN comment_page
    ON TRUE
LEFT JOIN app.user_profiles AS user_profile
    ON user_profile.user_id = comment_page.author_user_id
ORDER BY comment_page.created_at DESC NULLS LAST, comment_page.id DESC NULLS LAST;

-- name: CreateShortComment :one
WITH public_short AS (
    SELECT public_shorts.id
    FROM app.public_shorts
    WHERE public_shorts.id = sqlc.arg(short_id)
),
author_profile AS (
    SELECT
        user_id,
        display_name,
        handle,
        avatar_url
    FROM app.user_profiles
    WHERE user_id = sqlc.arg(author_user_id)
),
inserted AS (
    INSERT INTO app.short_comments (
        short_id,
        author_user_id,
        body
    )
    SELECT
        public_short.id,
        author_profile.user_id,
        sqlc.arg(body)
    FROM public_short
    CROSS JOIN author_profile
    RETURNING
        id,
        short_id,
        author_user_id,
        body,
        created_at
)
SELECT
    inserted.id,
    inserted.short_id,
    inserted.author_user_id,
    inserted.body,
    inserted.created_at,
    author_profile.display_name,
    author_profile.handle,
    author_profile.avatar_url
FROM inserted
JOIN author_profile
    ON author_profile.user_id = inserted.author_user_id;
