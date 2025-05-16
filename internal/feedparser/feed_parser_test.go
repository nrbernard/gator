package feedparser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchFeed(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		contentType    string
		wantErr        bool
		validateFeed   func(*testing.T, Feed)
	}{
		{
			name: "successful RSS feed fetch",
			serverResponse: `<?xml version="1.0" encoding="UTF-8"?>
				<rss version="2.0">
					<channel>
						<title>Test Feed</title>
						<link>https://example.com</link>
						<description>Test Description</description>
						<item>
							<title>Test Item</title>
							<link>https://example.com/item</link>
							<description>Test Item Description</description>
							<pubDate>Wed, 01 Jan 2024 12:00:00 GMT</pubDate>
						</item>
					</channel>
				</rss>`,
			statusCode:  http.StatusOK,
			contentType: "application/rss+xml",
			wantErr:     false,
			validateFeed: func(t *testing.T, feed Feed) {
				if feed.GetTitle() != "Test Feed" {
					t.Errorf("expected title 'Test Feed', got '%s'", feed.GetTitle())
				}
				if feed.GetLink() != "https://example.com" {
					t.Errorf("expected link 'https://example.com', got '%s'", feed.GetLink())
				}
				if len(feed.GetItems()) != 1 {
					t.Errorf("expected 1 item, got %d", len(feed.GetItems()))
				}
				if feed.GetItems()[0].GetTitle() != "Test Item" {
					t.Errorf("expected item title 'Test Item', got '%s'", feed.GetItems()[0].GetTitle())
				}
			},
		},
		{
			name: "successful Atom feed fetch",
			serverResponse: `<?xml version="1.0" encoding="UTF-8"?>
			<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom">
				<id>tag:world.hey.com,2005:/dhh/feed</id>
				<link rel="alternate" type="text/html" href="https://world.hey.com/dhh"/>
				<link rel="self" type="application/atom+xml" href="https://world.hey.com/dhh/feed.atom"/>
				<title>David Heinemeier Hansson</title>
				<updated>2025-04-28T06:00:48Z</updated>
				<entry>
					<id>tag:world.hey.com,2005:World::Post/42931</id>
					<published>2025-04-28T06:00:37Z</published>
					<updated>2025-04-28T06:00:48Z</updated>
					<link rel="alternate" type="text/html" href="https://world.hey.com/dhh/don-t-make-google-sell-chrome-93cefbc6"/>
					<title>Don't make Google sell Chrome</title>
					<content type="html">&lt;div class="trix-content"&gt;
				&lt;div&gt;The web will be far worse off &lt;a href="https://www.msn.com/en-us/technology/tech-companies/a-judge-could-force-google-to-sell-chrome-what-you-need-to-know/ar-AA1DjlGy"&gt;if Google is forced to sell Chrome&lt;/a&gt;, even if it's to atone for legitimate ad-market monopoly abuses.&lt;/div&gt;
				&lt;/div&gt;
				</content>
					<author>
					<name>David Heinemeier Hansson</name>
					<email>dhh@hey.com</email>
					</author>
				</entry>
			</feed>`,
			statusCode:  http.StatusOK,
			contentType: "application/atom+xml",
			wantErr:     false,
			validateFeed: func(t *testing.T, feed Feed) {
				if feed.GetTitle() != "David Heinemeier Hansson" {
					t.Errorf("expected title 'David Heinemeier Hansson', got '%s'", feed.GetTitle())
				}
				if feed.GetLink() != "https://world.hey.com/dhh/feed.atom" {
					t.Errorf("expected link 'https://world.hey.com/dhh/feed.atom', got '%s'", feed.GetLink())
				}
				if len(feed.GetItems()) != 1 {
					t.Errorf("expected 1 item, got %d", len(feed.GetItems()))
				}
				if feed.GetItems()[0].GetTitle() != "Don't make Google sell Chrome" {
					t.Errorf("expected item title 'Don't make Google sell Chrome', got '%s'", feed.GetItems()[0].GetTitle())
				}
				if feed.GetItems()[0].GetLink() != "https://world.hey.com/dhh/don-t-make-google-sell-chrome-93cefbc6" {
					t.Errorf("expected link 'https://world.hey.com/dhh/don-t-make-google-sell-chrome-93cefbc6', got '%s'", feed.GetItems()[0].GetLink())
				}
				if feed.GetItems()[0].GetDescription() != "The web will be far worse off if Google is forced to sell Chrome, even if it's to atone for legitimate ad-market monopoly abuses." {
					t.Errorf("expected description 'The web will be far worse off if Google is forced to sell Chrome, even if it's to atone for legitimate ad-market monopoly abuses.', got '%s'", feed.GetItems()[0].GetDescription())
				}
				if feed.GetItems()[0].GetDate() != time.Date(2025, 4, 28, 6, 0, 37, 0, time.UTC) {
					t.Errorf("expected date %v, got %v", time.Date(2025, 4, 28, 6, 0, 37, 0, time.UTC), feed.GetItems()[0].GetDate())
				}
			},
		},
		{
			name:           "server error",
			serverResponse: "",
			statusCode:     http.StatusInternalServerError,
			wantErr:        true,
			validateFeed:   nil,
		},
		{
			name:           "invalid XML",
			serverResponse: "invalid xml content",
			statusCode:     http.StatusOK,
			contentType:    "application/rss+xml",
			wantErr:        true,
			validateFeed:   nil,
		},
		{
			name: "HTML entities in content",
			serverResponse: `<?xml version="1.0" encoding="UTF-8"?>
				<rss version="2.0">
					<channel>
						<title>Test &amp; Feed</title>
						<link>https://example.com</link>
						<description>Test &lt;Description&gt;</description>
						<item>
							<title>Test &quot;Item&quot;</title>
							<link>https://example.com/item</link>
							<description>Test &apos;Item&apos; Description</description>
							<pubDate>Wed, 01 Jan 2024 12:00:00 GMT</pubDate>
						</item>
					</channel>
				</rss>`,
			statusCode:  http.StatusOK,
			contentType: "application/rss+xml",
			wantErr:     false,
			validateFeed: func(t *testing.T, feed Feed) {
				if feed.GetTitle() != "Test & Feed" {
					t.Errorf("expected title 'Test & Feed', got '%s'", feed.GetTitle())
				}
				if feed.GetDescription() != "Test <Description>" {
					t.Errorf("expected description 'Test <Description>', got '%s'", feed.GetDescription())
				}
				if feed.GetItems()[0].GetTitle() != "Test \"Item\"" {
					t.Errorf("expected item title 'Test \"Item\"', got '%s'", feed.GetItems()[0].GetTitle())
				}
				expectedDate := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
				if !feed.GetItems()[0].GetDate().Equal(expectedDate) {
					t.Errorf("expected date %v, got %v", expectedDate, feed.GetItems()[0].GetDate())
				}
			},
		},
		{
			name: "Atom feed with date parsing",
			serverResponse: `<?xml version="1.0" encoding="UTF-8"?>
			<feed xmlns="http://www.w3.org/2005/Atom">
				<title>Test Feed</title>
				<link href="https://example.com"/>
				<entry>
					<title>Test Item</title>
					<link href="https://example.com/item"/>
					<published>2024-01-01T12:00:00Z</published>
				</entry>
			</feed>`,
			statusCode:  http.StatusOK,
			contentType: "application/atom+xml",
			wantErr:     false,
			validateFeed: func(t *testing.T, feed Feed) {
				expectedDate := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
				if !feed.GetItems()[0].GetDate().Equal(expectedDate) {
					t.Errorf("expected date %v, got %v", expectedDate, feed.GetItems()[0].GetDate())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify User-Agent header
				if r.Header.Get("User-Agent") != "hello-lane/1.0" {
					t.Errorf("expected User-Agent 'hello-lane/1.0', got '%s'", r.Header.Get("User-Agent"))
				}
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call FetchFeed
			feed, err := FetchFeed(ctx, server.URL)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect an error, we're done
			if tt.wantErr {
				return
			}

			// Validate feed content if validation function is provided
			if tt.validateFeed != nil {
				tt.validateFeed(t, feed)
			}
		})
	}
}
