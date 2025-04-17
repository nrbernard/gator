package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/nrbernard/gator/internal/config"
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
	configFile, err := config.Read()
	if err != nil {
		fmt.Printf("Failed to read config: %s\n", err)
		os.Exit(1)
	}

	e := echo.New()
	e.Renderer = newTemplate()
	e.Use(echoMiddleware.Logger())
	e.Use(middleware.CurrentUser(configFile))

	db, err := sql.Open("postgres", configFile.DBUrl)
	if err != nil {
		fmt.Printf("Failed to connect to database: %s\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	userService := &service.UserService{Repo: dbQueries}
	postService := &service.PostService{Repo: dbQueries}
	feedService := &service.FeedService{Repo: dbQueries}

	postHandler := &handler.PostHandler{PostService: postService, UserService: userService}
	feedHandler := &handler.FeedHandler{FeedService: feedService, UserService: userService}

	e.GET("/", postHandler.Index)
	e.GET("/feeds", feedHandler.Index)
	e.POST("/feeds", feedHandler.Create)
	e.DELETE("/feeds/:id", feedHandler.Delete)

	e.Start(":8080")
}
