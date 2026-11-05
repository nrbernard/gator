package handler

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// pages are the app's full pages, which share the shell: header, assets, type.
var pages = []string{"/", "/reading-list", "/archive", "/trash", "/feeds"}

var headerPattern = regexp.MustCompile(`(?s)<header\b.*?</header>`)

func TestShell_HeaderLinksToTheQueueTheReadingListAndFeeds(t *testing.T) {
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			app := newTestApp(t)
			app.post(t, "A post", "", time.Now())

			body := app.request(t, http.MethodGet, page, nil).Body.String()

			header := headerPattern.FindString(body)
			if header == "" {
				t.Fatalf("page has no <header>; body:\n%s", body)
			}
			assertContains(t, header, `href="/"`, `href="/reading-list"`, `href="/archive"`, `href="/trash"`, `href="/feeds"`)
		})
	}
}

var (
	cssURLPattern    = regexp.MustCompile(`url\(\s*['"]?([^'")]+)['"]?\s*\)`)
	cssImportPattern = regexp.MustCompile(`@import\s+['"]([^'"]+)['"]`)
)

func TestShell_StylesheetLoadsOnlyAssetsTheAppServes(t *testing.T) {
	app := newTestApp(t)
	input, err := os.ReadFile("../../static/css/input.css")
	if err != nil {
		t.Fatalf("read Tailwind input: %v", err)
	}
	css := string(input)

	assertContains(t, css, `font-family: "Geist"`)
	for _, m := range cssImportPattern.FindAllStringSubmatch(css, -1) {
		if m[1] != "tailwindcss" {
			t.Errorf("stylesheet imports %q; only the Tailwind build is allowed", m[1])
		}
	}
	urls := cssURLPattern.FindAllStringSubmatch(css, -1)
	if len(urls) == 0 {
		t.Fatal("stylesheet loads no font files")
	}
	for _, m := range urls {
		if !strings.HasPrefix(m[1], "/static/") {
			t.Errorf("stylesheet loads %q, which the app doesn't serve", m[1])
			continue
		}
		app.request(t, http.MethodGet, m[1], nil)
	}
}

// tailwindOutput is the stylesheet the Tailwind build makes from static/css/input.css.
const tailwindOutput = "/static/css/output.css"

// resourcePattern matches what a page loads (scripts, styles, fonts, icons), not its links.
var resourcePattern = regexp.MustCompile(`<(?:script|link)\b[^>]*?\b(?:src|href)="([^"]*)"`)

func TestShell_PagesLoadOnlyAssetsTheAppServes(t *testing.T) {
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			app := newTestApp(t)
			app.post(t, "A post", "", time.Now())

			body := app.request(t, http.MethodGet, page, nil).Body.String()

			assertContains(t, body, `href="`+tailwindOutput+`"`)
			resources := resourcePattern.FindAllStringSubmatch(body, -1)
			for _, r := range resources {
				url := r[1]
				switch {
				case strings.HasPrefix(url, "data:"):
				case url == tailwindOutput:
					// A build artifact the tests don't build; TestShell_StylesheetLoadsOnlyAssetsTheAppServes covers its input.
				case strings.HasPrefix(url, "/static/"):
					app.request(t, http.MethodGet, url, nil)
				default:
					t.Errorf("page loads %q, which the app doesn't serve", url)
				}
			}
		})
	}
}
