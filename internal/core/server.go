package core

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HTTPServer struct {
	Config   *ApiServiceConfiguration
	Api      *echo.Echo
	Postgres *sql.DB
}

type ApiServiceConfiguration struct {
	DBConnectionString string
}

type HTTPServerContext struct {
	echo.Context
	S *HTTPServer
}

// TemplateRenderer is a custom renderer for Echo to use Go templates
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template with data
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (s *HTTPServer) Health(sc *HTTPServerContext) error {
	return sc.String(http.StatusOK, "ok")
}

func (s *HTTPServer) WrapHandler(h func(sc *HTTPServerContext) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()

		id := req.Header.Get(echo.HeaderXRequestID)
		if id == "" {
			res.Header().Get(echo.HeaderXRequestID)
		}
		return h(&HTTPServerContext{c, s})
	}
}

func (s *HTTPServer) Start() error {
	return s.Api.Start(":8080")
}
