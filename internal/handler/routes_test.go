package handler

import (
	"context"
	"database/sql"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/service"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
)

type templateRenderer struct{ tmpl *template.Template }

func (r *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.tmpl.ExecuteTemplate(w, name, data)
}

// testApp is the app's routes over a migrated in-memory database, with one
// user following one Feed.
type testApp struct {
	e       *echo.Echo
	queries *database.Queries
	states  *service.PostStateService
	userID  uuid.UUID
	feedID  uuid.UUID
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	provider, err := goose.NewProvider(goosedb.DialectSQLite3, db, os.DirFS("../../sql/schema"))
	if err != nil {
		t.Fatalf("create migration provider: %v", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	queries := database.New(db)

	userID, feedID := uuid.New(), uuid.New()
	if _, err := queries.CreateUser(ctx, database.CreateUserParams{ID: userID.String(), Name: "nick"}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := queries.CreateFeed(ctx, database.CreateFeedParams{
		ID: feedID.String(), Name: "Example Feed", Url: "http://example.com/feed.xml", UserID: userID.String(),
	}); err != nil {
		t.Fatalf("create feed: %v", err)
	}
	if _, err := queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID: uuid.NewString(), UserID: userID.String(), FeedID: feedID.String(),
	}); err != nil {
		t.Fatalf("create feed follow: %v", err)
	}

	app := &testApp{
		e:       echo.New(),
		queries: queries,
		states:  &service.PostStateService{Repo: queries},
		userID:  userID,
		feedID:  feedID,
	}
	app.e.Renderer = &templateRenderer{tmpl: template.Must(template.ParseGlob("../views/*.html"))}
	app.e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("userID", userID)
			return next(c)
		}
	})
	posts := &service.PostService{Repo: queries}
	feeds := &service.FeedService{Repo: queries}
	Register(app.e, Handlers{
		Queue:       &QueueHandler{PostService: posts, StateService: app.states, FeedService: feeds},
		ReadingList: &ReadingListHandler{PostService: posts, StateService: app.states},
		Post:        &PostHandler{PostService: posts},
		Feed:        &FeedHandler{FeedService: feeds, UserService: &service.UserService{Repo: queries}},
	})
	return app
}

func (a *testApp) post(t *testing.T, title, description string, publishedAt time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := a.queries.CreatePost(context.Background(), database.CreatePostParams{
		ID:          id.String(),
		Title:       title,
		Url:         "http://example.com/posts/" + id.String(),
		Description: sql.NullString{String: description, Valid: description != ""},
		PublishedAt: publishedAt,
		FeedID:      a.feedID.String(),
	}); err != nil {
		t.Fatalf("create post: %v", err)
	}
	return id
}

func (a *testApp) request(t *testing.T, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	if body != nil {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	}
	rec := httptest.NewRecorder()
	a.e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s %s: status %d; body:\n%s", method, target, rec.Code, rec.Body.String())
	}
	return rec
}

func (a *testApp) assertPlace(t *testing.T, postID uuid.UUID, want service.Place) {
	t.Helper()
	got, err := a.states.State(context.Background(), a.userID, postID)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if got.Place != want {
		t.Errorf("post is in %s, want %s", got.Place, want)
	}
}

func assertContains(t *testing.T, body string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(body, want) {
			t.Errorf("response is missing %q; body:\n%s", want, body)
		}
	}
}

func assertNotContains(t *testing.T, body string, unwanted ...string) {
	t.Helper()
	for _, u := range unwanted {
		if strings.Contains(body, u) {
			t.Errorf("response contains %q; body:\n%s", u, body)
		}
	}
}

