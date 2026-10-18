package versioning

import (
	"strings"
	"testing"
)

// O4: A field whose type changes between versions must be flagged as a
// breaking change (BreakingTypeChanged), not silently accepted.
func TestCompareOperationsDetectsTypeChange(t *testing.T) {
	old := map[string]interface{}{
		"info": map[string]interface{}{"version": "1.0"},
		"paths": map[string]interface{}{
			"/u": map[string]interface{}{
				"post": map[string]interface{}{
					"requestBody": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"properties": map[string]interface{}{
										"age": map[string]interface{}{"type": "integer"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{"200": map[string]interface{}{}},
				},
			},
		},
	}
	newSpec := map[string]interface{}{
		"info": map[string]interface{}{"version": "1.1"},
		"paths": map[string]interface{}{
			"/u": map[string]interface{}{
				"post": map[string]interface{}{
					"requestBody": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"properties": map[string]interface{}{
										"age": map[string]interface{}{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{"200": map[string]interface{}{}},
				},
			},
		},
	}

	diff, err := NewDiffer().Compare(old, newSpec)
	if err != nil {
		t.Fatalf("Compare error: %v", err)
	}
	if !diff.HasBreakingChanges() {
		t.Fatalf("expected breaking changes, got none")
	}
	found := false
	for _, c := range diff.Changes {
		if c.IsBreaking && strings.Contains(c.Description, "age") && strings.Contains(c.Description, "type changed") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected type-change entry for 'age', got: %#v", diff.Changes)
	}
}

func TestCompareOperationsDetectsParamTypeChange(t *testing.T) {
	mk := func(typ string) map[string]interface{} {
		return map[string]interface{}{
			"info": map[string]interface{}{"version": "1"},
			"paths": map[string]interface{}{
				"/u/{id}": map[string]interface{}{
					"get": map[string]interface{}{
						"parameters": []interface{}{
							map[string]interface{}{
								"name":   "id",
								"in":     "path",
								"schema": map[string]interface{}{"type": typ},
							},
						},
						"responses": map[string]interface{}{"200": map[string]interface{}{}},
					},
				},
			},
		}
	}
	diff, _ := NewDiffer().Compare(mk("integer"), mk("string"))
	if !diff.HasBreakingChanges() {
		t.Fatalf("expected breaking changes")
	}
}

// OpenAPI 3.1 nullable fields use a type array; the "null" member must not
// count as a type change, but a real type change inside the array must.
func TestSchemaTypeSignatureHandlesTypeArrays(t *testing.T) {
	cases := []struct {
		schema map[string]interface{}
		want   string
	}{
		{map[string]interface{}{"type": "string"}, "string"},
		{map[string]interface{}{"type": []interface{}{"string", "null"}}, "string"},
		{map[string]interface{}{"type": []interface{}{"integer", "null"}, "format": "int64"}, "integer/int64"},
		{map[string]interface{}{"type": []interface{}{"null"}}, ""},
		{map[string]interface{}{}, ""},
	}
	for _, c := range cases {
		if got := schemaTypeSignature(c.schema); got != c.want {
			t.Errorf("schemaTypeSignature(%v) = %q, want %q", c.schema, got, c.want)
		}
	}
}

func TestNullableAddedIsNotTypeChange(t *testing.T) {
	mk := func(typ interface{}) map[string]interface{} {
		return map[string]interface{}{
			"info": map[string]interface{}{"version": "1"},
			"paths": map[string]interface{}{
				"/u": map[string]interface{}{
					"post": map[string]interface{}{
						"requestBody": map[string]interface{}{
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"properties": map[string]interface{}{
											"name": map[string]interface{}{"type": typ},
										},
									},
								},
							},
						},
						"responses": map[string]interface{}{"200": map[string]interface{}{}},
					},
				},
			},
		}
	}
	diff, _ := NewDiffer().Compare(mk("string"), mk([]interface{}{"string", "null"}))
	if diff.HasBreakingChanges() {
		t.Fatalf("allowing null should not be breaking, got: %#v", diff.Changes)
	}
	diff, _ = NewDiffer().Compare(mk([]interface{}{"string", "null"}), mk([]interface{}{"integer", "null"}))
	if !diff.HasBreakingChanges() {
		t.Fatalf("string->integer inside a nullable type array should be breaking")
	}
}

// specWithUserRef builds a spec whose POST /u body is a $ref to the User
// component.
func specWithUserRef(version string, user map[string]interface{}, extra map[string]interface{}) map[string]interface{} {
	schemas := map[string]interface{}{"User": user}
	for k, v := range extra {
		schemas[k] = v
	}
	return map[string]interface{}{
		"info": map[string]interface{}{"version": version},
		"paths": map[string]interface{}{
			"/u": map[string]interface{}{
				"post": map[string]interface{}{
					"requestBody": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{"$ref": "#/components/schemas/User"},
							},
						},
					},
					"responses": map[string]interface{}{"200": map[string]interface{}{}},
				},
			},
		},
		"components": map[string]interface{}{"schemas": schemas},
	}
}

