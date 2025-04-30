package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
)

type SavedPostService struct {
	Repo *database.Queries
}

func (s *SavedPostService) SavePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	err := s.Repo.SavePost(ctx, database.SavePostParams{
		ID:     uuid.New(),
		PostID: postID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *SavedPostService) UnsavePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	err := s.Repo.UnsavePost(ctx, database.UnsavePostParams{
		PostID: postID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}
