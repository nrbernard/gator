package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
)

type ReadPostService struct {
	Repo *database.Queries
}

func (s *ReadPostService) Save(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	if err := s.Repo.SaveReadPost(ctx, database.SaveReadPostParams{
		ID:     uuid.New().String(),
		PostID: postID.String(),
		UserID: userID.String(),
	}); err != nil {
		return err
	}

	return nil
}
