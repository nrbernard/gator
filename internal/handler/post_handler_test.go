package handler

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/service"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
)

// failingRenderer renders the real views but fails the template named failOn,
// standing in for a template that is renamed or errors at execution.
type failingRenderer struct {
	tmpl   *template.Template
	failOn string
}

func (r *failingRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if name == r.failOn {
		return errors.New("injected render failure: " + name)
	}
	return r.tmpl.ExecuteTemplate(w, name, data)
}

func newTestPostHandler(t *testing.T) *PostHandler {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)

	provider, err := goose.NewProvider(goosedb.DialectSQLite3, db, os.DirFS("../../sql/schema"))
	if err != nil {
		t.Fatalf("Failed to create migration provider: %v", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	queries := database.New(db)
	return &PostHandler{
		PostService: &service.PostService{Repo: queries},
		UserService: &service.UserService{Repo: queries},
		FeedService: &service.FeedService{Repo: queries},
	}
}

// Index and Refresh render two fragments into one htmx response. If the first
// fragment fails, the request must fail rather than return the second alone.
func TestPostHandler_FirstFragmentRenderFailure(t *testing.T) {
	tmpl := template.Must(template.ParseGlob("../views/*.html"))

	tests := []struct {
		name   string
		method string
		target string
		failOn string
		route  func(h *PostHandler) echo.HandlerFunc
	}{
		{"index tabs", http.MethodGet, "/posts?status=saved", "tabs", func(h *PostHandler) echo.HandlerFunc { return h.Index }},
		{"refresh posts-refresh", http.MethodPost, "/posts/refresh", "posts-refresh", func(h *PostHandler) echo.HandlerFunc { return h.Refresh }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestPostHandler(t)
			e := echo.New()
			e.Renderer = &failingRenderer{tmpl: tmpl, failOn: tc.failOn}
			userID := uuid.New()
			e.Add(tc.method, strings.SplitN(tc.target, "?", 2)[0], tc.route(h), func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					c.Set("userID", userID)
					return next(c)
				}
			})

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.target, nil))

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d; body:\n%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), `id="posts"`) {
				t.Errorf("response contains the oob-posts fragment without %q", tc.failOn)
			}
		})
	}
}
