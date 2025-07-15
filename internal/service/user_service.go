package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/models"
)

type UserService struct {
	Repo *database.Queries
}

func (s *UserService) GetUser(ctx context.Context, userName string) (models.User, error) {
	dbUser, err := s.Repo.GetUser(ctx, userName)
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		ID:   uuid.MustParse(dbUser.ID),
		Name: dbUser.Name,
	}, nil
}
