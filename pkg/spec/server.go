package spec

// Server represents an OpenAPI server object
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents a server variable
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// NewServer creates a new server
func NewServer(url string) Server {
	return Server{URL: url}
}

// WithDescription sets the server description
func (s Server) WithDescription(desc string) Server {
	s.Description = desc
	return s
}

// WithVariable adds a variable to the server
func (s Server) WithVariable(name string, variable ServerVariable) Server {
	if s.Variables == nil {
		s.Variables = make(map[string]ServerVariable)
	}
	s.Variables[name] = variable
	return s
}

// NewServerVariable creates a new server variable
func NewServerVariable(defaultValue string) ServerVariable {
	return ServerVariable{Default: defaultValue}
}

// WithEnum sets the enum values for a server variable
func (v ServerVariable) WithEnum(values ...string) ServerVariable {
	v.Enum = values
	return v
}

// WithDescription sets the description for a server variable
func (v ServerVariable) WithDescription(desc string) ServerVariable {
	v.Description = desc
	return v
}

// CommonServers provides common server configurations
var CommonServers = struct {
	Localhost  func(port int) Server
	Production func(url string) Server
	Staging    func(url string) Server
}{
	Localhost: func(port int) Server {
		return Server{
			URL:         "http://localhost:" + string(rune(port)),
			Description: "Local development server",
		}
	},
	Production: func(url string) Server {
		return Server{
			URL:         url,
			Description: "Production server",
		}
	},
	Staging: func(url string) Server {
		return Server{
			URL:         url,
			Description: "Staging server",
		}
	},
}

// LocalhostServer creates a localhost server with the given port
func LocalhostServer(port int) Server {
	return Server{
		URL:         "http://localhost:" + portToString(port),
		Description: "Local development server",
	}
}

// ProductionServer creates a production server
func ProductionServer(url, description string) Server {
	if description == "" {
		description = "Production server"
	}
	return Server{
		URL:         url,
		Description: description,
	}
}

// StagingServer creates a staging server
func StagingServer(url string) Server {
	return Server{
		URL:         url,
		Description: "Staging server",
	}
}

func portToString(port int) string {
	result := ""
	if port == 0 {
		return "0"
	}
	for port > 0 {
		result = string(rune('0'+port%10)) + result
		port /= 10
	}
	return result
}
