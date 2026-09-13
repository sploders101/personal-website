-- name: GetUser :one
SELECT
    id,
    username,
    email,
    created_at
FROM users
WHERE
    id = $1;

-- name: ListUsers :many
SELECT
    id,
    username,
    email,
    created_at
FROM users
LIMIT $1
OFFSET $2;
