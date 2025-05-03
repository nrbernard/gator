package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

type ReadPostHandler struct {
	ReadPostService *service.ReadPostService
}

func NewReadPostHandler(readPostService *service.ReadPostService) (*ReadPostHandler, error) {
	if readPostService == nil {
		return nil, fmt.Errorf("all services must be provided")
	}

	return &ReadPostHandler{
		ReadPostService: readPostService,
	}, nil
}

func (h *ReadPostHandler) Save(c echo.Context) error {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return fmt.Errorf("failed to get user from context")
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse post ID: %s", err)
	}

	err = h.ReadPostService.Save(context.Background(), postID, userID)
	if err != nil {
		return fmt.Errorf("failed to save read post: %s", err)
	}

	return c.NoContent(http.StatusOK)
}
