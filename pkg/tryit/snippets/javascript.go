package snippets

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JavaScriptGenerator generates JavaScript fetch code snippets
type JavaScriptGenerator struct{}

// NewJavaScriptGenerator creates a new JavaScript generator
func NewJavaScriptGenerator() *JavaScriptGenerator {
	return &JavaScriptGenerator{}
}

// Generate creates a JavaScript fetch snippet for the given request
func (g *JavaScriptGenerator) Generate(req Request) string {
	var lines []string

	url := req.URL
	if len(req.QueryParams) > 0 {
		url += "?" + buildQueryString(req.QueryParams)
	}

	lines = append(lines, fmt.Sprintf("const response = await fetch('%s', {", url))
	lines = append(lines, fmt.Sprintf("  method: '%s',", req.Method))

	// Headers
	if len(req.Headers) > 0 {
		lines = append(lines, "  headers: {")
		headerLines := make([]string, 0, len(req.Headers))
		for key, value := range req.Headers {
			headerLines = append(headerLines, fmt.Sprintf("    '%s': '%s'", key, value))
		}
		lines = append(lines, strings.Join(headerLines, ",\n"))
		lines = append(lines, "  },")
	}

	// Body
	if req.Body != "" {
		var bodyObj interface{}
		if err := json.Unmarshal([]byte(req.Body), &bodyObj); err == nil {
			prettyBody, _ := json.MarshalIndent(bodyObj, "  ", "  ")
			lines = append(lines, fmt.Sprintf("  body: JSON.stringify(%s)", string(prettyBody)))
		} else {
			lines = append(lines, fmt.Sprintf("  body: '%s'", strings.ReplaceAll(req.Body, "'", "\\'")))
		}
	}

	lines = append(lines, "});")
	lines = append(lines, "")
	lines = append(lines, "const data = await response.json();")
	lines = append(lines, "console.log(data);")

	return strings.Join(lines, "\n")
}

// Language returns the language identifier
func (g *JavaScriptGenerator) Language() string {
	return "javascript"
}

// DisplayName returns the display name
func (g *JavaScriptGenerator) DisplayName() string {
	return "JavaScript"
}
