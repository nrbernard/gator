package middleware

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/service"
)

func CurrentUser(userService *service.UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, err := userService.GetUser(context.Background(), "nick")
			if err != nil {
				return fmt.Errorf("failed to get user from database: %w", err)
			}

			c.Set("userID", user.ID)
			return next(c)
		}
	}
}
