package feedparser

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
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
	GetDate() time.Time
}

type rssXML struct {
	Channel struct {
		Title       string       `xml:"title"`
		Link        string       `xml:"link"`
		Description string       `xml:"description"`
		Item        []rssItemXML `xml:"item"`
	} `xml:"channel"`
}

type rssItemXML struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Date        string `xml:"pubDate"`
}

type atomXML struct {
	Title string `xml:"title"`
	Links []struct {
		Rel  string `xml:"rel,attr"`
		URL  string `xml:"href,attr"`
		Type string `xml:"type,attr"`
	} `xml:"link"`
	Description string        `xml:"description"`
	Item        []atomItemXML `xml:"entry"`
}

type atomItemXML struct {
	Title string `xml:"title"`
	Links []struct {
		Rel  string `xml:"rel,attr"`
		URL  string `xml:"href,attr"`
		Type string `xml:"type,attr"`
	} `xml:"link"`
	Content struct {
		Type string `xml:"type,attr"`
		Data string `xml:",chardata"`
	} `xml:"content"`
	Date string `xml:"updated"`
}

type RSSFeed struct {
	title       string
	link        string
	description string
	items       []*RSSItem
}

type RSSItem struct {
	title       string
	link        string
	description string
	date        time.Time
}

type AtomFeed struct {
	title       string
	link        string
	description string
	items       []*AtomItem
}

type AtomItem struct {
	title       string
	link        string
	description string
	date        time.Time
}

func (f *RSSFeed) GetTitle() string {
	return f.title
}

func (f *RSSFeed) GetLink() string {
	return f.link
}

func (f *RSSFeed) GetDescription() string {
	return f.description
}

func (f *RSSFeed) GetItems() []Item {
	items := make([]Item, len(f.items))
	for i, item := range f.items {
		items[i] = item
	}
	return items
}

func (i *RSSItem) GetTitle() string {
	return i.title
}

func (i *RSSItem) GetLink() string {
	return i.link
}

func (i *RSSItem) GetDescription() string {
	return i.description
}

func (i *RSSItem) GetDate() time.Time {
	return i.date
}

func (f *AtomFeed) GetTitle() string {
	return f.title
}

func (f *AtomFeed) GetLink() string {
	return f.link
}

func (f *AtomFeed) GetDescription() string {
	return f.description
}

func (f *AtomFeed) GetItems() []Item {
	items := make([]Item, len(f.items))
	for i, item := range f.items {
		items[i] = item
	}
	return items
}

func (i *AtomItem) GetTitle() string {
	return i.title
}

func (i *AtomItem) GetLink() string {
	return i.link
}

func (i *AtomItem) GetDescription() string {
	return i.description
}

func (i *AtomItem) GetDate() time.Time {
	return i.date
}

func stripHTMLTags(htmlContent string) string {
	// First unescape HTML entities
	text := html.UnescapeString(htmlContent)

	// Remove HTML tags
	var result strings.Builder
	inTag := false
	for _, char := range text {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	// Clean up whitespace
	return strings.TrimSpace(result.String())
}

func parseDate(date string) (time.Time, error) {
	// Try RFC1123Z first (with timezone offset)
	parsed, err := time.Parse(time.RFC1123Z, date)
	if err == nil {
		return parsed, nil
	}

	// Try RFC1123 (with GMT)
	parsed, err = time.Parse(time.RFC1123, date)
	if err == nil {
		return parsed, nil
	}

	// Try ISO8601 format
	parsed, err = time.Parse(time.RFC3339, date)
	if err == nil {
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("failed to parse date: %s", date)
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

	if strings.Contains(resp.Header.Get("Content-Type"), "atom") {
		return parseAtomFeed(body)
	} else {
		return parseRSSFeed(body)
	}
}

func parseAtomFeed(body []byte) (Feed, error) {
	var xmlFeed atomXML
	if err := xml.Unmarshal(body, &xmlFeed); err != nil {
		return nil, err
	}

	feed := &AtomFeed{
		title:       html.UnescapeString(xmlFeed.Title),
		description: html.UnescapeString(xmlFeed.Description),
		items:       make([]*AtomItem, len(xmlFeed.Item)),
	}

	for _, link := range xmlFeed.Links {
		if link.Rel == "self" {
			feed.link = link.URL
			break
		}
	}

	for i, item := range xmlFeed.Item {
		parsedDate, err := parseDate(item.Date)
		if err != nil {
			return nil, err
		}

		feed.items[i] = &AtomItem{
			title:       html.UnescapeString(item.Title),
			description: stripHTMLTags(item.Content.Data),
			date:        parsedDate,
		}

		for _, link := range item.Links {
			if link.Rel == "alternate" {
				feed.items[i].link = link.URL
				break
			}
		}
	}

	return feed, nil
}

func parseRSSFeed(body []byte) (Feed, error) {
	var xmlFeed rssXML
	if err := xml.Unmarshal(body, &xmlFeed); err != nil {
		return nil, err
	}

	feed := &RSSFeed{
		title:       html.UnescapeString(xmlFeed.Channel.Title),
		link:        xmlFeed.Channel.Link,
		description: html.UnescapeString(xmlFeed.Channel.Description),
		items:       make([]*RSSItem, len(xmlFeed.Channel.Item)),
	}

	for i, item := range xmlFeed.Channel.Item {
		parsedDate, err := parseDate(item.Date)
		if err != nil {
			return nil, err
		}

		feed.items[i] = &RSSItem{
			title:       html.UnescapeString(item.Title),
			link:        item.Link,
			description: html.UnescapeString(item.Description),
			date:        parsedDate,
		}
	}

	return feed, nil
}
