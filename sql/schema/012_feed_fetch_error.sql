-- +goose Up
ALTER TABLE feeds ADD COLUMN last_fetch_error TEXT;
ALTER TABLE feeds ADD COLUMN last_fetch_error_at TIMESTAMP;

-- +goose Down
ALTER TABLE feeds DROP COLUMN last_fetch_error_at;
ALTER TABLE feeds DROP COLUMN last_fetch_error;
