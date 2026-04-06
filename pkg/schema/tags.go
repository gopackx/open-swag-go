package schema

import (
	"reflect"
	"strings"
)

// ParseFieldTags parses struct field tags into schema
func ParseFieldTags(field reflect.StructField, schema *Schema) {
	// Parse example tag
	if example := field.Tag.Get("example"); example != "" {
		schema.Example = example
	}

	// Parse description tag
	if desc := field.Tag.Get("description"); desc != "" {
		schema.Description = desc
	}

	// Parse format tag
	if format := field.Tag.Get("format"); format != "" {
		schema.Format = format
	}

	// Parse swagger tag
	if swagger := field.Tag.Get("swagger"); swagger != "" {
		parseSwaggerTag(swagger, schema)
	}
}

func parseSwaggerTag(tag string, schema *Schema) {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		key := strings.TrimSpace(kv[0])

		switch key {
		case "format":
			if len(kv) > 1 {
				schema.Format = kv[1]
			}
		case "description":
			if len(kv) > 1 {
				schema.Description = kv[1]
			}
		case "example":
			if len(kv) > 1 {
				schema.Example = kv[1]
			}
		}
	}
}

// IsRequired checks if a field is required based on tags
func IsRequired(field reflect.StructField) bool {
	if swagger := field.Tag.Get("swagger"); strings.Contains(swagger, "required") {
		return true
	}
	if validate := field.Tag.Get("validate"); strings.Contains(validate, "required") {
		return true
	}
	if binding := field.Tag.Get("binding"); strings.Contains(binding, "required") {
		return true
	}
	return false
}
