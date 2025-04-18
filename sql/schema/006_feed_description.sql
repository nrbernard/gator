-- +goose Up
ALTER TABLE feeds ADD COLUMN description TEXT;

-- +goose Down
ALTER TABLE feeds DROP COLUMN description;