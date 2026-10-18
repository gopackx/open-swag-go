package schema

import (
	"reflect"
	"strconv"
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
//
// Every struct is inlined. Use FromTypeWithRegistry to emit named structs as
// $ref components instead.
func FromType(t interface{}) *Schema {
	return FromTypeWithRegistry(t, nil)
}

// FromReflectType converts a reflect.Type to JSON Schema. Unlike FromType, a
// pointer type here yields a nullable schema. Every struct is inlined.
func FromReflectType(t reflect.Type) *Schema {
	return FromReflectTypeWithRegistry(t, nil)
}

// FromTypeWithRegistry is FromType with named schema components. When reg is
// non-nil, every named struct type reached from v (including v's own type) is
// converted once, stored in reg, and referenced as
// {"$ref": "#/components/schemas/<Name>"}; anonymous structs stay inline.
// Self-referencing structs therefore produce a real $ref back to themselves.
// A nil reg behaves exactly like FromType.
func FromTypeWithRegistry(v interface{}, reg *Registry) *Schema {
	if v == nil {
		return &Schema{Type: "object"}
	}
	return newWalker(reg).walk(derefType(reflect.TypeOf(v)))
}

// FromReflectTypeWithRegistry is FromReflectType with named schema
// components; see FromTypeWithRegistry.
func FromReflectTypeWithRegistry(t reflect.Type, reg *Registry) *Schema {
	return newWalker(reg).walk(t)
}

// RefPrefix is the JSON pointer prefix of every $ref produced via a Registry.
const RefPrefix = "#/components/schemas/"

// Registry collects the named struct schemas produced by
// FromTypeWithRegistry, keyed by component name, in registration order.
//
// A component is named after its Go type, sanitised to the OpenAPI component
// key alphabet [A-Za-z0-9_.-]: package paths inside generic type arguments are
// reduced to the package name and other disallowed runes become '_', so
// Page[github.com/acme/dto.User] is named Page_dto.User.
//
// When a different type with an already-taken name is registered, it is named
// with its package name as a prefix (dto.User -> dtoUser);
// if that is taken too, a numeric suffix is appended (dtoUser2, dtoUser3, ...).
// The first type registered keeps the bare name, so names are deterministic
// for a given registration order.
//
// A Registry is not safe for concurrent use.
type Registry struct {
	names   []string
	schemas map[string]*Schema
	byType  map[reflect.Type]string
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		schemas: make(map[string]*Schema),
		byType:  make(map[reflect.Type]string),
	}
}

// Names returns the component names in registration order.
func (r *Registry) Names() []string {
	if r == nil {
		return nil
	}
	return append([]string(nil), r.names...)
}

// Schema returns the component schema registered under name, or nil.
func (r *Registry) Schema(name string) *Schema {
	if r == nil {
		return nil
	}
	return r.schemas[name]
}

// Len returns the number of registered components.
func (r *Registry) Len() int {
	if r == nil {
		return 0
	}
	return len(r.names)
}

// NameOf returns the component name assigned to t, if t is registered.
func (r *Registry) NameOf(t reflect.Type) (string, bool) {
	if r == nil {
		return "", false
	}
	name, ok := r.byType[t]
	return name, ok
}

// register reserves a unique name for t and stores s under it.
func (r *Registry) register(t reflect.Type, s *Schema) string {
	base := sanitizeComponentName(t.Name())
	name := base
	if _, taken := r.schemas[name]; taken {
		prefixed := sanitizeComponentName(pkgName(t)) + base
		name = prefixed
		for i := 2; ; i++ {
			if _, taken := r.schemas[name]; !taken {
				break
			}
			name = prefixed + strconv.Itoa(i)
		}
	}
	r.names = append(r.names, name)
	r.schemas[name] = s
	r.byType[t] = name
	return name
}

// isComponentType reports whether t is a named struct type that a Registry
// turns into a component. time.Time is a string, not a component.
func isComponentType(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && t.Name() != "" && t.PkgPath() != "" &&
		t != reflect.TypeOf(time.Time{})
}

