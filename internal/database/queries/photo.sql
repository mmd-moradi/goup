-- name: CreatePhoto :one
INSERT INTO photos (user_id, title, description, s3_key)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPhotoByID :one
SELECT * FROM photos
WHERE id = $1
LIMIT 1;

-- name: ListUserPhotos :many
SELECT * FROM photos
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePhotoDetails :one
UPDATE photos
SET title = $1, description = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: DeletePhoto :exec
DELETE from photos
WHERE id = $1;