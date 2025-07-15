-- name: CreatePost :one
INSERT INTO posts (id, title, url, description, published_at, feed_id)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPostsByUser :many
SELECT * FROM posts WHERE feed_id IN (SELECT feed_id FROM feed_follows WHERE user_id = @user_id) ORDER BY published_at DESC LIMIT @limit;

-- name: SearchPostsByUser :many
SELECT posts.id as id, title, posts.url as url, posts.description as description, published_at, feeds.name as feed_name, feeds.id as feed_id, post_saves.created_at as saved_at, post_reads.created_at as read_at FROM posts
JOIN feeds ON posts.feed_id = feeds.id
LEFT JOIN post_saves ON posts.id = post_saves.post_id AND post_saves.user_id = @user_id
LEFT JOIN post_reads ON posts.id = post_reads.post_id AND post_reads.user_id = @user_id
WHERE feed_id IN (SELECT feed_id FROM feed_follows WHERE feed_follows.user_id = @user_id) 
AND ( CAST(sqlc.arg('search_text') AS TEXT) = ''
      OR posts.title LIKE '%' || CAST(sqlc.arg('search_text') AS TEXT) || '%'
      OR posts.description LIKE '%' || CAST(sqlc.arg('search_text') AS TEXT) || '%'
    )
AND ( CAST(sqlc.arg('filter_by_unread') AS BOOLEAN) = false OR post_reads.id  IS NULL )
AND ( CAST(sqlc.arg('filter_by_saved') AS BOOLEAN)  = false OR post_saves.id IS NOT NULL )
ORDER BY published_at DESC LIMIT sqlc.arg('limit_count');