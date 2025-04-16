package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/models"
	"github.com/nrbernard/gator/internal/service"
)

type FeedHandler struct {
	FeedService *service.FeedService
	UserService *service.UserService
}

type FormData struct {
	Errors map[string]string
	Values map[string]string
}

func NewFormData() FormData {
	return FormData{
		Errors: map[string]string{},
		Values: map[string]string{},
	}
}

type PageData struct {
	FormData FormData
	Feeds    []models.Feed
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

	return c.Render(http.StatusOK, "feeds-index.html", PageData{
		FormData: NewFormData(),
		Feeds:    feeds,
	})
}

func (h *FeedHandler) Create(c echo.Context) error {
	userName, ok := c.Get("userName").(string)
	if !ok {
		return fmt.Errorf("failed to get user name")
	}

	user, err := h.UserService.GetUser(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("failed to get user: %s", err)
	}

	feed, err := h.FeedService.CreateFeed(c.Request().Context(), service.CreateFeedParams{
		Name:   c.FormValue("name"),
		Url:    c.FormValue("url"),
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed: %s", err)
	}

	formData := NewFormData()
	renderErr := c.Render(http.StatusOK, "feed-form", formData)
	if renderErr != nil {
		return renderErr
	}

	return c.Render(http.StatusOK, "oob-feed", feed)
}
