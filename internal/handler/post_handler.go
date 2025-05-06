package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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

func (h *PostHandler) fetchPosts(c echo.Context, options service.SearchOptions) ([]models.Post, error) {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to get user from context")
	}

	posts, err := h.PostService.SearchPosts(c.Request().Context(), userID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %s", err)
	}

	return posts, nil
}

func formatSearchOption(statusParam string) bool {
	switch statusParam {
	case "", "unread":
		return false
	}
	return true
}

func (h *PostHandler) Index(c echo.Context) error {
	statusParam := c.QueryParam("status")
	fmt.Printf("status: %s\n", statusParam)
	posts, err := h.fetchPosts(c, service.SearchOptions{
		Query: nil,
		Read:  formatSearchOption(statusParam),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %s", err)
	}

	if statusParam == "" {
		return c.Render(http.StatusOK, "posts-index.html", map[string]interface{}{
			"Posts":    posts,
			"Selected": "unread",
		})
	} else {
		c.Render(http.StatusOK, "tabs", map[string]interface{}{
			"Selected": statusParam,
		})

		return c.Render(http.StatusOK, "oob-posts", map[string]interface{}{
			"Posts": posts,
		})
	}
}

func (h *PostHandler) Search(c echo.Context) error {
	query := c.FormValue("search")

	posts, err := h.fetchPosts(c, service.SearchOptions{
		Query: &query,
		Read:  false,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %s", err)
	}

	return c.Render(http.StatusOK, "posts-list", map[string]interface{}{
		"Posts": posts,
		"Query": query,
	})
}

func (h *PostHandler) Refresh(c echo.Context) error {
	if err := h.FeedService.ScrapeFeeds(c.Request().Context()); err != nil {
		return fmt.Errorf("failed to scrape feeds: %s", err)
	}

	posts, err := h.fetchPosts(c, service.SearchOptions{
		Query: nil,
		Read:  false,
	})
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
