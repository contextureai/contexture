package output

import (
	"encoding/json"
	"fmt"

	"github.com/contextureai/contexture/internal/domain"
)

// JSONWriter implements Writer interface for JSON output format
type JSONWriter struct{}

// NewJSONWriter creates a new JSON writer
func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

// JSONRulesListOutput represents the JSON structure for rules list output
type JSONRulesListOutput struct {
	Command  string         `json:"command"`
	Version  string         `json:"version"`
	Metadata ListMetadata   `json:"metadata"`
	Rules    []*domain.Rule `json:"rules"`
}

// WriteRulesList writes rules list in JSON format to stdout
func (w *JSONWriter) WriteRulesList(rules []*domain.Rule, metadata ListMetadata) error {
	// Set version for the JSON schema
	metadata.Version = "1.0"

	output := JSONRulesListOutput{
		Command:  metadata.Command,
		Version:  metadata.Version,
		Metadata: metadata,
		Rules:    rules,
	}

	// Marshal with indentation for readability
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules to JSON: %w", err)
	}

	// Print to stdout
	fmt.Println(string(jsonData))
	return nil
}
