package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

type SavedPostHandler struct {
	SavedPostService *service.SavedPostService
	UserService      *service.UserService
}

func NewSavedPostHandler(savedPostService *service.SavedPostService, userService *service.UserService) (*SavedPostHandler, error) {
	if savedPostService == nil || userService == nil {
		return nil, fmt.Errorf("all services must be provided")
	}

	return &SavedPostHandler{
		SavedPostService: savedPostService,
		UserService:      userService,
	}, nil
}

func (h *SavedPostHandler) Save(c echo.Context) error {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return fmt.Errorf("failed to get user from context")
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse post ID: %s", err)
	}

	err = h.SavedPostService.SavePost(context.Background(), postID, userID)
	if err != nil {
		return fmt.Errorf("failed to save post: %s", err)
	}

	return c.Render(http.StatusOK, "saved-post", map[string]interface{}{
		"ID": postID,
	})
}

func (h *SavedPostHandler) Delete(c echo.Context) error {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return fmt.Errorf("failed to get user from context")
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse post ID: %s", err)
	}

	err = h.SavedPostService.UnsavePost(context.Background(), postID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete post save: %s", err)
	}

	return c.Render(http.StatusOK, "save-post", map[string]interface{}{
		"ID": postID,
	})
}
