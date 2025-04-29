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
	var dbPosts []database.Post
	var err error

	if query == nil {
		dbPosts, err = s.Repo.GetPostsByUser(ctx, database.GetPostsByUserParams{
			UserID: userID,
			Limit:  100,
		})
	} else {
		dbPosts, err = s.Repo.SearchPosts(ctx, database.SearchPostsParams{
			UserID:  userID,
			Column2: sql.NullString{String: *query, Valid: true},
			Limit:   100,
		})
	}
	if err != nil {
		return nil, err
	}

	posts := make([]models.Post, 0, len(dbPosts))
	for _, dbPost := range dbPosts {
		posts = append(posts, models.Post{
			Title:       dbPost.Title,
			Link:        dbPost.Url,
			Description: dbPost.Description.String,
		})
	}
	return posts, nil
}
