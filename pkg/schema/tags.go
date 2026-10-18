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

	// Map go-playground/validator rules (used by Gin's `binding` tag too)
	// onto JSON Schema constraints.
	for _, tagName := range []string{"validate", "binding"} {
		if rules := field.Tag.Get(tagName); rules != "" {
			applyValidatorRules(rules, field.Type, schema)
		}
	}

	// Explicit enum tag: `enum:"a,b,c"`. Takes precedence over `oneof`.
	if enum := field.Tag.Get("enum"); enum != "" {
		var values []string
		for _, v := range strings.Split(enum, ",") {
			if v = strings.TrimSpace(v); v != "" {
				values = append(values, v)
			}
		}
		setEnum(schema, values, field.Type)
	}

	// Parse swagger tag
	if swagger := field.Tag.Get("swagger"); swagger != "" {
		parseSwaggerTag(swagger, schema)
	}
}

// validatorFormats maps validator rule names to OpenAPI formats.
var validatorFormats = map[string]string{
	"email":    "email",
	"uuid":     "uuid",
	"uuid3":    "uuid",
	"uuid4":    "uuid",
	"uuid5":    "uuid",
	"url":      "uri",
	"uri":      "uri",
	"ipv4":     "ipv4",
	"ipv6":     "ipv6",
	"hostname": "hostname",
}

// applyValidatorRules maps a validator tag (e.g. "required,min=1,max=10") to
// schema constraints for a value of type t. Rules after "dive" apply to the
// element type of a slice/array and are applied to schema.Items.
//
// Unknown rules, rules with "|" alternatives and parameters that do not parse
// as numbers are ignored: the schema stays less strict rather than wrong.
func applyValidatorRules(rules string, t reflect.Type, schema *Schema) {
	if schema == nil || t == nil {
		return
	}
	t = derefType(t)

	parts := strings.Split(rules, ",")
	for i, rule := range parts {
		rule = strings.TrimSpace(rule)
		if rule == "dive" {
			if (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) && schema.Items != nil {
				applyValidatorRules(strings.Join(parts[i+1:], ","), t.Elem(), schema.Items)
			}
			return
		}
		if rule == "" || strings.Contains(rule, "|") {
			continue
		}

		key, param, _ := strings.Cut(rule, "=")
		switch key {
		case "oneof":
			setEnum(schema, splitOneOf(param), t)
		case "min", "gte":
			if v, ok := parseNumber(param); ok {
				setLowerBound(schema, t, v, 0)
			}
		case "gt":
			if v, ok := parseNumber(param); ok {
				setLowerBound(schema, t, v, 1)
			}
		case "max", "lte":
			if v, ok := parseNumber(param); ok {
				setUpperBound(schema, t, v, 0)
			}
		case "lt":
			if v, ok := parseNumber(param); ok {
				setUpperBound(schema, t, v, 1)
			}
		case "len":
			if v, ok := parseNumber(param); ok {
				setLowerBound(schema, t, v, 0)
				setUpperBound(schema, t, v, 0)
			}
		default:
			if format, ok := validatorFormats[key]; ok && schema.Format == "" {
				schema.Format = format
			}
		}
	}
}

// boundKind classifies how min/max-style rules apply to a type.
func boundKind(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "length"
	case reflect.Slice, reflect.Array:
		if t == reflect.TypeOf([]byte{}) {
			return "" // encoded as a base64 string; byte counts do not map
		}
		return "items"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	}
	return ""
}

// setLowerBound applies a lower bound. For lengths and item counts, which are
// integers, an exclusive bound (gt) is expressed exactly as v+1. For numbers,
// an exclusive bound is approximated by the inclusive Minimum.
func setLowerBound(schema *Schema, t reflect.Type, v float64, exclusiveStep int) {
	switch boundKind(t) {
	case "number":
		schema.Minimum = &v
	case "length":
		n := int(v) + exclusiveStep
		schema.MinLength = &n
	case "items":
		n := int(v) + exclusiveStep
		schema.MinItems = &n
	}
}

// setUpperBound is the counterpart of setLowerBound (lt on counts is v-1).
func setUpperBound(schema *Schema, t reflect.Type, v float64, exclusiveStep int) {
	switch boundKind(t) {
	case "number":
		schema.Maximum = &v
	case "length":
		n := int(v) - exclusiveStep
		schema.MaxLength = &n
	case "items":
		n := int(v) - exclusiveStep
		schema.MaxItems = &n
	}
}

func parseNumber(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v, err == nil
}

// splitOneOf splits a validator oneof parameter on spaces, honouring
// single-quoted values that contain spaces: oneof='red green' blue.
func splitOneOf(param string) []string {
	var values []string
	var cur strings.Builder
	inQuote, hasToken := false, false
	for _, r := range param {
		switch {
		case r == '\'':
			inQuote = !inQuote
			hasToken = true
		case r == ' ' && !inQuote:
			if hasToken {
				values = append(values, cur.String())
				cur.Reset()
				hasToken = false
			}
		default:
			cur.WriteRune(r)
			hasToken = true
		}
	}
	if hasToken {
		values = append(values, cur.String())
	}
	return values
}

// setEnum sets schema.Enum from string values, converting each to the field's
// type (so an int field gets integer enum values). If the schema is nullable,
// null is added so that the enum does not reject it.
func setEnum(schema *Schema, values []string, t reflect.Type) {
	if len(values) == 0 {
		return
	}
	enum := make([]interface{}, 0, len(values)+1)
	for _, v := range values {
		enum = append(enum, ConvertExampleToType(v, t))
	}
	if schema.Nullable {
		enum = append(enum, nil)
	}
	schema.Enum = enum
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

// IsRequired checks if a field is required based on tags: a bare "required"
// rule in the swagger, validate or binding tag. Conditional validator rules
// such as required_if or required_without do not make a field required.
func IsRequired(field reflect.StructField) bool {
	for _, tagName := range []string{"swagger", "validate", "binding"} {
		for _, rule := range strings.Split(field.Tag.Get(tagName), ",") {
			if strings.TrimSpace(rule) == "required" {
				return true
			}
		}
	}
	return false
}
