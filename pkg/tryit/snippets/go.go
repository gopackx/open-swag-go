package snippets

import (
	"fmt"
	"strings"
)

// GoGenerator generates Go code snippets
type GoGenerator struct{}

// NewGoGenerator creates a new Go generator
func NewGoGenerator() *GoGenerator {
	return &GoGenerator{}
}

// Generate creates a Go snippet for the given request
func (g *GoGenerator) Generate(req Request) string {
	var lines []string

	lines = append(lines, "package main")
	lines = append(lines, "")
	lines = append(lines, "import (")
	lines = append(lines, `	"fmt"`)
	lines = append(lines, `	"io"`)
	lines = append(lines, `	"net/http"`)
	if req.Body != "" {
		lines = append(lines, `	"strings"`)
	}
	lines = append(lines, ")")
	lines = append(lines, "")
	lines = append(lines, "func main() {")

	url := req.URL
	if len(req.QueryParams) > 0 {
		url += "?" + buildQueryString(req.QueryParams)
	}

	// Body
	if req.Body != "" {
		escapedBody := strings.ReplaceAll(req.Body, "`", "` + \"`\" + `")
		lines = append(lines, fmt.Sprintf("\tbody := strings.NewReader(`%s`)", escapedBody))
		lines = append(lines, fmt.Sprintf("\treq, err := http.NewRequest(\"%s\", \"%s\", body)", req.Method, url))
	} else {
		lines = append(lines, fmt.Sprintf("\treq, err := http.NewRequest(\"%s\", \"%s\", nil)", req.Method, url))
	}

	lines = append(lines, "\tif err != nil {")
	lines = append(lines, "\t\tpanic(err)")
	lines = append(lines, "\t}")
	lines = append(lines, "")

	// Headers
	for key, value := range req.Headers {
		lines = append(lines, fmt.Sprintf("\treq.Header.Set(\"%s\", \"%s\")", key, value))
	}

	if len(req.Headers) > 0 {
		lines = append(lines, "")
	}

	lines = append(lines, "\tclient := &http.Client{}")
	lines = append(lines, "\tresp, err := client.Do(req)")
	lines = append(lines, "\tif err != nil {")
	lines = append(lines, "\t\tpanic(err)")
	lines = append(lines, "\t}")
	lines = append(lines, "\tdefer resp.Body.Close()")
	lines = append(lines, "")
	lines = append(lines, "\tdata, err := io.ReadAll(resp.Body)")
	lines = append(lines, "\tif err != nil {")
	lines = append(lines, "\t\tpanic(err)")
	lines = append(lines, "\t}")
	lines = append(lines, "")
	lines = append(lines, "\tfmt.Println(string(data))")
	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

// Language returns the language identifier
func (g *GoGenerator) Language() string {
	return "go"
}

// DisplayName returns the display name
func (g *GoGenerator) DisplayName() string {
	return "Go"
}
