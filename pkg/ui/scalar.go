package ui

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed templates/scalar.html
var scalarTemplate string

// ScalarConfig configures the Scalar UI
type ScalarConfig struct {
	Theme             string   `json:"theme"`
	Layout            string   `json:"layout"`
	DarkMode          bool     `json:"darkMode"`
	ShowSidebar       bool     `json:"showSidebar"`
	SearchHotKey      string   `json:"searchHotKey,omitempty"`
	HiddenClients     []string `json:"hiddenClients,omitempty"`
	DefaultHTTPClient string   `json:"defaultHttpClient,omitempty"`
	CustomCSS         string   `json:"-"`
}

// DefaultScalarConfig returns the default Scalar configuration
func DefaultScalarConfig() ScalarConfig {
	return ScalarConfig{
		Theme:        "purple",
		Layout:       "modern",
		DarkMode:     true,
		ShowSidebar:  true,
		SearchHotKey: "k",
	}
}

// Scalar represents the Scalar UI renderer
type Scalar struct {
	config  ScalarConfig
	specURL string
	title   string
}

// NewScalar creates a new Scalar UI instance
func NewScalar(specURL, title string, config ScalarConfig) *Scalar {
	return &Scalar{
		config:  config,
		specURL: specURL,
		title:   title,
	}
}

// Render generates the HTML for the Scalar UI
func (s *Scalar) Render() (string, error) {
	configJSON, err := json.Marshal(s.config)
	if err != nil {
		return "", err
	}

	html := scalarTemplate
	html = strings.ReplaceAll(html, "{{SPEC_URL}}", s.specURL)
	html = strings.ReplaceAll(html, "{{CONFIG}}", string(configJSON))
	html = strings.ReplaceAll(html, "{{TITLE}}", s.title)

	if s.config.CustomCSS != "" {
		html = strings.ReplaceAll(html, "{{CUSTOM_CSS}}", s.config.CustomCSS)
	} else {
		html = strings.ReplaceAll(html, "{{CUSTOM_CSS}}", "")
	}

	return html, nil
}

// SetTheme sets the UI theme
func (s *Scalar) SetTheme(theme string) {
	s.config.Theme = theme
}

// SetDarkMode enables or disables dark mode
func (s *Scalar) SetDarkMode(enabled bool) {
	s.config.DarkMode = enabled
}

// SetLayout sets the UI layout
func (s *Scalar) SetLayout(layout string) {
	s.config.Layout = layout
}

// SetCustomCSS sets custom CSS styles
func (s *Scalar) SetCustomCSS(css string) {
	s.config.CustomCSS = css
}

// GetTemplate returns the raw Scalar HTML template
func GetTemplate() string {
	return scalarTemplate
}
