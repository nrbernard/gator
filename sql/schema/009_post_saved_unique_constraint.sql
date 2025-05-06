-- +goose Up
ALTER TABLE post_saves ADD CONSTRAINT unique_post_user UNIQUE (post_id, user_id);

-- +goose Down
ALTER TABLE post_saves DROP CONSTRAINT unique_post_user;