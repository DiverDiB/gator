-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1, -- id
    $2, -- created_at
    $3, -- updated_at
    $4  -- name
)
RETURNING *;