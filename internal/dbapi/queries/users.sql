-- name: CreateUser :one
INSERT INTO users (
    username,
    email
) VALUES ($1, $2)
RETURNING *;

-- name: CreateUserSession :exec
INSERT INTO users__sessions (
    token_hash,
    user_id,
    expires
) VALUES ($1, $2, $3);

-- name: CreateOidcIdentity :one
INSERT INTO users__oidc_identities (
    user_id,
    issuer,
    subject
) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUser :one
SELECT
    users.*
FROM users
WHERE
    id = $1;

-- name: ListUsers :many
SELECT
    users.*
FROM users
LIMIT $1
OFFSET $2;

-- name: GetUserByOidc :one
SELECT
    sqlc.embed(users),
    sqlc.embed(oidc)
FROM users__oidc_identities oidc
INNER JOIN users ON oidc.user_id = users.id
WHERE
    oidc.issuer = $1
    AND oidc.subject = $2;
