package chi

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	openswag "github.com/gopackx/open-swag-go"
)

// Mount mounts the documentation on a Chi router
func Mount(r chi.Router, docs *openswag.Docs, basePath string) {
	// Ensure basePath ends with /
	baseWithSlash := basePath
	if !strings.HasSuffix(baseWithSlash, "/") {
		baseWithSlash += "/"
	}
	baseWithoutSlash := strings.TrimSuffix(baseWithSlash, "/")

	// Redirect /docs to /docs/ to fix relative URL resolution
	r.Get(baseWithoutSlash, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, baseWithSlash, 301)
	})
	r.Get(baseWithSlash, docs.Handler())
	r.Get(baseWithSlash+"openapi.json", docs.SpecHandler())
}
