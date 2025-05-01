package middleware

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/config"
	"github.com/nrbernard/gator/internal/service"
)

func CurrentUser(config *config.Config, userService *service.UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userName := config.GetUser()
			if userName == "" {
				return fmt.Errorf("failed to get user")
			}

			user, err := userService.GetUser(context.Background(), userName)
			if err != nil {
				return fmt.Errorf("failed to get user from database: %w", err)
			}

			c.Set("userID", user.ID)
			return next(c)
		}
	}
}
