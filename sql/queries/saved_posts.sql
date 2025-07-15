-- name: SaveSavedPost :exec
INSERT INTO post_saves (id, post_id, user_id) VALUES (?, ?, ?);

-- name: DeleteSavedPost :exec
DELETE FROM post_saves WHERE post_id = ? AND user_id = ?; 
