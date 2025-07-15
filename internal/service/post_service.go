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
		UserID:         userID.String(),
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
		isRead := dbPost.ReadAt.Valid
		if options.Saved {
			isRead = false
		}

		posts = append(posts, models.Post{
			ID:          uuid.MustParse(dbPost.ID),
			Title:       dbPost.Title,
			Link:        dbPost.Url,
			Description: dbPost.Description.String,
			PublishedAt: dbPost.PublishedAt,
			FeedID:      uuid.MustParse(dbPost.FeedID),
			FeedName:    dbPost.FeedName,
			IsSaved:     dbPost.SavedAt.Valid,
			IsRead:      isRead,
		})
	}

	return posts, nil
}
