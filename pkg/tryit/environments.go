package tryit

import (
	"encoding/json"
	"strings"
)

// Environment represents a set of variables for API testing
type Environment struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	IsActive  bool              `json:"isActive"`
}

// EnvironmentConfig configures environment management
type EnvironmentConfig struct {
	Enabled    bool   `json:"enabled"`
	Storage    string `json:"storage"`
	StorageKey string `json:"storageKey"`
}

// EnvironmentManager manages multiple environments
type EnvironmentManager struct {
	config       EnvironmentConfig
	environments []Environment
	active       string
}

// DefaultEnvironmentConfig returns the default environment configuration
func DefaultEnvironmentConfig() EnvironmentConfig {
	return EnvironmentConfig{
		Enabled:    true,
		Storage:    "localStorage",
		StorageKey: "openswag_environments",
	}
}

// NewEnvironmentManager creates a new environment manager
func NewEnvironmentManager(config EnvironmentConfig) *EnvironmentManager {
	return &EnvironmentManager{
		config:       config,
		environments: make([]Environment, 0),
	}
}

// Add adds a new environment
func (m *EnvironmentManager) Add(env Environment) {
	m.environments = append(m.environments, env)
}

// Get returns all environments
func (m *EnvironmentManager) Get() []Environment {
	return m.environments
}

// GetByName returns an environment by name
func (m *EnvironmentManager) GetByName(name string) (Environment, bool) {
	for _, env := range m.environments {
		if env.Name == name {
			return env, true
		}
	}
	return Environment{}, false
}

// SetActive sets the active environment
func (m *EnvironmentManager) SetActive(name string) bool {
	for i := range m.environments {
		m.environments[i].IsActive = m.environments[i].Name == name
		if m.environments[i].Name == name {
			m.active = name
		}
	}
	return m.active == name
}

// GetActive returns the active environment
func (m *EnvironmentManager) GetActive() (Environment, bool) {
	for _, env := range m.environments {
		if env.IsActive {
			return env, true
		}
	}
	return Environment{}, false
}

// Delete removes an environment
func (m *EnvironmentManager) Delete(name string) bool {
	for i, env := range m.environments {
		if env.Name == name {
			m.environments = append(m.environments[:i], m.environments[i+1:]...)
			return true
		}
	}
	return false
}

// Update updates an environment's variables
func (m *EnvironmentManager) Update(name string, variables map[string]string) bool {
	for i, env := range m.environments {
		if env.Name == name {
			m.environments[i].Variables = variables
			return true
		}
	}
	return false
}

// Interpolate replaces {{variable}} placeholders with environment values
func (m *EnvironmentManager) Interpolate(input string) string {
	env, ok := m.GetActive()
	if !ok {
		return input
	}

	result := input
	for key, value := range env.Variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ToJSON serializes environments for client-side storage
func (m *EnvironmentManager) ToJSON() (string, error) {
	data, err := json.Marshal(m.environments)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON deserializes environments from client-side storage
func (m *EnvironmentManager) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), &m.environments)
}

// GetConfig returns the environment configuration
func (m *EnvironmentManager) GetConfig() EnvironmentConfig {
	return m.config
}

// CreateDefaultEnvironments creates common default environments
func CreateDefaultEnvironments() []Environment {
	return []Environment{
		{
			Name: "Development",
			Variables: map[string]string{
				"baseUrl": "http://localhost:8080",
				"apiKey":  "dev-api-key",
			},
			IsActive: true,
		},
		{
			Name: "Staging",
			Variables: map[string]string{
				"baseUrl": "https://staging.example.com",
				"apiKey":  "staging-api-key",
			},
		},
		{
			Name: "Production",
			Variables: map[string]string{
				"baseUrl": "https://api.example.com",
				"apiKey":  "",
			},
		},
	}
}
