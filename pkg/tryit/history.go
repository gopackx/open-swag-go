package tryit

import (
	"encoding/json"
	"time"
)

// HistoryConfig configures request history
type HistoryConfig struct {
	Enabled    bool   `json:"enabled"`
	MaxEntries int    `json:"maxEntries"`
	Storage    string `json:"storage"`
	StorageKey string `json:"storageKey"`
}

// HistoryEntry represents a single request history entry
type HistoryEntry struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	StatusCode  int               `json:"statusCode"`
	Response    string            `json:"response,omitempty"`
	Duration    int64             `json:"duration"`
	OperationID string            `json:"operationId,omitempty"`
}

// History manages request history
type History struct {
	config  HistoryConfig
	entries []HistoryEntry
}

// DefaultHistoryConfig returns the default history configuration
func DefaultHistoryConfig() HistoryConfig {
	return HistoryConfig{
		Enabled:    true,
		MaxEntries: 50,
		Storage:    "localStorage",
		StorageKey: "openswag_history",
	}
}

// NewHistory creates a new history manager
func NewHistory(config HistoryConfig) *History {
	return &History{
		config:  config,
		entries: make([]HistoryEntry, 0),
	}
}

// Add adds an entry to the history
func (h *History) Add(entry HistoryEntry) {
	if !h.config.Enabled {
		return
	}

	if entry.ID == "" {
		entry.ID = generateID()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	h.entries = append([]HistoryEntry{entry}, h.entries...)

	if len(h.entries) > h.config.MaxEntries {
		h.entries = h.entries[:h.config.MaxEntries]
	}
}

// Get returns all history entries
func (h *History) Get() []HistoryEntry {
	return h.entries
}

// GetByID returns a specific history entry
func (h *History) GetByID(id string) (HistoryEntry, bool) {
	for _, entry := range h.entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return HistoryEntry{}, false
}

// Clear removes all history entries
func (h *History) Clear() {
	h.entries = make([]HistoryEntry, 0)
}

// Delete removes a specific history entry
func (h *History) Delete(id string) bool {
	for i, entry := range h.entries {
		if entry.ID == id {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			return true
		}
	}
	return false
}

// ToJSON serializes history for client-side storage
func (h *History) ToJSON() (string, error) {
	data, err := json.Marshal(h.entries)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON deserializes history from client-side storage
func (h *History) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), &h.entries)
}

// GetConfig returns the history configuration
func (h *History) GetConfig() HistoryConfig {
	return h.config
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
