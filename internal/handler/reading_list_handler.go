package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

// ReadingListHandler serves the Reading list and its per-Post actions. Each
// action responds with the updated list for htmx to swap in place.
type ReadingListHandler struct {
	PostService  *service.PostService
	StateService *service.PostStateService
}

func (h *ReadingListHandler) Index(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	posts, err := h.PostService.ReadingList(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "reading-list-index.html", map[string]any{"Posts": posts})
}

// Open marks the Post Read; the browser opens its link alongside.
func (h *ReadingListHandler) Open(c echo.Context) error {
	return h.act(c, h.StateService.Open)
}

// ToggleRead marks the Post Read, or clears its Read mark.
func (h *ReadingListHandler) ToggleRead(c echo.Context) error {
	return h.act(c, h.StateService.ToggleRead)
}

// Archive moves the Post to the Archive.
func (h *ReadingListHandler) Archive(c echo.Context) error {
	return h.act(c, h.StateService.Archive)
}

// Trash moves the Post to the Trash, from the Reading list.
func (h *ReadingListHandler) Trash(c echo.Context) error {
	return h.act(c, func(ctx context.Context, userID, postID uuid.UUID) error {
		return h.StateService.Trash(ctx, userID, postID, service.PlaceReadingList)
	})
}

func (h *ReadingListHandler) act(c echo.Context, move func(context.Context, uuid.UUID, uuid.UUID) error) error {
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
	posts, err := h.PostService.ReadingList(ctx, userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "reading-list", posts)
}
