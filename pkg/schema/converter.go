package schema

import (
	"reflect"
	"strings"
	"time"
)

// Schema represents a JSON Schema
type Schema struct {
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Description          string             `json:"description,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	Example              interface{}        `json:"example,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MinItems             *int               `json:"minItems,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
	Nullable             bool               `json:"nullable,omitempty"`
	ReadOnly             bool               `json:"readOnly,omitempty"`
	WriteOnly            bool               `json:"writeOnly,omitempty"`
	Deprecated           bool               `json:"deprecated,omitempty"`
	AllOf                []*Schema          `json:"allOf,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty"`
	AnyOf                []*Schema          `json:"anyOf,omitempty"`
}

// FromType converts a Go type to JSON Schema
func FromType(t interface{}) *Schema {
	if t == nil {
		return &Schema{Type: "object"}
	}
	return fromReflectType(reflect.TypeOf(t))
}

// FromReflectType converts a reflect.Type to JSON Schema
func FromReflectType(t reflect.Type) *Schema {
	return fromReflectType(t)
}

func fromReflectType(t reflect.Type) *Schema {
	if t == nil {
		return &Schema{Type: "object"}
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		return fromReflectType(t.Elem())
	}

	// Handle time.Time specially
	if t == reflect.TypeOf(time.Time{}) {
		return &Schema{Type: "string", Format: "date-time"}
	}

	// Handle time.Duration
	if t == reflect.TypeOf(time.Duration(0)) {
		return &Schema{Type: "string", Format: "duration"}
	}

	// Handle []byte
	if t == reflect.TypeOf([]byte{}) {
		return &Schema{Type: "string", Format: "byte"}
	}

	// Handle uuid.UUID (array of 16 uint8 named UUID)
	if t.Kind() == reflect.Array && t.Len() == 16 && t.Elem().Kind() == reflect.Uint8 && t.Name() == "UUID" {
		return &Schema{Type: "string", Format: "uuid"}
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16:
		return &Schema{Type: "integer"}
	case reflect.Int32:
		return &Schema{Type: "integer", Format: "int32"}
	case reflect.Int64:
		return &Schema{Type: "integer", Format: "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16:
		return &Schema{Type: "integer"}
	case reflect.Uint32:
		return &Schema{Type: "integer", Format: "int32"}
	case reflect.Uint64:
		return &Schema{Type: "integer", Format: "int64"}
	case reflect.Float32:
		return &Schema{Type: "number", Format: "float"}
	case reflect.Float64:
		return &Schema{Type: "number", Format: "double"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: fromReflectType(t.Elem()),
		}
	case reflect.Struct:
		return fromStruct(t)
	case reflect.Map:
		return &Schema{
			Type:                 "object",
			AdditionalProperties: fromReflectType(t.Elem()),
		}
	case reflect.Interface:
		return &Schema{} // empty schema = accepts any type
	default:
		return &Schema{Type: "string"}
	}
}

func fromStruct(t reflect.Type) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		// Handle embedded (anonymous) structs — flatten properties into parent
		if field.Anonymous {
			ft := field.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct && ft != reflect.TypeOf(time.Time{}) {
				embedded := fromReflectType(ft)
				if embedded.Properties != nil {
					for k, v := range embedded.Properties {
						schema.Properties[k] = v
					}
				}
				if embedded.Required != nil {
					schema.Required = append(schema.Required, embedded.Required...)
				}
				continue
			}
		}

		// Get field name from json tag first, then form tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		jsonParts := strings.Split(jsonTag, ",")
		name := jsonParts[0]
		hasOmitempty := false
		for _, part := range jsonParts[1:] {
			if part == "omitempty" {
				hasOmitempty = true
				break
			}
		}

		if name == "" {
			// Fallback to form tag
			formTag := field.Tag.Get("form")
			if formTag != "" && formTag != "-" {
				name = strings.Split(formTag, ",")[0]
			}
		}
		if name == "" {
			name = field.Name
		}

		// Build schema from field type
		fieldSchema := fromReflectType(field.Type)

		// Parse additional tags
		ParseFieldTags(field, fieldSchema)

		schema.Properties[name] = fieldSchema

		// Check if required — fields with omitempty should not be required
		if IsRequired(field) && !hasOmitempty {
			schema.Required = append(schema.Required, name)
		}
	}

	// Remove empty required array
	if len(schema.Required) == 0 {
		schema.Required = nil
	}

	return schema
}
