-- name: GetPostState :one
-- One row per existing Post; NULL place means the Post is in the user's Queue.
SELECT posts.id AS post_id, post_states.place, post_states.trashed_from, post_states.read_at
FROM posts
LEFT JOIN post_states ON post_states.post_id = posts.id AND post_states.user_id = @user_id
WHERE posts.id = @post_id;

-- name: GetNextQueuePost :one
-- The Queue is the Posts with no state from Feeds the user follows and hasn't
-- removed; every Queue query below repeats that feed_id filter.
-- The newest Post from a followed Feed that has no state and that gator
-- fetched at or before the cutoff: the top of the Queue. Later Posts are new,
-- behind the chip. datetime() compares at whole seconds, as trash all does.
SELECT posts.id, posts.title, posts.url, posts.description, posts.published_at, posts.feed_id, feeds.name AS feed_name
FROM posts
JOIN feeds ON feeds.id = posts.feed_id
WHERE posts.feed_id IN (
    SELECT feed_follows.feed_id FROM feed_follows
    JOIN feeds AS followed ON followed.id = feed_follows.feed_id
    WHERE feed_follows.user_id = @user_id AND followed.removed_at IS NULL
)
AND NOT EXISTS (SELECT 1 FROM post_states WHERE post_states.post_id = posts.id AND post_states.user_id = @user_id)
AND datetime(posts.created_at) <= datetime(@cutoff)
ORDER BY posts.published_at DESC
LIMIT 1;

-- name: CountQueuePostsCreatedAfter :one
-- The Queue Posts gator fetched after the cutoff: the new Posts the chip counts.
SELECT COUNT(*)
FROM posts
WHERE posts.feed_id IN (
    SELECT feed_follows.feed_id FROM feed_follows
    JOIN feeds AS followed ON followed.id = feed_follows.feed_id
    WHERE feed_follows.user_id = @user_id AND followed.removed_at IS NULL
)
AND NOT EXISTS (SELECT 1 FROM post_states WHERE post_states.post_id = posts.id AND post_states.user_id = @user_id)
AND datetime(posts.created_at) > datetime(@cutoff);

-- name: GetLastFetchedAt :one
-- When gator last fetched any of the user's Feeds that they haven't removed,
-- for the empty Queue.
SELECT feeds.last_fetched_at
FROM feeds
WHERE feeds.id IN (SELECT feed_follows.feed_id FROM feed_follows WHERE feed_follows.user_id = @user_id)
AND feeds.removed_at IS NULL
AND feeds.last_fetched_at IS NOT NULL
ORDER BY feeds.last_fetched_at DESC
LIMIT 1;

-- name: CountQueuePostsCreatedBy :one
-- The Queue Posts gator fetched at or before the cutoff: what trash all trashes.
-- datetime() compares at whole seconds, CURRENT_TIMESTAMP's resolution.
SELECT COUNT(*)
FROM posts
WHERE posts.feed_id IN (
    SELECT feed_follows.feed_id FROM feed_follows
    JOIN feeds AS followed ON followed.id = feed_follows.feed_id
    WHERE feed_follows.user_id = @user_id AND followed.removed_at IS NULL
)
AND NOT EXISTS (SELECT 1 FROM post_states WHERE post_states.post_id = posts.id AND post_states.user_id = @user_id)
AND datetime(posts.created_at) <= datetime(@cutoff);

-- name: TrashQueuePostsCreatedBy :execrows
-- Trashes, from the Queue, the Posts CountQueuePostsCreatedBy counts.
INSERT INTO post_states (user_id, post_id, place, trashed_from, changed_at)
SELECT @user_id, posts.id, 'trash', 'queue', @changed_at
FROM posts
WHERE posts.feed_id IN (
    SELECT feed_follows.feed_id FROM feed_follows
    JOIN feeds AS followed ON followed.id = feed_follows.feed_id
    WHERE feed_follows.user_id = @user_id AND followed.removed_at IS NULL
)
AND NOT EXISTS (SELECT 1 FROM post_states WHERE post_states.post_id = posts.id AND post_states.user_id = @user_id)
AND datetime(posts.created_at) <= datetime(@cutoff);

-- name: GetPostsInPlace :many
-- The Posts in the Reading list, the Archive, or the Trash, most recently
-- moved there first (changed_at is when the Post entered its Place); on the
-- Reading list, unread first.
SELECT posts.id, posts.title, posts.url, posts.description, posts.published_at, posts.feed_id, feeds.name AS feed_name, post_states.read_at
FROM post_states
JOIN posts ON posts.id = post_states.post_id
JOIN feeds ON feeds.id = posts.feed_id
WHERE post_states.user_id = @user_id AND post_states.place = @place
ORDER BY (post_states.place = 'reading_list' AND post_states.read_at IS NOT NULL), post_states.changed_at DESC;

-- name: RestoreTrashedPostToQueue :execrows
-- A Post with no row is in the Queue. Only a trashed Post is restored, so a
-- stale screen can't send a Post elsewhere back to the Queue.
DELETE FROM post_states
WHERE user_id = @user_id AND post_id = @post_id AND place = 'trash' AND trashed_from = 'queue';

-- name: SetPostState :exec
-- changed_at is when the Post entered its Place, so a change to the Read mark
-- alone keeps it.
INSERT INTO post_states (user_id, post_id, place, trashed_from, read_at, changed_at)
VALUES (@user_id, @post_id, @place, @trashed_from, @read_at, @changed_at)
ON CONFLICT (user_id, post_id) DO UPDATE SET
    place = excluded.place,
    trashed_from = excluded.trashed_from,
    read_at = excluded.read_at,
    changed_at = CASE WHEN post_states.place = excluded.place THEN post_states.changed_at ELSE excluded.changed_at END;
