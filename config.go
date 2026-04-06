package openswag

// Config is the main configuration for the documentation
type Config struct {
	Info     Info      `json:"info"`
	Servers  []Server  `json:"servers,omitempty"`
	Tags     []Tag     `json:"tags,omitempty"`
	UI       UIConfig  `json:"ui"`
	DocsAuth *DocsAuth `json:"docsAuth,omitempty"`
}

// Predefined security scheme names for use in Endpoint.Security
const (
	SecurityBearerAuth  = "bearerAuth"  // JWT Bearer token
	SecurityBasicAuth   = "basicAuth"   // HTTP Basic auth
	SecurityApiKey      = "apiKeyAuth"  // API Key in header (X-API-Key)
	SecurityApiKeyQuery = "apiKeyQuery" // API Key in query param (?api_key=)
	SecurityOAuth2      = "oauth2"      // OAuth2
)

// DocsAuth configures basic auth protection for the docs UI
type DocsAuth struct {
	Enabled  bool   `json:"enabled"`
	Username string `json:"username"`
	Password string `json:"password"`
	Realm    string `json:"realm,omitempty"`
	// Alternative: use API key in query param (?key=xxx)
	APIKey string `json:"apiKey,omitempty"`
}

// Info represents OpenAPI info object
type Info struct {
	Title          string   `json:"title"`
	Version        string   `json:"version"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License represents license information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents an API server
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Tag represents a tag for grouping operations
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UIConfig configures the documentation UI
type UIConfig struct {
	Theme              string `json:"theme"`
	DarkMode           bool   `json:"darkMode"`
	Layout             string `json:"layout"`
	ShowSidebar        bool   `json:"showSidebar"`
	SidebarSearch      bool   `json:"sidebarSearch"`
	TagGrouping        bool   `json:"tagGrouping"`
	CollapsibleSchemas bool   `json:"collapsibleSchemas"`
	CustomCSS          string `json:"customCss,omitempty"`
}
