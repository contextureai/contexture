// Package output provides extensible output formatting for CLI commands
package output

import (
	"time"

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
}

// ListMetadata contains contextual information for rules list commands
type ListMetadata struct {
	Command       string    `json:"command"`
	Version       string    `json:"version"`
	Pattern       string    `json:"pattern,omitempty"`
	TotalRules    int       `json:"totalRules"`
	FilteredRules int       `json:"filteredRules"`
	Timestamp     time.Time `json:"timestamp"`
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

// UnsupportedFormatError represents an error for unsupported output formats
type UnsupportedFormatError struct {
	Format string
}

func (e *UnsupportedFormatError) Error() string {
	return "unsupported output format: " + e.Format + " (supported formats: default, json)"
}
