package openswag

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/gopackx/open-swag-go/pkg/schema"
	"github.com/gopackx/open-swag-go/pkg/spec"
)

// Docs is the main documentation instance.
//
// Add never fails, so mistakes in endpoint definitions (a method+path
// registered twice, an unsupported HTTP method) do not surface on their own:
// BuildSpec keeps the last of duplicate endpoints and skips unsupported
// methods. Call Validate once at startup, after registering all endpoints,
// and treat a non-nil error as fatal:
//
//	if err := docs.Validate(); err != nil {
//		log.Fatal(err)
//	}
type Docs struct {
	config    Config
	endpoints []Endpoint
	openapi   *spec.OpenAPI
	mu        sync.RWMutex
}

// Endpoint represents an API endpoint definition
type Endpoint struct {
	Method string
	// Path may use OpenAPI style (/users/{id}) or router style (/users/:id,
	// /files/*path); router-style segments are normalised to {name} in the
	// generated spec.
	Path string
	// OperationID overrides the generated operationId. When empty, one is
	// derived from the method and path, e.g. GET /users/{id} -> getUsersById.
	OperationID string
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

// supportedMethods are the HTTP methods an OpenAPI 3.1 path item can hold.
var supportedMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
	"HEAD": true, "OPTIONS": true, "TRACE": true,
}

func (d *Docs) addEndpointToSpec(openapi *spec.OpenAPI, ep Endpoint) {
	method := strings.ToUpper(strings.TrimSpace(ep.Method))
	if !supportedMethods[method] {
		// Reported by Validate; never emit an empty path item.
		return
	}
	ep.Method = method
	ep.Path = normalizePath(ep.Path)

	pathItem := openapi.Paths[ep.Path]
	if pathItem == nil {
		pathItem = spec.NewPathItem()
	}

	operation := d.buildOperation(ep)

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
	case "HEAD":
		pathItem.SetHead(operation)
	case "OPTIONS":
		pathItem.SetOptions(operation)
	case "TRACE":
		pathItem.SetTrace(operation)
	}

	openapi.AddPath(ep.Path, pathItem)
}

// Validate checks the registered endpoints for definitions that cannot be
// represented faithfully in the spec. It reports, in one joined error:
//   - every method+path pair registered more than once (compared after
//     upper-casing the method and normalising /users/:id to /users/{id});
//   - every endpoint whose method is not an OpenAPI method (GET, POST, PUT,
//     PATCH, DELETE, HEAD, OPTIONS, TRACE);
//   - every operationId (explicit or generated) used by more than one
//     operation, since OpenAPI requires operationIds to be unique.
//
// It returns nil when all definitions are valid. Call it once at startup.
func (d *Docs) Validate() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var errs []error
	routes := make(map[string]int)       // "METHOD path" -> first endpoint index
	operationIDs := make(map[string]int) // operationId -> first endpoint index

	for i, ep := range d.endpoints {
		method := strings.ToUpper(strings.TrimSpace(ep.Method))
		path := normalizePath(ep.Path)

		if !supportedMethods[method] {
			errs = append(errs, fmt.Errorf("openswag: endpoint #%d %q %s: unsupported HTTP method", i, ep.Method, path))
			continue
		}

		key := method + " " + path
		if first, dup := routes[key]; dup {
			errs = append(errs, fmt.Errorf("openswag: duplicate endpoint %s (endpoints #%d and #%d)", key, first, i))
			continue
		}
		routes[key] = i

		opID := ep.OperationID
		if opID == "" {
			opID = generateOperationID(method, path)
		}
		if first, dup := operationIDs[opID]; dup {
			errs = append(errs, fmt.Errorf("openswag: duplicate operationId %q (endpoints #%d and #%d); set Endpoint.OperationID to disambiguate", opID, first, i))
			continue
		}
		operationIDs[opID] = i
	}

	return errors.Join(errs...)
}

// normalizePath converts router-style path parameters to OpenAPI style:
// "/users/:id" -> "/users/{id}", "/files/*path" -> "/files/{path}", and drops
// chi-style regex constraints: "/users/{id:[0-9]+}" -> "/users/{id}".
func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		switch {
		case len(part) > 1 && (part[0] == ':' || part[0] == '*'):
			parts[i] = "{" + part[1:] + "}"
		case strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}"):
			if name, _, found := strings.Cut(part[1:len(part)-1], ":"); found {
				parts[i] = "{" + name + "}"
			}
		}
	}
	return strings.Join(parts, "/")
}

// generateOperationID derives a lowerCamelCase operationId from a method and
// a normalised path: GET /users/{id}/orders -> getUsersByIdOrders.
// Non-alphanumeric characters act as word separators and are dropped.
func generateOperationID(method, path string) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(method))

	wrote := false
	for _, seg := range strings.Split(path, "/") {
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			b.WriteString("By")
			seg = seg[1 : len(seg)-1]
		}
		words := strings.FieldsFunc(seg, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		for _, word := range words {
			r := []rune(word)
			r[0] = unicode.ToUpper(r[0])
			b.WriteString(string(r))
			wrote = true
		}
	}
	if !wrote {
		b.WriteString("Root")
	}
	return b.String()
}

func (d *Docs) buildOperation(ep Endpoint) *spec.Operation {
	operationID := ep.OperationID
	if operationID == "" {
		operationID = generateOperationID(ep.Method, ep.Path)
	}

	op := spec.NewOperation(ep.Summary).
		WithOperationID(operationID).
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
		if param.Example != nil {
			p.WithExample(param.Example)
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

		// Build schema from field type, then apply the same tag rules used
		// for body schemas (description, format, example, swagger tag,
		// validator constraints, enum).
		fieldSchema := schema.FromReflectType(field.Type)
		// A pointer here means "optional", not "may be null": a query or path
		// parameter cannot carry a JSON null.
		fieldSchema.Nullable = false
		schema.ParseFieldTags(field, fieldSchema)

		// Description and example live on the parameter object rather than
		// being duplicated inside its schema.
		description := fieldSchema.Description
		if description == "" {
			description = field.Tag.Get("doc")
		}
		example := fieldSchema.Example
		fieldSchema.Description = ""
		fieldSchema.Example = nil

		p := spec.NewParameter(name, location).
			WithDescription(description).
			WithSchema(convertSchema(fieldSchema)).
			SetRequired(schema.IsRequired(field))

		if example != nil {
			p.WithExample(example)
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
