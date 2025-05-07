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
		wantErr        bool
		validateFeed   func(*testing.T, *RSSFeed)
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
			statusCode: http.StatusOK,
			wantErr:    false,
			validateFeed: func(t *testing.T, feed *RSSFeed) {
				if feed.Channel.Title != "Test Feed" {
					t.Errorf("expected title 'Test Feed', got '%s'", feed.Channel.Title)
				}
				if feed.Channel.Link != "https://example.com" {
					t.Errorf("expected link 'https://example.com', got '%s'", feed.Channel.Link)
				}
				if len(feed.Channel.Item) != 1 {
					t.Errorf("expected 1 item, got %d", len(feed.Channel.Item))
				}
				if feed.Channel.Item[0].Title != "Test Item" {
					t.Errorf("expected item title 'Test Item', got '%s'", feed.Channel.Item[0].Title)
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
			statusCode: http.StatusOK,
			wantErr:    false,
			validateFeed: func(t *testing.T, feed *RSSFeed) {
				if feed.Channel.Title != "Test & Feed" {
					t.Errorf("expected title 'Test & Feed', got '%s'", feed.Channel.Title)
				}
				if feed.Channel.Description != "Test <Description>" {
					t.Errorf("expected description 'Test <Description>', got '%s'", feed.Channel.Description)
				}
				if feed.Channel.Item[0].Title != "Test \"Item\"" {
					t.Errorf("expected item title 'Test \"Item\"', got '%s'", feed.Channel.Item[0].Title)
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
