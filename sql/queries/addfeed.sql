-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1, -- id
    $2, -- created_at
    $3, -- updated_at
    $4, -- name
    $5, -- url
    $6  -- user_id
)
RETURNING *;