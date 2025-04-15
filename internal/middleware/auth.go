package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/nrbernard/gator/internal/config"
)

func CurrentUser(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userName := config.GetUser()
			if userName == "" {
				return fmt.Errorf("failed to get user")
			}

			c.Set("userName", userName)
			return next(c)
		}
	}
}
