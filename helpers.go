package openswag

import "github.com/gopackx/open-swag-go/pkg/spec"

// Body wraps a struct as a JSON request body. The body is marked required —
// use BodyWithDesc or build a *RequestBody literal for finer control.
func Body(schema interface{}) *RequestBody {
	return &RequestBody{
		Required:    true,
		Schema:      schema,
		ContentType: "application/json",
	}
}

// BodyWithDesc is Body with a human-readable description attached.
func BodyWithDesc(description string, schema interface{}) *RequestBody {
	return &RequestBody{
		Description: description,
		Required:    true,
		Schema:      schema,
		ContentType: "application/json",
	}
}

// FormBody declares a multipart/form-data request body. Use it for file
// uploads and traditional HTML form posts.
func FormBody(schema interface{}) *RequestBody {
	return &RequestBody{
		Required:    true,
		Schema:      schema,
		ContentType: "multipart/form-data",
	}
}

// FormURLEncodedBody declares an application/x-www-form-urlencoded body.
func FormURLEncodedBody(schema interface{}) *RequestBody {
	return &RequestBody{
		Required:    true,
		Schema:      schema,
		ContentType: "application/x-www-form-urlencoded",
	}
}

// Response builds a ResponseSpec. The schema is optional — pass nil for
// empty responses (e.g. 204 No Content).
func Response(description string, schema ...interface{}) ResponseSpec {
	r := ResponseSpec{Description: description}
	if len(schema) > 0 {
		r.Schema = schema[0]
	}
	return r
}

// PathParam builds a required string path parameter.
func PathParam(name, description string) Parameter {
	return Parameter{
		Name:        name,
		In:          "path",
		Description: description,
		Required:    true,
		Schema:      spec.NewSchema("string"),
	}
}

// QueryParam builds an optional string query parameter.
func QueryParam(name, description string) Parameter {
	return Parameter{
		Name:        name,
		In:          "query",
		Description: description,
		Schema:      spec.NewSchema("string"),
	}
}

// RequiredQueryParam builds a required string query parameter.
func RequiredQueryParam(name, description string) Parameter {
	p := QueryParam(name, description)
	p.Required = true
	return p
}

// HeaderParam builds an optional string header parameter.
func HeaderParam(name, description string) Parameter {
	return Parameter{
		Name:        name,
		In:          "header",
		Description: description,
		Schema:      spec.NewSchema("string"),
	}
}

// CookieParam builds an optional string cookie parameter.
func CookieParam(name, description string) Parameter {
	return Parameter{
		Name:        name,
		In:          "cookie",
		Description: description,
		Schema:      spec.NewSchema("string"),
	}
}

// BearerAuth registers a JWT bearer scheme under the given name.
func BearerAuth(name string) AuthScheme {
	return AuthScheme{
		Name:         name,
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Bearer token authentication",
	}
}

// APIKeyAuth registers an apiKey scheme transported in a request header.
// headerName is the HTTP header field used to carry the key (e.g. "X-API-Key").
func APIKeyAuth(name, headerName string) AuthScheme {
	return AuthScheme{
		Name:        name,
		Type:        "apiKey",
		In:          "header",
		ParamName:   headerName,
		Description: "API key in " + headerName + " header",
	}
}

// APIKeyQueryAuth registers an apiKey scheme transported as a query parameter.
func APIKeyQueryAuth(name, paramName string) AuthScheme {
	return AuthScheme{
		Name:        name,
		Type:        "apiKey",
		In:          "query",
		ParamName:   paramName,
		Description: "API key in ?" + paramName + " query parameter",
	}
}

// BasicAuth registers an HTTP Basic scheme.
func BasicAuth(name string) AuthScheme {
	return AuthScheme{
		Name:        name,
		Type:        "http",
		Scheme:      "basic",
		Description: "HTTP Basic authentication",
	}
}

// CookieAuth registers an apiKey scheme transported via a cookie.
// cookieName is the cookie that carries the session/key (e.g. "session_id").
func CookieAuth(name, cookieName string) AuthScheme {
	return AuthScheme{
		Name:        name,
		Type:        "apiKey",
		In:          "cookie",
		ParamName:   cookieName,
		Description: "Session cookie authentication via " + cookieName,
	}
}
