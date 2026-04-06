package auth

// PlaygroundConfig configures the auth playground in the UI
type PlaygroundConfig struct {
	Enabled            bool              `json:"enabled"`
	DefaultScheme      string            `json:"defaultScheme,omitempty"`
	PersistCredentials bool              `json:"persistCredentials"`
	Schemes            map[string]Scheme `json:"schemes,omitempty"`
	PrefilledValues    map[string]string `json:"prefilledValues,omitempty"`
}

// PlaygroundOption is a functional option for PlaygroundConfig
type PlaygroundOption func(*PlaygroundConfig)

// NewPlayground creates a new auth playground configuration
func NewPlayground(opts ...PlaygroundOption) *PlaygroundConfig {
	cfg := &PlaygroundConfig{
		Enabled:            true,
		PersistCredentials: true,
		Schemes:            make(map[string]Scheme),
		PrefilledValues:    make(map[string]string),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithScheme adds a security scheme to the playground
func WithScheme(name string, scheme Scheme) PlaygroundOption {
	return func(cfg *PlaygroundConfig) {
		cfg.Schemes[name] = scheme
	}
}

// WithDefaultScheme sets the default scheme
func WithDefaultScheme(name string) PlaygroundOption {
	return func(cfg *PlaygroundConfig) {
		cfg.DefaultScheme = name
	}
}

// WithPersistence enables or disables credential persistence
func WithPersistence(enabled bool) PlaygroundOption {
	return func(cfg *PlaygroundConfig) {
		cfg.PersistCredentials = enabled
	}
}

// WithPrefilledValue sets a prefilled value for a scheme
func WithPrefilledValue(key, value string) PlaygroundOption {
	return func(cfg *PlaygroundConfig) {
		cfg.PrefilledValues[key] = value
	}
}

// Disable disables the auth playground
func Disable() PlaygroundOption {
	return func(cfg *PlaygroundConfig) {
		cfg.Enabled = false
	}
}
