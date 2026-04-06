package fiber

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"

	openswag "github.com/gopackx/open-swag-go"
)

// Mount mounts the documentation on a Fiber app
func Mount(app *fiber.App, docs *openswag.Docs, basePath string) {
	// Ensure basePath ends with /
	baseWithSlash := basePath
	if !strings.HasSuffix(baseWithSlash, "/") {
		baseWithSlash += "/"
	}
	baseWithoutSlash := strings.TrimSuffix(baseWithSlash, "/")

	// Redirect /docs to /docs/ to fix relative URL resolution
	app.Get(baseWithoutSlash, func(c *fiber.Ctx) error {
		return c.Redirect(baseWithSlash, 301)
	})
	app.Get(baseWithSlash, adaptor.HTTPHandlerFunc(docs.Handler()))
	app.Get(baseWithSlash+"openapi.json", adaptor.HTTPHandlerFunc(docs.SpecHandler()))
}

// MountGroup mounts the documentation on a Fiber router group
func MountGroup(g fiber.Router, docs *openswag.Docs) {
	g.Get("/", adaptor.HTTPHandlerFunc(docs.Handler()))
	g.Get("/openapi.json", adaptor.HTTPHandlerFunc(docs.SpecHandler()))
}
