package snippets

import (
	"net/url"
	"strings"
)

// Request represents an HTTP request for snippet generation
type Request struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	QueryParams map[string]string `json:"queryParams"`
}

// Generator is the interface for code snippet generators
type Generator interface {
	Generate(req Request) string
	Language() string
	DisplayName() string
}

// Manager manages multiple snippet generators
type Manager struct {
	generators map[string]Generator
}

// NewManager creates a new snippet manager with default generators
func NewManager() *Manager {
	m := &Manager{
		generators: make(map[string]Generator),
	}

	// Register default generators
	m.Register(NewCurlGenerator())
	m.Register(NewJavaScriptGenerator())
	m.Register(NewGoGenerator())
	m.Register(NewPythonGenerator())

	return m
}

// Register adds a generator to the manager
func (m *Manager) Register(gen Generator) {
	m.generators[gen.Language()] = gen
}

// Generate creates a snippet for the given language
func (m *Manager) Generate(language string, req Request) (string, bool) {
	gen, exists := m.generators[language]
	if !exists {
		return "", false
	}
	return gen.Generate(req), true
}

// GenerateAll creates snippets for all registered languages
func (m *Manager) GenerateAll(req Request) map[string]string {
	result := make(map[string]string)
	for lang, gen := range m.generators {
		result[lang] = gen.Generate(req)
	}
	return result
}

// Languages returns all registered language identifiers
func (m *Manager) Languages() []string {
	langs := make([]string, 0, len(m.generators))
	for lang := range m.generators {
		langs = append(langs, lang)
	}
	return langs
}

// GetGenerator returns a specific generator
func (m *Manager) GetGenerator(language string) (Generator, bool) {
	gen, exists := m.generators[language]
	return gen, exists
}

// buildQueryString builds a URL query string from parameters
func buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	values := url.Values{}
	for key, value := range params {
		values.Add(key, value)
	}
	return values.Encode()
}

// escapeString escapes special characters in a string
func escapeString(s string, quote rune) string {
	var result strings.Builder
	for _, c := range s {
		switch c {
		case quote:
			result.WriteRune('\\')
			result.WriteRune(c)
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			result.WriteRune(c)
		}
	}
	return result.String()
}
