-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    ?,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    ?,
    ?
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: DeleteUsers :exec
DELETE FROM users;
