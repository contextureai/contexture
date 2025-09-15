package output

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewTerminalWriter(t *testing.T) {
	writer := NewTerminalWriter()
	assert.NotNil(t, writer)
	assert.Implements(t, (*Writer)(nil), writer)
}

func TestTerminalWriter_WriteRulesList_EmptyRules(t *testing.T) {
	writer := NewTerminalWriter()
	metadata := ListMetadata{
		TotalRules:    0,
		FilteredRules: 0,
	}

	// Test with empty rules - should not error
	err := writer.WriteRulesList([]*domain.Rule{}, metadata)
	assert.NoError(t, err)
}

func TestTerminalWriter_WriteRulesList_WithRules(t *testing.T) {
	writer := NewTerminalWriter()

	rules := []*domain.Rule{
		{
			ID:          "test-rule",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"testing"},
			Content:     "Rule content",
		},
	}

	metadata := ListMetadata{
		TotalRules:    1,
		FilteredRules: 1,
	}

	// Should delegate to existing display logic without error
	err := writer.WriteRulesList(rules, metadata)
	assert.NoError(t, err)
}

func TestTerminalWriter_WriteRulesList_WithPattern(t *testing.T) {
	writer := NewTerminalWriter()

	rules := []*domain.Rule{
		{
			ID:          "testing-rule",
			Title:       "Testing Rule",
			Description: "A rule for testing",
			Tags:        []string{"testing", "validation"},
		},
	}

	metadata := ListMetadata{
		Pattern:       "testing",
		TotalRules:    1,
		FilteredRules: 1,
	}

	// Should pass pattern to display options
	err := writer.WriteRulesList(rules, metadata)
	assert.NoError(t, err)
}

func TestTerminalWriter_WriteRulesList_MultipleRules(t *testing.T) {
	writer := NewTerminalWriter()

	rules := []*domain.Rule{
		{
			ID:          "rule-1",
			Title:       "First Rule",
			Description: "First test rule",
			Tags:        []string{"testing"},
		},
		{
			ID:          "rule-2",
			Title:       "Second Rule",
			Description: "Second test rule",
			Tags:        []string{"validation"},
		},
	}

	metadata := ListMetadata{
		TotalRules:    2,
		FilteredRules: 2,
	}

	// Should handle multiple rules without error
	err := writer.WriteRulesList(rules, metadata)
	assert.NoError(t, err)
}
