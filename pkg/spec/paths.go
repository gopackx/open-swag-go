package spec

// PathItem represents an OpenAPI path item
type PathItem struct {
	Ref         string       `json:"$ref,omitempty"`
	Summary     string       `json:"summary,omitempty"`
	Description string       `json:"description,omitempty"`
	Get         *Operation   `json:"get,omitempty"`
	Put         *Operation   `json:"put,omitempty"`
	Post        *Operation   `json:"post,omitempty"`
	Delete      *Operation   `json:"delete,omitempty"`
	Options     *Operation   `json:"options,omitempty"`
	Head        *Operation   `json:"head,omitempty"`
	Patch       *Operation   `json:"patch,omitempty"`
	Trace       *Operation   `json:"trace,omitempty"`
	Servers     []Server     `json:"servers,omitempty"`
	Parameters  []*Parameter `json:"parameters,omitempty"`
}

// Operation represents an OpenAPI operation
type Operation struct {
	Tags         []string              `json:"tags,omitempty"`
	Summary      string                `json:"summary,omitempty"`
	Description  string                `json:"description,omitempty"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty"`
	OperationID  string                `json:"operationId,omitempty"`
	Parameters   []*Parameter          `json:"parameters,omitempty"`
	RequestBody  *RequestBody          `json:"requestBody,omitempty"`
	Responses    map[string]*Response  `json:"responses"`
	Callbacks    map[string]*Callback  `json:"callbacks,omitempty"`
	Deprecated   bool                  `json:"deprecated,omitempty"`
	Security     []SecurityRequirement `json:"security,omitempty"`
	Servers      []Server              `json:"servers,omitempty"`
}

// Callback represents an OpenAPI callback
type Callback map[string]*PathItem

// NewPathItem creates a new path item
func NewPathItem() *PathItem {
	return &PathItem{}
}

// SetGet sets the GET operation
func (p *PathItem) SetGet(op *Operation) *PathItem {
	p.Get = op
	return p
}

// SetPost sets the POST operation
func (p *PathItem) SetPost(op *Operation) *PathItem {
	p.Post = op
	return p
}

// SetPut sets the PUT operation
func (p *PathItem) SetPut(op *Operation) *PathItem {
	p.Put = op
	return p
}

// SetDelete sets the DELETE operation
func (p *PathItem) SetDelete(op *Operation) *PathItem {
	p.Delete = op
	return p
}

// SetPatch sets the PATCH operation
func (p *PathItem) SetPatch(op *Operation) *PathItem {
	p.Patch = op
	return p
}

// AddParameter adds a parameter to the path item
func (p *PathItem) AddParameter(param *Parameter) *PathItem {
	p.Parameters = append(p.Parameters, param)
	return p
}

// NewOperation creates a new operation
func NewOperation(summary string) *Operation {
	return &Operation{
		Summary:   summary,
		Responses: make(map[string]*Response),
	}
}

// WithDescription sets the operation description
func (o *Operation) WithDescription(desc string) *Operation {
	o.Description = desc
	return o
}

// WithTags sets the operation tags
func (o *Operation) WithTags(tags ...string) *Operation {
	o.Tags = tags
	return o
}

// WithOperationID sets the operation ID
func (o *Operation) WithOperationID(id string) *Operation {
	o.OperationID = id
	return o
}

// AddParameter adds a parameter to the operation
func (o *Operation) AddParameter(param *Parameter) *Operation {
	o.Parameters = append(o.Parameters, param)
	return o
}

// WithRequestBody sets the request body
func (o *Operation) WithRequestBody(body *RequestBody) *Operation {
	o.RequestBody = body
	return o
}

// AddResponse adds a response to the operation
func (o *Operation) AddResponse(code string, resp *Response) *Operation {
	if o.Responses == nil {
		o.Responses = make(map[string]*Response)
	}
	o.Responses[code] = resp
	return o
}

// WithSecurity sets security requirements
func (o *Operation) WithSecurity(requirements ...SecurityRequirement) *Operation {
	o.Security = requirements
	return o
}

// SetDeprecated marks the operation as deprecated
func (o *Operation) SetDeprecated(deprecated bool) *Operation {
	o.Deprecated = deprecated
	return o
}
