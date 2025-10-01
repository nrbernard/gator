package service

import (
	"context"
	"database/sql"
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
			ID:          uuid.MustParse(dbFeed.ID),
			Name:        dbFeed.Name,
			Description: &dbFeed.Description.String,
			Url:         dbFeed.Url,
		})
	}
	return feeds, nil
}

func (s *FeedService) CreateFeed(ctx context.Context, params CreateFeedParams) (models.Feed, error) {
	feedUrl := params.Url
	_, err := s.Repo.GetFeedByUrl(ctx, feedUrl)
	if err == nil {
		return models.Feed{}, fmt.Errorf("a feed with URL %s already exists", feedUrl)
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

func (s *FeedService) DeleteFeed(ctx context.Context, id uuid.UUID) error {
	if err := s.Repo.DeleteFeed(ctx, id.String()); err != nil {
		return err
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
			return fmt.Errorf("failed to fetch feed: %s", err)
		}

		// Handle 304 Not Modified response
		if result.NotModified {
			fmt.Printf("Feed %s not modified, updating headers only\n", feed.Name)
			if err := s.Repo.UpdateFeedConditionalHeadersNoFetch(context.Background(), database.UpdateFeedConditionalHeadersNoFetchParams{
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
