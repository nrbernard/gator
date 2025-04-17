package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/models"
)

type FeedService struct {
	Repo *database.Queries
}

type CreateFeedParams struct {
	Name   string
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
			ID:   dbFeed.ID,
			Name: dbFeed.Name,
			Url:  dbFeed.Url,
		})
	}
	return feeds, nil
}

func (s *FeedService) CreateFeed(ctx context.Context, params CreateFeedParams) (models.Feed, error) {
	_, err := s.Repo.GetFeedByUrl(ctx, params.Url)
	if err == nil {
		return models.Feed{}, fmt.Errorf("a feed with URL %s already exists", params.Url)
	}

	dbFeed, err := s.Repo.CreateFeed(ctx, database.CreateFeedParams{
		ID:     uuid.New(),
		Name:   params.Name,
		Url:    params.Url,
		UserID: params.UserID,
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
		ID:   dbFeed.ID,
		Name: dbFeed.Name,
		Url:  dbFeed.Url,
	}

	return feed, nil
}

func (s *FeedService) DeleteFeed(ctx context.Context, id uuid.UUID) error {
	if err := s.Repo.DeleteFeed(ctx, id); err != nil {
		return err
	}

	return nil
}
