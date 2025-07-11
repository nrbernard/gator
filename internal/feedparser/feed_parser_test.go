package feedparser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Helper function to compare feeds properly
func compareFeeds(expected, actual Feed) bool {
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	if expected.GetTitle() != actual.GetTitle() {
		return false
	}
	if expected.GetLink() != actual.GetLink() {
		return false
	}
	if expected.GetDescription() != actual.GetDescription() {
		return false
	}

	expectedItems := expected.GetItems()
	actualItems := actual.GetItems()

	if len(expectedItems) != len(actualItems) {
		return false
	}

	for i := range expectedItems {
		if !compareItems(expectedItems[i], actualItems[i]) {
			return false
		}
	}

	return true
}

// Helper function to compare items
func compareItems(expected, actual Item) bool {
	if expected.GetTitle() != actual.GetTitle() {
		return false
	}
	if expected.GetLink() != actual.GetLink() {
		return false
	}
	if expected.GetDescription() != actual.GetDescription() {
		return false
	}
	if !expected.GetDate().Equal(actual.GetDate()) {
		return false
	}
	return true
}

func TestFetchFeed(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
		expectedFeed   Feed
	}{
		{
			name: "RSS feed fetch",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/rss+xml")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
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
				</rss>`))
			},
			expectedError: false,
			expectedFeed: &RSSFeed{
				title:       "Test Feed",
				link:        "https://example.com",
				description: "Test Description",
				items: []*RSSItem{
					{
						title:       "Test Item",
						link:        "https://example.com/item",
						description: "Test Item Description",
						date:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "atom feed fetch",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/atom+xml")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
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
			</feed>`))
			},
			expectedError: false,
			expectedFeed: &AtomFeed{
				title:       "David Heinemeier Hansson",
				link:        "https://world.hey.com/dhh/feed.atom",
				description: "",
				items: []*AtomItem{
					{
						title:       "Don't make Google sell Chrome",
						link:        "https://world.hey.com/dhh/don-t-make-google-sell-chrome-93cefbc6",
						description: "The web will be far worse off if Google is forced to sell Chrome, even if it's to atone for legitimate ad-market monopoly abuses.",
						date:        time.Date(2025, 4, 28, 6, 0, 37, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: true,
			expectedFeed:  nil,
		},
		{
			name: "invalid XML",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid xml content"))
			},
			expectedError: true,
			expectedFeed:  nil,
		},
		{
			name: "HTML entities in content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/rss+xml")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
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
				</rss>`))
			},
			expectedError: false,
			expectedFeed: &RSSFeed{
				title:       "Test & Feed",
				link:        "https://example.com",
				description: "Test <Description>",
				items: []*RSSItem{
					{
						title:       "Test \"Item\"",
						link:        "https://example.com/item",
						description: "Test 'Item' Description",
						date:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.serverResponse))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			feed, err := FetchFeed(ctx, server.URL)

			if tc.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tc.expectedFeed != nil {
				if !compareFeeds(tc.expectedFeed, feed) {
					t.Errorf("expected feed %v, got %v", tc.expectedFeed, feed)
				}
			}
		})
	}
}
