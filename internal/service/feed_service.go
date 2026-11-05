package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/feedparser"
	"github.com/nrbernard/gator/internal/models"
)

type FeedService struct {
	Repo *database.Queries
}

type CreateFeedParams struct {
	Url    string
	UserID uuid.UUID
}

func (s *FeedService) ListFeeds(ctx context.Context, userID uuid.UUID) ([]models.Feed, error) {
	dbFeeds, err := s.Repo.GetFeeds(ctx)
	if err != nil {
		return nil, err
	}

	var feeds []models.Feed
	for _, dbFeed := range dbFeeds {
		feeds = append(feeds, models.Feed{
			ID:               uuid.MustParse(dbFeed.ID),
			Name:             dbFeed.Name,
			Description:      &dbFeed.Description.String,
			Url:              dbFeed.Url,
			LastFetchedAt:    timePtr(dbFeed.LastFetchedAt),
			LastFetchError:   dbFeed.LastFetchError.String,
			LastFetchErrorAt: timePtr(dbFeed.LastFetchErrorAt),
		})
	}
	return feeds, nil
}

func timePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func (s *FeedService) CreateFeed(ctx context.Context, params CreateFeedParams) (models.Feed, error) {
	feedUrl := params.Url
	existing, err := s.Repo.GetFeedByUrl(ctx, feedUrl)
	if err == nil {
		if !existing.RemovedAt.Valid {
			return models.Feed{}, fmt.Errorf("a feed with URL %s already exists", feedUrl)
		}
		// Feed URLs are unique, so re-adding a removed Feed un-removes it.
		dbFeed, err := s.Repo.UnremoveFeed(ctx, existing.ID)
		if err != nil {
			return models.Feed{}, fmt.Errorf("unremove feed: %w", err)
		}
		return models.Feed{
			ID:          uuid.MustParse(dbFeed.ID),
			Name:        dbFeed.Name,
			Description: &dbFeed.Description.String,
			Url:         dbFeed.Url,
		}, nil
	}

	feedData, err := feedparser.FetchFeed(context.Background(), feedUrl)
	if err != nil {
		return models.Feed{}, fmt.Errorf("failed to fetch feed: %s", err)
	}

	dbFeed, err := s.Repo.CreateFeed(ctx, database.CreateFeedParams{
		ID:          uuid.New().String(),
		Name:        feedData.GetTitle(),
		Description: sql.NullString{String: feedData.GetDescription(), Valid: true},
		Url:         feedUrl,
		UserID:      params.UserID.String(),
	})
	if err != nil {
		return models.Feed{}, err
	}

	if _, err := s.Repo.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New().String(),
		UserID: params.UserID.String(),
		FeedID: dbFeed.ID,
	}); err != nil {
		return models.Feed{}, err
	}

	feed := models.Feed{
		ID:          uuid.MustParse(dbFeed.ID),
		Name:        dbFeed.Name,
		Description: &dbFeed.Description.String,
		Url:         dbFeed.Url,
	}

	return feed, nil
}

// ErrFeedNotFound is returned when removing a Feed that doesn't exist or is
// already removed.
var ErrFeedNotFound = errors.New("feed not found")

// RemoveFeed stops following the Feed: it is no longer fetched or listed and
// its Queue Posts leave the Queue, but its row stays, so its Reading list,
// Archive, and Trash Posts keep their Feed.
func (s *FeedService) RemoveFeed(ctx context.Context, id uuid.UUID) error {
	n, err := s.Repo.RemoveFeed(ctx, id.String())
	if err != nil {
		return fmt.Errorf("remove feed: %w", err)
	}
	if n == 0 {
		return ErrFeedNotFound
	}
	return nil
}

func (s *FeedService) ScrapeFeeds(ctx context.Context) error {
	// Change from 24 hours to 1 hour to respect the "once per hour" limit
	cutoff := time.Now().Add(-1 * time.Hour)
	feeds, err := s.Repo.GetFeedsToFetch(ctx, sql.NullTime{Time: cutoff, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get feeds: %s", err)
	}

	if len(feeds) == 0 {
		fmt.Println("no feeds to fetch")
		return nil
	}

	for _, feed := range feeds {
		fmt.Printf("fetching feed: %s\n", feed.Name)

		// Extract conditional headers from database
		var etag, lastModified *string
		if feed.Etag.Valid && feed.Etag.String != "" {
			etag = &feed.Etag.String
		}
		if feed.LastModified.Valid && feed.LastModified.String != "" {
			lastModified = &feed.LastModified.String
		}

		// Use conditional request
		result, err := feedparser.FetchFeedWithConditionals(context.Background(), feed.Url, etag, lastModified)
		if err != nil {
			// Handle rate limiting (429) with exponential backoff
			if strings.Contains(err.Error(), "status code: 429") {
				fmt.Printf("Rate limited for feed %s, skipping for now\n", feed.Name)
				continue
			}
			// Record the failure on this Feed and carry on with the rest.
			fmt.Printf("failed to fetch feed %s: %s\n", feed.Name, err)
			if err := s.Repo.RecordFeedFetchError(ctx, database.RecordFeedFetchErrorParams{
				LastFetchError: sql.NullString{String: err.Error(), Valid: true},
				ID:             feed.ID,
			}); err != nil {
				return fmt.Errorf("failed to record fetch error for feed %s: %s", feed.Name, err)
			}
			continue
		}

		// A 304 Not Modified counts as a fetch for the hourly limit.
		if result.NotModified {
			fmt.Printf("Feed %s not modified, updating headers only\n", feed.Name)
			if err := s.Repo.UpdateFeedConditionalHeaders(context.Background(), database.UpdateFeedConditionalHeadersParams{
				Etag:         sql.NullString{String: result.ETag, Valid: result.ETag != ""},
				LastModified: sql.NullString{String: result.LastModified, Valid: result.LastModified != ""},
				ID:           feed.ID,
			}); err != nil {
				return fmt.Errorf("failed to update feed headers: %s", err)
			}
			continue
		}

		// Handle successful response with new content
		if result.Feed == nil {
			fmt.Printf("No feed data received for %s\n", feed.Name)
			continue
		}

		// Update conditional headers and fetch timestamp
		if err := s.Repo.UpdateFeedConditionalHeaders(context.Background(), database.UpdateFeedConditionalHeadersParams{
			Etag:         sql.NullString{String: result.ETag, Valid: result.ETag != ""},
			LastModified: sql.NullString{String: result.LastModified, Valid: result.LastModified != ""},
			ID:           feed.ID,
		}); err != nil {
			return fmt.Errorf("failed to update feed headers: %s", err)
		}

		// Process new posts
		for _, item := range result.Feed.GetItems() {
			post, err := s.Repo.CreatePost(ctx, database.CreatePostParams{
				ID:          uuid.New().String(),
				Title:       item.GetTitle(),
				Url:         item.GetLink(),
				Description: sql.NullString{String: *item.GetDescription(), Valid: item.GetDescription() != nil},
				PublishedAt: item.GetDate(),
				FeedID:      feed.ID,
			})
			if err != nil {
				if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					fmt.Printf("post already exists: %s\n", item.GetLink())
				}
			} else {
				fmt.Printf("created post with URL %s\n", post.Url)
			}
		}
	}

	return nil
}
