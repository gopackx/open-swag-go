package openswag

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gopackx/open-swag-go/pkg/ui"
)

// basicAuth wraps a handler with basic authentication or API key
func (d *Docs) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d.config.DocsAuth == nil || !d.config.DocsAuth.Enabled {
			next(w, r)
			return
		}

		// Option 1: API Key in query param (?key=xxx)
		if d.config.DocsAuth.APIKey != "" {
			key := r.URL.Query().Get("key")
			if subtle.ConstantTimeCompare([]byte(key), []byte(d.config.DocsAuth.APIKey)) == 1 {
				next(w, r)
				return
			}
		}

		// Option 2: Basic Auth
		if d.config.DocsAuth.Username != "" && d.config.DocsAuth.Password != "" {
			username, password, ok := r.BasicAuth()
			if ok {
				usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(d.config.DocsAuth.Username)) == 1
				passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(d.config.DocsAuth.Password)) == 1
				if usernameMatch && passwordMatch {
					next(w, r)
					return
				}
			}
		}

		d.unauthorized(w)
	}
}

func (d *Docs) unauthorized(w http.ResponseWriter) {
	realm := d.config.DocsAuth.Realm
	if realm == "" {
		realm = "API Documentation"
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`", charset="UTF-8"`)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// Handler returns the documentation UI handler
func (d *Docs) Handler() http.HandlerFunc {
	return d.basicAuth(func(w http.ResponseWriter, r *http.Request) {
		config := ui.ScalarConfig{
			Theme:       d.config.UI.Theme,
			Layout:      d.config.UI.Layout,
			DarkMode:    d.config.UI.DarkMode,
			ShowSidebar: d.config.UI.ShowSidebar,
			CustomCSS:   d.config.UI.CustomCSS,
		}

		scalar := ui.NewScalar("./openapi.json", d.config.Info.Title, config)
		html, err := scalar.Render()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})
}

// SpecHandler returns the OpenAPI spec JSON handler
func (d *Docs) SpecHandler() http.HandlerFunc {
	return d.basicAuth(func(w http.ResponseWriter, r *http.Request) {
		specJSON, err := d.SpecJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(specJSON)
	})
}

// Mount registers both handlers on a mux
func (d *Docs) Mount(mux *http.ServeMux, basePath string) {
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	mux.HandleFunc(basePath, d.Handler())
	mux.HandleFunc(basePath+"openapi.json", d.SpecHandler())
}

// GetUIConfig returns the UI configuration as JSON for client-side use
func (d *Docs) GetUIConfig() (string, error) {
	config := map[string]interface{}{
		"theme":       d.config.UI.Theme,
		"layout":      d.config.UI.Layout,
		"darkMode":    d.config.UI.DarkMode,
		"showSidebar": d.config.UI.ShowSidebar,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
