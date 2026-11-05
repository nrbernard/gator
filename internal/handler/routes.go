package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handlers are the handlers behind the app's routes.
type Handlers struct {
	Queue       *QueueHandler
	ReadingList *ReadingListHandler
	Post        *PostHandler
	Feed        *FeedHandler
}

// Register adds the app's routes to e.
func Register(e *echo.Echo, h Handlers) {
	e.GET("/", h.Queue.Index)
	e.GET("/posts", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/")
	})
	e.POST("/queue/refresh", h.Queue.Refresh)
	e.POST("/queue/:id/keep", h.Queue.Keep)
	e.POST("/queue/:id/trash", h.Queue.Trash)

	e.GET("/reading-list", h.ReadingList.Index)

	e.POST("/search", h.Post.Search)

	e.GET("/feeds", h.Feed.Index)
	e.POST("/feeds", h.Feed.Create)
	e.DELETE("/feeds/:id", h.Feed.Delete)
}
