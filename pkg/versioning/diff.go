package versioning

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeRemoved  ChangeType = "removed"
	ChangeModified ChangeType = "modified"
)

// Change represents a single change between specs
type Change struct {
	Type        ChangeType `json:"type"`
	Path        string     `json:"path"`
	Method      string     `json:"method,omitempty"`
	Description string     `json:"description"`
	IsBreaking  bool       `json:"isBreaking"`
}

// BreakingChange represents a breaking change with migration info
type BreakingChange struct {
	Path      string `json:"path"`
	Method    string `json:"method"`
	Reason    string `json:"reason"`
	Migration string `json:"migration"`
}

// Summary of changes between specs
type Summary struct {
	AddedEndpoints    int `json:"addedEndpoints"`
	RemovedEndpoints  int `json:"removedEndpoints"`
	ModifiedEndpoints int `json:"modifiedEndpoints"`
	BreakingChanges   int `json:"breakingChanges"`
}

// Diff represents differences between two OpenAPI specs
type Diff struct {
	OldVersion string           `json:"oldVersion"`
	NewVersion string           `json:"newVersion"`
	Changes    []Change         `json:"changes"`
	Breaking   []BreakingChange `json:"breaking"`
	Summary    Summary          `json:"summary"`
}

// Differ compares OpenAPI specs
type Differ struct{}

// NewDiffer creates a new spec differ
func NewDiffer() *Differ {
	return &Differ{}
}

// CompareFiles compares two spec files
func (d *Differ) CompareFiles(oldPath, newPath string) (*Diff, error) {
	oldSpec, err := loadSpec(oldPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load old spec: %w", err)
	}

	newSpec, err := loadSpec(newPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load new spec: %w", err)
	}

	return d.Compare(oldSpec, newSpec)
}

// Compare compares two parsed specs
func (d *Differ) Compare(oldSpec, newSpec map[string]interface{}) (*Diff, error) {
	diff := &Diff{
		OldVersion: getVersion(oldSpec),
		NewVersion: getVersion(newSpec),
		Changes:    []Change{},
		Breaking:   []BreakingChange{},
	}

	oldPaths := getPaths(oldSpec)
	newPaths := getPaths(newSpec)

	// Find added endpoints
	for path, methods := range newPaths {
		oldMethods, pathExists := oldPaths[path]
		for method := range methods {
			if !pathExists || oldMethods[method] == nil {
				diff.Changes = append(diff.Changes, Change{
					Type:        ChangeAdded,
					Path:        path,
					Method:      method,
					Description: fmt.Sprintf("New endpoint: %s %s", method, path),
					IsBreaking:  false,
				})
				diff.Summary.AddedEndpoints++
			}
		}
	}

	// Find removed endpoints (breaking!)
	for path, methods := range oldPaths {
		newMethods, pathExists := newPaths[path]
		for method := range methods {
			if !pathExists || newMethods[method] == nil {
				diff.Changes = append(diff.Changes, Change{
					Type:        ChangeRemoved,
					Path:        path,
					Method:      method,
					Description: fmt.Sprintf("Removed endpoint: %s %s", method, path),
					IsBreaking:  true,
				})
				diff.Breaking = append(diff.Breaking, BreakingChange{
					Path:      path,
					Method:    method,
					Reason:    "Endpoint removed",
					Migration: "Update client code to use alternative endpoint or remove usage",
				})
				diff.Summary.RemovedEndpoints++
				diff.Summary.BreakingChanges++
			}
		}
	}

	// Find modified endpoints
	for path, oldMethods := range oldPaths {
		if newMethods, exists := newPaths[path]; exists {
			for method, oldOp := range oldMethods {
				if newOp, methodExists := newMethods[method]; methodExists {
					changes := d.compareOperations(oldSpec, newSpec, path, method, oldOp, newOp)
					diff.Changes = append(diff.Changes, changes...)

					for _, change := range changes {
						if change.IsBreaking {
							diff.Summary.BreakingChanges++
							diff.Breaking = append(diff.Breaking, BreakingChange{
								Path:      path,
								Method:    method,
								Reason:    change.Description,
								Migration: getMigrationGuide(change),
							})
						}
					}

					if len(changes) > 0 {
						diff.Summary.ModifiedEndpoints++
					}
				}
			}
		}
	}

	return diff, nil
}

