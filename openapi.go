package openswag

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/gopackx/open-swag-go/pkg/schema"
	"github.com/gopackx/open-swag-go/pkg/spec"
)

// Docs is the main documentation instance
type Docs struct {
	config    Config
	endpoints []Endpoint
	openapi   *spec.OpenAPI
	mu        sync.RWMutex
}

// Endpoint represents an API endpoint definition
type Endpoint struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	QueryParams interface{} // Struct with query parameters (uses form/query tags)
	PathParams  interface{} // Struct with path parameters
	RequestBody *RequestBody
	Responses   Responses
	Security    []string
	Deprecated  bool
}

// Responses maps HTTP status codes to response definitions.
type Responses = map[int]ResponseSpec

// Parameter represents an API parameter
type Parameter struct {
	Name        string
	In          string // "path", "query", "header", "cookie"
	Description string
	Required    bool
	Schema      *spec.Schema
	Example     interface{}
}

// RequestBody represents a request body
type RequestBody struct {
	Description string
	Required    bool
	Schema      interface{}
	ContentType string
}

// ResponseSpec represents an API response.
// Use the Response() helper for the most concise form.
type ResponseSpec struct {
	Description string
	Schema      interface{}
}

// New creates a new documentation instance
func New(config Config) *Docs {
	if config.UI.Theme == "" {
		config.UI.Theme = "purple"
	}
	if config.UI.Layout == "" {
		config.UI.Layout = "modern"
	}

	return &Docs{
		config:    config,
		endpoints: make([]Endpoint, 0),
	}
}

// Add registers an endpoint
func (d *Docs) Add(endpoint Endpoint) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.endpoints = append(d.endpoints, endpoint)
	d.openapi = nil
}

// AddAll registers multiple endpoints
func (d *Docs) AddAll(endpoints ...Endpoint) {
	for _, ep := range endpoints {
		d.Add(ep)
	}
}

// BuildSpec generates the OpenAPI spec
func (d *Docs) BuildSpec() *spec.OpenAPI {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.openapi != nil {
		return d.openapi
	}

	info := spec.NewInfo(d.config.Info.Title, d.config.Info.Version).
		WithDescription(d.config.Info.Description)

	if d.config.Info.Contact != nil {
		info = info.WithContact(
			d.config.Info.Contact.Name,
			d.config.Info.Contact.URL,
			d.config.Info.Contact.Email,
		)
	}

	if d.config.Info.License != nil {
		info = info.WithLicense(d.config.Info.License.Name, d.config.Info.License.URL)
	}

	openapi := spec.NewOpenAPI(info)

	// Add servers
	for _, srv := range d.config.Servers {
		openapi.AddServer(spec.NewServer(srv.URL).WithDescription(srv.Description))
	}

	// Add tags
	for _, tag := range d.config.Tags {
		openapi.AddTag(spec.Tag{Name: tag.Name, Description: tag.Description})
	}

	// Build paths from endpoints
	for _, ep := range d.endpoints {
		d.addEndpointToSpec(openapi, ep)
	}

	// Add predefined security schemes if any endpoint uses security
	d.addSecuritySchemes(openapi)

	d.openapi = openapi
	return openapi
}

// addSecuritySchemes wires security schemes into the spec. Resolution order:
//  1. Schemes explicitly registered via Config.Auth.Schemes (always win).
//  2. The predefined constants (SecurityBearerAuth, SecurityBasicAuth, etc.)
//     when an endpoint references them by name.
//
// Unknown names that are neither registered nor predefined are skipped — we no
// longer silently materialise them as http-bearer. This keeps the spec honest:
// the generated client will see exactly the schemes the user configured.
func (d *Docs) addSecuritySchemes(openapi *spec.OpenAPI) {
	usedSchemes := make(map[string]bool)
	for _, ep := range d.endpoints {
		for _, sec := range ep.Security {
			usedSchemes[sec] = true
		}
	}

	// Always register schemes the user declared — they may be referenced by
	// AuthConfig alone (e.g. for global Security requirements) without an
	// endpoint listing them.
	hasUserSchemes := len(d.config.Auth.Schemes) > 0
	if len(usedSchemes) == 0 && !hasUserSchemes {
		return
	}

	if openapi.Components == nil {
		openapi.Components = &spec.Components{}
	}
	if openapi.Components.SecuritySchemes == nil {
		openapi.Components.SecuritySchemes = make(map[string]*spec.SecurityScheme)
	}

	// 1. User-supplied schemes take priority.
	registered := make(map[string]bool)
	for _, s := range d.config.Auth.Schemes {
		if s.Name == "" {
			continue
		}
		openapi.Components.SecuritySchemes[s.Name] = s.toSpec()
		registered[s.Name] = true
	}

	// 2. Predefined fallbacks for endpoint-referenced names.
	for scheme := range usedSchemes {
		if registered[scheme] {
			continue
		}
		if predef := predefinedScheme(scheme); predef != nil {
			openapi.Components.SecuritySchemes[scheme] = predef
		}
		// else: unknown — caller must register it via Config.Auth.Schemes.
	}
}

