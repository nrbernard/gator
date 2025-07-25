-- name: CreateFeed :one
INSERT INTO feeds (id, name, url, description, user_id)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ? 
)
RETURNING *;

-- name: GetFeeds :many
SELECT f.id, f.name, f.url, f.description, u.name as user_name 
FROM feeds f
JOIN users u ON f.user_id = u.id
ORDER BY f.created_at DESC;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = ?;

-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, f.name as feed_name, u.name as user_name
FROM feed_follows
JOIN feeds f ON feed_follows.feed_id = f.id
JOIN users u ON feed_follows.user_id = u.id
WHERE feed_follows.user_id = ?;

-- name: DeleteFeed :exec
DELETE FROM feeds WHERE id = ?;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows WHERE feed_follows.user_id = ? AND feed_follows.feed_id = (SELECT id FROM feeds WHERE url = ?);

-- name: MarkFeedAsFetched :exec
UPDATE feeds SET last_fetched_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY (last_fetched_at IS NOT NULL), last_fetched_at ASC
LIMIT 1;

-- name: GetFeedsToFetch :many
SELECT * FROM feeds
WHERE last_fetched_at IS NULL
   OR last_fetched_at < ?
ORDER BY (last_fetched_at IS NOT NULL), last_fetched_at ASC;
