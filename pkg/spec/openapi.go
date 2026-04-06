package spec

import (
	"encoding/json"
)

// OpenAPI represents the root OpenAPI specification object
type OpenAPI struct {
	OpenAPI      string                `json:"openapi"`
	Info         Info                  `json:"info"`
	Servers      []Server              `json:"servers,omitempty"`
	Paths        map[string]*PathItem  `json:"paths"`
	Components   *Components           `json:"components,omitempty"`
	Security     []SecurityRequirement `json:"security,omitempty"`
	Tags         []Tag                 `json:"tags,omitempty"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty"`
}

// SecurityRequirement represents a security requirement
type SecurityRequirement map[string][]string

// ExternalDocs represents external documentation
type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// Tag represents a tag for API documentation
type Tag struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

// NewOpenAPI creates a new OpenAPI specification
func NewOpenAPI(info Info) *OpenAPI {
	return &OpenAPI{
		OpenAPI: "3.1.0",
		Info:    info,
		Paths:   make(map[string]*PathItem),
		Components: &Components{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]*SecurityScheme),
			Parameters:      make(map[string]*Parameter),
			RequestBodies:   make(map[string]*RequestBody),
			Responses:       make(map[string]*Response),
		},
	}
}

// AddServer adds a server to the specification
func (o *OpenAPI) AddServer(server Server) *OpenAPI {
	o.Servers = append(o.Servers, server)
	return o
}

// AddTag adds a tag to the specification
func (o *OpenAPI) AddTag(tag Tag) *OpenAPI {
	o.Tags = append(o.Tags, tag)
	return o
}

// AddPath adds a path item to the specification
func (o *OpenAPI) AddPath(path string, item *PathItem) *OpenAPI {
	o.Paths[path] = item
	return o
}

// AddSchema adds a schema to components
func (o *OpenAPI) AddSchema(name string, schema *Schema) *OpenAPI {
	if o.Components == nil {
		o.Components = &Components{Schemas: make(map[string]*Schema)}
	}
	o.Components.Schemas[name] = schema
	return o
}

// AddSecurityScheme adds a security scheme to components
func (o *OpenAPI) AddSecurityScheme(name string, scheme *SecurityScheme) *OpenAPI {
	if o.Components == nil {
		o.Components = &Components{SecuritySchemes: make(map[string]*SecurityScheme)}
	}
	o.Components.SecuritySchemes[name] = scheme
	return o
}

// SetSecurity sets global security requirements
func (o *OpenAPI) SetSecurity(requirements ...SecurityRequirement) *OpenAPI {
	o.Security = requirements
	return o
}

// ToJSON serializes the specification to JSON
func (o *OpenAPI) ToJSON() ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}

// ToJSONString serializes the specification to a JSON string
func (o *OpenAPI) ToJSONString() (string, error) {
	data, err := o.ToJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}
