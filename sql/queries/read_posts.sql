-- name: SaveReadPost :exec
INSERT INTO post_reads (id, post_id, user_id) VALUES (@id, @post_id, @user_id);

-- name: DeleteReadPost :exec
DELETE FROM post_reads WHERE user_id = @user_id AND post_id = @post_id;