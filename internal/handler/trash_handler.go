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
	return applyMove(c, h.StateService.Restore, h.list)
}

// list responds with the updated Trash.
func (h *TrashHandler) list(c echo.Context, userID, _ uuid.UUID) error {
	posts, err := h.PostService.Trash(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "trash", posts)
}
