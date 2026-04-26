package schema

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// ParseFieldTags parses struct field tags into schema
func ParseFieldTags(field reflect.StructField, schema *Schema) {
	// Parse example tag with proper type coercion
	if example := field.Tag.Get("example"); example != "" {
		schema.Example = ConvertExampleToType(example, field.Type)
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
				schema.Example = kv[1] // swagger tag examples remain string; field type not available here
			}
		}
	}
}

// ConvertExampleToType converts a string example value to the appropriate Go type
// based on the reflect.Type of the field.
func ConvertExampleToType(example string, t reflect.Type) interface{} {
	// Literal "null" — always serialize as JSON null, regardless of field type
	if example == "null" {
		return json.RawMessage("null")
	}

	// Unwrap pointer
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, err := strconv.ParseInt(example, 10, 64); err == nil {
			return v
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, err := strconv.ParseUint(example, 10, 64); err == nil {
			return v
		}
	case reflect.Float32, reflect.Float64:
		if v, err := strconv.ParseFloat(example, 64); err == nil {
			return v
		}
	case reflect.Bool:
		if v, err := strconv.ParseBool(example); err == nil {
			return v
		}
	case reflect.Slice, reflect.Array:
		var arr []interface{}
		if err := json.Unmarshal([]byte(example), &arr); err == nil {
			return arr
		}
	case reflect.Map:
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(example), &m); err == nil {
			return m
		}
	}

	return example
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
