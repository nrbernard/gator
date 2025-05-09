package feedparser

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type Feed interface {
	GetTitle() string
	GetLink() string
	GetDescription() string
	GetItems() []Item
}

type Item interface {
	GetTitle() string
	GetLink() string
	GetDescription() string
	GetDate() string
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Date        string `xml:"pubDate"`
}

func (i *RSSItem) GetTitle() string {
	return i.Title
}

func (i *RSSItem) GetLink() string {
	return i.Link
}

func (i *RSSItem) GetDescription() string {
	return i.Description
}

func (i *RSSItem) GetDate() string {
	return i.Date
}

func (f *RSSFeed) GetTitle() string {
	return f.Channel.Title
}

func (f *RSSFeed) GetLink() string {
	return f.Channel.Link
}

func (f *RSSFeed) GetDescription() string {
	return f.Channel.Description
}

func (f *RSSFeed) GetItems() []Item {
	items := make([]Item, len(f.Channel.Item))
	for i, item := range f.Channel.Item {
		items[i] = &RSSItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			Date:        item.Date,
		}
	}
	return items
}

func FetchFeed(ctx context.Context, feedURL string) (Feed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "hello-lane/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)

	for i, item := range feed.Channel.Item {
		feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
		feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
	}

	return &feed, nil
}
