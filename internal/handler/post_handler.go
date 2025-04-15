package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

type PostHandler struct {
	PostService *service.PostService
	UserService *service.UserService
}

func (h *PostHandler) Index(c echo.Context) error {
	userName, ok := c.Get("userName").(string)
	if !ok {
		return fmt.Errorf("failed to get user name")
	}

	user, err := h.UserService.GetUser(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("failed to get user: %s", err)
	}

	posts, err := h.PostService.ListPosts(c.Request().Context(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get posts: %s", err)
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Posts": posts,
	})
}
