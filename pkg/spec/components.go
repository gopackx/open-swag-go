package spec

// Components represents the OpenAPI components object
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty"`
	Examples        map[string]*Example        `json:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"`
	Headers         map[string]*Header         `json:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
	Links           map[string]*Link           `json:"links,omitempty"`
	Callbacks       map[string]*Callback       `json:"callbacks,omitempty"`
	PathItems       map[string]*PathItem       `json:"pathItems,omitempty"`
}

// Schema represents a JSON Schema
type Schema struct {
	Ref                  string             `json:"$ref,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              any                `json:"default,omitempty"`
	Example              any                `json:"example,omitempty"`
	Enum                 []any              `json:"enum,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
	AllOf                []*Schema          `json:"allOf,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty"`
	AnyOf                []*Schema          `json:"anyOf,omitempty"`
	Not                  *Schema            `json:"not,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MinItems             *int               `json:"minItems,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty"`
	Nullable             bool               `json:"nullable,omitempty"`
	ReadOnly             bool               `json:"readOnly,omitempty"`
	WriteOnly            bool               `json:"writeOnly,omitempty"`
	Deprecated           bool               `json:"deprecated,omitempty"`
}

// Response represents an OpenAPI response
type Response struct {
	Ref         string                `json:"$ref,omitempty"`
	Description string                `json:"description"`
	Headers     map[string]*Header    `json:"headers,omitempty"`
	Content     map[string]*MediaType `json:"content,omitempty"`
	Links       map[string]*Link      `json:"links,omitempty"`
}

// Parameter represents an OpenAPI parameter
type Parameter struct {
	Ref             string                `json:"$ref,omitempty"`
	Name            string                `json:"name"`
	In              string                `json:"in"`
	Description     string                `json:"description,omitempty"`
	Required        bool                  `json:"required,omitempty"`
	Deprecated      bool                  `json:"deprecated,omitempty"`
	AllowEmptyValue bool                  `json:"allowEmptyValue,omitempty"`
	Style           string                `json:"style,omitempty"`
	Explode         bool                  `json:"explode,omitempty"`
	AllowReserved   bool                  `json:"allowReserved,omitempty"`
	Schema          *Schema               `json:"schema,omitempty"`
	Example         any                   `json:"example,omitempty"`
	Examples        map[string]*Example   `json:"examples,omitempty"`
	Content         map[string]*MediaType `json:"content,omitempty"`
}

// RequestBody represents an OpenAPI request body
type RequestBody struct {
	Ref         string                `json:"$ref,omitempty"`
	Description string                `json:"description,omitempty"`
	Content     map[string]*MediaType `json:"content"`
	Required    bool                  `json:"required,omitempty"`
}

// MediaType represents an OpenAPI media type
type MediaType struct {
	Schema   *Schema              `json:"schema,omitempty"`
	Example  any                  `json:"example,omitempty"`
	Examples map[string]*Example  `json:"examples,omitempty"`
	Encoding map[string]*Encoding `json:"encoding,omitempty"`
}

// Example represents an OpenAPI example
type Example struct {
	Ref           string `json:"$ref,omitempty"`
	Summary       string `json:"summary,omitempty"`
	Description   string `json:"description,omitempty"`
	Value         any    `json:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty"`
}

// Header represents an OpenAPI header
type Header struct {
	Ref             string                `json:"$ref,omitempty"`
	Description     string                `json:"description,omitempty"`
	Required        bool                  `json:"required,omitempty"`
	Deprecated      bool                  `json:"deprecated,omitempty"`
	AllowEmptyValue bool                  `json:"allowEmptyValue,omitempty"`
	Style           string                `json:"style,omitempty"`
	Explode         bool                  `json:"explode,omitempty"`
	AllowReserved   bool                  `json:"allowReserved,omitempty"`
	Schema          *Schema               `json:"schema,omitempty"`
	Example         any                   `json:"example,omitempty"`
	Examples        map[string]*Example   `json:"examples,omitempty"`
	Content         map[string]*MediaType `json:"content,omitempty"`
}

// SecurityScheme represents an OpenAPI security scheme
type SecurityScheme struct {
	Ref              string      `json:"$ref,omitempty"`
	Type             string      `json:"type"`
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
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

// Link represents an OpenAPI link
type Link struct {
	Ref          string         `json:"$ref,omitempty"`
	OperationRef string         `json:"operationRef,omitempty"`
	OperationID  string         `json:"operationId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
	RequestBody  any            `json:"requestBody,omitempty"`
	Description  string         `json:"description,omitempty"`
	Server       *Server        `json:"server,omitempty"`
}

// Encoding represents encoding for a media type
type Encoding struct {
	ContentType   string             `json:"contentType,omitempty"`
	Headers       map[string]*Header `json:"headers,omitempty"`
	Style         string             `json:"style,omitempty"`
	Explode       bool               `json:"explode,omitempty"`
	AllowReserved bool               `json:"allowReserved,omitempty"`
}

// NewSchema creates a new schema
func NewSchema(schemaType string) *Schema {
	return &Schema{Type: schemaType}
}

// NewResponse creates a new response
func NewResponse(description string) *Response {
	return &Response{Description: description}
}

// WithContent adds content to a response
func (r *Response) WithContent(mediaType string, schema *Schema) *Response {
	if r.Content == nil {
		r.Content = make(map[string]*MediaType)
	}
	r.Content[mediaType] = &MediaType{Schema: schema}
	return r
}

// NewParameter creates a new parameter
func NewParameter(name, in string) *Parameter {
	return &Parameter{Name: name, In: in}
}

// PathParam creates a path parameter
func PathParam(name string) *Parameter {
	return &Parameter{Name: name, In: "path", Required: true}
}

// QueryParam creates a query parameter
func QueryParam(name string) *Parameter {
	return &Parameter{Name: name, In: "query"}
}

// HeaderParam creates a header parameter
func HeaderParam(name string) *Parameter {
	return &Parameter{Name: name, In: "header"}
}

// CookieParam creates a cookie parameter
func CookieParam(name string) *Parameter {
	return &Parameter{Name: name, In: "cookie"}
}

// WithDescription sets the parameter description
func (p *Parameter) WithDescription(desc string) *Parameter {
	p.Description = desc
	return p
}

// WithSchema sets the parameter schema
func (p *Parameter) WithSchema(schema *Schema) *Parameter {
	p.Schema = schema
	return p
}

// SetRequired sets whether the parameter is required
func (p *Parameter) SetRequired(required bool) *Parameter {
	p.Required = required
	return p
}

// WithExample sets the parameter example
func (p *Parameter) WithExample(example any) *Parameter {
	p.Example = example
	return p
}

// NewRequestBody creates a new request body
func NewRequestBody(description string, required bool) *RequestBody {
	return &RequestBody{
		Description: description,
		Required:    required,
		Content:     make(map[string]*MediaType),
	}
}

// WithJSONContent adds JSON content to a request body
func (rb *RequestBody) WithJSONContent(schema *Schema) *RequestBody {
	rb.Content["application/json"] = &MediaType{Schema: schema}
	return rb
}
