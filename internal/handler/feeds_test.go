package handler

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/service"
)

// followedFeed adds a second Feed the user follows and returns its ID.
func (a *testApp) followedFeed(t *testing.T, name string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	id := uuid.New()
	if _, err := a.queries.CreateFeed(ctx, database.CreateFeedParams{
		ID: id.String(), Name: name, Url: "http://example.com/" + id.String() + ".xml", UserID: a.userID.String(),
	}); err != nil {
		t.Fatalf("create feed: %v", err)
	}
	if _, err := a.queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID: uuid.NewString(), UserID: a.userID.String(), FeedID: id.String(),
	}); err != nil {
		t.Fatalf("follow feed: %v", err)
	}
	return id
}

// postIn adds a Post to the Feed feedID.
func (a *testApp) postIn(t *testing.T, feedID uuid.UUID, title string, publishedAt time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := a.queries.CreatePost(context.Background(), database.CreatePostParams{
		ID:          id.String(),
		Title:       title,
		Url:         "http://example.com/posts/" + id.String(),
		Description: sql.NullString{},
		PublishedAt: publishedAt,
		FeedID:      feedID.String(),
	}); err != nil {
		t.Fatalf("create post: %v", err)
	}
	return id
}

// removeFeed removes the fixture Feed, "Example Feed", through the Feeds screen.
func (a *testApp) removeFeed(t *testing.T) {
	t.Helper()
	a.request(t, http.MethodDelete, "/feeds/"+a.feedID.String(), nil)
}

func TestFeeds_RemovingAFeedTakesItsPostsOutOfTheQueue(t *testing.T) {
	app := newTestApp(t)
	now := time.Now()
	pageLoad := now.Add(-time.Minute)
	other := app.followedFeed(t, "Other Feed")
	kept := app.postIn(t, other, "Other post", now.Add(-2*time.Hour))
	app.fetchedAt(t, kept, pageLoad.Add(-time.Minute))
	removed := app.post(t, "Removed Feed post", "", now.Add(-time.Hour))
	app.fetchedAt(t, removed, pageLoad.Add(-time.Minute))
	app.post(t, "Removed Feed new post", "", now)

	app.removeFeed(t)

	card := app.request(t, http.MethodGet, "/queue/card?"+cutoffQuery(pageLoad), nil).Body.String()
	assertContains(t, card, "Other post", "1 left")
	assertNotContains(t, card, "Removed Feed post", " new<")

	confirm := app.request(t, http.MethodGet, "/queue/trash-all?"+cutoffQuery(pageLoad), nil).Body.String()
	assertContains(t, confirm, "Trash 1 post")

	form := url.Values{"cutoff": {pageLoad.UTC().Format(time.RFC3339Nano)}}.Encode()
	app.request(t, http.MethodPost, "/queue/trash-all", strings.NewReader(form))
	app.assertPlace(t, kept, service.PlaceTrash)
	app.assertPlace(t, removed, service.PlaceQueue)
}

func TestFeeds_RemovingAFeedKeepsItsReadingListArchiveAndTrashPostsWithItsName(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	kept := app.post(t, "Kept post", "", time.Now())
	archived := app.post(t, "Archived post", "", time.Now())
	trashed := app.post(t, "Trashed post", "", time.Now())
	for _, id := range []uuid.UUID{kept, archived} {
		if err := app.states.Keep(ctx, app.userID, id); err != nil {
			t.Fatalf("Keep: %v", err)
		}
	}
	if err := app.states.Archive(ctx, app.userID, archived); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if err := app.states.Trash(ctx, app.userID, trashed, service.PlaceQueue); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	app.removeFeed(t)

	for target, title := range map[string]string{
		"/reading-list": "Kept post",
		"/archive":      "Archived post",
		"/trash":        "Trashed post",
	} {
		body := app.request(t, http.MethodGet, target, nil).Body.String()
		assertContains(t, body, title, "Example Feed")
	}
}

