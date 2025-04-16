package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

type FeedHandler struct {
	FeedService *service.FeedService
	UserService *service.UserService
}

func (h *FeedHandler) Index(c echo.Context) error {
	userName, ok := c.Get("userName").(string)
	if !ok {
		return fmt.Errorf("failed to get user name")
	}

	user, err := h.UserService.GetUser(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("failed to get user: %s", err)
	}

	feeds, err := h.FeedService.ListFeeds(c.Request().Context(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %s", err)
	}

	return c.Render(http.StatusOK, "feeds.html", map[string]interface{}{
		"Feeds": feeds,
	})
}
