-- name: CreatePhoto :one
INSERT INTO photos (id, user_id, title, description, file_name, file_size, content_type, storage_path, public_URL, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetPhotoByID :one
SELECT * FROM photos
WHERE id = $1
LIMIT 1;

-- name: ListPhotosByUserID :many
SELECT * FROM photos
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPhotosByUserID :one
SELECT COUNT(*) FROM photos
WHERE user_id = $1;

-- name: UpdatePhoto :one
UPDATE photos
SET title = $2,
    description = $3,
    updated_at = $4
WHERE id = $1
RETURNING *;

-- name: UpdatePhotoStorageInfo :one
UPDATE photos
SET storage_path = $2,
    public_URL = $3,
    updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeletePhoto :exec
DELETE from photos
WHERE id = $1;