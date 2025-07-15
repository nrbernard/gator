-- +goose Up
CREATE UNIQUE INDEX unique_post_user ON post_saves (post_id, user_id);

-- +goose Down
DROP INDEX unique_post_user;