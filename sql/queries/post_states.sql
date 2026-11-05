-- name: GetPostState :one
-- One row per existing Post; NULL place means the Post is in the user's Queue.
SELECT posts.id AS post_id, post_states.place, post_states.trashed_from, post_states.read_at
FROM posts
LEFT JOIN post_states ON post_states.post_id = posts.id AND post_states.user_id = @user_id
WHERE posts.id = @post_id;

-- name: GetNextQueuePost :one
-- The newest Post from a followed Feed that has no state: the top of the Queue.
SELECT posts.id, posts.title, posts.url, posts.description, posts.published_at, posts.feed_id, feeds.name AS feed_name
FROM posts
JOIN feeds ON feeds.id = posts.feed_id
WHERE posts.feed_id IN (SELECT feed_follows.feed_id FROM feed_follows WHERE feed_follows.user_id = @user_id)
AND NOT EXISTS (SELECT 1 FROM post_states WHERE post_states.post_id = posts.id AND post_states.user_id = @user_id)
ORDER BY posts.published_at DESC
LIMIT 1;

-- name: CountQueuePosts :one
SELECT COUNT(*)
FROM posts
WHERE posts.feed_id IN (SELECT feed_follows.feed_id FROM feed_follows WHERE feed_follows.user_id = @user_id)
AND NOT EXISTS (SELECT 1 FROM post_states WHERE post_states.post_id = posts.id AND post_states.user_id = @user_id);

-- name: GetReadingListPosts :many
-- Unread first, then most recently kept: changed_at is when the Post entered
-- the Reading list.
SELECT posts.id, posts.title, posts.url, posts.description, posts.published_at, posts.feed_id, feeds.name AS feed_name, post_states.read_at
FROM post_states
JOIN posts ON posts.id = post_states.post_id
JOIN feeds ON feeds.id = posts.feed_id
WHERE post_states.user_id = @user_id AND post_states.place = 'reading_list'
ORDER BY post_states.read_at IS NOT NULL, post_states.changed_at DESC;

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
