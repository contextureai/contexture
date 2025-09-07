package domain

// FormatMetadata contains metadata about a format implementation
type FormatMetadata struct {
	Type        FormatType
	DisplayName string
	Description string
	IsDirectory bool // true if format outputs to directories, false for single files
}

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

	// GetOutputPath returns the output path for this format
	GetOutputPath(config *FormatConfig) string

	// CleanupEmptyDirectories handles cleanup of empty directories for this format
	CleanupEmptyDirectories(config *FormatConfig) error

	// CreateDirectories creates necessary directories for this format
	CreateDirectories(config *FormatConfig) error

	// GetMetadata returns metadata about this format
	GetMetadata() *FormatMetadata
}
