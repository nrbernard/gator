package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/models"
	"github.com/nrbernard/gator/internal/rss"
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
			ID:          dbFeed.ID,
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

	feedData, err := rss.FetchFeed(context.Background(), feedUrl)
	if err != nil {
		return models.Feed{}, fmt.Errorf("failed to fetch feed: %s", err)
	}

	dbFeed, err := s.Repo.CreateFeed(ctx, database.CreateFeedParams{
		ID:          uuid.New(),
		Name:        feedData.Channel.Title,
		Description: sql.NullString{String: feedData.Channel.Description, Valid: true},
		Url:         feedUrl,
		UserID:      params.UserID,
	})
	if err != nil {
		return models.Feed{}, err
	}

	if _, err := s.Repo.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: params.UserID,
		FeedID: dbFeed.ID,
	}); err != nil {
		return models.Feed{}, err
	}

	feed := models.Feed{
		ID:          dbFeed.ID,
		Name:        dbFeed.Name,
		Description: &dbFeed.Description.String,
		Url:         dbFeed.Url,
	}

	return feed, nil
}

func (s *FeedService) DeleteFeed(ctx context.Context, id uuid.UUID) error {
	if err := s.Repo.DeleteFeed(ctx, id); err != nil {
		return err
	}

	return nil
}

func parseDate(date string) time.Time {
	// Mon, 01 Jan 0001 00:00:00 +0000
	fmt.Printf("parsing date: %s\n", date)

	parsed, err := time.Parse(time.RFC1123Z, date)
	if err != nil {
		fmt.Printf("failed to parse date: %s\n", err)
		return time.Time{}
	}

	return parsed
}

func (s *FeedService) ScrapeFeeds(ctx context.Context) error {
	feed, err := s.Repo.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %s", err)
	}

	if err := s.Repo.MarkFeedAsFetched(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %s", err)
	}

	feedData, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %s", err)
	}

	for _, item := range feedData.Channel.Item {
		if _, err := s.Repo.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: parseDate(item.PubDate),
			FeedID:      feed.ID,
		}); err != nil {
			if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				return fmt.Errorf("post with URL %s already exists", item.Link)
			}
		}

		fmt.Printf("created post with URL %s\n", item.Link)
	}

	return nil
}
