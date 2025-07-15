package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/handler"
	"github.com/nrbernard/gator/internal/middleware"
	"github.com/nrbernard/gator/internal/models"
	"github.com/nrbernard/gator/internal/service"
)

type Template struct {
	tmpl *template.Template
}

func newTemplate() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("internal/views/*.html")),
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

type Page struct {
	Posts []models.Post
}

func main() {
	e := echo.New()
	e.Renderer = newTemplate()
	e.Use(echoMiddleware.Logger())
	e.Static("/static", "static")

	db, err := sql.Open("sqlite3", "./data/gator.db")
	if err != nil {
		fmt.Printf("Failed to connect to database: %s\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	userService := &service.UserService{Repo: dbQueries}
	postService := &service.PostService{Repo: dbQueries}
	feedService := &service.FeedService{Repo: dbQueries}
	savedPostService := &service.SavedPostService{Repo: dbQueries}
	readPostService := &service.ReadPostService{Repo: dbQueries}

	e.Use(middleware.CurrentUser(userService))

	postHandler, err := handler.NewPostHandler(postService, userService, feedService)
	if err != nil {
		fmt.Printf("Failed to create post handler: %s\n", err)
		os.Exit(1)
	}

	feedHandler, err := handler.NewFeedHandler(feedService, userService)
	if err != nil {
		fmt.Printf("Failed to create feed handler: %s\n", err)
		os.Exit(1)
	}

	savedPostHandler, err := handler.NewSavedPostHandler(savedPostService, userService)
	if err != nil {
		fmt.Printf("Failed to create feed handler: %s\n", err)
		os.Exit(1)
	}

	readPostHandler, err := handler.NewReadPostHandler(readPostService)
	if err != nil {
		fmt.Printf("Failed to create feed handler: %s\n", err)
		os.Exit(1)
	}

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(301, "/posts")
	})

	e.GET("/posts", postHandler.Index)

	e.POST("/saved-posts/:id", savedPostHandler.Save)
	e.DELETE("/saved-posts/:id", savedPostHandler.Delete)

	e.POST("/read-posts/:id", readPostHandler.Save)

	e.POST("/posts/refresh", postHandler.Refresh)

	e.POST("/search", postHandler.Search)

	e.GET("/feeds", feedHandler.Index)
	e.POST("/feeds", feedHandler.Create)
	e.DELETE("/feeds/:id", feedHandler.Delete)

	e.Start(":8080")
}
