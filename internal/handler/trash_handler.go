package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

// TrashHandler serves the Trash and Restore. Restore responds with the updated
// list for htmx to swap in place.
type TrashHandler struct {
	PostService  *service.PostService
	StateService *service.PostStateService
}

func (h *TrashHandler) Index(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	posts, err := h.PostService.Trash(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "trash-index.html", map[string]any{"Posts": posts})
}

// Restore returns the Post to where it was trashed from.
func (h *TrashHandler) Restore(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid post ID")
	}
	ctx := c.Request().Context()
	if err := h.StateService.Restore(ctx, userID, postID); err != nil {
		return moveError(err)
	}
	posts, err := h.PostService.Trash(ctx, userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "trash", posts)
}
