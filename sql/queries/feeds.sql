-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;