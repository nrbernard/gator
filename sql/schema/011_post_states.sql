-- +goose Up
-- Where each Post is for a user (docs/adr/0001-exclusive-post-states.md).
-- A Post with no row is in the user's Queue. Additive: post_saves and
-- post_reads stay for the previous release's code until a later migration.
CREATE TABLE post_states (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id TEXT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    place TEXT NOT NULL CHECK (place IN ('reading_list', 'archive', 'trash')),
    -- Where a trashed Post came from, so a restore returns it there.
    trashed_from TEXT CHECK (trashed_from IN ('queue', 'reading_list', 'archive')),
    read_at TIMESTAMP,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, post_id),
    CHECK ((place = 'trash') = (trashed_from IS NOT NULL))
);

-- Saved Posts go to the Reading list with their Read mark.
INSERT INTO post_states (user_id, post_id, place, read_at, changed_at)
SELECT post_saves.user_id, post_saves.post_id, 'reading_list', post_reads.created_at, post_saves.created_at
FROM post_saves
LEFT JOIN post_reads ON post_reads.post_id = post_saves.post_id AND post_reads.user_id = post_saves.user_id;

-- Unsaved, read Posts go to the Trash, from the Queue: old "read" marks are not
-- evidence the user read them (scripts/mark_read.go), so not the Archive.
INSERT INTO post_states (user_id, post_id, place, trashed_from, changed_at)
SELECT post_reads.user_id, post_reads.post_id, 'trash', 'queue', post_reads.created_at
FROM post_reads
WHERE NOT EXISTS (
    SELECT 1 FROM post_saves
    WHERE post_saves.post_id = post_reads.post_id AND post_saves.user_id = post_reads.user_id
);

-- +goose Down
DROP TABLE post_states;