// compareOperations compares one operation present in both specs. The full
// specs are passed so that $refs into components.schemas can be resolved.
func (d *Differ) compareOperations(oldSpec, newSpec map[string]interface{}, path, method string, oldOp, newOp map[string]interface{}) []Change {
	changes := []Change{}

	// Compare request body
	oldBody := getRequestBody(oldOp)
	newBody := getRequestBody(newOp)

	if oldBody != nil && newBody == nil {
		changes = append(changes, Change{
			Type:        ChangeModified,
			Path:        path,
			Method:      method,
			Description: "Request body removed",
			IsBreaking:  true,
		})
	} else if oldBody == nil && newBody != nil {
		// Adding required body is breaking
		if isRequired, ok := newBody["required"].(bool); ok && isRequired {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        path,
				Method:      method,
				Description: "Required request body added",
				IsBreaking:  true,
			})
		}
	}

	// Compare required fields in request body
	oldRequired := getRequiredFields(oldSpec, oldOp)
	newRequired := getRequiredFields(newSpec, newOp)

	for _, field := range newRequired {
		if !contains(oldRequired, field) {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("New required field: %s", field),
				IsBreaking:  true,
			})
		}
	}

	// Compare field types in the request body — changing the JSON type or
	// format of an existing field breaks client deserialization. Tagged with
	// BreakingTypeChanged in the breaking-rule catalog.
	oldFields := getBodyFieldTypes(oldSpec, oldOp)
	newFields := getBodyFieldTypes(newSpec, newOp)
	for name, oldType := range oldFields {
		if newType, exists := newFields[name]; exists && newType != oldType {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("Field '%s' type changed from %s to %s", name, oldType, newType),
				IsBreaking:  true,
			})
		}
	}

	// Compare parameter types — same rationale as body fields, but for
	// path/query/header/cookie params.
	oldParamTypes := getParamTypes(oldSpec, oldOp)
	newParamTypes := getParamTypes(newSpec, newOp)
	for name, oldType := range oldParamTypes {
		if newType, exists := newParamTypes[name]; exists && newType != oldType {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("Parameter '%s' type changed from %s to %s", name, oldType, newType),
				IsBreaking:  true,
			})
		}
	}

	// Compare response codes
	oldResponses := getResponseCodes(oldOp)
	newResponses := getResponseCodes(newOp)

	for _, code := range oldResponses {
		if !contains(newResponses, code) {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("Response code %s removed", code),
				IsBreaking:  true,
			})
		}
	}

	// Compare parameters
	oldParams := getParameters(oldOp)
	newParams := getParameters(newOp)

	// Check for removed parameters
	for name := range oldParams {
		if _, exists := newParams[name]; !exists {
			changes = append(changes, Change{
				Type:        ChangeModified,
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("Parameter '%s' removed", name),
				IsBreaking:  true,
			})
		}
	}

	// Check for new required parameters
	for name, param := range newParams {
		if _, exists := oldParams[name]; !exists {
			if isParamRequired(param) {
				changes = append(changes, Change{
					Type:        ChangeModified,
					Path:        path,
					Method:      method,
					Description: fmt.Sprintf("New required parameter: %s", name),
					IsBreaking:  true,
				})
			}
		}
	}

	return changes
}

// HasBreakingChanges returns true if there are any breaking changes
func (d *Diff) HasBreakingChanges() bool {
	return d.Summary.BreakingChanges > 0
}

