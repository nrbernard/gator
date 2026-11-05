-- name: GetPostState :one
-- One row per existing Post; NULL place means the Post is in the user's Queue.
SELECT posts.id AS post_id, post_states.place, post_states.trashed_from, post_states.read_at
FROM posts
LEFT JOIN post_states ON post_states.post_id = posts.id AND post_states.user_id = @user_id
WHERE posts.id = @post_id;

-- name: SetPostState :exec
INSERT INTO post_states (user_id, post_id, place, trashed_from, read_at, changed_at)
VALUES (@user_id, @post_id, @place, @trashed_from, @read_at, CURRENT_TIMESTAMP)
ON CONFLICT (user_id, post_id) DO UPDATE SET
    place = excluded.place,
    trashed_from = excluded.trashed_from,
    read_at = excluded.read_at,
    changed_at = excluded.changed_at;
