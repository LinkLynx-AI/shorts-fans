-- name: CreateUser :one
INSERT INTO app.users DEFAULT VALUES
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM app.users
WHERE id = $1
LIMIT 1;
