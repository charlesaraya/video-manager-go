-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, created_at, updated_at, expires_at, revoked_at)
VALUES (
    ?,
    ?,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    ?,
    NULL
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = ?;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE token = ?;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token = ?;

-- name: DeleteTokens :exec
DELETE FROM refresh_tokens;
