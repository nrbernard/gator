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
	err := s.Repo.SaveSavedPost(ctx, database.SaveSavedPostParams{
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
	err := s.Repo.DeleteSavedPost(ctx, database.DeleteSavedPostParams{
		PostID: postID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}