func TestFeeds_EachFeedShowsItsLastFetchAndLastError(t *testing.T) {
	app := newTestApp(t)
	broken := app.followedFeed(t, "Broken Feed")
	app.followedFeed(t, "Never fetched Feed")
	if _, err := app.db.Exec("UPDATE feeds SET last_fetched_at = ? WHERE id = ?", "2026-11-04 09:30:00", app.feedID.String()); err != nil {
		t.Fatalf("set last_fetched_at: %v", err)
	}
	if _, err := app.db.Exec(
		"UPDATE feeds SET last_fetched_at = ?, last_fetch_error = ?, last_fetch_error_at = ? WHERE id = ?",
		"2026-11-01 08:00:00", "unexpected status code: 500", "2026-11-04 10:15:00", broken.String(),
	); err != nil {
		t.Fatalf("set fetch error: %v", err)
	}

	body := app.request(t, http.MethodGet, "/feeds", nil).Body.String()

	assertContains(t, body,
		"Example Feed", `datetime="2026-11-04T09:30:00Z"`, "November 4, 2026 09:30",
		"Broken Feed", `datetime="2026-11-01T08:00:00Z"`, "unexpected status code: 500", `datetime="2026-11-04T10:15:00Z"`,
		"Never fetched Feed", "Not fetched yet",
	)
	if n := strings.Count(body, "unexpected status code"); n != 1 {
		t.Errorf("error shown %d times, want once, on the broken Feed only", n)
	}
}

func TestFeeds_ARemovedFeedIsNotListed(t *testing.T) {
	app := newTestApp(t)
	app.followedFeed(t, "Other Feed")

	app.removeFeed(t)

	body := app.request(t, http.MethodGet, "/feeds", nil).Body.String()
	assertContains(t, body, "Other Feed")
	assertNotContains(t, body, "Example Feed")
}

// serveFeed serves an RSS Feed, "Served Feed", with one item, "Served post".
func serveFeed(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Served Feed</title><link>http://example.com</link><description>d</description>
<item><title>Served post</title><link>http://example.com/served</link><description>Served today</description><pubDate>Mon, 02 Nov 2026 10:00:00 GMT</pubDate></item>
</channel></rss>`))
	}))
	t.Cleanup(server.Close)
	return server
}

// fetchNow runs the Queue's fetch, then responds with the Queue as a page
// loaded afterwards shows it.
func (a *testApp) fetchNow(t *testing.T) string {
	t.Helper()
	// The fixture Feed's URL isn't served; mark it fetched so the fetch skips it.
	if err := a.queries.MarkFeedAsFetched(context.Background(), a.feedID.String()); err != nil {
		t.Fatalf("MarkFeedAsFetched: %v", err)
	}
	form := url.Values{"cutoff": {time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)}}.Encode()
	a.request(t, http.MethodPost, "/queue/check", strings.NewReader(form))
	return a.request(t, http.MethodGet, "/queue/card", nil).Body.String()
}

func addFeedForm(feedURL string) io.Reader {
	return strings.NewReader(url.Values{"url": {feedURL}}.Encode())
}

func TestFeeds_ANewFeedsItemsEnterTheQueueOnTheNextFetch(t *testing.T) {
	app := newTestApp(t)
	server := serveFeed(t)

	added := app.request(t, http.MethodPost, "/feeds", addFeedForm(server.URL)).Body.String()
	assertContains(t, added, "Served Feed")

	assertContains(t, app.fetchNow(t), "Served post", "Served Feed", "1 left")
}

func TestFeeds_ReAddingARemovedFeedsURLUnRemovesIt(t *testing.T) {
	app := newTestApp(t)
	server := serveFeed(t)
	app.request(t, http.MethodPost, "/feeds", addFeedForm(server.URL))
	app.fetchNow(t)
	served, err := app.queries.GetFeedByUrl(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("GetFeedByUrl: %v", err)
	}
	app.request(t, http.MethodDelete, "/feeds/"+served.ID, nil)
	if body := app.request(t, http.MethodGet, "/queue/card", nil).Body.String(); strings.Contains(body, "Served post") {
		t.Fatalf("removed Feed's Post is still in the Queue; body:\n%s", body)
	}

	added := app.request(t, http.MethodPost, "/feeds", addFeedForm(server.URL)).Body.String()

	assertContains(t, added, "Served Feed")
	assertNotContains(t, added, "already exists")
	assertContains(t, app.request(t, http.MethodGet, "/feeds", nil).Body.String(), "Served Feed")
	assertContains(t, app.fetchNow(t), "Served post", "1 left")
}

func TestFeeds_AddingAFeedAlreadyFollowedSaysSo(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/feeds", addFeedForm("http://example.com/feed.xml"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	app.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	assertContains(t, rec.Body.String(), "already exists")
}

func TestFeeds_SearchFindsARemovedFeedsKeptPostsButNotItsQueuePosts(t *testing.T) {
	app := newTestApp(t)
	kept := app.post(t, "Kept needle", "", time.Now())
	app.post(t, "Queued needle", "", time.Now())
	if err := app.states.Keep(context.Background(), app.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	app.removeFeed(t)

	body := app.request(t, http.MethodPost, "/search", strings.NewReader("search=needle")).Body.String()
	assertContains(t, body, "Kept needle", "Example Feed")
	assertNotContains(t, body, "Queued needle")
}
