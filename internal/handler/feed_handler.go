package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/models"
	"github.com/nrbernard/gator/internal/service"
)

type FeedHandler struct {
	FeedService *service.FeedService
	UserService *service.UserService
}

func NewFeedHandler(feedService *service.FeedService, userService *service.UserService) (*FeedHandler, error) {
	if feedService == nil || userService == nil {
		return nil, fmt.Errorf("all services must be provided")
	}
	return &FeedHandler{FeedService: feedService, UserService: userService}, nil
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
		return fmt.Errorf("failed to get user %s: %s", userName, err)
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

	url := c.FormValue("url")

	feed, err := h.FeedService.CreateFeed(c.Request().Context(), service.CreateFeedParams{
		Url:    url,
		UserID: user.ID,
	})
	if err != nil {
		formData := FormData{
			Errors: map[string]string{
				"url": "There is already a feed with this URL",
			},
			Values: map[string]string{
				"url": url,
			},
		}

		return c.Render(http.StatusUnprocessableEntity, "feed-form", formData)
	}

	formData := NewFormData()
	renderErr := c.Render(http.StatusOK, "feed-form", formData)
	if renderErr != nil {
		return renderErr
	}

	return c.Render(http.StatusOK, "oob-feed", feed)
}

func (h *FeedHandler) Delete(c echo.Context) error {
	feedID := c.Param("id")
	if feedID == "" {
		return fmt.Errorf("failed to get feed id")
	}

	if err := h.FeedService.DeleteFeed(c.Request().Context(), uuid.MustParse(feedID)); err != nil {
		return c.String(http.StatusNotFound, "failed to delete feed")
	}

	return c.NoContent(http.StatusOK)
}
