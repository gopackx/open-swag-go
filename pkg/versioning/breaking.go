package versioning

// BreakingChangeType represents types of breaking changes
type BreakingChangeType string

const (
	BreakingEndpointRemoved    BreakingChangeType = "endpoint_removed"
	BreakingParameterRemoved   BreakingChangeType = "parameter_removed"
	BreakingRequiredAdded      BreakingChangeType = "required_field_added"
	BreakingResponseRemoved    BreakingChangeType = "response_removed"
	BreakingTypeChanged        BreakingChangeType = "type_changed"
	BreakingRequestBodyRemoved BreakingChangeType = "request_body_removed"
	BreakingSecurityAdded      BreakingChangeType = "security_added"
)

// BreakingChangeRule defines a rule for detecting breaking changes
type BreakingChangeRule struct {
	Type        BreakingChangeType
	Description string
	Severity    string // "error", "warning"
	Check       func(old, new map[string]interface{}) bool
}

// DefaultBreakingRules returns default breaking change rules
func DefaultBreakingRules() []BreakingChangeRule {
	return []BreakingChangeRule{
		{
			Type:        BreakingEndpointRemoved,
			Description: "Removing an endpoint breaks existing clients",
			Severity:    "error",
		},
		{
			Type:        BreakingParameterRemoved,
			Description: "Removing a parameter may break clients that send it",
			Severity:    "error",
		},
		{
			Type:        BreakingRequiredAdded,
			Description: "Adding a required field breaks clients not sending it",
			Severity:    "error",
		},
		{
			Type:        BreakingResponseRemoved,
			Description: "Removing a response code may break client error handling",
			Severity:    "warning",
		},
		{
			Type:        BreakingTypeChanged,
			Description: "Changing a field type breaks serialization",
			Severity:    "error",
		},
	}
}

// IsBreaking checks if a change type is considered breaking
func IsBreaking(changeType BreakingChangeType) bool {
	breakingTypes := map[BreakingChangeType]bool{
		BreakingEndpointRemoved:    true,
		BreakingParameterRemoved:   true,
		BreakingRequiredAdded:      true,
		BreakingResponseRemoved:    true,
		BreakingTypeChanged:        true,
		BreakingRequestBodyRemoved: true,
		BreakingSecurityAdded:      true,
	}
	return breakingTypes[changeType]
}
