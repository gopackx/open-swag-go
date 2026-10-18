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

// FromType converts a Go type to JSON Schema.
//
// A top-level pointer (e.g. FromType(&User{})) is treated as the pointed-to
// type: it only reflects how the caller passed the value, so the resulting
// schema is not marked nullable. Pointers nested inside the type (struct
// fields, slice elements, map values) do produce nullable schemas.
func FromType(t interface{}) *Schema {
	if t == nil {
		return &Schema{Type: "object"}
	}
	return fromReflectType(derefType(reflect.TypeOf(t)), map[reflect.Type]bool{})
}

// FromReflectType converts a reflect.Type to JSON Schema. Unlike FromType, a
// pointer type here yields a nullable schema.
func FromReflectType(t reflect.Type) *Schema {
	return fromReflectType(t, map[reflect.Type]bool{})
}

// derefType strips any number of pointer indirections.
func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func fromReflectType(t reflect.Type, visiting map[reflect.Type]bool) *Schema {
	if t == nil {
		return &Schema{Type: "object"}
	}

	// Pointer types: same schema as the element, but nil marshals as JSON
	// null, so the schema is nullable.
	if t.Kind() == reflect.Ptr {
		s := fromReflectType(t.Elem(), visiting)
		s.Nullable = true
		return s
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
			Items: fromReflectType(t.Elem(), visiting),
		}
	case reflect.Struct:
		return fromStruct(t, visiting)
	case reflect.Map:
		return &Schema{
			Type:                 "object",
			AdditionalProperties: fromReflectType(t.Elem(), visiting),
		}
	case reflect.Interface:
		return &Schema{} // empty schema = accepts any type
	default:
		return &Schema{Type: "string"}
	}
}

func fromStruct(t reflect.Type, visiting map[reflect.Type]bool) *Schema {
	// Cycle guard: a struct that references itself (directly or via slice/map/pointer)
	// would otherwise recurse forever. Return an open object schema for the back-edge.
	if visiting[t] {
		return &Schema{Type: "object", Description: "Recursive reference to " + t.String()}
	}
	visiting[t] = true
	defer delete(visiting, t)

	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
		Required:   []string{},
	}

	// Properties promoted from embedded structs. Kept separate so that, as in
	// encoding/json, a field declared directly on the outer struct always
	// wins over a promoted field with the same name, regardless of order.
	promoted := make(map[string]*Schema)
	var promotedOrder []string
	promotedRequired := make(map[string]bool)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Visibility rules follow encoding/json: an embedded field of an
		// unexported *struct* type is still traversed, because its exported
		// fields are promoted. Any other unexported field is ignored.
		if field.Anonymous {
			ft := derefType(field.Type)
			if !field.IsExported() && ft.Kind() != reflect.Struct {
				continue
			}
		} else if !field.IsExported() {
			continue
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

		// Embedded (anonymous) struct without a json name: flatten its
		// properties into the parent. With a json name (`json:"base"`),
		// encoding/json treats it as an ordinary named field instead, so it
		// falls through and becomes a nested property.
		if field.Anonymous && name == "" {
			ft := derefType(field.Type)
			if ft.Kind() == reflect.Struct && ft != reflect.TypeOf(time.Time{}) {
				embedded := fromReflectType(ft, visiting)
				for k, v := range embedded.Properties {
					if _, seen := promoted[k]; !seen {
						promotedOrder = append(promotedOrder, k)
					}
					promoted[k] = v
				}
				for _, r := range embedded.Required {
					promotedRequired[r] = true
				}
				continue
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
		fieldSchema := fromReflectType(field.Type, visiting)

		// Parse additional tags
		ParseFieldTags(field, fieldSchema)

		schema.Properties[name] = fieldSchema

		// Check if required — fields with omitempty should not be required
		if IsRequired(field) && !hasOmitempty {
			schema.Required = append(schema.Required, name)
		}
	}

	// Merge promoted fields that are not shadowed by direct fields.
	for _, k := range promotedOrder {
		if _, shadowed := schema.Properties[k]; shadowed {
			continue
		}
		schema.Properties[k] = promoted[k]
		if promotedRequired[k] {
			schema.Required = append(schema.Required, k)
		}
	}

	// Remove empty required array
	if len(schema.Required) == 0 {
		schema.Required = nil
	}

	return schema
}
