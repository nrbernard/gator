package service

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
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

	f := newFixture(t, database.New(db))
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	save := func(postID uuid.UUID) {
		exec(`INSERT INTO post_saves (id, post_id, user_id) VALUES (?, ?, ?)`, uuid.NewString(), postID.String(), f.userID.String())
	}
	read := func(postID uuid.UUID) {
		exec(`INSERT INTO post_reads (id, post_id, user_id) VALUES (?, ?, ?)`, uuid.NewString(), postID.String(), f.userID.String())
	}

	savedRead := f.post(t, "saved and read", time.Now())
	save(savedRead)
	read(savedRead)
	savedUnread := f.post(t, "saved, unread", time.Now())
	save(savedUnread)
	unsavedRead := f.post(t, "read, not saved", time.Now())
	read(unsavedRead)
	unsavedUnread := f.post(t, "neither", time.Now())

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("apply remaining migrations: %v", err)
	}

	s := &PostStateService{Repo: f.queries}
	assertState(t, s, f.userID, savedRead, PostState{Place: PlaceReadingList, Read: true})
	assertState(t, s, f.userID, savedUnread, PostState{Place: PlaceReadingList})
	assertState(t, s, f.userID, unsavedRead, PostState{Place: PlaceTrash, TrashedFrom: PlaceQueue})
	assertState(t, s, f.userID, unsavedUnread, PostState{Place: PlaceQueue})

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
