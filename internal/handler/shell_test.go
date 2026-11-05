package handler

import (
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// pages are the app's full pages, which share the shell: header, assets, type.
var pages = []string{"/", "/reading-list", "/archive", "/trash", "/search", "/feeds"}

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

var manifestLinkPattern = regexp.MustCompile(`<link\b[^>]*\brel="manifest"[^>]*\bhref="([^"]*)"`)

// webManifest is the part of the web app manifest the home screen reads.
type webManifest struct {
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	StartURL  string `json:"start_url"`
	Display   string `json:"display"`
	Icons     []struct {
		Src   string `json:"src"`
		Sizes string `json:"sizes"`
		Type  string `json:"type"`
	} `json:"icons"`
}

func TestShell_EveryPageLinksAManifestThatOpensTheQueueStandalone(t *testing.T) {
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			app := newTestApp(t)
			app.post(t, "A post", "", time.Now())

			body := app.request(t, http.MethodGet, page, nil).Body.String()

			link := manifestLinkPattern.FindStringSubmatch(body)
			if link == nil {
				t.Fatalf("page links no manifest; body:\n%s", body)
			}
			var m webManifest
			if err := json.Unmarshal(app.request(t, http.MethodGet, link[1], nil).Body.Bytes(), &m); err != nil {
				t.Fatalf("manifest at %s isn't JSON: %v", link[1], err)
			}
			if m.Name != "Gator" || m.ShortName != "Gator" {
				t.Errorf("manifest names the app %q (short %q), want Gator", m.Name, m.ShortName)
			}
			if m.StartURL != "/" {
				t.Errorf("manifest starts at %q, want the Queue at /", m.StartURL)
			}
			if m.Display != "standalone" {
				t.Errorf("manifest display is %q, want standalone", m.Display)
			}
			sizes := map[string]bool{}
			for _, icon := range m.Icons {
				if icon.Type != "image/png" {
					t.Errorf("manifest icon %s has type %q, want image/png", icon.Src, icon.Type)
				}
				assertPNG(t, app, icon.Src, icon.Sizes)
				sizes[icon.Sizes] = true
			}
			for _, want := range []string{"192x192", "512x512"} {
				if !sizes[want] {
					t.Errorf("manifest has no %s icon", want)
				}
			}
		})
	}
}

// iconLinkPattern matches a page's icons: the favicon and the home-screen icon.
var iconLinkPattern = regexp.MustCompile(`<link\b[^>]*\brel="(icon|apple-touch-icon)"[^>]*>`)

var (
	hrefAttr  = regexp.MustCompile(`\bhref="([^"]*)"`)
	sizesAttr = regexp.MustCompile(`\bsizes="([^"]*)"`)
)

func TestShell_EveryPageHasTheGatorIconAsFaviconAndHomeScreenIcon(t *testing.T) {
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			app := newTestApp(t)
			app.post(t, "A post", "", time.Now())

			body := app.request(t, http.MethodGet, page, nil).Body.String()

			found := map[string]bool{}
			for _, m := range iconLinkPattern.FindAllStringSubmatch(body, -1) {
				rel, tag := m[1], m[0]
				href, sizes := hrefAttr.FindStringSubmatch(tag), sizesAttr.FindStringSubmatch(tag)
				if href == nil || sizes == nil {
					t.Errorf("%s link needs an href and sizes: %s", rel, tag)
					continue
				}
				assertPNG(t, app, href[1], sizes[1])
				found[rel+" "+sizes[1]] = true
			}
			if !found["apple-touch-icon 180x180"] {
				t.Errorf("page has no 180x180 apple-touch-icon; has %v", found)
			}
			hasFavicon := false
			for k := range found {
				hasFavicon = hasFavicon || strings.HasPrefix(k, "icon ")
			}
			if !hasFavicon {
				t.Error("page has no PNG favicon")
			}
		})
	}
}

var scriptSrcPattern = regexp.MustCompile(`<script\b[^>]*\bsrc="([^"]*)"`)

// Installable without offline mode: nothing a page runs registers a service worker.
func TestShell_NoPageRegistersAServiceWorker(t *testing.T) {
	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			app := newTestApp(t)
			app.post(t, "A post", "", time.Now())

			body := app.request(t, http.MethodGet, page, nil).Body.String()

			assertNotContains(t, body, "serviceWorker")
			for _, m := range scriptSrcPattern.FindAllStringSubmatch(body, -1) {
				script := app.request(t, http.MethodGet, m[1], nil).Body.String()
				if strings.Contains(script, "serviceWorker.register") {
					t.Errorf("%s registers a service worker", m[1])
				}
			}
		})
	}
}

// assertPNG requests url and checks it is a PNG of the given "WxH" size.
func assertPNG(t *testing.T, app *testApp, url, size string) {
	t.Helper()
	cfg, err := png.DecodeConfig(app.request(t, http.MethodGet, url, nil).Body)
	if err != nil {
		t.Errorf("%s isn't a PNG: %v", url, err)
		return
	}
	if got := fmt.Sprintf("%dx%d", cfg.Width, cfg.Height); got != size {
		t.Errorf("%s is %s, want %s", url, got, size)
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
