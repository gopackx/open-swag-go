package openswag

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gopackx/open-swag-go/internal/testtypes"
	"github.com/gopackx/open-swag-go/pkg/schema"
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

func buildSpecMap(t *testing.T, docs *Docs) map[string]interface{} {
	t.Helper()
	js, err := docs.SpecJSON()
	if err != nil {
		t.Fatalf("SpecJSON error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(js, &m); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}
	return m
}

func specPaths(t *testing.T, m map[string]interface{}) map[string]interface{} {
	t.Helper()
	paths, ok := m["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec has no paths object: %v", m)
	}
	return paths
}

func operation(t *testing.T, m map[string]interface{}, path, method string) map[string]interface{} {
	t.Helper()
	item, ok := specPaths(t, m)[path].(map[string]interface{})
	if !ok {
		t.Fatalf("path %q missing", path)
	}
	op, ok := item[method].(map[string]interface{})
	if !ok {
		t.Fatalf("operation %s %s missing, path item: %v", method, path, item)
	}
	return op
}

// P0-1: router-style ":id" segments are written to paths as "{id}".
func TestColonPathParamsAreNormalized(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/users/:id", Responses: Responses{200: Response("ok")}})
	docs.Add(Endpoint{Method: "GET", Path: "/users/:id/orders/:orderId", Responses: Responses{200: Response("ok")}})

	m := buildSpecMap(t, docs)
	for key := range specPaths(t, m) {
		if strings.Contains(key, ":") {
			t.Errorf("path key %q still contains ':'", key)
		}
	}
	op := operation(t, m, "/users/{id}", "get")
	params, _ := op["parameters"].([]interface{})
	if len(params) != 1 || params[0].(map[string]interface{})["name"] != "id" {
		t.Errorf("expected single path param 'id', got %v", params)
	}
	operation(t, m, "/users/{id}/orders/{orderId}", "get")
}

func TestNormalizePath(t *testing.T) {
	cases := map[string]string{
		"/users/:id":         "/users/{id}",
		"/users/{id}":        "/users/{id}",
		"/files/*filepath":   "/files/{filepath}",
		"/users/{id:[0-9]+}": "/users/{id}",
		"/a/:b/c/:d":         "/a/{b}/c/{d}",
		"/static/*":          "/static/*",
		"/":                  "/",
	}
	for in, want := range cases {
		if got := normalizePath(in); got != want {
			t.Errorf("normalizePath(%q) = %q, want %q", in, got, want)
		}
	}
}

// P0-3: duplicates and unsupported methods are reported by Validate.
func TestValidateReportsDuplicatesAndUnknownMethods(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/users/{id}", Responses: Responses{200: Response("ok")}})
	docs.Add(Endpoint{Method: "get", Path: "/users/:id", Responses: Responses{200: Response("ok")}})
	docs.Add(Endpoint{Method: "FOO", Path: "/weird", Responses: Responses{200: Response("ok")}})

	err := docs.Validate()
	if err == nil {
		t.Fatal("expected Validate to return an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "duplicate endpoint GET /users/{id}") {
		t.Errorf("duplicate not reported: %s", msg)
	}
	if !strings.Contains(msg, `"FOO"`) || !strings.Contains(msg, "unsupported HTTP method") {
		t.Errorf("unsupported method not reported: %s", msg)
	}

	// The unsupported method must not produce an (empty) path item.
	m := buildSpecMap(t, docs)
	if item, present := specPaths(t, m)["/weird"]; present {
		t.Errorf("unsupported method produced a path item: %v", item)
	}
}

func TestValidateReportsDuplicateOperationIDs(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/a", OperationID: "same", Responses: Responses{200: Response("ok")}})
	docs.Add(Endpoint{Method: "GET", Path: "/b", OperationID: "same", Responses: Responses{200: Response("ok")}})
	err := docs.Validate()
	if err == nil || !strings.Contains(err.Error(), `duplicate operationId "same"`) {
		t.Fatalf("expected duplicate operationId error, got %v", err)
	}
}

func TestValidateOK(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/users/{id}", Responses: Responses{200: Response("ok")}})
	docs.Add(Endpoint{Method: "DELETE", Path: "/users/{id}", Responses: Responses{204: Response("gone")}})
	docs.Add(Endpoint{Method: "HEAD", Path: "/health", Responses: Responses{200: Response("ok")}})
	if err := docs.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeadOptionsTraceOperations(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "HEAD", Path: "/health", Responses: Responses{200: Response("ok")}})
	docs.Add(Endpoint{Method: "options", Path: "/health", Responses: Responses{204: Response("ok")}})
	docs.Add(Endpoint{Method: "TRACE", Path: "/health", Responses: Responses{200: Response("ok")}})

	m := buildSpecMap(t, docs)
	for _, method := range []string{"head", "options", "trace"} {
		operation(t, m, "/health", method)
	}
}

// P0-5: operationId is generated from method + path, and can be overridden.
func TestGenerateOperationID(t *testing.T) {
	cases := []struct{ method, path, want string }{
		{"GET", "/users/{id}", "getUsersById"},
		{"POST", "/users", "postUsers"},
		{"GET", "/users/{id}/orders/{orderId}", "getUsersByIdOrdersByOrderId"},
		{"DELETE", "/products/{id}", "deleteProductsById"},
		{"GET", "/api/v1/user-profiles", "getApiV1UserProfiles"},
		{"GET", "/", "getRoot"},
	}
	for _, c := range cases {
		if got := generateOperationID(c.method, c.path); got != c.want {
			t.Errorf("generateOperationID(%s, %s) = %q, want %q", c.method, c.path, got, c.want)
		}
	}
}

func TestOperationIDDefaultAndOverride(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/users/:id", Responses: Responses{200: Response("ok")}})
	docs.Add(Endpoint{Method: "POST", Path: "/users", OperationID: "createUser", Responses: Responses{201: Response("ok")}})

	m := buildSpecMap(t, docs)
	if got := operation(t, m, "/users/{id}", "get")["operationId"]; got != "getUsersById" {
		t.Errorf("default operationId = %v, want getUsersById", got)
	}
	if got := operation(t, m, "/users", "post")["operationId"]; got != "createUser" {
		t.Errorf("override operationId = %v, want createUser", got)
	}
}

// P1-9: query params use the same tag rules as body schemas.
func TestQueryParamTagsAreUnified(t *testing.T) {
	type query struct {
		Search string  `query:"search" swagger:"required" description:"Search term"`
		Page   int     `query:"page" doc:"Page number" example:"2" validate:"min=1"`
		Sort   string  `query:"sort" validate:"oneof=asc desc"`
		Limit  *int    `query:"limit"`
		Cond   string  `query:"cond" validate:"required_without=Search"`
		Tag    *string `query:"tag" binding:"required"`
	}
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/items", QueryParams: query{}, Responses: Responses{200: Response("ok")}})

	m := buildSpecMap(t, docs)
	params, _ := operation(t, m, "/items", "get")["parameters"].([]interface{})
	byName := map[string]map[string]interface{}{}
	for _, p := range params {
		pm := p.(map[string]interface{})
		byName[pm["name"].(string)] = pm
	}

	if byName["search"]["required"] != true {
		t.Errorf("swagger:\"required\" query param should be required: %v", byName["search"])
	}
	if byName["search"]["description"] != "Search term" {
		t.Errorf("description tag not applied: %v", byName["search"])
	}
	if byName["tag"]["required"] != true {
		t.Errorf("binding:\"required\" query param should be required: %v", byName["tag"])
	}
	if _, req := byName["cond"]["required"]; req {
		t.Errorf("required_without must not make a param required: %v", byName["cond"])
	}
	page := byName["page"]
	if page["description"] != "Page number" {
		t.Errorf("doc tag fallback not applied: %v", page)
	}
	if page["example"] != float64(2) {
		t.Errorf("example should be integer 2, got %v", page["example"])
	}
	if s := page["schema"].(map[string]interface{}); s["minimum"] != float64(1) {
		t.Errorf("validate min not applied to param schema: %v", s)
	}
	if s := byName["sort"]["schema"].(map[string]interface{}); len(s["enum"].([]interface{})) != 2 {
		t.Errorf("oneof not applied to param schema: %v", s)
	}
	// Pointer query params are optional, not nullable.
	if s := byName["limit"]["schema"].(map[string]interface{}); s["type"] != "integer" {
		t.Errorf("pointer query param should keep scalar type, got %v", s["type"])
	}
}

// P0-4: pointer fields become nullable using the OpenAPI 3.1 type array.
func TestPointerFieldsAreNullable31(t *testing.T) {
	type manager struct {
		ID string `json:"id"`
	}
	type body struct {
		Name     string   `json:"name"`
		Nickname *string  `json:"nickname"`
		Manager  *manager `json:"manager"`
		Any      *any     `json:"any"`
	}
	// InlineSchemas keeps the inline form asserted here; the components form
	// of a nullable struct is covered by TestNullableRefUsesOneOf.
	docs := New(Config{Info: Info{Title: "t", Version: "1"}, InlineSchemas: true})
	docs.Add(Endpoint{
		Method:      "POST",
		Path:        "/people",
		RequestBody: &RequestBody{Schema: &body{}},
		Responses:   Responses{200: Response("ok", body{})},
	})

	js, _ := docs.SpecJSON()
	if strings.Contains(string(js), `"nullable"`) {
		t.Errorf("OpenAPI 3.1 spec must not contain the nullable keyword: %s", js)
	}

	m := buildSpecMap(t, docs)
	op := operation(t, m, "/people", "post")
	content := op["requestBody"].(map[string]interface{})["content"].(map[string]interface{})
	reqSchema := content["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
	if reqSchema["type"] != "object" {
		t.Errorf("top-level pointer body should not be nullable, got type %v", reqSchema["type"])
	}
	props := reqSchema["properties"].(map[string]interface{})
	if got := props["name"].(map[string]interface{})["type"]; got != "string" {
		t.Errorf("plain string should keep type string, got %v", got)
	}
	nick, ok := props["nickname"].(map[string]interface{})["type"].([]interface{})
	if !ok || len(nick) != 2 || nick[0] != "string" || nick[1] != "null" {
		t.Errorf("*string should have type [string, null], got %v", props["nickname"])
	}
	mgr := props["manager"].(map[string]interface{})
	if typ, ok := mgr["type"].([]interface{}); !ok || typ[0] != "object" || typ[1] != "null" {
		t.Errorf("pointer-to-struct should have type [object, null], got %v", mgr["type"])
	}
	if _, ok := mgr["properties"].(map[string]interface{})["id"]; !ok {
		t.Errorf("pointer-to-struct should keep its properties: %v", mgr)
	}
	if anySchema := props["any"].(map[string]interface{}); len(anySchema) != 0 {
		t.Errorf("nullable any schema should stay empty, got %v", anySchema)
	}
}

// Parameter.Example on an explicitly declared parameter must reach the spec.
func TestExplicitParameterExampleIsEmitted(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{
		Method: "GET",
		Path:   "/items",
		Parameters: []Parameter{
			{Name: "page", In: "query", Example: 3},
			{Name: "q", In: "query"},
		},
		Responses: Responses{200: Response("ok")},
	})
	m := buildSpecMap(t, docs)
	params, _ := operation(t, m, "/items", "get")["parameters"].([]interface{})
	byName := map[string]map[string]interface{}{}
	for _, p := range params {
		pm := p.(map[string]interface{})
		byName[pm["name"].(string)] = pm
	}
	if byName["page"]["example"] != float64(3) {
		t.Errorf("expected example 3 on page, got %v", byName["page"])
	}
	if _, has := byName["q"]["example"]; has {
		t.Errorf("q must not have an example: %v", byName["q"])
	}
}

// --- Schema components ---

type Address struct {
	City string `json:"city"`
}

type User struct {
	Name    string   `json:"name" validate:"required"`
	Address *Address `json:"address" description:"Postal address"`
	Home    Address  `json:"home" description:"Home address"`
	Work    Address  `json:"work"`
}

// bodySchema returns the JSON media-type schema of an operation's request body.
func bodySchema(t *testing.T, op map[string]interface{}) map[string]interface{} {
	t.Helper()
	content := op["requestBody"].(map[string]interface{})["content"].(map[string]interface{})
	return content["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
}

// responseSchema returns the JSON schema of one of an operation's responses.
func responseSchema(t *testing.T, op map[string]interface{}, code string) map[string]interface{} {
	t.Helper()
	resp := op["responses"].(map[string]interface{})[code].(map[string]interface{})
	content := resp["content"].(map[string]interface{})
	return content["application/json"].(map[string]interface{})["schema"].(map[string]interface{})
}

func componentSchemas(t *testing.T, m map[string]interface{}) map[string]interface{} {
	t.Helper()
	comp, _ := m["components"].(map[string]interface{})
	schemas, ok := comp["schemas"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec has no components.schemas: %v", m["components"])
	}
	return schemas
}

func hasComponentSchemas(m map[string]interface{}) bool {
	comp, _ := m["components"].(map[string]interface{})
	_, has := comp["schemas"]
	return has
}

func TestNamedStructIsSharedComponent(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{
		Method:      "POST",
		Path:        "/users",
		RequestBody: &RequestBody{Schema: &User{}, Required: true},
		Responses:   Responses{201: Response("created", User{})},
	})
	docs.Add(Endpoint{
		Method:    "GET",
		Path:      "/users",
		Responses: Responses{200: Response("ok", []User{})},
	})

	m := buildSpecMap(t, docs)
	schemas := componentSchemas(t, m)
	if len(schemas) != 2 || schemas["User"] == nil || schemas["Address"] == nil {
		t.Fatalf("expected exactly the User and Address components, got %v", schemas)
	}

	ref := map[string]interface{}{"$ref": "#/components/schemas/User"}
	post := operation(t, m, "/users", "post")
	if got := bodySchema(t, post); !reflect.DeepEqual(got, ref) {
		t.Errorf("request body schema = %v, want %v", got, ref)
	}
	if got := responseSchema(t, post, "201"); !reflect.DeepEqual(got, ref) {
		t.Errorf("201 schema = %v, want %v", got, ref)
	}
	list := responseSchema(t, operation(t, m, "/users", "get"), "200")
	if list["type"] != "array" || !reflect.DeepEqual(list["items"], ref) {
		t.Errorf("list response = %v, want array of %v", list, ref)
	}

	user := schemas["User"].(map[string]interface{})
	if user["type"] != "object" || !reflect.DeepEqual(user["required"], []interface{}{"name"}) {
		t.Errorf("User component = %v", user)
	}
}

func TestNullableRefUsesOneOf(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/u", Responses: Responses{200: Response("ok", User{})}})

	js, _ := docs.SpecJSON()
	if strings.Contains(string(js), `"nullable"`) {
		t.Errorf("OpenAPI 3.1 spec must not contain the nullable keyword: %s", js)
	}
	user := componentSchemas(t, buildSpecMap(t, docs))["User"].(map[string]interface{})
	props := user["properties"].(map[string]interface{})

	wantNullable := map[string]interface{}{
		"description": "Postal address",
		"oneOf": []interface{}{
			map[string]interface{}{"$ref": "#/components/schemas/Address"},
			map[string]interface{}{"type": "null"},
		},
	}
	if got := props["address"]; !reflect.DeepEqual(got, wantNullable) {
		t.Errorf("*Address = %v, want %v", got, wantNullable)
	}
	wantDescribed := map[string]interface{}{"$ref": "#/components/schemas/Address", "description": "Home address"}
	if got := props["home"]; !reflect.DeepEqual(got, wantDescribed) {
		t.Errorf("Address with description = %v, want %v", got, wantDescribed)
	}
	if got := props["work"]; !reflect.DeepEqual(got, map[string]interface{}{"$ref": "#/components/schemas/Address"}) {
		t.Errorf("plain Address should be a bare $ref, got %v", got)
	}
}

func TestRecursiveStructUsesSelfRef(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/categories", Responses: Responses{200: Response("ok", []category{})}})

	m := buildSpecMap(t, docs)
	ref := map[string]interface{}{"$ref": "#/components/schemas/category"}
	if got := responseSchema(t, operation(t, m, "/categories", "get"), "200")["items"]; !reflect.DeepEqual(got, ref) {
		t.Errorf("response items = %v, want %v", got, ref)
	}
	comp := componentSchemas(t, m)["category"].(map[string]interface{})
	children := comp["properties"].(map[string]interface{})["children"].(map[string]interface{})
	if !reflect.DeepEqual(children["items"], ref) {
		t.Errorf("children.items = %v, want %v", children["items"], ref)
	}
	js, _ := docs.SpecJSON()
	if strings.Contains(string(js), "Recursive reference") {
		t.Errorf("recursion should use $ref, not the inline fallback: %s", js)
	}
}

func TestAnonymousStructStaysInline(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{
		Method: "POST",
		Path:   "/login",
		RequestBody: &RequestBody{Schema: struct {
			Username string `json:"username"`
		}{}},
		Responses: Responses{200: Response("ok")},
	})
	m := buildSpecMap(t, docs)
	body := bodySchema(t, operation(t, m, "/login", "post"))
	if body["type"] != "object" || body["properties"].(map[string]interface{})["username"] == nil {
		t.Errorf("anonymous body should be inline, got %v", body)
	}
	if hasComponentSchemas(m) {
		t.Errorf("no components.schemas expected, got %v", m["components"])
	}
}

