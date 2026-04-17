-- name: CreateCreatorCapability :one
INSERT INTO app.creator_capabilities (
    user_id,
    state,
    rejection_reason_code,
    is_resubmit_eligible,
    is_support_review_required,
    self_serve_resubmit_count,
    kyc_provider_case_ref,
    payout_provider_account_ref,
    submitted_at,
    approved_at,
    rejected_at,
    suspended_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(state),
    sqlc.narg(rejection_reason_code),
    sqlc.arg(is_resubmit_eligible),
    sqlc.arg(is_support_review_required),
    sqlc.arg(self_serve_resubmit_count),
    sqlc.narg(kyc_provider_case_ref),
    sqlc.narg(payout_provider_account_ref),
    sqlc.narg(submitted_at),
    sqlc.narg(approved_at),
    sqlc.narg(rejected_at),
    sqlc.narg(suspended_at)
)
RETURNING *;

-- name: GetCreatorCapabilityByUserID :one
SELECT *
FROM app.creator_capabilities
WHERE user_id = $1
LIMIT 1;

-- name: GetCreatorCapabilityByUserIDForUpdate :one
SELECT *
FROM app.creator_capabilities
WHERE user_id = $1
LIMIT 1
FOR UPDATE;

-- name: UpdateCreatorCapabilityState :one
UPDATE app.creator_capabilities
SET
    state = sqlc.arg(state),
    rejection_reason_code = sqlc.narg(rejection_reason_code),
    is_resubmit_eligible = sqlc.arg(is_resubmit_eligible),
    is_support_review_required = sqlc.arg(is_support_review_required),
    self_serve_resubmit_count = sqlc.arg(self_serve_resubmit_count),
    kyc_provider_case_ref = sqlc.narg(kyc_provider_case_ref),
    payout_provider_account_ref = sqlc.narg(payout_provider_account_ref),
    submitted_at = sqlc.narg(submitted_at),
    approved_at = sqlc.narg(approved_at),
    rejected_at = sqlc.narg(rejected_at),
    suspended_at = sqlc.narg(suspended_at),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)
RETURNING *;
