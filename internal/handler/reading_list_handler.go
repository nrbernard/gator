package handler

import (
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
	return applyMove(c, h.StateService.Open, h.list)
}

// ToggleRead marks the Post Read, or clears its Read mark.
func (h *ReadingListHandler) ToggleRead(c echo.Context) error {
	return applyMove(c, h.StateService.ToggleRead, h.list)
}

// Archive moves the Post to the Archive.
func (h *ReadingListHandler) Archive(c echo.Context) error {
	return applyMove(c, h.StateService.Archive, h.list)
}

// Trash moves the Post to the Trash, from the Reading list.
func (h *ReadingListHandler) Trash(c echo.Context) error {
	return applyMove(c, trashFrom(h.StateService, service.PlaceReadingList), h.list)
}

// list responds with the updated Reading list.
func (h *ReadingListHandler) list(c echo.Context, userID, _ uuid.UUID) error {
	posts, err := h.PostService.ReadingList(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "reading-list", posts)
}
