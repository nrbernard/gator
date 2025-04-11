package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func main() {
	e := echo.New()

	e.Renderer = newTemplate()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"Posts": []map[string]interface{}{
				{
					"Title": "Post 1",
					"Link":  "https://example.com/post1",
				},
				{
					"Title": "Post 2",
					"Link":  "https://example.com/post2",
				},
			},
		})
	})

	e.Start(":8080")
}
