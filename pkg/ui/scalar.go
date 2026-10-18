package ui

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"strings"
)

//go:embed templates/scalar.html
var scalarTemplate string

// DefaultScalarVersion is the @scalar/api-reference release loaded from the
// CDN when ScalarConfig.ScalarVersion is empty. It is pinned so that upstream
// releases do not silently change the docs UI for every user.
const DefaultScalarVersion = "1.75.0"

// scalarHTMLTemplate is the embedded template with its {{PLACEHOLDER}} markers
// rewritten to html/template actions, so every value is escaped for the
// context it appears in (HTML text, JS string, URL, CSS).
var scalarHTMLTemplate = template.Must(template.New("scalar").Parse(
	strings.NewReplacer(
		"{{TITLE}}", "{{.Title}}",
		"{{SPEC_URL}}", "{{.SpecURL}}",
		"{{CONFIG}}", "{{.Config}}",
		"{{CUSTOM_CSS}}", "{{.CustomCSS}}",
		"{{SCALAR_VERSION}}", "{{.ScalarVersion}}",
	).Replace(scalarTemplate),
))

// scalarTemplateData is the data passed to scalarHTMLTemplate.
type scalarTemplateData struct {
	Title         string
	SpecURL       string
	Config        template.JS
	CustomCSS     template.CSS
	ScalarVersion string
}

// ScalarConfig configures the Scalar UI
type ScalarConfig struct {
	Theme             string   `json:"theme"`
	Layout            string   `json:"layout"`
	DarkMode          bool     `json:"darkMode"`
	ShowSidebar       bool     `json:"showSidebar"`
	SearchHotKey      string   `json:"searchHotKey,omitempty"`
	HiddenClients     []string `json:"hiddenClients,omitempty"`
	DefaultHTTPClient string   `json:"defaultHttpClient,omitempty"`
	// CustomCSS is inserted verbatim into a <style> element. It must come
	// from a trusted source (developer configuration).
	CustomCSS string `json:"-"`
	// ScalarVersion selects the @scalar/api-reference version loaded from the
	// CDN. Empty means DefaultScalarVersion.
	ScalarVersion string `json:"-"`
}

// DefaultScalarConfig returns the default Scalar configuration
func DefaultScalarConfig() ScalarConfig {
	return ScalarConfig{
		Theme:         "purple",
		Layout:        "modern",
		DarkMode:      true,
		ShowSidebar:   true,
		SearchHotKey:  "k",
		ScalarVersion: DefaultScalarVersion,
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
	if config.ScalarVersion == "" {
		config.ScalarVersion = DefaultScalarVersion
	}
	return &Scalar{
		config:  config,
		specURL: specURL,
		title:   title,
	}
}

// Render generates the HTML for the Scalar UI
func (s *Scalar) Render() (string, error) {
	// encoding/json escapes <, > and & by default, so the JSON is safe to
	// embed directly in a <script> element.
	configJSON, err := json.Marshal(s.config)
	if err != nil {
		return "", err
	}

	version := s.config.ScalarVersion
	if version == "" {
		version = DefaultScalarVersion
	}

	var buf bytes.Buffer
	err = scalarHTMLTemplate.Execute(&buf, scalarTemplateData{
		Title:         s.title,
		SpecURL:       s.specURL,
		Config:        template.JS(configJSON),
		CustomCSS:     template.CSS(s.config.CustomCSS),
		ScalarVersion: version,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
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
