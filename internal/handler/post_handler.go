package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

// PostHandler serves search across every Place.
type PostHandler struct {
	PostService *service.PostService
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
