-- name: GetFeeds :many
SELECT feeds.*, users.name AS user_name
FROM feeds
JOIN users ON feeds.user_id = users.id;