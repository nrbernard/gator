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

// QueueHandler serves the Queue, the home screen, and its triage actions.
type QueueHandler struct {
	PostService  *service.PostService
	StateService *service.PostStateService
	FeedService  *service.FeedService
}

func currentUserID(c echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("failed to get user from context")
	}
	return userID, nil
}

func (h *QueueHandler) Index(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	queue, err := h.PostService.Queue(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue-index.html", map[string]any{"Queue": queue})
}

// Refresh fetches the Feeds due a fetch and responds with the top of the Queue.
func (h *QueueHandler) Refresh(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if err := h.FeedService.ScrapeFeeds(ctx); err != nil {
		return fmt.Errorf("failed to scrape feeds: %w", err)
	}
	queue, err := h.PostService.Queue(ctx, userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue", queue)
}

// Keep moves the Post to the Reading list and responds with the next card.
func (h *QueueHandler) Keep(c echo.Context) error {
	return h.triage(c, h.StateService.Keep)
}

// Open Keeps the Post marked Read, as the user clicks through to its site, and
// responds with the next card.
func (h *QueueHandler) Open(c echo.Context) error {
	return h.triage(c, h.StateService.Open)
}

// Trash moves the Post to the Trash and responds with the next card.
func (h *QueueHandler) Trash(c echo.Context) error {
	return h.triage(c, func(ctx context.Context, userID, postID uuid.UUID) error {
		return h.StateService.Trash(ctx, userID, postID, service.PlaceQueue)
	})
}

func (h *QueueHandler) triage(c echo.Context, move func(context.Context, uuid.UUID, uuid.UUID) error) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post ID")
	}
	ctx := c.Request().Context()
	if err := move(ctx, userID, postID); err != nil {
		return moveError(err)
	}
	queue, err := h.PostService.Queue(ctx, userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue", queue)
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
