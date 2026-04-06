package auth

import (
	"encoding/json"
	"time"
)

// StorageType represents the type of credential storage
type StorageType string

const (
	StorageLocal   StorageType = "localStorage"
	StorageSession StorageType = "sessionStorage"
	StorageNone    StorageType = "none"
)

// PersistConfig configures credential persistence
type PersistConfig struct {
	Storage    StorageType   `json:"storage"`
	Key        string        `json:"key"`
	Expiration time.Duration `json:"expiration,omitempty"`
	Encrypt    bool          `json:"encrypt"`
}

// Credential represents a stored credential
type Credential struct {
	Scheme    string    `json:"scheme"`
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// CredentialStore manages credential persistence
type CredentialStore struct {
	config      PersistConfig
	credentials map[string]Credential
}

// NewCredentialStore creates a new credential store
func NewCredentialStore(config PersistConfig) *CredentialStore {
	if config.Key == "" {
		config.Key = "openswag_credentials"
	}
	if config.Storage == "" {
		config.Storage = StorageLocal
	}

	return &CredentialStore{
		config:      config,
		credentials: make(map[string]Credential),
	}
}

// Set stores a credential
func (s *CredentialStore) Set(scheme, value string) {
	cred := Credential{
		Scheme:    scheme,
		Value:     value,
		CreatedAt: time.Now(),
	}

	if s.config.Expiration > 0 {
		cred.ExpiresAt = time.Now().Add(s.config.Expiration)
	}

	s.credentials[scheme] = cred
}

// Get retrieves a credential
func (s *CredentialStore) Get(scheme string) (string, bool) {
	cred, exists := s.credentials[scheme]
	if !exists {
		return "", false
	}

	if !cred.ExpiresAt.IsZero() && time.Now().After(cred.ExpiresAt) {
		delete(s.credentials, scheme)
		return "", false
	}

	return cred.Value, true
}

// Delete removes a credential
func (s *CredentialStore) Delete(scheme string) {
	delete(s.credentials, scheme)
}

// Clear removes all credentials
func (s *CredentialStore) Clear() {
	s.credentials = make(map[string]Credential)
}

// ToJSON serializes credentials for client-side storage
func (s *CredentialStore) ToJSON() (string, error) {
	data, err := json.Marshal(s.credentials)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON deserializes credentials from client-side storage
func (s *CredentialStore) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), &s.credentials)
}

// GetConfig returns the persistence configuration for client-side use
func (s *CredentialStore) GetConfig() PersistConfig {
	return s.config
}
