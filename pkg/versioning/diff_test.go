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
