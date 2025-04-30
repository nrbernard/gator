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

func (s *PostService) SearchPosts(ctx context.Context, userID uuid.UUID, query *string) ([]models.Post, error) {
	var queryStr sql.NullString
	if query != nil {
		queryStr = sql.NullString{String: *query, Valid: true}
	} else {
		queryStr = sql.NullString{Valid: false}
	}

	dbPosts, err := s.Repo.SearchPostsByUser(ctx, database.SearchPostsByUserParams{
		UserID:  userID,
		Column2: queryStr.String,
		Limit:   100,
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
