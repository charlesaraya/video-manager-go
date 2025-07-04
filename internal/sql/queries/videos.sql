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

-- name: UpdateVideoThumbnail :one
UPDATE videos
SET thumbnail_url = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING * ;

-- name: UpdateVideoUrl :one
UPDATE videos
SET video_url = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING * ;

-- name: DeleteVideo :exec
DELETE FROM videos 
WHERE id = ? AND user_id = ?;

-- name: DeleteAllVideos :exec
DELETE FROM videos;
