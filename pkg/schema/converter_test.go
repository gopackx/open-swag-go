package schema

import (
	"encoding/json"
	"reflect"
	"testing"
)

type TokenRequest struct {
	GrantType    string `json:"grant_type" validate:"required"`
	ClientID     string `json:"client_id" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"omitempty"`
}

type UserQuery struct {
	Page    int    `form:"page" description:"Page number" example:"1"`
	PerPage int    `form:"per_page" description:"Items per page" example:"20"`
	Search  string `form:"search" description:"Search term"`
}

func TestFromType_Struct(t *testing.T) {
	schema := FromType(TokenRequest{})

	if schema.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", schema.Type)
	}

	if len(schema.Properties) != 3 {
		t.Errorf("expected 3 properties, got %d", len(schema.Properties))
	}

	// Check grant_type
	if prop, ok := schema.Properties["grant_type"]; ok {
		if prop.Type != "string" {
			t.Errorf("expected grant_type type 'string', got '%s'", prop.Type)
		}
	} else {
		t.Error("missing property 'grant_type'")
	}

	// Check required fields
	if len(schema.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(schema.Required))
	}
}

func TestFromType_BasicTypes(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
		format   string
	}{
		{"string", "string", ""},
		{int(1), "integer", ""},
		{int32(1), "integer", "int32"},
		{int64(1), "integer", "int64"},
		{float32(1.0), "number", "float"},
		{float64(1.0), "number", "double"},
		{true, "boolean", ""},
	}

	for _, tt := range tests {
		schema := FromType(tt.input)
		if schema.Type != tt.expected {
			t.Errorf("FromType(%T): expected type '%s', got '%s'", tt.input, tt.expected, schema.Type)
		}
		if tt.format != "" && schema.Format != tt.format {
			t.Errorf("FromType(%T): expected format '%s', got '%s'", tt.input, tt.format, schema.Format)
		}
	}
}

func TestFromType_Array(t *testing.T) {
	schema := FromType([]string{})

	if schema.Type != "array" {
		t.Errorf("expected type 'array', got '%s'", schema.Type)
	}

	if schema.Items == nil {
		t.Error("expected Items to be set")
	} else if schema.Items.Type != "string" {
		t.Errorf("expected Items.Type 'string', got '%s'", schema.Items.Type)
	}
}

func TestFromType_JSON(t *testing.T) {
	schema := FromType(TokenRequest{})

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Generated schema:\n%s", string(data))

	// Verify it's valid JSON with proper structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	props, ok := result["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be an object")
	}

	grantType, ok := props["grant_type"].(map[string]interface{})
	if !ok {
		t.Fatal("grant_type should be an object")
	}

	if grantType["type"] != "string" {
		t.Errorf("grant_type.type should be 'string', got %v", grantType["type"])
	}

	// No auto-generated example — only set when user provides example tag
	if _, hasExample := grantType["example"]; hasExample {
		t.Errorf("grant_type should not have auto-generated example, got %v", grantType["example"])
	}
}

func TestFromType_Examples(t *testing.T) {
	type TestStruct struct {
		Name   string  `json:"name"`
		Age    int     `json:"age"`
		Score  float64 `json:"score"`
		Active bool    `json:"active"`
		Custom string  `json:"custom" example:"my-custom-value"`
	}

	schema := FromType(TestStruct{})

	// Untagged fields should NOT have auto-generated examples
	if schema.Properties["name"].Example != nil {
		t.Errorf("name.example should be nil (no tag), got %v", schema.Properties["name"].Example)
	}
	if schema.Properties["age"].Example != nil {
		t.Errorf("age.example should be nil (no tag), got %v", schema.Properties["age"].Example)
	}
	if schema.Properties["score"].Example != nil {
		t.Errorf("score.example should be nil (no tag), got %v", schema.Properties["score"].Example)
	}
	if schema.Properties["active"].Example != nil {
		t.Errorf("active.example should be nil (no tag), got %v", schema.Properties["active"].Example)
	}
	// Tagged field should have its example value
	if schema.Properties["custom"].Example != "my-custom-value" {
		t.Errorf("custom.example should be 'my-custom-value', got %v", schema.Properties["custom"].Example)
	}
}

// P0-2: embedded structs of unexported types are flattened, like encoding/json.
type base struct {
	ID        string `json:"id" validate:"required"`
	CreatedAt string `json:"created_at"`
	internal  string //nolint:unused // unexported, must be skipped
}

type audit struct {
	By string `json:"by"`
}

type secretID string

type withEmbedded struct {
	base
	*audit
	secretID        // unexported non-struct embedded: skipped
	Name     string `json:"name"`
}

func TestFromType_EmbeddedUnexportedStruct(t *testing.T) {
	s := FromType(withEmbedded{})
	for _, want := range []string{"id", "created_at", "by", "name"} {
		if _, ok := s.Properties[want]; !ok {
			t.Errorf("missing promoted property %q; got %v", want, keys(s.Properties))
		}
	}
	for _, unwanted := range []string{"internal", "secretID", "base", "audit"} {
		if _, ok := s.Properties[unwanted]; ok {
			t.Errorf("unexpected property %q", unwanted)
		}
	}
	if len(s.Required) != 1 || s.Required[0] != "id" {
		t.Errorf("expected promoted required [id], got %v", s.Required)
	}
}

func TestFromType_EmbeddedWithJSONTagIsNested(t *testing.T) {
	type outer struct {
		base `json:"base"`
		Name string `json:"name"`
	}
	s := FromType(outer{})
	nested, ok := s.Properties["base"]
	if !ok {
		t.Fatalf("expected nested property 'base', got %v", keys(s.Properties))
	}
	if _, ok := nested.Properties["id"]; !ok {
		t.Errorf("nested base should have 'id', got %v", keys(nested.Properties))
	}
	if _, ok := s.Properties["id"]; ok {
		t.Error("tagged embedded struct must not be flattened")
	}
}

