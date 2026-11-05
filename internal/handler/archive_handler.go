package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

// ArchiveHandler serves the Archive and its per-Post actions. Each action
// responds with the updated list for htmx to swap in place.
type ArchiveHandler struct {
	PostService  *service.PostService
	StateService *service.PostStateService
}

func (h *ArchiveHandler) Index(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	posts, err := h.PostService.Archive(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "archive-index.html", map[string]any{"Posts": posts})
}

// Unarchive moves the Post back to the Reading list, to reread it.
func (h *ArchiveHandler) Unarchive(c echo.Context) error {
	return h.act(c, h.StateService.Unarchive)
}

// Trash moves the Post to the Trash, from the Archive.
func (h *ArchiveHandler) Trash(c echo.Context) error {
	return h.act(c, func(ctx context.Context, userID, postID uuid.UUID) error {
		return h.StateService.Trash(ctx, userID, postID, service.PlaceArchive)
	})
}

func (h *ArchiveHandler) act(c echo.Context, move func(context.Context, uuid.UUID, uuid.UUID) error) error {
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
	posts, err := h.PostService.Archive(ctx, userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "archive", posts)
}
