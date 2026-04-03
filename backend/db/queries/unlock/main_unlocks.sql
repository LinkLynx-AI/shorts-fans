-- name: CreateMainUnlock :one
INSERT INTO app.main_unlocks (
    user_id,
    main_id,
    payment_provider_purchase_ref,
    purchased_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(main_id),
    sqlc.narg(payment_provider_purchase_ref),
    COALESCE(sqlc.narg(purchased_at)::timestamptz, CURRENT_TIMESTAMP)
)
RETURNING *;

-- name: GetMainUnlockByUserIDAndMainID :one
SELECT *
FROM app.main_unlocks
WHERE user_id = sqlc.arg(user_id)
    AND main_id = sqlc.arg(main_id)
LIMIT 1;

-- name: ListUnlockedMainIDsByUserID :many
SELECT main_id
FROM app.main_unlocks
WHERE user_id = $1
ORDER BY purchased_at DESC, created_at DESC, main_id DESC;
