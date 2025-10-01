package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrbernard/gator/internal/database"
)

func setupTestDB(t *testing.T) *database.Queries {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	createTables := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		name VARCHAR(255) NOT NULL
	);
	
	CREATE TABLE feeds (
		id TEXT PRIMARY KEY,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		name VARCHAR(255) NOT NULL,
		url VARCHAR(255) NOT NULL UNIQUE,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		last_fetched_at TIMESTAMP,
		description TEXT,
		etag TEXT,
		last_modified TEXT
	);
	
	CREATE TABLE posts (
		id TEXT PRIMARY KEY,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		title VARCHAR(255) NOT NULL,
		url VARCHAR(255) NOT NULL UNIQUE,
		description TEXT,
		published_at TIMESTAMP NOT NULL,
		feed_id TEXT NOT NULL REFERENCES feeds(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(createTables); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	return database.New(db)
}

func TestFeedService_ScrapeFeeds_WithConditionalRequests(t *testing.T) {
	queries := setupTestDB(t)

	ctx := context.Background()

	// Create a test user
	userID := uuid.New().String()
	_, err := queries.CreateUser(ctx, database.CreateUserParams{
		ID:   userID,
		Name: "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test feed with existing conditional headers
	feedID := uuid.New().String()
	_, err = queries.CreateFeed(ctx, database.CreateFeedParams{
		ID:          feedID,
		Name:        "Test Feed",
		Url:         "http://example.com/feed.xml",
		UserID:      userID,
		Description: sql.NullString{String: "Test Description", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create feed: %v", err)
	}

	// Set up conditional headers for the feed
	err = queries.UpdateFeedConditionalHeadersNoFetch(ctx, database.UpdateFeedConditionalHeadersNoFetchParams{
		Etag:         sql.NullString{String: `"test-etag"`, Valid: true},
		LastModified: sql.NullString{String: "Wed, 21 Oct 2015 07:28:00 GMT", Valid: true},
		ID:           feedID,
	})
	if err != nil {
		t.Fatalf("Failed to set conditional headers: %v", err)
	}

	// Test that the feed is included in feeds to fetch (new feeds should be fetched)
	cutoff := time.Now().Add(-1 * time.Hour)
	feeds, err := queries.GetFeedsToFetch(ctx, sql.NullTime{Time: cutoff, Valid: true})
	if err != nil {
		t.Fatalf("Failed to get feeds to fetch: %v", err)
	}

	if len(feeds) != 1 {
		t.Fatalf("Expected 1 feed to fetch, got %d", len(feeds))
	}

	if feeds[0].Etag.String != `"test-etag"` {
		t.Errorf("Expected ETag %q, got %q", `"test-etag"`, feeds[0].Etag.String)
	}

	if feeds[0].LastModified.String != "Wed, 21 Oct 2015 07:28:00 GMT" {
		t.Errorf("Expected LastModified %q, got %q", "Wed, 21 Oct 2015 07:28:00 GMT", feeds[0].LastModified.String)
	}
}

func TestFeedService_ScrapeFeeds_PollingFrequency(t *testing.T) {
	queries := setupTestDB(t)

	ctx := context.Background()

	// Create a test user
	userID := uuid.New().String()
	_, err := queries.CreateUser(ctx, database.CreateUserParams{
		ID:   userID,
		Name: "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test feed
	feedID := uuid.New().String()
	_, err = queries.CreateFeed(ctx, database.CreateFeedParams{
		ID:     feedID,
		Name:   "Test Feed",
		Url:    "http://example.com/feed.xml",
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("Failed to create feed: %v", err)
	}

	// Mark feed as fetched recently (within 1 hour)
	err = queries.MarkFeedAsFetched(ctx, feedID)
	if err != nil {
		t.Fatalf("Failed to mark feed as fetched: %v", err)
	}

	// Test that the feed is NOT included in feeds to fetch (due to 1-hour limit)
	cutoff := time.Now().Add(-1 * time.Hour)
	feeds, err := queries.GetFeedsToFetch(ctx, sql.NullTime{Time: cutoff, Valid: true})
	if err != nil {
		t.Fatalf("Failed to get feeds to fetch: %v", err)
	}

	if len(feeds) != 0 {
		t.Errorf("Expected 0 feeds to fetch (recently fetched), got %d", len(feeds))
	}
}
