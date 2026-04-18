-- name: ListUserPaymentMethodsByUserID :many
SELECT *
FROM app.user_payment_methods
WHERE user_id = sqlc.arg(user_id)
ORDER BY last_used_at DESC, id DESC;

-- name: GetUserPaymentMethodByIDAndUserID :one
SELECT *
FROM app.user_payment_methods
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: UpsertUserPaymentMethod :one
INSERT INTO app.user_payment_methods (
    user_id,
    provider,
    provider_payment_token_ref,
    provider_payment_account_ref,
    brand,
    last4,
    last_used_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(provider),
    sqlc.arg(provider_payment_token_ref),
    sqlc.arg(provider_payment_account_ref),
    sqlc.arg(brand),
    sqlc.arg(last4),
    COALESCE(sqlc.narg(last_used_at)::timestamptz, CURRENT_TIMESTAMP)
)
ON CONFLICT (user_id, provider_payment_account_ref) DO UPDATE
SET provider_payment_token_ref = EXCLUDED.provider_payment_token_ref,
    brand = EXCLUDED.brand,
    last4 = EXCLUDED.last4,
    last_used_at = EXCLUDED.last_used_at,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: TouchUserPaymentMethodLastUsedAt :one
UPDATE app.user_payment_methods
SET last_used_at = COALESCE(sqlc.narg(last_used_at)::timestamptz, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
RETURNING *;
