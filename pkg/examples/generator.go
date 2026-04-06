package examples

import (
	"reflect"
	"strings"
	"time"
)

// Config for example generation
type Config struct {
	UseFaker     bool
	TypeExamples map[string]interface{}
}

// Generator generates example values from Go types
type Generator struct {
	config Config
}

// New creates a new example generator
func New(config Config) *Generator {
	if config.TypeExamples == nil {
		config.TypeExamples = DefaultTypeExamples()
	}
	return &Generator{config: config}
}

// DefaultTypeExamples returns default examples for common formats
func DefaultTypeExamples() map[string]interface{} {
	return map[string]interface{}{
		"uuid":      "550e8400-e29b-41d4-a716-446655440000",
		"email":     "user@example.com",
		"uri":       "https://example.com",
		"url":       "https://example.com",
		"hostname":  "api.example.com",
		"ipv4":      "192.168.1.1",
		"ipv6":      "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"date":      "2024-01-15",
		"date-time": "2024-01-15T10:30:00Z",
		"time":      "10:30:00",
		"password":  "********",
		"byte":      "U3dhZ2dlciByb2Nrcw==",
		"binary":    "<binary data>",
	}
}

// Generate creates an example from a Go type
func (g *Generator) Generate(t interface{}) interface{} {
	if t == nil {
		return nil
	}
	return g.generateFromType(reflect.TypeOf(t))
}

func (g *Generator) generateFromType(t reflect.Type) interface{} {
	if t == nil {
		return nil
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		return g.generateFromType(t.Elem())
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return 42
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return 42
	case reflect.Float32, reflect.Float64:
		return 3.14
	case reflect.Bool:
		return true
	case reflect.Slice, reflect.Array:
		elem := g.generateFromType(t.Elem())
		return []interface{}{elem}
	case reflect.Map:
		return map[string]interface{}{"key": "value"}
	case reflect.Struct:
		// Handle time.Time specially
		if t == reflect.TypeOf(time.Time{}) {
			return "2024-01-15T10:30:00Z"
		}
		return g.generateFromStruct(t)
	default:
		return nil
	}
}

func (g *Generator) generateFromStruct(t reflect.Type) map[string]interface{} {
	result := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag
		name := g.getJSONFieldName(field)
		if name == "-" {
			continue
		}

		// Check for explicit example tag first
		if example := field.Tag.Get("example"); example != "" {
			result[name] = example
			continue
		}

		// Check format for type examples
		if format := field.Tag.Get("format"); format != "" {
			if example, ok := g.config.TypeExamples[format]; ok {
				result[name] = example
				continue
			}
		}

		// Check swagger tag for format
		if swagger := field.Tag.Get("swagger"); swagger != "" {
			if format := g.extractFormat(swagger); format != "" {
				if example, ok := g.config.TypeExamples[format]; ok {
					result[name] = example
					continue
				}
			}
		}

		// Generate based on field name heuristics
		if example := g.guessFromFieldName(name, field.Type); example != nil {
			result[name] = example
			continue
		}

		// Generate based on type
		result[name] = g.generateFromType(field.Type)
	}

	return result
}

func (g *Generator) getJSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		return field.Name
	}
	return parts[0]
}

func (g *Generator) extractFormat(swagger string) string {
	parts := strings.Split(swagger, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == "format" {
			return strings.TrimSpace(kv[1])
		}
	}
	return ""
}

// guessFromFieldName tries to guess example based on common field names
func (g *Generator) guessFromFieldName(name string, t reflect.Type) interface{} {
	lower := strings.ToLower(name)

	// Common field name patterns
	switch {
	case strings.Contains(lower, "email"):
		return "user@example.com"
	case strings.Contains(lower, "phone"):
		return "+1-555-123-4567"
	case lower == "id" || strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "id"):
		if t.Kind() == reflect.String {
			return "550e8400-e29b-41d4-a716-446655440000"
		}
		return 1
	case strings.Contains(lower, "name"):
		return "John Doe"
	case strings.Contains(lower, "url") || strings.Contains(lower, "link"):
		return "https://example.com"
	case strings.Contains(lower, "password"):
		return "********"
	case strings.Contains(lower, "token"):
		return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
	case strings.Contains(lower, "created") || strings.Contains(lower, "updated"):
		return "2024-01-15T10:30:00Z"
	case strings.Contains(lower, "date"):
		return "2024-01-15"
	case strings.Contains(lower, "time"):
		return "10:30:00"
	case strings.Contains(lower, "age"):
		return 25
	case strings.Contains(lower, "count") || strings.Contains(lower, "total"):
		return 100
	case strings.Contains(lower, "price") || strings.Contains(lower, "amount"):
		return 99.99
	case strings.Contains(lower, "description"):
		return "A sample description"
	case strings.Contains(lower, "title"):
		return "Sample Title"
	case strings.Contains(lower, "address"):
		return "123 Main St, City, Country"
	case strings.Contains(lower, "status"):
		return "active"
	case strings.Contains(lower, "type"):
		return "default"
	}

	return nil
}

// GenerateJSON generates example and returns as map suitable for JSON
func (g *Generator) GenerateJSON(t interface{}) map[string]interface{} {
	result := g.Generate(t)
	if m, ok := result.(map[string]interface{}); ok {
		return m
	}
	return nil
}
