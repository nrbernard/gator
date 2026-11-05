package service

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/clock"
	"github.com/nrbernard/gator/internal/database"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
)

// lastVersionBeforePostStates is the schema the previous release ran against.
const lastVersionBeforePostStates = 10

func TestPostStatesMigration_BackfillsFromSavesAndReads(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)

	provider, err := goose.NewProvider(goosedb.DialectSQLite3, db, os.DirFS(schemaDir))
	if err != nil {
		t.Fatalf("create migration provider: %v", err)
	}
	if _, err := provider.UpTo(ctx, lastVersionBeforePostStates); err != nil {
		t.Fatalf("migrate to %d: %v", lastVersionBeforePostStates, err)
	}

	// Seed with plain SQL: the generated queries track the latest schema.
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	userID, feedID := uuid.New(), uuid.New()
	exec(`INSERT INTO users (id, name) VALUES (?, 'nick')`, userID.String())
	exec(`INSERT INTO feeds (id, name, url, user_id) VALUES (?, 'Feed', 'http://example.com/feed', ?)`, feedID.String(), userID.String())
	exec(`INSERT INTO feed_follows (id, user_id, feed_id) VALUES (?, ?, ?)`, uuid.NewString(), userID.String(), feedID.String())
	post := func() uuid.UUID {
		id := uuid.New()
		exec(`INSERT INTO posts (id, title, url, published_at, feed_id) VALUES (?, 'Post', ?, ?, ?)`,
			id.String(), "http://example.com/"+id.String(), time.Now(), feedID.String())
		return id
	}
	save := func(postID uuid.UUID) {
		exec(`INSERT INTO post_saves (id, post_id, user_id) VALUES (?, ?, ?)`, uuid.NewString(), postID.String(), userID.String())
	}
	read := func(postID uuid.UUID) {
		exec(`INSERT INTO post_reads (id, post_id, user_id) VALUES (?, ?, ?)`, uuid.NewString(), postID.String(), userID.String())
	}

	savedRead := post()
	save(savedRead)
	read(savedRead)
	savedUnread := post()
	save(savedUnread)
	unsavedRead := post()
	read(unsavedRead)
	unsavedUnread := post()

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("apply remaining migrations: %v", err)
	}

	s := &PostStateService{Repo: database.New(db), Clock: clock.New()}
	assertState(t, s, userID, savedRead, PostState{Place: PlaceReadingList, Read: true})
	assertState(t, s, userID, savedUnread, PostState{Place: PlaceReadingList})
	assertState(t, s, userID, unsavedRead, PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue})
	assertState(t, s, userID, unsavedUnread, PostState{Place: PlaceQueue})

	// Additive: the previous release's tables keep their rows.
	for table, want := range map[string]int{"post_saves": 2, "post_reads": 2} {
		var got int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&got); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if got != want {
			t.Errorf("%s has %d rows after migrating, want %d", table, got, want)
		}
	}
}
