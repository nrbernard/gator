package handler

import (
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
	return applyMove(c, h.StateService.Unarchive, h.list)
}

// Trash moves the Post to the Trash, from the Archive.
func (h *ArchiveHandler) Trash(c echo.Context) error {
	return applyMove(c, trashFrom(h.StateService, service.PlaceArchive), h.list)
}

// list responds with the updated Archive.
func (h *ArchiveHandler) list(c echo.Context, userID, _ uuid.UUID) error {
	posts, err := h.PostService.Archive(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "archive", posts)
}
