// Package output provides extensible output formatting for CLI commands
package output

import (
	"github.com/contextureai/contexture/internal/domain"
)

// Format represents the output format type
type Format string

const (
	// FormatDefault represents the default terminal output format
	FormatDefault Format = "default"
	// FormatJSON represents JSON output format
	FormatJSON Format = "json"
)

// Writer interface for different output formats
type Writer interface {
	WriteRulesList(rules []*domain.Rule, metadata ListMetadata) error
	WriteRulesAdd(metadata AddMetadata) error
	WriteRulesRemove(metadata RemoveMetadata) error
	WriteRulesUpdate(metadata UpdateMetadata) error
}

// ListMetadata contains contextual information for rules list commands
type ListMetadata struct {
	Pattern       string `json:"pattern,omitempty"`
	TotalRules    int    `json:"totalRules"`
	FilteredRules int    `json:"filteredRules"`
}

// AddMetadata contains contextual information for rules add commands
type AddMetadata struct {
	RulesAdded []string `json:"rulesAdded"`
}

// RemoveMetadata contains contextual information for rules remove commands
type RemoveMetadata struct {
	RulesRemoved []string `json:"rulesRemoved"`
}

// UpdateMetadata contains contextual information for rules update commands
type UpdateMetadata struct {
	RulesUpdated  []string `json:"rulesUpdated"`
	RulesUpToDate []string `json:"rulesUpToDate,omitempty"`
	RulesFailed   []string `json:"rulesFailed,omitempty"`
}

// Manager handles output format selection and writing
type Manager struct {
	format Format
	writer Writer
}

// NewManager creates a new output manager for the specified format
func NewManager(format Format) (*Manager, error) {
	var writer Writer
	var err error

	switch format {
	case FormatDefault, "":
		writer = NewTerminalWriter()
	case FormatJSON:
		writer = NewJSONWriter()
	default:
		return nil, &UnsupportedFormatError{Format: string(format)}
	}

	return &Manager{
		format: format,
		writer: writer,
	}, err
}

// WriteRulesList writes the rules list using the configured format
func (m *Manager) WriteRulesList(rules []*domain.Rule, metadata ListMetadata) error {
	return m.writer.WriteRulesList(rules, metadata)
}

// WriteRulesAdd writes the rules add result using the configured format
func (m *Manager) WriteRulesAdd(metadata AddMetadata) error {
	return m.writer.WriteRulesAdd(metadata)
}

// WriteRulesRemove writes the rules remove result using the configured format
func (m *Manager) WriteRulesRemove(metadata RemoveMetadata) error {
	return m.writer.WriteRulesRemove(metadata)
}

// WriteRulesUpdate writes the rules update result using the configured format
func (m *Manager) WriteRulesUpdate(metadata UpdateMetadata) error {
	return m.writer.WriteRulesUpdate(metadata)
}

// UnsupportedFormatError represents an error for unsupported output formats
type UnsupportedFormatError struct {
	Format string
}

func (e *UnsupportedFormatError) Error() string {
	return "unsupported output format: " + e.Format + " (supported formats: default, json)"
}
