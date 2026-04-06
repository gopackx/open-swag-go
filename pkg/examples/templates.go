package examples

// Template represents an example template
type Template struct {
	Name        string
	Description string
	Value       any
}

// TemplateRegistry holds predefined example templates
type TemplateRegistry struct {
	templates map[string]Template
}

// NewTemplateRegistry creates a new template registry with defaults
func NewTemplateRegistry() *TemplateRegistry {
	r := &TemplateRegistry{
		templates: make(map[string]Template),
	}
	r.registerDefaults()
	return r
}

// Register adds a template to the registry
func (r *TemplateRegistry) Register(name string, template Template) {
	r.templates[name] = template
}

// Get retrieves a template by name
func (r *TemplateRegistry) Get(name string) (Template, bool) {
	t, exists := r.templates[name]
	return t, exists
}

// GetValue retrieves just the value of a template
func (r *TemplateRegistry) GetValue(name string) (any, bool) {
	t, exists := r.templates[name]
	if !exists {
		return nil, false
	}
	return t.Value, true
}

// All returns all registered templates
func (r *TemplateRegistry) All() map[string]Template {
	return r.templates
}

func (r *TemplateRegistry) registerDefaults() {
	// User templates
	r.Register("user", Template{
		Name:        "user",
		Description: "A typical user object",
		Value: map[string]any{
			"id":        "550e8400-e29b-41d4-a716-446655440000",
			"name":      "John Doe",
			"email":     "john.doe@example.com",
			"createdAt": "2024-01-15T10:30:00Z",
		},
	})

	r.Register("userList", Template{
		Name:        "userList",
		Description: "A list of users",
		Value: []map[string]any{
			{
				"id":    "550e8400-e29b-41d4-a716-446655440000",
				"name":  "John Doe",
				"email": "john.doe@example.com",
			},
			{
				"id":    "550e8400-e29b-41d4-a716-446655440001",
				"name":  "Jane Smith",
				"email": "jane.smith@example.com",
			},
		},
	})

	// Error templates
	r.Register("error", Template{
		Name:        "error",
		Description: "A standard error response",
		Value: map[string]any{
			"error":   "Bad Request",
			"message": "The request was invalid",
			"code":    400,
		},
	})

	r.Register("validationError", Template{
		Name:        "validationError",
		Description: "A validation error response",
		Value: map[string]any{
			"error":   "Validation Error",
			"message": "One or more fields are invalid",
			"details": []map[string]any{
				{"field": "email", "message": "Invalid email format"},
				{"field": "name", "message": "Name is required"},
			},
		},
	})

	r.Register("notFound", Template{
		Name:        "notFound",
		Description: "A not found error response",
		Value: map[string]any{
			"error":   "Not Found",
			"message": "The requested resource was not found",
			"code":    404,
		},
	})

	r.Register("unauthorized", Template{
		Name:        "unauthorized",
		Description: "An unauthorized error response",
		Value: map[string]any{
			"error":   "Unauthorized",
			"message": "Authentication required",
			"code":    401,
		},
	})

	// Pagination templates
	r.Register("pagination", Template{
		Name:        "pagination",
		Description: "Pagination metadata",
		Value: map[string]any{
			"page":       1,
			"perPage":    20,
			"total":      100,
			"totalPages": 5,
		},
	})

	r.Register("paginatedList", Template{
		Name:        "paginatedList",
		Description: "A paginated list response",
		Value: map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Item 1"},
				{"id": 2, "name": "Item 2"},
			},
			"pagination": map[string]any{
				"page":       1,
				"perPage":    20,
				"total":      100,
				"totalPages": 5,
			},
		},
	})

	// Auth templates
	r.Register("loginRequest", Template{
		Name:        "loginRequest",
		Description: "A login request body",
		Value: map[string]any{
			"email":    "user@example.com",
			"password": "********",
		},
	})

	r.Register("loginResponse", Template{
		Name:        "loginResponse",
		Description: "A login response with token",
		Value: map[string]any{
			"token":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			"expiresIn": 3600,
			"tokenType": "Bearer",
		},
	})

	// CRUD templates
	r.Register("createRequest", Template{
		Name:        "createRequest",
		Description: "A generic create request",
		Value: map[string]any{
			"name":        "New Item",
			"description": "A description of the item",
		},
	})

	r.Register("updateRequest", Template{
		Name:        "updateRequest",
		Description: "A generic update request",
		Value: map[string]any{
			"name":        "Updated Item",
			"description": "An updated description",
		},
	})

	r.Register("deleteResponse", Template{
		Name:        "deleteResponse",
		Description: "A delete confirmation response",
		Value: map[string]any{
			"success": true,
			"message": "Resource deleted successfully",
		},
	})
}
