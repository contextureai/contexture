package output

import (
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/ui/rules"
)

// TerminalWriter implements Writer interface for terminal output format
type TerminalWriter struct{}

// NewTerminalWriter creates a new terminal writer
func NewTerminalWriter() *TerminalWriter {
	return &TerminalWriter{}
}

// WriteRulesList writes rules list in terminal format using existing display logic
func (w *TerminalWriter) WriteRulesList(rulesSlice []*domain.Rule, metadata ListMetadata) error {
	// Create display options from metadata
	options := rules.DefaultDisplayOptions()

	// Apply pattern from metadata if provided
	if metadata.Pattern != "" {
		options.Pattern = metadata.Pattern
	}

	// Delegate to existing display logic
	return rules.DisplayRuleList(rulesSlice, options)
}