func TestSameNameDifferentPackages(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}})
	docs.Add(Endpoint{Method: "GET", Path: "/a", Responses: Responses{200: Response("ok", User{})}})
	docs.Add(Endpoint{Method: "GET", Path: "/b", Responses: Responses{200: Response("ok", testtypes.User{})}})

	m := buildSpecMap(t, docs)
	if got := responseSchema(t, operation(t, m, "/a", "get"), "200")["$ref"]; got != "#/components/schemas/User" {
		t.Errorf("/a ref = %v", got)
	}
	if got := responseSchema(t, operation(t, m, "/b", "get"), "200")["$ref"]; got != "#/components/schemas/testtypesUser" {
		t.Errorf("/b ref = %v", got)
	}
	other := componentSchemas(t, m)["testtypesUser"].(map[string]interface{})
	if other["properties"].(map[string]interface{})["email"] == nil {
		t.Errorf("testtypesUser should be testtypes.User, got %v", other)
	}
}

// InlineSchemas: true reproduces the pre-components output: no
// components.schemas, no $ref, and every body schema equal to the plain
// inline conversion.
func TestInlineSchemasRestoresInlining(t *testing.T) {
	docs := New(Config{Info: Info{Title: "t", Version: "1"}, InlineSchemas: true})
	docs.Add(Endpoint{
		Method:      "POST",
		Path:        "/users",
		RequestBody: &RequestBody{Schema: &User{}},
		Responses:   Responses{200: Response("ok", []category{})},
	})

	m := buildSpecMap(t, docs)
	if hasComponentSchemas(m) {
		t.Fatalf("InlineSchemas must not emit components.schemas, got %v", m["components"])
	}
	js, _ := docs.SpecJSON()
	if strings.Contains(string(js), "$ref") {
		t.Errorf("InlineSchemas output must not contain $ref: %s", js)
	}

	inline := func(v interface{}) map[string]interface{} {
		raw, err := json.Marshal(convertSchema(schema.FromType(v)))
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]interface{}
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	op := operation(t, m, "/users", "post")
	if got, want := bodySchema(t, op), inline(&User{}); !reflect.DeepEqual(got, want) {
		t.Errorf("inline body = %v, want %v", got, want)
	}
	if got, want := responseSchema(t, op, "200"), inline([]category{}); !reflect.DeepEqual(got, want) {
		t.Errorf("inline response = %v, want %v", got, want)
	}
}
