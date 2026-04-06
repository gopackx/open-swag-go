package versioning

import (
	"encoding/json"
	"fmt"
	"os"
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
					changes := d.compareOperations(path, method, oldOp, newOp)
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

func (d *Differ) compareOperations(path, method string, oldOp, newOp map[string]interface{}) []Change {
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
	oldRequired := getRequiredFields(oldOp)
	newRequired := getRequiredFields(newOp)

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

func getRequiredFields(op map[string]interface{}) []string {
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
	switch {
	case change.Description == "Request body removed":
		return "Remove request body from client calls"
	case change.Description == "Required request body added":
		return "Add required request body to client calls"
	case contains([]string{"New required field"}, change.Description[:18]):
		return "Add the new required field to request payload"
	case contains([]string{"Response code"}, change.Description[:13]):
		return "Update client to handle the removed response code"
	case contains([]string{"Parameter"}, change.Description[:9]):
		return "Update client to remove usage of the deleted parameter"
	case contains([]string{"New required parameter"}, change.Description[:21]):
		return "Add the new required parameter to client calls"
	default:
		return "Review the change and update client code accordingly"
	}
}