// predefinedScheme returns the spec for a known SecurityXxx constant, or nil
// if the name is not one of the built-ins.
func predefinedScheme(name string) *spec.SecurityScheme {
	switch name {
	case SecurityBearerAuth:
		return &spec.SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT Bearer token authentication",
		}
	case SecurityBasicAuth:
		return &spec.SecurityScheme{
			Type:        "http",
			Scheme:      "basic",
			Description: "HTTP Basic authentication",
		}
	case SecurityApiKey:
		return &spec.SecurityScheme{
			Type:        "apiKey",
			In:          "header",
			Name:        "X-API-Key",
			Description: "API key in X-API-Key header",
		}
	case SecurityApiKeyQuery:
		return &spec.SecurityScheme{
			Type:        "apiKey",
			In:          "query",
			Name:        "api_key",
			Description: "API key in query parameter",
		}
	case SecurityCookieAuth:
		return &spec.SecurityScheme{
			Type:        "apiKey",
			In:          "cookie",
			Name:        "session_id",
			Description: "Session cookie authentication",
		}
	case SecurityOAuth2:
		return &spec.SecurityScheme{
			Type:        "oauth2",
			Description: "OAuth2 authentication",
			Flows: &spec.OAuthFlows{
				AuthorizationCode: &spec.OAuthFlow{
					AuthorizationURL: "/oauth2/authorize",
					TokenURL:         "/oauth2/token",
					Scopes: map[string]string{
						"read":  "Read access",
						"write": "Write access",
					},
				},
			},
		}
	}
	return nil
}

func (d *Docs) addEndpointToSpec(openapi *spec.OpenAPI, ep Endpoint) {
	pathItem := openapi.Paths[ep.Path]
	if pathItem == nil {
		pathItem = spec.NewPathItem()
	}

	operation := d.buildOperation(ep)

	method := strings.ToUpper(ep.Method)
	switch method {
	case "GET":
		pathItem.SetGet(operation)
	case "POST":
		pathItem.SetPost(operation)
	case "PUT":
		pathItem.SetPut(operation)
	case "PATCH":
		pathItem.SetPatch(operation)
	case "DELETE":
		pathItem.SetDelete(operation)
	}

	openapi.AddPath(ep.Path, pathItem)
}

func (d *Docs) buildOperation(ep Endpoint) *spec.Operation {
	op := spec.NewOperation(ep.Summary).
		WithDescription(ep.Description).
		WithTags(ep.Tags...).
		SetDeprecated(ep.Deprecated)

	// Build explicit parameters
	for _, param := range ep.Parameters {
		p := spec.NewParameter(param.Name, param.In).
			WithDescription(param.Description).
			SetRequired(param.Required)

		if param.Schema != nil {
			p.WithSchema(param.Schema)
		} else {
			p.WithSchema(spec.NewSchema("string"))
		}

		op.AddParameter(p)
	}

	// Build query parameters from struct
	if ep.QueryParams != nil {
		params := d.buildParamsFromStruct(ep.QueryParams, "query")
		for _, p := range params {
			op.AddParameter(p)
		}
	}

	// Build path parameters from struct
	if ep.PathParams != nil {
		params := d.buildParamsFromStruct(ep.PathParams, "path")
		for _, p := range params {
			p.SetRequired(true) // Path params are always required
			op.AddParameter(p)
		}
	}

	// Auto-extract path params from path like /users/:id or /users/{id}
	pathParams := extractPathParams(ep.Path)
	for _, paramName := range pathParams {
		// Skip if already defined
		if !hasParam(op.Parameters, paramName) {
			p := spec.NewParameter(paramName, "path").
				SetRequired(true).
				WithSchema(spec.NewSchema("string"))
			op.AddParameter(p)
		}
	}

	// Build request body
	if ep.RequestBody != nil {
		contentType := ep.RequestBody.ContentType
		if contentType == "" {
			contentType = "application/json"
		}

		var s *spec.Schema
		if ep.RequestBody.Schema != nil {
			schemaResult := schema.FromType(ep.RequestBody.Schema)
			s = convertSchema(schemaResult)
		}

		rb := spec.NewRequestBody(ep.RequestBody.Description, ep.RequestBody.Required).
			WithContent(contentType, s)
		op.WithRequestBody(rb)
	}

	// Build responses
	for code, resp := range ep.Responses {
		r := spec.NewResponse(resp.Description)

		if resp.Schema != nil {
			schemaResult := schema.FromType(resp.Schema)
			s := convertSchema(schemaResult)
			r.WithContent("application/json", s)
		}

		op.AddResponse(intToString(code), r)
	}

	// Build security
	for _, secName := range ep.Security {
		op.WithSecurity(spec.SecurityRequirement{secName: {}})
	}

	return op
}

