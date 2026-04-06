package tryit

// ConsoleConfig configures the Try-It console
type ConsoleConfig struct {
	Enabled          bool              `json:"enabled"`
	DefaultServer    string            `json:"defaultServer,omitempty"`
	RequestTimeout   int               `json:"requestTimeout"`
	ShowCodeSnippets bool              `json:"showCodeSnippets"`
	EnabledLanguages []string          `json:"enabledLanguages"`
	CustomHeaders    map[string]string `json:"customHeaders,omitempty"`
	ProxyURL         string            `json:"proxyUrl,omitempty"`
	CORSProxy        bool              `json:"corsProxy"`
}

// ConsoleOption is a functional option for ConsoleConfig
type ConsoleOption func(*ConsoleConfig)

// DefaultConsoleConfig returns the default console configuration
func DefaultConsoleConfig() ConsoleConfig {
	return ConsoleConfig{
		Enabled:          true,
		RequestTimeout:   30000,
		ShowCodeSnippets: true,
		EnabledLanguages: []string{"curl", "javascript", "go", "python", "php"},
		CORSProxy:        false,
	}
}

// NewConsole creates a new console configuration
func NewConsole(opts ...ConsoleOption) *ConsoleConfig {
	cfg := DefaultConsoleConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	return &cfg
}

// WithTimeout sets the request timeout in milliseconds
func WithTimeout(ms int) ConsoleOption {
	return func(cfg *ConsoleConfig) {
		cfg.RequestTimeout = ms
	}
}

// WithDefaultServer sets the default server URL
func WithDefaultServer(url string) ConsoleOption {
	return func(cfg *ConsoleConfig) {
		cfg.DefaultServer = url
	}
}

// WithLanguages sets the enabled code snippet languages
func WithLanguages(languages ...string) ConsoleOption {
	return func(cfg *ConsoleConfig) {
		cfg.EnabledLanguages = languages
	}
}

// WithCustomHeader adds a custom header to all requests
func WithCustomHeader(key, value string) ConsoleOption {
	return func(cfg *ConsoleConfig) {
		if cfg.CustomHeaders == nil {
			cfg.CustomHeaders = make(map[string]string)
		}
		cfg.CustomHeaders[key] = value
	}
}

// WithProxy sets a proxy URL for requests
func WithProxy(url string) ConsoleOption {
	return func(cfg *ConsoleConfig) {
		cfg.ProxyURL = url
	}
}

// WithCORSProxy enables CORS proxy for browser requests
func WithCORSProxy(enabled bool) ConsoleOption {
	return func(cfg *ConsoleConfig) {
		cfg.CORSProxy = enabled
	}
}

// DisableSnippets disables code snippet generation
func DisableSnippets() ConsoleOption {
	return func(cfg *ConsoleConfig) {
		cfg.ShowCodeSnippets = false
	}
}

// Disable disables the Try-It console
func Disable() ConsoleOption {
	return func(cfg *ConsoleConfig) {
		cfg.Enabled = false
	}
}
