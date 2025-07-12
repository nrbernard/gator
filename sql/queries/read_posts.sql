-- name: SaveReadPost :exec
INSERT INTO post_reads (id, post_id, user_id) VALUES ($1, $2, $3);

-- name: DeleteReadPost :exec
DELETE FROM post_reads WHERE user_id = $1 AND post_id = $2;