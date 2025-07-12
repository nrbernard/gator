package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

	feedData, err := feedparser.FetchFeed(context.Background(), feedUrl)
	if err != nil {
		return models.Feed{}, fmt.Errorf("failed to fetch feed: %s", err)
	}

	dbFeed, err := s.Repo.CreateFeed(ctx, database.CreateFeedParams{
		ID:          uuid.New(),
		Name:        feedData.GetTitle(),
		Description: sql.NullString{String: feedData.GetDescription(), Valid: true},
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

func (s *FeedService) ScrapeFeeds(ctx context.Context) error {
	feeds, err := s.Repo.GetFeedsToFetch(ctx, "24 hours")
	if err != nil {
		return fmt.Errorf("failed to get feeds: %s", err)
	}

	if len(feeds) == 0 {
		fmt.Println("no feeds to fetch")
		return nil
	}

	for _, feed := range feeds {
		fmt.Printf("fetching feed: %s\n", feed.Name)

		feedData, err := feedparser.FetchFeed(context.Background(), feed.Url)
		if err != nil {
			return fmt.Errorf("failed to fetch feed: %s", err)
		}

		if err := s.Repo.MarkFeedAsFetched(context.Background(), feed.ID); err != nil {
			return fmt.Errorf("failed to mark feed as fetched: %s", err)
		}

		for _, item := range feedData.GetItems() {
			post, err := s.Repo.CreatePost(ctx, database.CreatePostParams{
				ID:          uuid.New(),
				Title:       item.GetTitle(),
				Url:         item.GetLink(),
				Description: sql.NullString{String: *item.GetDescription(), Valid: item.GetDescription() != nil},
				PublishedAt: item.GetDate(),
				FeedID:      feed.ID,
			})
			if err != nil {
				if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					return fmt.Errorf("failed to create post: %s", err.Error())
				}
			} else {
				fmt.Printf("created post with URL %s\n", post.Url)
			}
		}
	}

	return nil
}
