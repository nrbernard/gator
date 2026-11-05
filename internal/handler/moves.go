package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

// postMove is one of PostStateService's moves of a Post for a user.
type postMove func(ctx context.Context, userID, postID uuid.UUID) error

// moveResponse responds to a move of the Post postID, once it has moved.
type moveResponse func(c echo.Context, userID, postID uuid.UUID) error

// trashFrom is the move that trashes a Post from the Place the screen showed
// it in.
func trashFrom(states *service.PostStateService, from service.Place) postMove {
	return func(ctx context.Context, userID, postID uuid.UUID) error {
		return states.Trash(ctx, userID, postID, from)
	}
}

// applyMove makes the move on the route's :id Post for the current user, then
// responds with respond. A rejected move responds with its HTTP status.
func applyMove(c echo.Context, move postMove, respond moveResponse) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post ID")
	}
	if err := move(c.Request().Context(), userID, postID); err != nil {
		return moveError(err)
	}
	return respond(c, userID, postID)
}

func currentUserID(c echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("failed to get user from context")
	}
	return userID, nil
}

// moveError maps a rejected move to its HTTP status.
func moveError(err error) error {
	switch {
	case errors.Is(err, service.ErrPostNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error()).SetInternal(err)
	case errors.Is(err, service.ErrInvalidMove):
		return echo.NewHTTPError(http.StatusConflict, err.Error()).SetInternal(err)
	default:
		return err
	}
}
