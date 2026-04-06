package gin

import (
	"strings"

	"github.com/gin-gonic/gin"

	openswag "github.com/gopackx/open-swag-go"
)

// Mount mounts the documentation on a Gin router
func Mount(r *gin.Engine, docs *openswag.Docs, basePath string) {
	// Ensure basePath ends with /
	baseWithSlash := basePath
	if !strings.HasSuffix(baseWithSlash, "/") {
		baseWithSlash += "/"
	}
	baseWithoutSlash := strings.TrimSuffix(baseWithSlash, "/")

	// Redirect /docs to /docs/ to fix relative URL resolution
	r.GET(baseWithoutSlash, func(c *gin.Context) {
		c.Redirect(301, baseWithSlash)
	})
	r.GET(baseWithSlash, gin.WrapF(docs.Handler()))
	r.GET(baseWithSlash+"openapi.json", gin.WrapF(docs.SpecHandler()))
}

// MountGroup mounts the documentation on a Gin router group
func MountGroup(rg *gin.RouterGroup, docs *openswag.Docs) {
	rg.GET("", func(c *gin.Context) {
		c.Redirect(301, c.Request.URL.Path+"/")
	})
	rg.GET("/", gin.WrapF(docs.Handler()))
	rg.GET("/openapi.json", gin.WrapF(docs.SpecHandler()))
}