// Helper functions
func loadSpec(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func getVersion(spec map[string]interface{}) string {
	if info, ok := spec["info"].(map[string]interface{}); ok {
		if version, ok := info["version"].(string); ok {
			return version
		}
	}
	return "unknown"
}

func getPaths(spec map[string]interface{}) map[string]map[string]map[string]interface{} {
	result := make(map[string]map[string]map[string]interface{})

	if paths, ok := spec["paths"].(map[string]interface{}); ok {
		for path, methods := range paths {
			result[path] = make(map[string]map[string]interface{})
			if methodMap, ok := methods.(map[string]interface{}); ok {
				for method, op := range methodMap {
					if opMap, ok := op.(map[string]interface{}); ok {
						result[path][method] = opMap
					}
				}
			}
		}
	}

	return result
}

func getRequestBody(op map[string]interface{}) map[string]interface{} {
	if body, ok := op["requestBody"].(map[string]interface{}); ok {
		return body
	}
	return nil
}

// getRequiredFields returns the required property names of the request-body
// schema, resolving a top-level $ref against spec's components.schemas.
func getRequiredFields(spec, op map[string]interface{}) []string {
	var required []string

	body := getRequestBody(op)
	if body == nil {
		return required
	}

	content, ok := body["content"].(map[string]interface{})
	if !ok {
		return required
	}

	for _, mediaType := range content {
		if mt, ok := mediaType.(map[string]interface{}); ok {
			if schema, ok := mt["schema"].(map[string]interface{}); ok {
				schema = resolveSchema(spec, schema)
				if req, ok := schema["required"].([]interface{}); ok {
					for _, r := range req {
						if s, ok := r.(string); ok {
							required = append(required, s)
						}
					}
				}
			}
		}
	}

	return required
}

func getResponseCodes(op map[string]interface{}) []string {
	var codes []string

	if responses, ok := op["responses"].(map[string]interface{}); ok {
		for code := range responses {
			codes = append(codes, code)
		}
	}

	return codes
}

func getParameters(op map[string]interface{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	if params, ok := op["parameters"].([]interface{}); ok {
		for _, p := range params {
			if param, ok := p.(map[string]interface{}); ok {
				if name, ok := param["name"].(string); ok {
					result[name] = param
				}
			}
		}
	}

	return result
}

// getBodyFieldTypes returns a map of request-body property name → "type[/format]".
// Format is included because changing int32→int64 (same JSON type, different
// format) still breaks strongly-typed clients. A top-level $ref, and $refs on
// the properties themselves, are resolved against spec's components.schemas.
func getBodyFieldTypes(spec, op map[string]interface{}) map[string]string {
	result := make(map[string]string)

	body := getRequestBody(op)
	if body == nil {
		return result
	}
	content, ok := body["content"].(map[string]interface{})
	if !ok {
		return result
	}
	for _, mediaType := range content {
		mt, ok := mediaType.(map[string]interface{})
		if !ok {
			continue
		}
		schema, ok := mt["schema"].(map[string]interface{})
		if !ok {
			continue
		}
		schema = resolveSchema(spec, schema)
		props, ok := schema["properties"].(map[string]interface{})
		if !ok {
			continue
		}
		for name, p := range props {
			if pm, ok := p.(map[string]interface{}); ok {
				result[name] = resolvedTypeSignature(spec, pm)
			}
		}
		// First media type is enough; multiple media types with diverging
		// schemas would require per-content-type tracking which is out of
		// scope here.
		break
	}
	return result
}

// getParamTypes returns a map of parameter name → "type[/format]".
func getParamTypes(spec, op map[string]interface{}) map[string]string {
	result := make(map[string]string)
	params, ok := op["parameters"].([]interface{})
	if !ok {
		return result
	}
	for _, p := range params {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := pm["name"].(string)
		if name == "" {
			continue
		}
		if schema, ok := pm["schema"].(map[string]interface{}); ok {
			result[name] = resolvedTypeSignature(spec, schema)
		}
	}
	return result
}

// schemaTypeSignature returns "type[/format]" for a schema. OpenAPI 3.1
// may express a nullable field as a type array such as ["string", "null"];
// the "null" member is ignored so that merely allowing null (a loosening of
// the request contract) is not reported as a type change.
func schemaTypeSignature(schema map[string]interface{}) string {
	var t string
	switch v := schema["type"].(type) {
	case string:
		t = v
	case []interface{}:
		var types []string
		for _, item := range v {
			if s, ok := item.(string); ok && s != "null" {
				types = append(types, s)
			}
		}
		sort.Strings(types)
		t = strings.Join(types, "|")
	}
	f, _ := schema["format"].(string)
	if f != "" {
		return t + "/" + f
	}
	return t
}

// componentRefPrefix is the only $ref form the differ resolves: a local
// pointer to a schema component.
const componentRefPrefix = "#/components/schemas/"

// maxRefDepth bounds how many chained $refs resolveSchema follows, so a
// malformed spec with a ref cycle cannot loop forever.
const maxRefDepth = 8

// resolveSchema follows "#/components/schemas/<Name>" $refs in spec until it
// reaches a schema without one. A $ref that cannot be resolved (external,
// unknown name, too deep) is returned as is.
func resolveSchema(spec, schema map[string]interface{}) map[string]interface{} {
	for i := 0; i < maxRefDepth; i++ {
		ref, ok := schema["$ref"].(string)
		if !ok {
			return schema
		}
		target := componentSchema(spec, ref)
		if target == nil {
			return schema
		}
		schema = target
	}
	return schema
}

// componentSchema returns the components.schemas entry a local $ref points
// to, or nil.
func componentSchema(spec map[string]interface{}, ref string) map[string]interface{} {
	name, ok := strings.CutPrefix(ref, componentRefPrefix)
	if !ok || name == "" {
		return nil
	}
	components, _ := spec["components"].(map[string]interface{})
	schemas, _ := components["schemas"].(map[string]interface{})
	target, _ := schemas[name].(map[string]interface{})
	return target
}

// resolvedTypeSignature is schemaTypeSignature for a schema that may be a
// $ref or a nullable $ref ({"oneOf": [{"$ref": ...}, {"type": "null"}]}).
// The $ref is resolved so that moving a struct into components, or making
// it nullable, is not reported as a type change: a nullable ref to an object
// component and an inline ["object", "null"] both give "object". A $ref
// that cannot be resolved yields "$ref:<ref>".
func resolvedTypeSignature(spec, schema map[string]interface{}) string {
	if branch := nonNullBranch(schema); branch != nil {
		schema = branch
	}
	schema = resolveSchema(spec, schema)
	if ref, ok := schema["$ref"].(string); ok {
		return "$ref:" + ref
	}
	return schemaTypeSignature(schema)
}

// nonNullBranch returns X for an untyped {"oneOf"|"anyOf": [X, {"type":
// "null"}]} union (the OpenAPI 3.1 form of a nullable $ref), or nil.
func nonNullBranch(schema map[string]interface{}) map[string]interface{} {
	if _, typed := schema["type"]; typed {
		return nil
	}
	for _, key := range []string{"oneOf", "anyOf"} {
		branches, ok := schema[key].([]interface{})
		if !ok {
			continue
		}
		var found map[string]interface{}
		for _, b := range branches {
			bm, ok := b.(map[string]interface{})
			if !ok {
				return nil
			}
			if t, _ := bm["type"].(string); t == "null" && len(bm) == 1 {
				continue
			}
			if found != nil {
				return nil // more than one non-null branch
			}
			found = bm
		}
		return found
	}
	return nil
}

func isParamRequired(param map[string]interface{}) bool {
	if required, ok := param["required"].(bool); ok {
		return required
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getMigrationGuide(change Change) string {
	desc := change.Description
	switch {
	case desc == "Request body removed":
		return "Remove request body from client calls"
	case desc == "Required request body added":
		return "Add required request body to client calls"
	case strings.HasPrefix(desc, "New required field"):
		return "Add the new required field to request payload"
	case strings.HasPrefix(desc, "Response code"):
		return "Update client to handle the removed response code"
	case strings.HasPrefix(desc, "New required parameter"):
		return "Add the new required parameter to client calls"
	case strings.HasPrefix(desc, "Parameter"):
		// Either "Parameter 'x' removed" or "Parameter 'x' type changed from..."
		if strings.Contains(desc, "type changed") {
			return "Update client field type to match the new schema (regenerate types if you use code generation)"
		}
		return "Update client to remove usage of the deleted parameter"
	case strings.HasPrefix(desc, "Field"):
		return "Update client field type to match the new schema (regenerate types if you use code generation)"
	default:
		return "Review the change and update client code accordingly"
	}
}
