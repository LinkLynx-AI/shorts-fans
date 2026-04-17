-- name: CreateMainPurchaseAttempt :one
INSERT INTO app.main_purchase_attempts (
    user_id,
    main_id,
    from_short_id,
    provider,
    payment_method_mode,
    user_payment_method_id,
    provider_payment_token_ref,
    idempotency_key,
    status,
    requested_price_jpy,
    requested_currency_code,
    accepted_age,
    accepted_terms
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(main_id),
    sqlc.arg(from_short_id),
    sqlc.arg(provider),
    sqlc.arg(payment_method_mode),
    sqlc.narg(user_payment_method_id),
    sqlc.arg(provider_payment_token_ref),
    sqlc.arg(idempotency_key),
    sqlc.arg(status),
    sqlc.arg(requested_price_jpy),
    sqlc.arg(requested_currency_code),
    sqlc.arg(accepted_age),
    sqlc.arg(accepted_terms)
)
RETURNING *;

-- name: AcquireMainPurchaseLock :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(sqlc.arg(user_key), 0),
    hashtextextended(sqlc.arg(main_key), 0)
);

-- name: GetMainPurchaseAttemptByID :one
SELECT *
FROM app.main_purchase_attempts
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetMainPurchaseAttemptByIDForUpdate :one
SELECT *
FROM app.main_purchase_attempts
WHERE id = sqlc.arg(id)
LIMIT 1
FOR UPDATE;

-- name: GetMainPurchaseAttemptByIdempotencyKeyForUpdate :one
SELECT *
FROM app.main_purchase_attempts
WHERE idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1
FOR UPDATE;

-- name: GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate :one
SELECT *
FROM app.main_purchase_attempts
WHERE user_id = sqlc.arg(user_id)
    AND main_id = sqlc.arg(main_id)
    AND status IN ('processing', 'pending')
ORDER BY created_at DESC, id DESC
LIMIT 1
FOR UPDATE;

-- name: GetLatestSucceededMainPurchaseAttemptByUserIDAndMainIDForUpdate :one
SELECT *
FROM app.main_purchase_attempts
WHERE user_id = sqlc.arg(user_id)
    AND main_id = sqlc.arg(main_id)
    AND status = 'succeeded'
ORDER BY provider_processed_at DESC NULLS LAST, created_at DESC, id DESC
LIMIT 1
FOR UPDATE;

-- name: GetMainPurchaseAttemptByProviderPurchaseRefForUpdate :one
SELECT *
FROM app.main_purchase_attempts
WHERE provider_purchase_ref = sqlc.arg(provider_purchase_ref)
LIMIT 1
FOR UPDATE;

-- name: GetMainPurchaseAttemptByProviderTransactionRefForUpdate :one
SELECT *
FROM app.main_purchase_attempts
WHERE provider_transaction_ref = sqlc.arg(provider_transaction_ref)
LIMIT 1
FOR UPDATE;

-- name: UpdateMainPurchaseAttemptOutcome :one
UPDATE app.main_purchase_attempts
SET status = sqlc.arg(status),
    failure_reason = sqlc.narg(failure_reason),
    pending_reason = sqlc.narg(pending_reason),
    provider_payment_token_ref = COALESCE(sqlc.narg(provider_payment_token_ref), provider_payment_token_ref),
    provider_purchase_ref = COALESCE(sqlc.narg(provider_purchase_ref), provider_purchase_ref),
    provider_transaction_ref = COALESCE(sqlc.narg(provider_transaction_ref), provider_transaction_ref),
    provider_session_ref = COALESCE(sqlc.narg(provider_session_ref), provider_session_ref),
    provider_payment_unique_ref = COALESCE(sqlc.narg(provider_payment_unique_ref), provider_payment_unique_ref),
    provider_decline_code = COALESCE(sqlc.narg(provider_decline_code), provider_decline_code),
    provider_decline_text = COALESCE(sqlc.narg(provider_decline_text), provider_decline_text),
    provider_processed_at = COALESCE(sqlc.narg(provider_processed_at)::timestamptz, provider_processed_at),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;
