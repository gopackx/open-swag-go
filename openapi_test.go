package openswag

import (
	"encoding/json"
	"strings"
	"testing"
)

// O1: Generating a schema for a self-referential struct must not stack-overflow.
type category struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Children []category `json:"children,omitempty"`
}

func TestSelfReferentialStructDoesNotPanic(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{
		Method:    "GET",
		Path:      "/categories",
		Responses: Responses{200: Response("ok", category{})},
	})
	js, err := docs.SpecJSON()
	if err != nil {
		t.Fatalf("SpecJSON error: %v", err)
	}
	if !strings.Contains(string(js), "/categories") {
		t.Fatalf("expected /categories in spec, got: %s", js)
	}
}

// O2: RequestBody.ContentType must be honored.
func TestRequestBodyContentTypeIsHonored(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{
		Method: "POST",
		Path:   "/upload",
		RequestBody: &RequestBody{
			Schema:      struct{ File string }{},
			ContentType: "multipart/form-data",
			Required:    true,
		},
		Responses: Responses{200: Response("ok")},
	})
	js, _ := docs.SpecJSON()
	if !strings.Contains(string(js), "multipart/form-data") {
		t.Fatalf("expected multipart/form-data, got: %s", js)
	}
	if strings.Contains(string(js), `"application/json"`) {
		t.Fatalf("application/json should not appear when ContentType is multipart: %s", js)
	}
}

// O3: A custom scheme name with no Config.Auth registration must NOT silently
// become http-bearer. It should simply be absent from components.securitySchemes
// (so a downstream linter would flag the unresolved security requirement).
func TestUnknownSchemeIsNotAutoBearer(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{
		Method:    "GET",
		Path:      "/x",
		Security:  []string{"weirdCustomScheme"},
		Responses: Responses{200: Response("ok")},
	})
	js, _ := docs.SpecJSON()
	var spec map[string]interface{}
	_ = json.Unmarshal(js, &spec)
	comp, _ := spec["components"].(map[string]interface{})
	sec, _ := comp["securitySchemes"].(map[string]interface{})
	if _, present := sec["weirdCustomScheme"]; present {
		t.Fatalf("unknown scheme should not be auto-registered, got: %s", js)
	}
}

// O3 (positive): A custom scheme registered via Config.Auth.Schemes IS exposed
// with the user-provided type, not coerced to bearer.
func TestCustomSchemeViaAuthConfigIsHonored(t *testing.T) {
	docs := New(Config{
		Info: Info{Title: "t", Version: "1"},
		Auth: AuthConfig{
			Schemes: []AuthScheme{APIKeyAuth("myKey", "X-Custom-Key")},
		},
	})
	docs.Add(Endpoint{
		Method:    "GET",
		Path:      "/x",
		Security:  []string{"myKey"},
		Responses: Responses{200: Response("ok")},
	})
	js, _ := docs.SpecJSON()
	if !strings.Contains(string(js), `"X-Custom-Key"`) {
		t.Fatalf("expected X-Custom-Key in spec, got: %s", js)
	}
	if strings.Contains(string(js), `"bearerFormat"`) {
		t.Fatalf("scheme should be apiKey, not bearer, got: %s", js)
	}
}
