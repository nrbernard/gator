-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsByUser :many
SELECT * FROM posts WHERE feed_id IN (SELECT feed_id FROM feed_follows WHERE user_id = $1) ORDER BY published_at DESC LIMIT $2;

-- name: SearchPostsByUser :many
SELECT posts.id as id, title, posts.url as url, posts.description as description, published_at, feeds.name as feed_name, feeds.id as feed_id, post_saves.created_at as saved_at, post_reads.created_at as read_at FROM posts
JOIN feeds ON posts.feed_id = feeds.id
LEFT JOIN post_saves ON posts.id = post_saves.post_id AND post_saves.user_id = @user_id
LEFT JOIN post_reads ON posts.id = post_reads.post_id AND post_reads.user_id = @user_id
WHERE feed_id IN (SELECT feed_id FROM feed_follows WHERE feed_follows.user_id = @user_id) 
AND (@search_text::TEXT IS NULL OR @search_text::TEXT = '' OR (posts.title ILIKE '%' || @search_text::TEXT || '%' OR posts.description ILIKE '%' || @search_text::TEXT || '%'))
AND (NOT @filter_by_unread::BOOLEAN OR post_reads.id IS NULL)
AND (NOT @filter_by_saved::BOOLEAN OR post_saves.id IS NOT NULL)
ORDER BY published_at DESC LIMIT @limit_count;
