package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/models"
)

type PostService struct {
	Repo *database.Queries
}

type SearchOptions struct {
	Query  *string
	Unread bool
	Saved  bool
}

func (s *PostService) SearchPosts(ctx context.Context, userID uuid.UUID, options SearchOptions) ([]models.Post, error) {
	var queryStr sql.NullString
	if options.Query != nil {
		queryStr = sql.NullString{String: *options.Query, Valid: true}
	} else {
		queryStr = sql.NullString{Valid: false}
	}

	dbPosts, err := s.Repo.SearchPostsByUser(ctx, database.SearchPostsByUserParams{
		UserID:         userID,
		SearchText:     queryStr.String,
		FilterByUnread: options.Unread,
		FilterBySaved:  options.Saved,
		LimitCount:     100,
	})
	if err != nil {
		return nil, err
	}

	posts := make([]models.Post, 0, len(dbPosts))
	for _, dbPost := range dbPosts {
		posts = append(posts, models.Post{
			ID:          dbPost.ID,
			Title:       dbPost.Title,
			Link:        dbPost.Url,
			Description: dbPost.Description.String,
			PublishedAt: dbPost.PublishedAt,
			FeedID:      dbPost.FeedID,
			FeedName:    dbPost.FeedName,
			IsSaved:     dbPost.SavedAt.Valid,
		})
	}

	return posts, nil
}
