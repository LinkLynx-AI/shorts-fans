-- name: CreateMain :one
INSERT INTO app.mains (
    creator_user_id,
    media_asset_id,
    state,
    review_reason_code,
    post_report_state,
    price_minor,
    currency_code,
    ownership_confirmed,
    consent_confirmed,
    approved_for_unlock_at
) VALUES (
    sqlc.arg(creator_user_id),
    sqlc.arg(media_asset_id),
    sqlc.arg(state),
    sqlc.narg(review_reason_code),
    sqlc.narg(post_report_state),
    sqlc.narg(price_minor),
    sqlc.narg(currency_code),
    sqlc.arg(ownership_confirmed),
    sqlc.arg(consent_confirmed),
    sqlc.narg(approved_for_unlock_at)
)
RETURNING *;

-- name: GetMainByID :one
SELECT *
FROM app.mains
WHERE id = $1
LIMIT 1;

-- name: GetMainByMediaAssetID :one
SELECT *
FROM app.mains
WHERE media_asset_id = $1
LIMIT 1;

-- name: ListMainsByCreatorUserID :many
SELECT *
FROM app.mains
WHERE creator_user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: UpdateMainState :one
UPDATE app.mains
SET
    state = sqlc.arg(state),
    review_reason_code = sqlc.narg(review_reason_code),
    post_report_state = sqlc.narg(post_report_state),
    price_minor = sqlc.narg(price_minor),
    currency_code = sqlc.narg(currency_code),
    ownership_confirmed = sqlc.arg(ownership_confirmed),
    consent_confirmed = sqlc.arg(consent_confirmed),
    approved_for_unlock_at = sqlc.narg(approved_for_unlock_at),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetUnlockableMainByID :one
SELECT *
FROM app.unlockable_mains
WHERE id = $1
LIMIT 1;
