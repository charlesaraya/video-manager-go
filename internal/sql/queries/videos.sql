-- name: CreateVideo :one
INSERT INTO videos(id, created_at, updated_at, title, description, user_id)
VALUES (
    ?,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    ?,
    ?,
    ?
) RETURNING *;

-- name: GetVideosByUser :many
SELECT * FROM videos WHERE user_id = ?;

-- name: GetVideo :one
SELECT * FROM videos WHERE id = ?;

-- name: DeleteVideo :exec
DELETE FROM videos 
WHERE id = ? AND user_id = ?;

-- name: DeleteAllVideos :exec
DELETE FROM videos;
