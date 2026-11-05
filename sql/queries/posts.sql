-- name: CreatePost :one
INSERT INTO posts (id, title, url, description, published_at, feed_id)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPostsByUser :many
SELECT * FROM posts WHERE feed_id IN (SELECT feed_id FROM feed_follows WHERE user_id = @user_id) ORDER BY published_at DESC LIMIT @limit;

-- name: SearchPostsByUser :many
-- Posts in every Place; a NULL place means the Post is in the Queue.
SELECT posts.id as id, title, posts.url as url, posts.description as description, published_at, feeds.name as feed_name, feeds.id as feed_id,
       post_states.place, post_states.read_at
FROM posts
JOIN feeds ON posts.feed_id = feeds.id
LEFT JOIN post_states ON post_states.post_id = posts.id AND post_states.user_id = @user_id
WHERE feed_id IN (SELECT feed_id FROM feed_follows WHERE feed_follows.user_id = @user_id)
AND ( posts.title LIKE '%' || CAST(sqlc.arg('search_text') AS TEXT) || '%'
      OR posts.description LIKE '%' || CAST(sqlc.arg('search_text') AS TEXT) || '%'
    )
ORDER BY published_at DESC LIMIT sqlc.arg('limit_count');