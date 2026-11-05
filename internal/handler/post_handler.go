package handler

import (
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

// Index renders the Search page; its box posts to Search as the user types.
func (h *PostHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "search-index.html", map[string]any{})
}

// Search responds with the results for the query, for htmx to swap in.
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
	return applyMove(c, h.StateService.Keep, h.result)
}

// Open Keeps a Queue Post, or marks a Reading list Post, Read; the browser
// opens its link alongside.
func (h *PostHandler) Open(c echo.Context) error {
	return applyMove(c, h.StateService.Open, h.result)
}

// ToggleRead marks a Reading list Post Read, or clears its Read mark.
func (h *PostHandler) ToggleRead(c echo.Context) error {
	return applyMove(c, h.StateService.ToggleRead, h.result)
}

// Archive moves a Reading list Post to the Archive.
func (h *PostHandler) Archive(c echo.Context) error {
	return applyMove(c, h.StateService.Archive, h.result)
}

// Unarchive moves an archived Post back to the Reading list.
func (h *PostHandler) Unarchive(c echo.Context) error {
	return applyMove(c, h.StateService.Unarchive, h.result)
}

// Restore returns a trashed Post to where it was trashed from.
func (h *PostHandler) Restore(c echo.Context) error {
	return applyMove(c, h.StateService.Restore, h.result)
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
	return applyMove(c, trashFrom(h.StateService, from), h.result)
}

// result responds with the moved Post's result, showing its new Place.
func (h *PostHandler) result(c echo.Context, userID, postID uuid.UUID) error {
	result, err := h.PostService.SearchResult(c.Request().Context(), userID, postID)
	if err != nil {
		return moveError(err)
	}
	return c.Render(http.StatusOK, "search-result", result)
}
