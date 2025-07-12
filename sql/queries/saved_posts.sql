-- name: SaveSavedPost :exec
INSERT INTO post_saves (id, post_id, user_id) VALUES ($1, $2, $3);

-- name: DeleteSavedPost :exec
DELETE FROM post_saves WHERE post_id = $1 AND user_id = $2; 
