-- name: ListShortsByCanonicalMainID :many
SELECT *
FROM app.shorts
WHERE canonical_main_id = $1
ORDER BY created_at DESC, id DESC;

-- name: GetCanonicalMainIDByShortID :one
SELECT canonical_main_id
FROM app.shorts
WHERE id = $1
LIMIT 1;