func TestFromType_OuterFieldShadowsPromoted(t *testing.T) {
	type outer struct {
		ID int `json:"id"` // declared before the embedded struct
		base
	}
	s := FromType(outer{})
	if s.Properties["id"].Type != "integer" {
		t.Errorf("outer field should win over promoted field, got type %q", s.Properties["id"].Type)
	}
	for _, r := range s.Required {
		if r == "id" {
			t.Error("shadowed promoted field must not contribute to required")
		}
	}
}

// P0-4: pointers are nullable.
func TestFromType_PointerIsNullable(t *testing.T) {
	type T struct {
		Name *string `json:"name"`
		Age  int     `json:"age"`
		Next *T      `json:"next"`
	}
	s := FromType(&T{})
	if s.Nullable {
		t.Error("top-level pointer passed to FromType must not be nullable")
	}
	if !s.Properties["name"].Nullable || s.Properties["name"].Type != "string" {
		t.Errorf("*string should be nullable string, got %+v", s.Properties["name"])
	}
	if s.Properties["age"].Nullable {
		t.Error("int must not be nullable")
	}
	if !s.Properties["next"].Nullable || s.Properties["next"].Type != "object" {
		t.Errorf("*T should be nullable object, got %+v", s.Properties["next"])
	}
}

// P1-10: validator tags map to schema constraints.
func TestParseFieldTags_Validator(t *testing.T) {
	type T struct {
		Status   string   `json:"status" validate:"required,oneof=active inactive 'on hold'"`
		Level    int      `json:"level" binding:"oneof=1 2 3"`
		Age      int      `json:"age" validate:"min=18,max=99"`
		Score    float64  `json:"score" validate:"gte=0,lt=100"`
		Username string   `json:"username" validate:"min=3,max=20"`
		Code     string   `json:"code" validate:"len=6"`
		Email    string   `json:"email" validate:"required,email"`
		Site     string   `json:"site" validate:"omitempty,url"`
		Custom   string   `json:"custom" format:"my-format" validate:"email"`
		Tags     []string `json:"tags" validate:"min=1,max=5,dive,min=2"`
		Color    string   `json:"color" enum:"red, green,blue"`
		Prio     int      `json:"prio" enum:"1,2"`
		Nick     *string  `json:"nick" validate:"omitempty,oneof=a b"`
		Either   string   `json:"either" validate:"email|url"`
	}
	s := FromType(T{})
	p := s.Properties

	assertEnum(t, "status", p["status"].Enum, []interface{}{"active", "inactive", "on hold"})
	assertEnum(t, "level", p["level"].Enum, []interface{}{int64(1), int64(2), int64(3)})
	assertEnum(t, "color", p["color"].Enum, []interface{}{"red", "green", "blue"})
	assertEnum(t, "prio", p["prio"].Enum, []interface{}{int64(1), int64(2)})
	assertEnum(t, "nick", p["nick"].Enum, []interface{}{"a", "b", nil})

	assertFloat(t, "age.minimum", p["age"].Minimum, 18)
	assertFloat(t, "age.maximum", p["age"].Maximum, 99)
	if p["age"].MinLength != nil {
		t.Error("int field must not get minLength")
	}
	assertFloat(t, "score.minimum", p["score"].Minimum, 0)
	assertFloat(t, "score.maximum", p["score"].Maximum, 100)

	assertInt(t, "username.minLength", p["username"].MinLength, 3)
	assertInt(t, "username.maxLength", p["username"].MaxLength, 20)
	if p["username"].Minimum != nil {
		t.Error("string field must not get minimum")
	}
	assertInt(t, "code.minLength", p["code"].MinLength, 6)
	assertInt(t, "code.maxLength", p["code"].MaxLength, 6)

	assertInt(t, "tags.minItems", p["tags"].MinItems, 1)
	assertInt(t, "tags.maxItems", p["tags"].MaxItems, 5)
	assertInt(t, "tags.items.minLength", p["tags"].Items.MinLength, 2)

	if p["email"].Format != "email" {
		t.Errorf("email format: got %q", p["email"].Format)
	}
	if p["site"].Format != "uri" {
		t.Errorf("url format: got %q", p["site"].Format)
	}
	if p["custom"].Format != "my-format" {
		t.Errorf("explicit format tag must win, got %q", p["custom"].Format)
	}
	if p["either"].Format != "" {
		t.Errorf("alternatives must not set a format, got %q", p["either"].Format)
	}

	if !contains(s.Required, "status") || !contains(s.Required, "email") {
		t.Errorf("required should include status and email, got %v", s.Required)
	}
}

func TestIsRequired_ExactRule(t *testing.T) {
	type T struct {
		A string `validate:"required"`
		B string `validate:"required_if=A x"`
		C string `binding:"omitempty,required_without=A"`
		D string `swagger:"required,format=email"`
	}
	rt := reflect.TypeOf(T{})
	want := map[string]bool{"A": true, "B": false, "C": false, "D": true}
	for name, w := range want {
		f, _ := rt.FieldByName(name)
		if got := IsRequired(f); got != w {
			t.Errorf("IsRequired(%s) = %v, want %v", name, got, w)
		}
	}
}

func keys(m map[string]*Schema) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func assertEnum(t *testing.T, name string, got, want []interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s enum = %#v, want %#v", name, got, want)
	}
}

func assertFloat(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func assertInt(t *testing.T, name string, got *int, want int) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}
