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

	dbPath := os.Getenv("DATABASE_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to connect to database: %s\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	userService := &service.UserService{Repo: dbQueries}
	postService := &service.PostService{Repo: dbQueries}
	feedService := &service.FeedService{Repo: dbQueries}
	stateService := &service.PostStateService{Repo: dbQueries}

	e.Use(middleware.CurrentUser(userService))

	feedHandler, err := handler.NewFeedHandler(feedService, userService)
	if err != nil {
		fmt.Printf("Failed to create feed handler: %s\n", err)
		os.Exit(1)
	}

	handler.Register(e, handler.Handlers{
		Queue:       &handler.QueueHandler{PostService: postService, StateService: stateService, FeedService: feedService},
		ReadingList: &handler.ReadingListHandler{PostService: postService, StateService: stateService},
		Post:        &handler.PostHandler{PostService: postService},
		Feed:        feedHandler,
		StaticDir:   "static",
	})

	if err := e.Start(":8080"); err != nil {
		fmt.Printf("Server stopped: %s\n", err)
		os.Exit(1)
	}
}
