package snippets

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PythonGenerator generates Python requests code snippets
type PythonGenerator struct{}

// NewPythonGenerator creates a new Python generator
func NewPythonGenerator() *PythonGenerator {
	return &PythonGenerator{}
}

// Generate creates a Python snippet for the given request
func (g *PythonGenerator) Generate(req Request) string {
	var lines []string

	lines = append(lines, "import requests")
	lines = append(lines, "")

	url := req.URL
	if len(req.QueryParams) > 0 {
		url += "?" + buildQueryString(req.QueryParams)
	}

	lines = append(lines, fmt.Sprintf("url = '%s'", url))

	// Headers
	if len(req.Headers) > 0 {
		lines = append(lines, "headers = {")
		headerLines := make([]string, 0, len(req.Headers))
		for key, value := range req.Headers {
			headerLines = append(headerLines, fmt.Sprintf("    '%s': '%s'", key, value))
		}
		lines = append(lines, strings.Join(headerLines, ",\n"))
		lines = append(lines, "}")
	}

	// Body
	if req.Body != "" {
		var bodyObj interface{}
		if err := json.Unmarshal([]byte(req.Body), &bodyObj); err == nil {
			prettyBody, _ := json.MarshalIndent(bodyObj, "", "    ")
			bodyStr := strings.ReplaceAll(string(prettyBody), "\"", "'")
			bodyStr = strings.ReplaceAll(bodyStr, "null", "None")
			bodyStr = strings.ReplaceAll(bodyStr, "true", "True")
			bodyStr = strings.ReplaceAll(bodyStr, "false", "False")
			lines = append(lines, fmt.Sprintf("data = %s", bodyStr))
		} else {
			lines = append(lines, fmt.Sprintf("data = '%s'", strings.ReplaceAll(req.Body, "'", "\\'")))
		}
	}

	lines = append(lines, "")

	// Request call
	method := strings.ToLower(req.Method)
	args := []string{"url"}

	if len(req.Headers) > 0 {
		args = append(args, "headers=headers")
	}
	if req.Body != "" {
		args = append(args, "json=data")
	}

	lines = append(lines, fmt.Sprintf("response = requests.%s(%s)", method, strings.Join(args, ", ")))
	lines = append(lines, "")
	lines = append(lines, "print(response.status_code)")
	lines = append(lines, "print(response.json())")

	return strings.Join(lines, "\n")
}

// Language returns the language identifier
func (g *PythonGenerator) Language() string {
	return "python"
}

// DisplayName returns the display name
func (g *PythonGenerator) DisplayName() string {
	return "Python"
}
