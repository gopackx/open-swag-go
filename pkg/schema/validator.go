package schema

import (
	"fmt"
)

// ValidationError represents a schema validation error
type ValidationError struct {
	Path    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// Validator validates schemas
type Validator struct{}

// NewValidator creates a new schema validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a schema
func (v *Validator) Validate(schema *Schema) []ValidationError {
	errors := []ValidationError{}

	if schema.Type == "" && schema.Ref == "" {
		errors = append(errors, ValidationError{
			Path:    "type",
			Message: "type or $ref is required",
		})
	}

	if schema.Type == "array" && schema.Items == nil {
		errors = append(errors, ValidationError{
			Path:    "items",
			Message: "items is required for array type",
		})
	}

	if schema.Type == "object" && schema.Properties != nil {
		for name, prop := range schema.Properties {
			propErrors := v.Validate(prop)
			for _, err := range propErrors {
				errors = append(errors, ValidationError{
					Path:    fmt.Sprintf("properties.%s.%s", name, err.Path),
					Message: err.Message,
				})
			}
		}
	}

	return errors
}

// ValidateValue validates a value against a schema
func (v *Validator) ValidateValue(value interface{}, schema *Schema) []ValidationError {
	errors := []ValidationError{}

	if value == nil {
		return errors
	}

	switch schema.Type {
	case "string":
		if _, ok := value.(string); !ok {
			errors = append(errors, ValidationError{
				Path:    "",
				Message: "expected string",
			})
		}
	case "integer":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// valid
		default:
			errors = append(errors, ValidationError{
				Path:    "",
				Message: "expected integer",
			})
		}
	case "number":
		switch value.(type) {
		case float32, float64, int, int64:
			// valid
		default:
			errors = append(errors, ValidationError{
				Path:    "",
				Message: "expected number",
			})
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			errors = append(errors, ValidationError{
				Path:    "",
				Message: "expected boolean",
			})
		}
	}

	return errors
}
