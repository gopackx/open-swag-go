package schema

import (
	"encoding/json"
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

	if grantType["example"] != "string" {
		t.Errorf("grant_type.example should be 'string', got %v", grantType["example"])
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

	// Check default examples
	if schema.Properties["name"].Example != "string" {
		t.Errorf("name.example should be 'string', got %v", schema.Properties["name"].Example)
	}
	if schema.Properties["age"].Example != 0 {
		t.Errorf("age.example should be 0, got %v", schema.Properties["age"].Example)
	}
	if schema.Properties["score"].Example != 0.0 {
		t.Errorf("score.example should be 0.0, got %v", schema.Properties["score"].Example)
	}
	if schema.Properties["active"].Example != false {
		t.Errorf("active.example should be false, got %v", schema.Properties["active"].Example)
	}
	// Check custom example from tag
	if schema.Properties["custom"].Example != "my-custom-value" {
		t.Errorf("custom.example should be 'my-custom-value', got %v", schema.Properties["custom"].Example)
	}
}
