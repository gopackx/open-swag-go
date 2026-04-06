package snippets

import (
	"fmt"
	"strings"
)

// CurlGenerator generates curl code snippets
type CurlGenerator struct{}

// NewCurlGenerator creates a new curl generator
func NewCurlGenerator() *CurlGenerator {
	return &CurlGenerator{}
}

// Generate creates a curl command for the given request
func (g *CurlGenerator) Generate(req Request) string {
	var parts []string

	parts = append(parts, "curl")

	// Method
	if req.Method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", req.Method))
	}

	// URL
	url := req.URL
	if len(req.QueryParams) > 0 {
		url += "?" + buildQueryString(req.QueryParams)
	}
	parts = append(parts, fmt.Sprintf("'%s'", url))

	// Headers
	for key, value := range req.Headers {
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", key, value))
	}

	// Body
	if req.Body != "" {
		escapedBody := strings.ReplaceAll(req.Body, "'", "'\\''")
		parts = append(parts, fmt.Sprintf("-d '%s'", escapedBody))
	}

	return strings.Join(parts, " \\\n  ")
}

// Language returns the language identifier
func (g *CurlGenerator) Language() string {
	return "curl"
}

// DisplayName returns the display name
func (g *CurlGenerator) DisplayName() string {
	return "cURL"
}
