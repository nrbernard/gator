package feedparser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchFeedWithConditionals(t *testing.T) {
	// Create a test server that responds with different headers based on conditional requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for conditional headers
		ifNoneMatch := r.Header.Get("If-None-Match")
		ifModifiedSince := r.Header.Get("If-Modified-Since")

		// If both conditional headers are present and match our test values, return 304
		if ifNoneMatch == `"test-etag"` && ifModifiedSince == "Wed, 21 Oct 2015 07:28:00 GMT" {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Otherwise return 200 with test headers
		w.Header().Set("ETag", `"test-etag"`)
		w.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		w.Header().Set("Content-Type", "application/rss+xml")

		// Return a simple RSS feed
		rssContent := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>http://example.com</link>
    <description>Test Description</description>
    <item>
      <title>Test Item</title>
      <link>http://example.com/item1</link>
      <description>Test Description</description>
      <pubDate>Wed, 21 Oct 2015 07:28:00 GMT</pubDate>
    </item>
  </channel>
</rss>`
		w.Write([]byte(rssContent))
	}))
	defer server.Close()

	ctx := context.Background()

	t.Run("First request without conditionals", func(t *testing.T) {
		result, err := FetchFeedWithConditionals(ctx, server.URL, nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", result.StatusCode)
		}

		if result.ETag != `"test-etag"` {
			t.Errorf("Expected ETag %q, got %q", `"test-etag"`, result.ETag)
		}

		if result.LastModified != "Wed, 21 Oct 2015 07:28:00 GMT" {
			t.Errorf("Expected Last-Modified %q, got %q", "Wed, 21 Oct 2015 07:28:00 GMT", result.LastModified)
		}

		if result.NotModified {
			t.Error("Expected NotModified to be false")
		}

		if result.Feed == nil {
			t.Error("Expected Feed to be non-nil")
		}
	})

	t.Run("Conditional request with matching headers", func(t *testing.T) {
		etag := `"test-etag"`
		lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
		result, err := FetchFeedWithConditionals(ctx, server.URL, &etag, &lastModified)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.StatusCode != http.StatusNotModified {
			t.Errorf("Expected status 304, got %d", result.StatusCode)
		}

		if !result.NotModified {
			t.Error("Expected NotModified to be true")
		}

		if result.Feed != nil {
			t.Error("Expected Feed to be nil for 304 response")
		}
	})

	t.Run("Conditional request with non-matching headers", func(t *testing.T) {
		etag := `"old-etag"`
		lastModified := "Wed, 20 Oct 2015 07:28:00 GMT"
		result, err := FetchFeedWithConditionals(ctx, server.URL, &etag, &lastModified)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", result.StatusCode)
		}

		if result.NotModified {
			t.Error("Expected NotModified to be false")
		}

		if result.Feed == nil {
			t.Error("Expected Feed to be non-nil")
		}
	})
}

func TestFetchFeedWithConditionals_ErrorHandling(t *testing.T) {
	// Test server that returns 429 Too Many Requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limited"))
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := FetchFeedWithConditionals(ctx, server.URL, nil, nil)

	// Should return an error for non-200/304 status codes
	if err == nil {
		t.Error("Expected error for 429 status code")
	}

	if result != nil {
		t.Error("Expected result to be nil on error")
	}
}
