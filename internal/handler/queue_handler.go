package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/models"
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

// Index renders the Queue page, which records the moment it loaded as its
// cutoff.
func (h *QueueHandler) Index(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	queue, err := h.queue(c.Request().Context(), userID, h.PostService.Cutoff())
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue-index.html", map[string]any{"Queue": queue})
}

// Card responds with the top of the Queue, keeping the page's cutoff.
func (h *QueueHandler) Card(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	cutoff, ok := requestCutoff(c)
	if !ok {
		cutoff = h.PostService.Cutoff()
	}
	queue, err := h.queue(c.Request().Context(), userID, cutoff)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue", queue)
}

// ConfirmTrashAll asks before trashing the Queue, with the number of Posts
// trash all would trash. It moves nothing.
func (h *QueueHandler) ConfirmTrashAll(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	cutoff, ok := requestCutoff(c)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "missing or invalid cutoff")
	}
	count, err := h.PostService.CountQueueCreatedBy(c.Request().Context(), userID, cutoff)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue-trash-all", models.Queue{Count: count, Cutoff: cutoff})
}

// TrashAll trashes the Queue Posts fetched by the page's cutoff, sparing any
// fetched since, and responds with the top of what's left.
func (h *QueueHandler) TrashAll(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	cutoff, ok := requestCutoff(c)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "missing or invalid cutoff")
	}
	ctx := c.Request().Context()
	if _, err := h.StateService.TrashAll(ctx, userID, cutoff); err != nil {
		return err
	}
	queue, err := h.queue(ctx, userID, cutoff)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue", queue)
}

// waitPastSecond waits until the second after t's. gator records when it
// fetched a Post to the second, and the Queue shows Posts fetched in its
// cutoff's second, so a fetch in that second would change the card.
func waitPastSecond(ctx context.Context, t time.Time) error {
	wait := time.Until(t.Truncate(time.Second).Add(time.Second))
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// requestCutoff is the cutoff the Queue page sent, if it sent a valid one.
func requestCutoff(c echo.Context) (time.Time, bool) {
	cutoff, err := time.Parse(time.RFC3339Nano, c.FormValue("cutoff"))
	return cutoff, err == nil
}

// queue is the top of the user's Queue, for a page whose cutoff is cutoff.
func (h *QueueHandler) queue(ctx context.Context, userID uuid.UUID, cutoff time.Time) (models.Queue, error) {
	return h.PostService.Queue(ctx, userID, cutoff)
}

// Check fetches the Feeds due a fetch and responds with the chip. What it
// fetched waits behind the chip, so the card doesn't change. The Queue page
// sends it when it loads, and check now sends it on demand.
func (h *QueueHandler) Check(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	cutoff, ok := requestCutoff(c)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "missing or invalid cutoff")
	}
	ctx := c.Request().Context()
	if err := waitPastSecond(ctx, cutoff); err != nil {
		return err
	}
	if err := h.FeedService.ScrapeFeeds(ctx); err != nil {
		return fmt.Errorf("failed to scrape feeds: %w", err)
	}
	queue, err := h.queue(ctx, userID, cutoff)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue-new", queue)
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
	cutoff, ok := requestCutoff(c)
	if !ok {
		cutoff = h.PostService.Cutoff()
	}
	queue, err := h.queue(ctx, userID, cutoff)
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