func TestQueue_ShowsTheNewestQueuePostAndTheCountLeft(t *testing.T) {
	app := newTestApp(t)
	now := time.Now()
	app.post(t, "Older post", "", now.Add(-2*time.Hour))
	app.post(t, "Newest post", "The newest description", now.Add(-1*time.Hour))
	kept := app.post(t, "Already kept", "", now)
	if err := app.states.Keep(context.Background(), app.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	body := app.request(t, http.MethodGet, "/", nil).Body.String()

	assertContains(t, body, "Newest post", "The newest description", "Example Feed", "2 left", `href="/reading-list"`)
	assertNotContains(t, body, "Older post", "Already kept")
}

func TestQueue_TriageMovesThePostAndRespondsWithTheNextCard(t *testing.T) {
	tests := []struct {
		action string
		want   service.Place
	}{
		{"keep", service.PlaceReadingList},
		{"trash", service.PlaceTrash},
	}

	for _, tc := range tests {
		t.Run(tc.action, func(t *testing.T) {
			app := newTestApp(t)
			now := time.Now()
			next := app.post(t, "Next post", "", now.Add(-1*time.Hour))
			top := app.post(t, "Top post", "", now)

			body := app.request(t, http.MethodPost, "/queue/"+top.String()+"/"+tc.action, nil).Body.String()

			app.assertPlace(t, top, tc.want)
			app.assertPlace(t, next, service.PlaceQueue)
			assertContains(t, body, `id="queue"`, "Next post", "1 left", "/queue/"+next.String()+"/keep")
			assertNotContains(t, body, "Top post", "<html")
		})
	}
}

func TestQueue_TriagingTheLastPostRespondsWithAnEmptyQueue(t *testing.T) {
	app := newTestApp(t)
	last := app.post(t, "Last post", "", time.Now())

	body := app.request(t, http.MethodPost, "/queue/"+last.String()+"/keep", nil).Body.String()

	app.assertPlace(t, last, service.PlaceReadingList)
	assertContains(t, body, "The Queue is empty", `href="/reading-list"`)
	assertNotContains(t, body, "Last post")
}

func TestReadingList_ListsKeptPosts(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	kept := app.post(t, "Kept post", "", time.Now())
	trashed := app.post(t, "Trashed post", "", time.Now())
	app.post(t, "Queue post", "", time.Now())
	if err := app.states.Keep(ctx, app.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	if err := app.states.Trash(ctx, app.userID, trashed, service.PlaceQueue); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	body := app.request(t, http.MethodGet, "/reading-list", nil).Body.String()

	assertContains(t, body, "Kept post", "Example Feed", "http://example.com/posts/"+kept.String())
	assertNotContains(t, body, "Trashed post", "Queue post")
}

func TestReadingList_EachPostOffersOpenReadAndTrashOverHtmx(t *testing.T) {
	app := newTestApp(t)
	kept := app.post(t, "Kept post", "", time.Now())
	if err := app.states.Keep(context.Background(), app.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	body := app.request(t, http.MethodGet, "/reading-list", nil).Body.String()

	id := kept.String()
	assertContains(t, body, `id="reading-list"`, "htmx",
		`hx-post="/reading-list/`+id+`/open"`, `hx-post="/reading-list/`+id+`/read"`, `hx-post="/reading-list/`+id+`/trash"`,
		">Mark read<")
}

func TestReadingList_ActionsUpdateTheListInPlace(t *testing.T) {
	tests := []struct {
		action   string
		want     service.PostState
		wantBody []string
		unwanted []string
	}{
		// A read Post stays on the Reading list, offering to mark it unread.
		{"open", service.PostState{Place: service.PlaceReadingList, Read: true}, []string{"Acted on", ">Mark unread<"}, nil},
		{"read", service.PostState{Place: service.PlaceReadingList, Read: true}, []string{"Acted on", ">Mark unread<"}, nil},
		{"trash", service.PostState{Place: service.PlaceTrash, TrashedFrom: service.PlaceReadingList}, nil, []string{"Acted on"}},
	}

	for _, tc := range tests {
		t.Run(tc.action, func(t *testing.T) {
			app := newTestApp(t)
			ctx := context.Background()
			other := app.post(t, "Other kept post", "", time.Now())
			acted := app.post(t, "Acted on", "", time.Now())
			for _, id := range []uuid.UUID{other, acted} {
				if err := app.states.Keep(ctx, app.userID, id); err != nil {
					t.Fatalf("Keep: %v", err)
				}
			}

			body := app.request(t, http.MethodPost, "/reading-list/"+acted.String()+"/"+tc.action, nil).Body.String()

			got, err := app.states.State(ctx, app.userID, acted)
			if err != nil {
				t.Fatalf("State: %v", err)
			}
			if got != tc.want {
				t.Errorf("state = %+v, want %+v", got, tc.want)
			}
			assertContains(t, body, append([]string{`id="reading-list"`, "Other kept post"}, tc.wantBody...)...)
			assertNotContains(t, body, append([]string{"<html"}, tc.unwanted...)...)
		})
	}
}

func TestReadingList_ReadTogglesBackToUnread(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	kept := app.post(t, "Opened by accident", "", time.Now())
	if err := app.states.Keep(ctx, app.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	if err := app.states.Open(ctx, app.userID, kept); err != nil {
		t.Fatalf("Open: %v", err)
	}

	body := app.request(t, http.MethodPost, "/reading-list/"+kept.String()+"/read", nil).Body.String()

	got, err := app.states.State(ctx, app.userID, kept)
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if got.Read {
		t.Errorf("post is still read")
	}
	assertContains(t, body, "Opened by accident", ">Mark read<")
}

func TestReadingList_RejectsActionsOnAPostNotOnTheReadingList(t *testing.T) {
	app := newTestApp(t)
	queued := app.post(t, "Queue post", "", time.Now())

	for _, action := range []string{"read", "trash"} {
		rec := httptest.NewRecorder()
		app.e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/reading-list/"+queued.String()+"/"+action, nil))

		if rec.Code != http.StatusConflict {
			t.Errorf("%s: status = %d, want %d", action, rec.Code, http.StatusConflict)
		}
	}
	app.assertPlace(t, queued, service.PlaceQueue)
}

func TestSearch_FindsPostsByTitleAndDescriptionInEveryPlace(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	app.post(t, "Needle in the title", "", time.Now())
	kept := app.post(t, "Kept post", "a needle in the description", time.Now())
	trashed := app.post(t, "Trashed needle", "", time.Now())
	app.post(t, "Haystack", "", time.Now())
	if err := app.states.Keep(ctx, app.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	if err := app.states.Trash(ctx, app.userID, trashed, service.PlaceQueue); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	body := app.request(t, http.MethodPost, "/search", strings.NewReader("search=needle")).Body.String()

	assertContains(t, body, "Needle in the title", "Kept post", "Trashed needle")
	assertNotContains(t, body, "Haystack", "<button", "/saved-posts/", "/read-posts/")
}

func TestQueue_RefreshFetchesFeedsAndRespondsWithTheQueue(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	// The fixture Feed's URL isn't served; mark it fetched so Refresh skips it.
	if err := app.queries.MarkFeedAsFetched(ctx, app.feedID.String()); err != nil {
		t.Fatalf("MarkFeedAsFetched: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>Served Feed</title><link>http://example.com</link><description>d</description>
<item><title>Freshly fetched</title><link>http://example.com/fresh</link><description>New today</description><pubDate>Mon, 02 Nov 2026 10:00:00 GMT</pubDate></item>
</channel></rss>`))
	}))
	t.Cleanup(server.Close)
	servedID := uuid.NewString()
	if _, err := app.queries.CreateFeed(ctx, database.CreateFeedParams{ID: servedID, Name: "Served Feed", Url: server.URL, UserID: app.userID.String()}); err != nil {
		t.Fatalf("create feed: %v", err)
	}
	if _, err := app.queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{ID: uuid.NewString(), UserID: app.userID.String(), FeedID: servedID}); err != nil {
		t.Fatalf("follow feed: %v", err)
	}

	body := app.request(t, http.MethodPost, "/queue/refresh", nil).Body.String()

	assertContains(t, body, `id="queue"`, "Freshly fetched", "Served Feed", "1 left")
}

func TestQueue_HasNoTabsOrToggleButtons(t *testing.T) {
	app := newTestApp(t)
	app.post(t, "A post", "", time.Now())

	body := app.request(t, http.MethodGet, "/", nil).Body.String()

	assertNotContains(t, body, `id="tabs"`, "status=saved", "/saved-posts/", "/read-posts/")
	for _, target := range []string{"/posts?status=saved", "/saved-posts/" + uuid.NewString(), "/read-posts/" + uuid.NewString()} {
		rec := httptest.NewRecorder()
		app.e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, target, nil))
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s: status %d, want the route gone", target, rec.Code)
		}
	}
}

func TestQueue_RejectsTriageOfAPostNotInTheQueue(t *testing.T) {
	app := newTestApp(t)
	kept := app.post(t, "Kept post", "", time.Now())
	if err := app.states.Keep(context.Background(), app.userID, kept); err != nil {
		t.Fatalf("Keep: %v", err)
	}

	rec := httptest.NewRecorder()
	app.e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/queue/"+kept.String()+"/trash", nil))

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	app.assertPlace(t, kept, service.PlaceReadingList)
}
