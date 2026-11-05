package service

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrbernard/gator/internal/clock"
	"github.com/nrbernard/gator/internal/database"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
)

// schemaDir holds the goose migrations, relative to this package.
const schemaDir = "../../sql/schema"

func setupTestDB(t *testing.T) *database.Queries {
	t.Helper()
	return database.New(setupTestSQLDB(t))
}

// setupTestSQLDB is setupTestDB's database handle, for seeding what the
// queries can't, such as a Post's created_at.
func setupTestSQLDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	// Each connection to :memory: is a separate database; pin to one so
	// migrations and queries see the same schema.
	db.SetMaxOpenConns(1)

	provider, err := goose.NewProvider(goosedb.DialectSQLite3, db, os.DirFS(schemaDir))
	if err != nil {
		t.Fatalf("Failed to create migration provider: %v", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func TestFeedService_ScrapeFeeds_WithConditionalRequests(t *testing.T) {
	db := setupTestSQLDB(t)
	queries := database.New(db)

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

	// Set up conditional headers for the feed, without marking it fetched
	if _, err := db.Exec("UPDATE feeds SET etag = ?, last_modified = ? WHERE id = ?",
		`"test-etag"`, "Wed, 21 Oct 2015 07:28:00 GMT", feedID); err != nil {
		t.Fatalf("Failed to set conditional headers: %v", err)
	}

	// Test that the feed is included in feeds to fetch (new feeds should be fetched)
	cutoff := time.Now().Add(-1 * time.Hour).UTC()
	feeds, err := queries.GetFeedsToFetch(ctx, cutoff)
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

const testRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Working Feed</title>
		<link>https://example.com</link>
		<description>Works</description>
		<item>
			<title>A Post</title>
			<link>https://example.com/a-post</link>
			<description>Hello</description>
			<pubDate>Wed, 01 Jan 2024 12:00:00 GMT</pubDate>
		</item>
	</channel>
</rss>`

func serveRSS(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(testRSS))
	}))
	t.Cleanup(server.Close)
	return server
}

func serveStatus(t *testing.T, status int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	return server
}

// createFollowedFeed creates a user-followed Feed at url and returns its ID.
func createFollowedFeed(t *testing.T, queries *database.Queries, userID, name, url string) string {
	t.Helper()
	ctx := context.Background()
	feedID := uuid.New().String()
	if _, err := queries.CreateFeed(ctx, database.CreateFeedParams{
		ID:     feedID,
		Name:   name,
		Url:    url,
		UserID: userID,
	}); err != nil {
		t.Fatalf("Failed to create feed: %v", err)
	}
	if _, err := queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:     uuid.New().String(),
		UserID: userID,
		FeedID: feedID,
	}); err != nil {
		t.Fatalf("Failed to follow feed: %v", err)
	}
	return feedID
}

func createTestUser(t *testing.T, queries *database.Queries) string {
	t.Helper()
	userID := uuid.New().String()
	if _, err := queries.CreateUser(context.Background(), database.CreateUserParams{
		ID:   userID,
		Name: "Test User",
	}); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	return userID
}

func TestFeedService_ScrapeFeeds_FailingFeedDoesNotStopOthers(t *testing.T) {
	queries := setupTestDB(t)
	ctx := context.Background()
	userID := createTestUser(t, queries)

	failing := serveStatus(t, http.StatusInternalServerError)
	working := serveRSS(t)
	// Created first so it is fetched first, ahead of the working Feed.
	createFollowedFeed(t, queries, userID, "Failing", failing.URL)
	workingID := createFollowedFeed(t, queries, userID, "Working", working.URL)

	before := time.Now().Add(-time.Second)
	svc := &FeedService{Repo: queries, Clock: clock.New()}
	if err := svc.ScrapeFeeds(ctx); err != nil {
		t.Fatalf("ScrapeFeeds returned error: %v", err)
	}

	posts, err := queries.GetPostsByUser(ctx, database.GetPostsByUserParams{UserID: userID, Limit: 10})
	if err != nil {
		t.Fatalf("Failed to get posts: %v", err)
	}
	if len(posts) != 1 || posts[0].FeedID != workingID || posts[0].Title != "A Post" {
		t.Fatalf("Expected the working Feed's one Post, got %+v", posts)
	}

	failed, err := queries.GetFeedByUrl(ctx, failing.URL)
	if err != nil {
		t.Fatalf("Failed to get failing feed: %v", err)
	}
	if !failed.LastFetchError.Valid || !strings.Contains(failed.LastFetchError.String, "500") {
		t.Errorf("Expected the failing Feed's error to mention 500, got %+v", failed.LastFetchError)
	}
	if !failed.LastFetchErrorAt.Valid || failed.LastFetchErrorAt.Time.Before(before) {
		t.Errorf("Expected the failing Feed's error time to be set to now, got %+v", failed.LastFetchErrorAt)
	}

	ok, err := queries.GetFeedByUrl(ctx, working.URL)
	if err != nil {
		t.Fatalf("Failed to get working feed: %v", err)
	}
	if ok.LastFetchError.Valid || ok.LastFetchErrorAt.Valid {
		t.Errorf("Expected the working Feed to have no error, got %+v at %+v", ok.LastFetchError, ok.LastFetchErrorAt)
	}
}

func TestFeedService_ScrapeFeeds_SuccessClearsError(t *testing.T) {
	tests := []struct {
		name   string
		server func(t *testing.T) *httptest.Server
	}{
		{"new content", serveRSS},
		{"not modified", func(t *testing.T) *httptest.Server { return serveStatus(t, http.StatusNotModified) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestSQLDB(t)
			queries := database.New(db)
			ctx := context.Background()
			userID := createTestUser(t, queries)
			server := tt.server(t)
			feedID := createFollowedFeed(t, queries, userID, "Recovered", server.URL)
			// Failed over an hour ago, so the Feed is due another attempt.
			if _, err := db.Exec(
				"UPDATE feeds SET last_fetch_error = 'status code: 500', last_fetch_error_at = datetime('now', '-2 hours') WHERE id = ?",
				feedID,
			); err != nil {
				t.Fatalf("Failed to record fetch error: %v", err)
			}

			svc := &FeedService{Repo: queries, Clock: clock.New()}
			if err := svc.ScrapeFeeds(ctx); err != nil {
				t.Fatalf("ScrapeFeeds returned error: %v", err)
			}

			feed, err := queries.GetFeedByUrl(ctx, server.URL)
			if err != nil {
				t.Fatalf("Failed to get feed: %v", err)
			}
			if feed.LastFetchError.Valid || feed.LastFetchErrorAt.Valid {
				t.Errorf("Expected the error to be cleared, got %+v at %+v", feed.LastFetchError, feed.LastFetchErrorAt)
			}
		})
	}
}

// serveCountingStatus answers every request with status and counts them.
func serveCountingStatus(t *testing.T, status int) (*httptest.Server, *int) {
	t.Helper()
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	return server, &hits
}

func TestFeedService_ScrapeFeeds_FailedFeedWaitsOutTheHour(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantError string
	}{
		{"fetch error", http.StatusInternalServerError, "500"},
		{"rate limited", http.StatusTooManyRequests, "rate limited"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := setupTestDB(t)
			ctx := context.Background()
			userID := createTestUser(t, queries)
			server, hits := serveCountingStatus(t, tt.status)
			createFollowedFeed(t, queries, userID, "Failing", server.URL)

			svc := &FeedService{Repo: queries, Clock: clock.New()}
			for range 2 {
				if err := svc.ScrapeFeeds(ctx); err != nil {
					t.Fatalf("ScrapeFeeds returned error: %v", err)
				}
			}

			if *hits != 1 {
				t.Errorf("Expected a failed Feed to be fetched once within the hour, got %d requests", *hits)
			}
			feed, err := queries.GetFeedByUrl(ctx, server.URL)
			if err != nil {
				t.Fatalf("Failed to get feed: %v", err)
			}
			if !feed.LastFetchError.Valid || !strings.Contains(feed.LastFetchError.String, tt.wantError) || !feed.LastFetchErrorAt.Valid {
				t.Errorf("Expected the error to mention %q, got %+v at %+v", tt.wantError, feed.LastFetchError, feed.LastFetchErrorAt)
			}
			if feed.LastFetchedAt.Valid {
				t.Errorf("Expected a failed fetch not to count as fetched, got %+v", feed.LastFetchedAt)
			}
		})
	}
}

func TestFeedService_ScrapeFeeds_RemovedFeedIsSkipped(t *testing.T) {
	queries := setupTestDB(t)
	ctx := context.Background()
	userID := createTestUser(t, queries)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(testRSS))
	}))
	t.Cleanup(server.Close)
	feedID := createFollowedFeed(t, queries, userID, "Removed", server.URL)

	svc := &FeedService{Repo: queries, Clock: clock.New()}
	if err := svc.RemoveFeed(ctx, uuid.MustParse(feedID)); err != nil {
		t.Fatalf("RemoveFeed: %v", err)
	}
	if err := svc.ScrapeFeeds(ctx); err != nil {
		t.Fatalf("ScrapeFeeds returned error: %v", err)
	}

	if requests != 0 {
		t.Errorf("Expected the removed Feed not to be fetched, got %d requests", requests)
	}
}

func TestFeedService_ScrapeFeeds_NotModifiedCountsAsFetch(t *testing.T) {
	queries := setupTestDB(t)
	ctx := context.Background()
	userID := createTestUser(t, queries)
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNotModified)
	}))
	t.Cleanup(server.Close)
	createFollowedFeed(t, queries, userID, "Unchanged", server.URL)

	svc := &FeedService{Repo: queries, Clock: clock.New()}
	for range 2 {
		if err := svc.ScrapeFeeds(ctx); err != nil {
			t.Fatalf("ScrapeFeeds returned error: %v", err)
		}
	}

	if hits != 1 {
		t.Errorf("Expected a 304 Feed to be fetched once within the hour, got %d requests", hits)
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
	cutoff := time.Now().Add(-1 * time.Hour).UTC()
	feeds, err := queries.GetFeedsToFetch(ctx, cutoff)
	if err != nil {
		t.Fatalf("Failed to get feeds to fetch: %v", err)
	}

	if len(feeds) != 0 {
		t.Errorf("Expected 0 feeds to fetch (recently fetched), got %d", len(feeds))
	}
}
