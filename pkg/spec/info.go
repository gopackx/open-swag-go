package spec

// Info represents the OpenAPI info object
type Info struct {
	Title          string   `json:"title"`
	Version        string   `json:"version"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Summary        string   `json:"summary,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License represents license information
type License struct {
	Name       string `json:"name"`
	URL        string `json:"url,omitempty"`
	Identifier string `json:"identifier,omitempty"`
}

// NewInfo creates a new Info object
func NewInfo(title, version string) Info {
	return Info{
		Title:   title,
		Version: version,
	}
}

// WithDescription sets the description
func (i Info) WithDescription(desc string) Info {
	i.Description = desc
	return i
}

// WithTermsOfService sets the terms of service URL
func (i Info) WithTermsOfService(url string) Info {
	i.TermsOfService = url
	return i
}

// WithContact sets the contact information
func (i Info) WithContact(name, url, email string) Info {
	i.Contact = &Contact{
		Name:  name,
		URL:   url,
		Email: email,
	}
	return i
}

// WithLicense sets the license information
func (i Info) WithLicense(name, url string) Info {
	i.License = &License{
		Name: name,
		URL:  url,
	}
	return i
}

// WithSummary sets the summary
func (i Info) WithSummary(summary string) Info {
	i.Summary = summary
	return i
}
