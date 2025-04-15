package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/models"
)

type PostService struct {
	Repo *database.Queries
}

func (s *PostService) ListPosts(ctx context.Context, userID uuid.UUID) ([]models.Post, error) {
	dbPosts, err := s.Repo.GetPostsByUser(ctx, database.GetPostsByUserParams{
		UserID: userID,
		Limit:  10,
	})
	if err != nil {
		return nil, err
	}

	var posts []models.Post
	for _, dbPost := range dbPosts {
		posts = append(posts, models.Post{
			Title:       dbPost.Title,
			Link:        dbPost.Url,
			Description: dbPost.Description.String,
		})
	}
	return posts, nil
}
