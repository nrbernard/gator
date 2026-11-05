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
-- The Feeds the user hasn't removed, with how their last fetches went.
SELECT f.id, f.name, f.url, f.description, f.last_fetched_at, f.last_fetch_error, f.last_fetch_error_at, u.name as user_name
FROM feeds f
JOIN users u ON f.user_id = u.id
WHERE f.removed_at IS NULL
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

-- name: RemoveFeed :execrows
-- Keeps the row, so the Feed's kept, archived, and trashed Posts stay.
UPDATE feeds SET removed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND removed_at IS NULL;

-- name: UnremoveFeed :one
-- Re-adding a removed Feed's URL. Clearing the fetch state makes the next fetch
-- request the whole Feed, so its current items enter the Queue as a new Feed's do.
UPDATE feeds
SET removed_at = NULL, last_fetched_at = NULL, etag = NULL, last_modified = NULL,
    last_fetch_error = NULL, last_fetch_error_at = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND removed_at IS NOT NULL
RETURNING *;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows WHERE feed_follows.user_id = ? AND feed_follows.feed_id = (SELECT id FROM feeds WHERE url = ?);

-- name: MarkFeedAsFetched :exec
UPDATE feeds SET last_fetched_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY (last_fetched_at IS NOT NULL), last_fetched_at ASC
LIMIT 1;

-- name: GetFeedsToFetch :many
-- The Feeds due a fetch: neither fetched nor failed since the cutoff, so a
-- failing Feed waits out the same hourly limit as one that fetched.
SELECT * FROM feeds
WHERE removed_at IS NULL
AND (last_fetched_at IS NULL OR datetime(last_fetched_at) < datetime(@cutoff))
AND (last_fetch_error_at IS NULL OR datetime(last_fetch_error_at) < datetime(@cutoff))
ORDER BY (last_fetched_at IS NOT NULL), last_fetched_at ASC;

-- name: UpdateFeedConditionalHeaders :exec
UPDATE feeds 
SET etag = ?, last_modified = ?, last_fetched_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP,
    last_fetch_error = NULL, last_fetch_error_at = NULL
WHERE id = ?;

-- name: RecordFeedFetchError :exec
UPDATE feeds
SET last_fetch_error = ?, last_fetch_error_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
