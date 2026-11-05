package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handlers are the handlers behind the app's routes.
type Handlers struct {
	Queue       *QueueHandler
	ReadingList *ReadingListHandler
	Archive     *ArchiveHandler
	Trash       *TrashHandler
	Post        *PostHandler
	Feed        *FeedHandler
	// StaticDir is the directory served at /static: fonts, scripts, and the Tailwind build.
	StaticDir string
}

// Register adds the app's routes to e.
func Register(e *echo.Echo, h Handlers) {
	e.Static("/static", h.StaticDir)

	e.GET("/", h.Queue.Index)
	e.GET("/posts", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/")
	})
	e.POST("/queue/check", h.Queue.Check)
	e.GET("/queue/card", h.Queue.Card)
	e.GET("/queue/trash-all", h.Queue.ConfirmTrashAll)
	e.POST("/queue/trash-all", h.Queue.TrashAll)
	e.POST("/queue/:id/keep", h.Queue.Keep)
	e.POST("/queue/:id/trash", h.Queue.Trash)
	e.POST("/queue/:id/open", h.Queue.Open)

	e.GET("/reading-list", h.ReadingList.Index)
	e.POST("/reading-list/:id/open", h.ReadingList.Open)
	e.POST("/reading-list/:id/read", h.ReadingList.ToggleRead)
	e.POST("/reading-list/:id/archive", h.ReadingList.Archive)
	e.POST("/reading-list/:id/trash", h.ReadingList.Trash)

	e.GET("/archive", h.Archive.Index)
	e.POST("/archive/:id/unarchive", h.Archive.Unarchive)
	e.POST("/archive/:id/trash", h.Archive.Trash)

	e.GET("/trash", h.Trash.Index)
	e.POST("/trash/:id/restore", h.Trash.Restore)

	e.GET("/search", h.Post.Index)
	e.POST("/search", h.Post.Search)
	e.POST("/search/:id/keep", h.Post.Keep)
	e.POST("/search/:id/open", h.Post.Open)
	e.POST("/search/:id/read", h.Post.ToggleRead)
	e.POST("/search/:id/archive", h.Post.Archive)
	e.POST("/search/:id/unarchive", h.Post.Unarchive)
	e.POST("/search/:id/trash", h.Post.Trash)
	e.POST("/search/:id/restore", h.Post.Restore)

	e.GET("/feeds", h.Feed.Index)
	e.POST("/feeds", h.Feed.Create)
	e.DELETE("/feeds/:id", h.Feed.Delete)
}
