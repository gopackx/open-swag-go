package nethttp

import (
	"net/http"
	"strings"

	openswag "github.com/gopackx/open-swag-go"
)

// Mount mounts the documentation on a net/http ServeMux
func Mount(mux *http.ServeMux, docs *openswag.Docs, basePath string) {
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	mux.HandleFunc(basePath, docs.Handler())
	mux.HandleFunc(basePath+"openapi.json", docs.SpecHandler())
}

// MountWithPrefix mounts with a custom prefix handler
func MountWithPrefix(mux *http.ServeMux, docs *openswag.Docs, basePath string) {
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	// Handle both with and without trailing slash
	mux.HandleFunc(strings.TrimSuffix(basePath, "/"), func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath, http.StatusMovedPermanently)
	})
	mux.HandleFunc(basePath, docs.Handler())
	mux.HandleFunc(basePath+"openapi.json", docs.SpecHandler())
}
