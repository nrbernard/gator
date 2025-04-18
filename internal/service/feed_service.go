package service

import (
	"context"
	"database/sql"
	"fmt"

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
