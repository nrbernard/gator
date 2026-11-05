package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

// ReadingListHandler serves the Reading list.
type ReadingListHandler struct {
	PostService *service.PostService
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
