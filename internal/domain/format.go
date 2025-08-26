package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	contextureerrors "github.com/contextureai/contexture/internal/errors"
)

// FormatType represents the type of output format
type FormatType string

const (
	// FormatClaude represents the Claude AI assistant format (CLAUDE.md)
	FormatClaude FormatType = "claude"
	// FormatCursor represents the Cursor IDE format (.cursorrules/)
	FormatCursor FormatType = "cursor"
	// FormatWindsurf represents the Windsurf IDE format (.windsurf/rules/)
	FormatWindsurf FormatType = "windsurf"
)

// String returns the string representation of the format type
func (ft FormatType) String() string {
	return string(ft)
}

// FormatConfig represents the core format configuration
type FormatConfig struct {
	Type    FormatType `yaml:"type"    json:"type"    validate:"required,oneof=claude cursor windsurf"`
	Enabled bool       `yaml:"enabled" json:"enabled"`
	BaseDir string     `yaml:"-"       json:"-"` // Runtime option, not serialized
}

// FormatSpecificRule represents a rule with format-specific configuration
type FormatSpecificRule struct {
	ID        string         `yaml:"id"                  json:"id"                  validate:"required"`
	Filename  string         `yaml:"filename,omitempty"  json:"filename,omitempty"`
	Variables map[string]any `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// TransformedRule represents a rule that has been transformed for a specific format
type TransformedRule struct {
	// Original rule
	Rule *Rule

	// Transformed content
	Content string

	// Format-specific metadata
	Metadata map[string]any

	// Target filename (for multi-file formats)
	Filename string

	// Relative path within format directory
	RelativePath string

	// Transformation timestamp
	TransformedAt time.Time

	// Content hash for change detection (when persisted)
	ContentHash string

	// File size (when persisted)
	Size int64
}

// ValidationResult represents the result of rule validation for a format
type ValidationResult struct {
	// Whether the rule is valid for this format
	Valid bool

	// Validation errors
	Errors []error

	// Validation warnings
	Warnings []ValidationWarning

	// Format-specific validation metadata
	Metadata map[string]any
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// InstalledRule represents a rule that is currently installed in a format
// This is an alias for TransformedRule with additional installation metadata
type InstalledRule struct {
	*TransformedRule

	// Installation timestamp
	InstalledAt time.Time
}

// NewInstalledRule creates an InstalledRule from a TransformedRule
func NewInstalledRule(transformed *TransformedRule) *InstalledRule {
	return &InstalledRule{
		TransformedRule: transformed,
		InstalledAt:     time.Now(),
	}
}

// ID returns the rule ID
func (ir *InstalledRule) ID() string {
	if ir.TransformedRule != nil && ir.Rule != nil {
		return ir.Rule.ID
	}
	return ""
}

// Title returns the rule title
func (ir *InstalledRule) Title() string {
	if ir.TransformedRule != nil && ir.Rule != nil {
		return ir.Rule.Title
	}
	return ""
}

// Source returns the rule source
func (ir *InstalledRule) Source() string {
	if ir.TransformedRule != nil && ir.Rule != nil {
		return ir.Rule.Source
	}
	return ""
}

// Ref returns the rule ref (branch/tag/commit)
func (ir *InstalledRule) Ref() string {
	if ir.TransformedRule != nil && ir.Rule != nil {
		return ir.Rule.Ref
	}
	return ""
}

// Error returns a formatted error string for ValidationResult
func (vr *ValidationResult) Error() string {
	if vr.Valid {
		return ""
	}

	if len(vr.Errors) == 0 {
		return "validation failed"
	}

	return vr.Errors[0].Error()
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings returns true if there are validation warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(field, message, code string) {
	vr.Errors = append(
		vr.Errors,
		contextureerrors.ValidationErrorf(field, "%s (code: %s)", message, code),
	)
	vr.Valid = false
}

// AddWarning adds a validation warning
func (vr *ValidationResult) AddWarning(field, message, code string) {
	vr.Warnings = append(vr.Warnings, ValidationWarning{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// GetContentHash calculates a hash of the transformed content
func (tr *TransformedRule) GetContentHash() string {
	// Use SHA256 for proper content hashing
	h := sha256.New()
	h.Write([]byte(tr.Content))
	return hex.EncodeToString(h.Sum(nil))
}

// GetAbsolutePath returns the absolute path for the transformed rule
func (tr *TransformedRule) GetAbsolutePath(baseDir string) string {
	if tr.RelativePath == "" {
		return baseDir + "/" + tr.Filename
	}
	return baseDir + "/" + tr.RelativePath
}
