
-- name: CreateUser :one
INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * from users
WHERE id = $1
LIMIT 1;

-- name: GetUserByUserName :one
SELECT * FROM users
WHERE username = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET username = $2,
    email = $3,
    updated_at = $4
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2,
updated_at = $3
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1; 