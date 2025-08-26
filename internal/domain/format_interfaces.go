package domain

// Format defines the interface for output format implementations
type Format interface {
	// Transform converts a processed rule to format-specific representation
	Transform(processedRule *ProcessedRule) (*TransformedRule, error)

	// Validate checks if a rule is valid for this format
	Validate(rule *Rule) (*ValidationResult, error)

	// Write outputs transformed rules to the target location
	Write(rules []*TransformedRule, config *FormatConfig) error

	// Remove deletes a specific rule from the target location
	Remove(ruleID string, config *FormatConfig) error

	// List returns all currently installed rules for this format
	List(config *FormatConfig) ([]*InstalledRule, error)
}
