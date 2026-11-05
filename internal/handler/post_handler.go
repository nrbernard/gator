package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

// PostHandler serves search across every Place, and the actions on a search
// result. Each action responds with the result, re-rendered with the Post's
// new Place, for htmx to swap in place.
type PostHandler struct {
	PostService  *service.PostService
	StateService *service.PostStateService
}

func (h *PostHandler) Search(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	query := c.FormValue("search")

	var posts []service.SearchResult
	if query != "" {
		posts, err = h.PostService.SearchPosts(c.Request().Context(), userID, query)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, "search-results", map[string]any{
		"Posts": posts,
		"Query": query,
	})
}

// Keep moves a Queue Post to the Reading list.
func (h *PostHandler) Keep(c echo.Context) error {
	return h.act(c, h.StateService.Keep)
}

// Open Keeps a Queue Post, or marks a Reading list Post, Read; the browser
// opens its link alongside.
func (h *PostHandler) Open(c echo.Context) error {
	return h.act(c, h.StateService.Open)
}

// ToggleRead marks a Reading list Post Read, or clears its Read mark.
func (h *PostHandler) ToggleRead(c echo.Context) error {
	return h.act(c, h.StateService.ToggleRead)
}

// Archive moves a Reading list Post to the Archive.
func (h *PostHandler) Archive(c echo.Context) error {
	return h.act(c, h.StateService.Archive)
}

// Unarchive moves an archived Post back to the Reading list.
func (h *PostHandler) Unarchive(c echo.Context) error {
	return h.act(c, h.StateService.Unarchive)
}

// Restore returns a trashed Post to where it was trashed from.
func (h *PostHandler) Restore(c echo.Context) error {
	return h.act(c, h.StateService.Restore)
}

// Trash moves the Post to the Trash from the Place the result showed it in,
// sent as from, so a stale result can't trash a Post that has since moved.
func (h *PostHandler) Trash(c echo.Context) error {
	from := service.Place(c.FormValue("from"))
	switch from {
	case service.PlaceQueue, service.PlaceReadingList, service.PlaceArchive:
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "missing or invalid from")
	}
	return h.act(c, func(ctx context.Context, userID, postID uuid.UUID) error {
		return h.StateService.Trash(ctx, userID, postID, from)
	})
}

func (h *PostHandler) act(c echo.Context, move func(context.Context, uuid.UUID, uuid.UUID) error) error {
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
	result, err := h.PostService.SearchResult(ctx, userID, postID)
	if err != nil {
		return moveError(err)
	}
	return c.Render(http.StatusOK, "search-result", result)
}
