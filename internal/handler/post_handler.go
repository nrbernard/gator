package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/models"
	"github.com/nrbernard/gator/internal/service"
)

type PostHandler struct {
	PostService *service.PostService
	UserService *service.UserService
	FeedService *service.FeedService
}

func NewPostHandler(postService *service.PostService, userService *service.UserService, feedService *service.FeedService) (*PostHandler, error) {
	if postService == nil || userService == nil || feedService == nil {
		return nil, fmt.Errorf("all services must be provided")
	}
	return &PostHandler{
		PostService: postService,
		UserService: userService,
		FeedService: feedService,
	}, nil
}

func (h *PostHandler) fetchPosts(c echo.Context) ([]models.Post, error) {
	userName, ok := c.Get("userName").(string)
	if !ok {
		return nil, fmt.Errorf("failed to get user name")
	}

	user, err := h.UserService.GetUser(context.Background(), userName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %s", err)
	}

	posts, err := h.PostService.ListPosts(c.Request().Context(), user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %s", err)
	}

	return posts, nil
}

func (h *PostHandler) Index(c echo.Context) error {
	posts, err := h.fetchPosts(c)
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %s", err)
	}

	return c.Render(http.StatusOK, "posts-index.html", map[string]interface{}{
		"Posts": posts,
	})
}

func (h *PostHandler) Refresh(c echo.Context) error {
	if err := h.FeedService.ScrapeFeeds(c.Request().Context()); err != nil {
		return fmt.Errorf("failed to scrape feeds: %s", err)
	}

	posts, err := h.fetchPosts(c)
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %s", err)
	}

	c.Render(http.StatusOK, "posts-refresh", map[string]interface{}{
		"LastRefresh": time.Now().Format(time.RFC3339),
	})

	return c.Render(http.StatusOK, "oob-posts", map[string]interface{}{
		"Posts": posts,
	})
}
