package auth

// SchemeType represents the type of security scheme
type SchemeType string

const (
	SchemeTypeHTTP   SchemeType = "http"
	SchemeTypeAPIKey SchemeType = "apiKey"
	SchemeTypeOAuth2 SchemeType = "oauth2"
	SchemeTypeOpenID SchemeType = "openIdConnect"
)

// APIKeyLocation represents where the API key is sent
type APIKeyLocation string

const (
	APIKeyInHeader APIKeyLocation = "header"
	APIKeyInQuery  APIKeyLocation = "query"
	APIKeyInCookie APIKeyLocation = "cookie"
)

// Scheme represents a security scheme
type Scheme struct {
	Type             SchemeType     `json:"type"`
	Description      string         `json:"description,omitempty"`
	Name             string         `json:"name,omitempty"`
	In               APIKeyLocation `json:"in,omitempty"`
	Scheme           string         `json:"scheme,omitempty"`
	BearerFormat     string         `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows    `json:"flows,omitempty"`
	OpenIDConnectURL string         `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents OAuth2 flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents a single OAuth2 flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// BearerAuth creates a bearer token authentication scheme
func BearerAuth(description string) Scheme {
	return Scheme{
		Type:         SchemeTypeHTTP,
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  description,
	}
}

// BasicAuth creates a basic authentication scheme
func BasicAuth(description string) Scheme {
	return Scheme{
		Type:        SchemeTypeHTTP,
		Scheme:      "basic",
		Description: description,
	}
}

// APIKeyAuth creates an API key authentication scheme
func APIKeyAuth(name string, in APIKeyLocation, description string) Scheme {
	return Scheme{
		Type:        SchemeTypeAPIKey,
		Name:        name,
		In:          in,
		Description: description,
	}
}

// APIKeyHeader creates an API key in header scheme
func APIKeyHeader(name, description string) Scheme {
	return APIKeyAuth(name, APIKeyInHeader, description)
}

// APIKeyQuery creates an API key in query scheme
func APIKeyQuery(name, description string) Scheme {
	return APIKeyAuth(name, APIKeyInQuery, description)
}

// CookieAuth creates a cookie-based authentication scheme
func CookieAuth(name, description string) Scheme {
	return Scheme{
		Type:        SchemeTypeAPIKey,
		Name:        name,
		In:          APIKeyInCookie,
		Description: description,
	}
}
