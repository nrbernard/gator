-- name: CreateFeed :one
INSERT INTO feeds (id, name, url, description, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetFeeds :many
SELECT f.id, f.name, f.url, f.description, u.name as user_name 
FROM feeds f
JOIN users u ON f.user_id = u.id
ORDER BY f.created_at DESC;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT
    inserted_feed_follow.*,
    f.name AS feed_name,
    u.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds f ON inserted_feed_follow.feed_id = f.id
INNER JOIN users u ON inserted_feed_follow.user_id = u.id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, f.name as feed_name, u.name as user_name
FROM feed_follows
JOIN feeds f ON feed_follows.feed_id = f.id
JOIN users u ON feed_follows.user_id = u.id
WHERE feed_follows.user_id = $1;

-- name: DeleteFeed :exec
DELETE FROM feeds WHERE id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows WHERE feed_follows.user_id = $1 AND feed_follows.feed_id = (SELECT id FROM feeds WHERE url = $2);

-- name: MarkFeedAsFetched :exec
UPDATE feeds SET last_fetched_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;