// buildParamsFromStruct extracts parameters from a struct using reflection
func (d *Docs) buildParamsFromStruct(v interface{}, location string) []*spec.Parameter {
	var params []*spec.Parameter

	t := reflect.TypeOf(v)
	if t == nil {
		return params
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return params
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		// Get parameter name from tags — prefer location-specific tag first
		var name string
		switch location {
		case "query":
			name = field.Tag.Get("query")
			if name == "" {
				name = field.Tag.Get("form")
			}
		case "path":
			name = field.Tag.Get("path")
			if name == "" {
				name = field.Tag.Get("param")
			}
		default:
			name = field.Tag.Get("form")
			if name == "" {
				name = field.Tag.Get("query")
			}
		}
		if name == "" {
			name = field.Tag.Get("param")
		}
		if name == "" {
			name = field.Tag.Get("json")
		}
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		name = strings.Split(name, ",")[0]

		if name == "-" {
			continue
		}

		// Build schema from field type using reflect.Type directly
		fieldSchema := schema.FromReflectType(field.Type)
		specSchema := convertSchema(fieldSchema)

		// Get description and example from tags
		description := field.Tag.Get("description")
		if description == "" {
			description = field.Tag.Get("doc")
		}

		p := spec.NewParameter(name, location).
			WithDescription(description).
			WithSchema(specSchema)

		// Check if required
		validate := field.Tag.Get("validate")
		binding := field.Tag.Get("binding")
		if strings.Contains(validate, "required") || strings.Contains(binding, "required") {
			p.SetRequired(true)
		}

		// Check for example tag with type coercion
		if example := field.Tag.Get("example"); example != "" {
			p.WithExample(schema.ConvertExampleToType(example, field.Type))
		}

		params = append(params, p)
	}

	return params
}

// extractPathParams extracts parameter names from path like /users/:id or /users/{id}
func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, strings.TrimPrefix(part, ":"))
		} else if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			params = append(params, strings.Trim(part, "{}"))
		}
	}

	return params
}

// hasParam checks if a parameter with the given name already exists
func hasParam(params []*spec.Parameter, name string) bool {
	for _, p := range params {
		if p.Name == name {
			return true
		}
	}
	return false
}

func convertSchema(s *schema.Schema) *spec.Schema {
	if s == nil {
		return nil
	}

	result := &spec.Schema{
		Type:        s.Type,
		Format:      s.Format,
		Description: s.Description,
		Example:     s.Example,
		Default:     s.Default,
		Enum:        s.Enum,
		Required:    s.Required,
		Pattern:     s.Pattern,
		Minimum:     s.Minimum,
		Maximum:     s.Maximum,
		MinLength:   s.MinLength,
		MaxLength:   s.MaxLength,
		Ref:         s.Ref,
		Nullable:    s.Nullable,
		ReadOnly:    s.ReadOnly,
		WriteOnly:   s.WriteOnly,
		Deprecated:  s.Deprecated,
		MinItems:    s.MinItems,
		MaxItems:    s.MaxItems,
		UniqueItems: s.UniqueItems,
	}

	// AdditionalProperties
	if s.AdditionalProperties != nil {
		result.AdditionalProperties = convertSchema(s.AdditionalProperties)
	}

	// Items (array)
	if s.Items != nil {
		result.Items = convertSchema(s.Items)
	}

	// Properties (object)
	if len(s.Properties) > 0 {
		result.Properties = make(map[string]*spec.Schema)
		for k, v := range s.Properties {
			result.Properties[k] = convertSchema(v)
		}
	}

	// Composition
	for _, sub := range s.AllOf {
		result.AllOf = append(result.AllOf, convertSchema(sub))
	}
	for _, sub := range s.OneOf {
		result.OneOf = append(result.OneOf, convertSchema(sub))
	}
	for _, sub := range s.AnyOf {
		result.AnyOf = append(result.AnyOf, convertSchema(sub))
	}

	return result
}

func intToString(n int) string {
	return strconv.Itoa(n)
}

// SpecJSON returns the OpenAPI spec as JSON
func (d *Docs) SpecJSON() ([]byte, error) {
	openapi := d.BuildSpec()
	return json.MarshalIndent(openapi, "", "  ")
}
