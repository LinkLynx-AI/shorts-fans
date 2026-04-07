-- name: CountCreatorFollowsByUserID :one
SELECT COUNT(*)::bigint
FROM app.creator_follows
WHERE user_id = $1;

-- name: CountPinnedShortsByUserID :one
SELECT COUNT(*)::bigint
FROM app.pinned_shorts
WHERE user_id = $1;

-- name: CountUnlockedMainsByUserID :one
SELECT COUNT(*)::bigint
FROM app.main_unlocks
WHERE user_id = $1;
