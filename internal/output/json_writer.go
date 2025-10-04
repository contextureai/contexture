package output

import (
	"encoding/json"
	"fmt"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
)

// JSONWriter implements Writer interface for JSON output format
type JSONWriter struct{}

// NewJSONWriter creates a new JSON writer
func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

// JSONRule represents a rule in JSON output (without timestamps)
type JSONRule struct {
	ID               string              `json:"id"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Tags             []string            `json:"tags"`
	Trigger          *domain.RuleTrigger `json:"trigger,omitempty"`
	Languages        []string            `json:"languages,omitempty"`
	Frameworks       []string            `json:"frameworks,omitempty"`
	Content          string              `json:"content"`
	Variables        map[string]any      `json:"variables,omitempty"`
	DefaultVariables map[string]any      `json:"defaultVariables,omitempty"`
	FilePath         string              `json:"filePath"`
	Source           string              `json:"source"`
	Ref              string              `json:"ref,omitempty"`
}

// JSONRulesListOutput represents the JSON structure for rules list output
type JSONRulesListOutput struct {
	Metadata ListMetadata `json:"metadata"`
	Rules    []*JSONRule  `json:"rules"`
}

// JSONRulesAddOutput represents the JSON structure for rules add output
type JSONRulesAddOutput struct {
	Metadata AddMetadata `json:"metadata"`
}

// JSONRulesRemoveOutput represents the JSON structure for rules remove output
type JSONRulesRemoveOutput struct {
	Metadata RemoveMetadata `json:"metadata"`
}

// JSONRulesUpdateOutput represents the JSON structure for rules update output
type JSONRulesUpdateOutput struct {
	Metadata UpdateMetadata `json:"metadata"`
}

// WriteRulesList writes rules list in JSON format to stdout
func (w *JSONWriter) WriteRulesList(rules []*domain.Rule, metadata ListMetadata) error {
	// Convert domain.Rule to JSONRule (without timestamps)
	jsonRules := make([]*JSONRule, len(rules))
	for i, rule := range rules {
		jsonRules[i] = &JSONRule{
			ID:               rule.ID,
			Title:            rule.Title,
			Description:      rule.Description,
			Tags:             rule.Tags,
			Trigger:          rule.Trigger,
			Languages:        rule.Languages,
			Frameworks:       rule.Frameworks,
			Content:          rule.Content,
			Variables:        rule.Variables,
			DefaultVariables: rule.DefaultVariables,
			FilePath:         rule.FilePath,
			Source:           rule.Source,
			Ref:              rule.Ref,
		}
	}

	output := JSONRulesListOutput{
		Metadata: metadata,
		Rules:    jsonRules,
	}

	// Marshal with indentation for readability
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return contextureerrors.Wrap(err, "marshal rules to JSON")
	}

	// Print to stdout
	fmt.Println(string(jsonData))
	return nil
}

// WriteRulesAdd writes rules add result in JSON format to stdout
func (w *JSONWriter) WriteRulesAdd(metadata AddMetadata) error {
	output := JSONRulesAddOutput{
		Metadata: metadata,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return contextureerrors.Wrap(err, "marshal add result to JSON")
	}

	fmt.Println(string(jsonData))
	return nil
}

// WriteRulesRemove writes rules remove result in JSON format to stdout
func (w *JSONWriter) WriteRulesRemove(metadata RemoveMetadata) error {
	output := JSONRulesRemoveOutput{
		Metadata: metadata,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return contextureerrors.Wrap(err, "marshal remove result to JSON")
	}

	fmt.Println(string(jsonData))
	return nil
}

// WriteRulesUpdate writes rules update result in JSON format to stdout
func (w *JSONWriter) WriteRulesUpdate(metadata UpdateMetadata) error {
	output := JSONRulesUpdateOutput{
		Metadata: metadata,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return contextureerrors.Wrap(err, "marshal update result to JSON")
	}

	fmt.Println(string(jsonData))
	return nil
}
