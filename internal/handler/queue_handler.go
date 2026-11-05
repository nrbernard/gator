package handler

import (
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

// Index renders the Queue page, which records the cutoff of the moment it
// loaded.
func (h *QueueHandler) Index(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	queue, err := h.PostService.Queue(c.Request().Context(), userID, h.PostService.Cutoff())
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
	return h.card(c, userID, uuid.Nil)
}

// card responds with the top of the Queue at the cutoff the page sent, or, if
// it sent none, a page loading now's.
func (h *QueueHandler) card(c echo.Context, userID, _ uuid.UUID) error {
	cutoff, ok := requestCutoff(c)
	if !ok {
		cutoff = h.PostService.Cutoff()
	}
	queue, err := h.PostService.Queue(c.Request().Context(), userID, cutoff)
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
	count, err := h.PostService.CountQueueAtOrBefore(c.Request().Context(), userID, cutoff)
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
	if err := h.StateService.TrashAll(ctx, userID, cutoff); err != nil {
		return err
	}
	queue, err := h.PostService.Queue(ctx, userID, cutoff)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue", queue)
}

// requestCutoff is the cutoff the Queue page sent, if it sent a valid one.
func requestCutoff(c echo.Context) (time.Time, bool) {
	cutoff, err := time.Parse(time.RFC3339Nano, c.FormValue("cutoff"))
	return cutoff, err == nil
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
	if err := h.FeedService.ScrapeFeeds(ctx); err != nil {
		return fmt.Errorf("failed to scrape feeds: %w", err)
	}
	// What the scrape fetched in the current second counts once it ends;
	// wait for that, so the chip counts everything the check fetched.
	if err := h.PostService.AwaitSecond(ctx); err != nil {
		return err
	}
	queue, err := h.PostService.Queue(ctx, userID, cutoff)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "queue-new", queue)
}

// Keep moves the Post to the Reading list and responds with the next card.
func (h *QueueHandler) Keep(c echo.Context) error {
	return applyMove(c, h.StateService.Keep, h.card)
}

// Open Keeps the Post marked Read, as the user clicks through to its site, and
// responds with the next card.
func (h *QueueHandler) Open(c echo.Context) error {
	return applyMove(c, h.StateService.Open, h.card)
}

// Trash moves the Post to the Trash and responds with the next card.
func (h *QueueHandler) Trash(c echo.Context) error {
	return applyMove(c, trashFrom(h.StateService, service.PlaceQueue), h.card)
}
