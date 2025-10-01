-- +goose Up
ALTER TABLE feeds ADD COLUMN etag TEXT;
ALTER TABLE feeds ADD COLUMN last_modified TEXT;

-- +goose Down
ALTER TABLE feeds DROP COLUMN etag;
ALTER TABLE feeds DROP COLUMN last_modified;