func breakingContaining(diff *Diff, substr string) bool {
	for _, c := range diff.Changes {
		if c.IsBreaking && strings.Contains(c.Description, substr) {
			return true
		}
	}
	return false
}

func TestCompareResolvesRefForRequiredFields(t *testing.T) {
	user := func(required ...interface{}) map[string]interface{} {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name":  map[string]interface{}{"type": "string"},
				"email": map[string]interface{}{"type": "string"},
			},
			"required": required,
		}
	}
	diff, err := NewDiffer().Compare(
		specWithUserRef("1", user("name"), nil),
		specWithUserRef("2", user("name", "email"), nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !breakingContaining(diff, "New required field: email") {
		t.Fatalf("expected new required field through $ref, got %#v", diff.Changes)
	}
	if breakingContaining(diff, "New required field: name") {
		t.Fatalf("name was already required: %#v", diff.Changes)
	}
}

func TestCompareResolvesRefForTypeChange(t *testing.T) {
	user := func(ageType string) map[string]interface{} {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"age": map[string]interface{}{"type": ageType},
			},
		}
	}
	diff, _ := NewDiffer().Compare(specWithUserRef("1", user("integer"), nil), specWithUserRef("2", user("string"), nil))
	if !breakingContaining(diff, "Field 'age' type changed from integer to string") {
		t.Fatalf("expected type change through $ref, got %#v", diff.Changes)
	}
}

// A property that is itself a (nullable) $ref resolves to the component's
// type, so moving an inline object into components is not a type change,
// while pointing the field at a component of another type is.
func TestCompareResolvesPropertyRefs(t *testing.T) {
	inline := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"address": map[string]interface{}{"type": []interface{}{"object", "null"}},
		},
	}
	viaRef := func(target string) map[string]interface{} {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"address": map[string]interface{}{
					"oneOf": []interface{}{
						map[string]interface{}{"$ref": "#/components/schemas/" + target},
						map[string]interface{}{"type": "null"},
					},
				},
			},
		}
	}
	extra := map[string]interface{}{
		"Address": map[string]interface{}{"type": "object"},
		"Code":    map[string]interface{}{"type": "string"},
	}

	diff, _ := NewDiffer().Compare(specWithUserRef("1", inline, nil), specWithUserRef("2", viaRef("Address"), extra))
	if diff.HasBreakingChanges() {
		t.Fatalf("inline object -> nullable $ref to object must not be breaking, got %#v", diff.Changes)
	}
	diff, _ = NewDiffer().Compare(specWithUserRef("1", viaRef("Address"), extra), specWithUserRef("2", viaRef("Code"), extra))
	if !breakingContaining(diff, "Field 'address' type changed from object to string") {
		t.Fatalf("expected type change object -> string via $ref, got %#v", diff.Changes)
	}
}

func TestResolvedTypeSignature(t *testing.T) {
	spec := map[string]interface{}{
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"ID":    map[string]interface{}{"type": "integer", "format": "int64"},
				"Alias": map[string]interface{}{"$ref": "#/components/schemas/ID"},
			},
		},
	}
	cases := []struct {
		schema map[string]interface{}
		want   string
	}{
		{map[string]interface{}{"$ref": "#/components/schemas/ID"}, "integer/int64"},
		{map[string]interface{}{"$ref": "#/components/schemas/Alias"}, "integer/int64"},
		{map[string]interface{}{"$ref": "#/components/schemas/Missing"}, "$ref:#/components/schemas/Missing"},
		{map[string]interface{}{"$ref": "other.yaml#/X"}, "$ref:other.yaml#/X"},
		{map[string]interface{}{"oneOf": []interface{}{
			map[string]interface{}{"$ref": "#/components/schemas/ID"},
			map[string]interface{}{"type": "null"},
		}}, "integer/int64"},
		{map[string]interface{}{"type": "string"}, "string"},
	}
	for _, c := range cases {
		if got := resolvedTypeSignature(spec, c.schema); got != c.want {
			t.Errorf("resolvedTypeSignature(%v) = %q, want %q", c.schema, got, c.want)
		}
	}
}
