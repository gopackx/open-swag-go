package echo

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	openswag "github.com/gopackx/open-swag-go"
)

// Mount mounts the documentation on an Echo router
func Mount(e *echo.Echo, docs *openswag.Docs, basePath string) {
	// Ensure basePath ends with /
	baseWithSlash := basePath
	if !strings.HasSuffix(baseWithSlash, "/") {
		baseWithSlash += "/"
	}
	baseWithoutSlash := strings.TrimSuffix(baseWithSlash, "/")

	// Redirect /docs to /docs/ to fix relative URL resolution
	e.GET(baseWithoutSlash, func(c echo.Context) error {
		return c.Redirect(301, baseWithSlash)
	})
	e.GET(baseWithSlash, echo.WrapHandler(http.HandlerFunc(docs.Handler())))
	e.GET(baseWithSlash+"openapi.json", echo.WrapHandler(http.HandlerFunc(docs.SpecHandler())))
}

// MountGroup mounts the documentation on an Echo group
func MountGroup(g *echo.Group, docs *openswag.Docs) {
	g.GET("", func(c echo.Context) error {
		return c.Redirect(301, c.Request().URL.Path+"/")
	})
	g.GET("/", echo.WrapHandler(http.HandlerFunc(docs.Handler())))
	g.GET("/openapi.json", echo.WrapHandler(http.HandlerFunc(docs.SpecHandler())))
}
