-- name: GetCreatorRegistrationIntakeByUserID :one
SELECT *
FROM app.creator_registration_intakes
WHERE user_id = $1
LIMIT 1;

-- name: UpsertCreatorRegistrationIntake :one
INSERT INTO app.creator_registration_intakes (
    user_id,
    legal_name,
    birth_date,
    payout_recipient_type,
    payout_recipient_name,
    declares_no_prohibited_category,
    accepts_consent_responsibility
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(legal_name),
    sqlc.narg(birth_date),
    sqlc.narg(payout_recipient_type),
    sqlc.arg(payout_recipient_name),
    sqlc.arg(declares_no_prohibited_category),
    sqlc.arg(accepts_consent_responsibility)
)
ON CONFLICT (user_id) DO UPDATE
SET
    legal_name = EXCLUDED.legal_name,
    birth_date = EXCLUDED.birth_date,
    payout_recipient_type = EXCLUDED.payout_recipient_type,
    payout_recipient_name = EXCLUDED.payout_recipient_name,
    declares_no_prohibited_category = EXCLUDED.declares_no_prohibited_category,
    accepts_consent_responsibility = EXCLUDED.accepts_consent_responsibility,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: ListCreatorRegistrationEvidencesByUserID :many
SELECT *
FROM app.creator_registration_evidences
WHERE user_id = $1
ORDER BY kind ASC;

-- name: ListCreatorRegistrationReviewCasesByState :many
SELECT
    c.user_id,
    c.state,
    c.submitted_at,
    c.approved_at,
    c.rejected_at,
    c.suspended_at,
    u.display_name,
    u.handle,
    u.avatar_url,
    COALESCE(p.bio, '') AS creator_bio,
    COALESCE(i.legal_name, '') AS legal_name
FROM app.creator_capabilities AS c
JOIN app.user_profiles AS u
    ON u.user_id = c.user_id
LEFT JOIN app.creator_profiles AS p
    ON p.user_id = c.user_id
LEFT JOIN app.creator_registration_intakes AS i
    ON i.user_id = c.user_id
WHERE c.state = sqlc.arg(state)
ORDER BY
    COALESCE(c.submitted_at, c.approved_at, c.rejected_at, c.suspended_at) DESC,
    c.user_id ASC;

-- name: UpsertCreatorRegistrationEvidence :one
INSERT INTO app.creator_registration_evidences (
    user_id,
    kind,
    file_name,
    mime_type,
    file_size_bytes,
    storage_bucket,
    storage_key,
    uploaded_at
) VALUES (
    sqlc.arg(user_id),
    sqlc.arg(kind),
    sqlc.arg(file_name),
    sqlc.arg(mime_type),
    sqlc.arg(file_size_bytes),
    sqlc.arg(storage_bucket),
    sqlc.arg(storage_key),
    sqlc.arg(uploaded_at)
)
ON CONFLICT (user_id, kind) DO UPDATE
SET
    file_name = EXCLUDED.file_name,
    mime_type = EXCLUDED.mime_type,
    file_size_bytes = EXCLUDED.file_size_bytes,
    storage_bucket = EXCLUDED.storage_bucket,
    storage_key = EXCLUDED.storage_key,
    uploaded_at = EXCLUDED.uploaded_at,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;
