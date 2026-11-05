-- +goose Up
-- When the user removed the Feed. A removed Feed keeps its row, so its Reading
-- list, Archive, and Trash Posts stay; it is not fetched or listed, and its
-- Queue Posts leave the Queue. Additive: the previous release ignores it.
ALTER TABLE feeds ADD COLUMN removed_at TIMESTAMP;

-- +goose Down
ALTER TABLE feeds DROP COLUMN removed_at;