// pkgName returns the Go package name of t (the identifier used in source,
// e.g. "openswag" for github.com/gopackx/open-swag-go), which is the prefix
// of t.String() before the type name. It differs from the last path element
// when the directory and package names differ.
func pkgName(t reflect.Type) string {
	if name := strings.TrimSuffix(t.String(), "."+t.Name()); name != t.String() {
		return name
	}
	pkgPath := t.PkgPath()
	if i := strings.LastIndexByte(pkgPath, '/'); i >= 0 {
		return pkgPath[i+1:]
	}
	return pkgPath
}

// sanitizeComponentName turns a Go type name into a valid OpenAPI component
// key ([A-Za-z0-9_.-]). Package paths inside generic type arguments are
// reduced to the package name, every other disallowed rune becomes '_', and
// trailing underscores (from a closing ']') are dropped.
func sanitizeComponentName(name string) string {
	var b []byte
	tokenStart := 0 // start in b of the current path-like token
	for _, r := range name {
		switch {
		case r == '/':
			b = b[:tokenStart] // drop the package path written so far
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9',
			r == '_', r == '.', r == '-':
			b = append(b, byte(r))
		default:
			b = append(b, '_')
			tokenStart = len(b)
		}
	}
	out := strings.TrimRight(string(b), "_")
	if out == "" {
		return "Schema"
	}
	return out
}

// walker carries the state of one conversion: the optional component
// registry and the inline cycle guard.
type walker struct {
	reg      *Registry
	visiting map[reflect.Type]bool
}

func newWalker(reg *Registry) *walker {
	return &walker{reg: reg, visiting: map[reflect.Type]bool{}}
}

// derefType strips any number of pointer indirections.
func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (w *walker) walk(t reflect.Type) *Schema {
	if t == nil {
		return &Schema{Type: "object"}
	}

	// Pointer types: same schema as the element, but nil marshals as JSON
	// null, so the schema is nullable.
	if t.Kind() == reflect.Ptr {
		s := w.walk(t.Elem())
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
			Items: w.walk(t.Elem()),
		}
	case reflect.Struct:
		if w.reg != nil && isComponentType(t) {
			return w.ref(t)
		}
		return w.fromStruct(t)
	case reflect.Map:
		return &Schema{
			Type:                 "object",
			AdditionalProperties: w.walk(t.Elem()),
		}
	case reflect.Interface:
		return &Schema{} // empty schema = accepts any type
	default:
		return &Schema{Type: "string"}
	}
}

// ref returns a $ref to the component for the named struct t, converting and
// registering it first if needed. The name is reserved before the fields are
// walked, so a field that refers back to t resolves to the same $ref.
func (w *walker) ref(t reflect.Type) *Schema {
	name, ok := w.reg.NameOf(t)
	if !ok {
		placeholder := &Schema{}
		name = w.reg.register(t, placeholder)
		*placeholder = *w.fromStruct(t)
	}
	return &Schema{Ref: RefPrefix + name}
}

func (w *walker) fromStruct(t reflect.Type) *Schema {
	// Cycle guard for inline conversion: a struct that references itself
	// (directly or via slice/map/pointer) would otherwise recurse forever.
	// Return an open object schema for the back-edge. With a registry, named
	// structs never get here twice (they become $refs), so this only fires
	// without one.
	if w.visiting[t] {
		return &Schema{Type: "object", Description: "Recursive reference to " + t.String()}
	}
	w.visiting[t] = true
	defer delete(w.visiting, t)

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
				// Always inline: the properties are flattened into the
				// parent, so the embedded type itself never becomes a $ref
				// (its own named-struct fields still do).
				embedded := w.fromStruct(ft)
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
		fieldSchema := w.walk(field.Type)

		// Parse additional tags
		ParseFieldTags(field, fieldSchema)
		fieldSchema = bareRef(fieldSchema)
		fieldSchema.Items = bareRef(fieldSchema.Items)
		fieldSchema.AdditionalProperties = bareRef(fieldSchema.AdditionalProperties)

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

// bareRef strips the keywords that field tags may have added next to a $ref
// (examples, enums, formats, bounds): they would constrain the referenced
// component from the outside and most tools ignore or reject them. Only the
// field's description and nullability are kept. Non-$ref schemas are returned
// unchanged.
func bareRef(s *Schema) *Schema {
	if s == nil || s.Ref == "" {
		return s
	}
	return &Schema{Ref: s.Ref, Description: s.Description, Nullable: s.Nullable}
}
