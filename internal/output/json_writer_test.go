package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout captures stdout during test execution
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	// Save original stdout
	originalStdout := os.Stdout
	defer func() {
		os.Stdout = originalStdout
	}()

	// Create a pipe to capture output
	r, w, err := os.Pipe()
	require.NoError(t, err)

	// Replace stdout with our pipe writer
	os.Stdout = w

	// Run the function in a goroutine
	done := make(chan bool)
	var output string
	go func() {
		defer close(done)
		buf := bytes.NewBuffer(nil)
		_, copyErr := io.Copy(buf, r)
		assert.NoError(t, copyErr)
		output = buf.String()
	}()

	// Execute the function
	fn()

	// Close writer and wait for reader to finish
	_ = w.Close()
	<-done
	_ = r.Close()

	return strings.TrimSpace(output)
}

func TestJSONWriter_WriteRulesList_EmptyRules(t *testing.T) {
	writer := NewJSONWriter()
	metadata := ListMetadata{
		Command:       "rules list",
		Pattern:       "",
		TotalRules:    0,
		FilteredRules: 0,
		Timestamp:     time.Date(2025, 9, 14, 19, 0, 0, 0, time.UTC),
	}

	output := captureStdout(t, func() {
		err := writer.WriteRulesList([]*domain.Rule{}, metadata)
		require.NoError(t, err)
	})

	// Parse and verify JSON
	var result JSONRulesListOutput
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "rules list", result.Command)
	assert.Equal(t, "1.0", result.Version)
	assert.Empty(t, result.Metadata.Pattern)
	assert.Equal(t, 0, result.Metadata.TotalRules)
	assert.Equal(t, 0, result.Metadata.FilteredRules)
	assert.Empty(t, result.Rules)
}

func TestJSONWriter_WriteRulesList_SingleRule(t *testing.T) {
	writer := NewJSONWriter()

	rule := &domain.Rule{
		ID:          "test-rule-id",
		Title:       "Test Rule",
		Description: "A test rule for validation",
		Tags:        []string{"testing", "validation"},
		Languages:   []string{"go"},
		Frameworks:  []string{"testing"},
		Content:     "Rule content here",
		Variables:   map[string]any{"key": "value"},
		FilePath:    "test/rule.md",
		Source:      "test-source",
	}

	metadata := ListMetadata{
		Command:       "rules list",
		Pattern:       "testing",
		TotalRules:    1,
		FilteredRules: 1,
		Timestamp:     time.Date(2025, 9, 14, 19, 0, 0, 0, time.UTC),
	}

	output := captureStdout(t, func() {
		err := writer.WriteRulesList([]*domain.Rule{rule}, metadata)
		require.NoError(t, err)
	})

	// Parse and verify JSON
	var result JSONRulesListOutput
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "rules list", result.Command)
	assert.Equal(t, "1.0", result.Version)
	assert.Equal(t, "testing", result.Metadata.Pattern)
	assert.Equal(t, 1, result.Metadata.TotalRules)
	assert.Equal(t, 1, result.Metadata.FilteredRules)
	assert.Len(t, result.Rules, 1)

	// Verify rule content
	outputRule := result.Rules[0]
	assert.Equal(t, "test-rule-id", outputRule.ID)
	assert.Equal(t, "Test Rule", outputRule.Title)
	assert.Equal(t, "A test rule for validation", outputRule.Description)
	assert.Equal(t, []string{"testing", "validation"}, outputRule.Tags)
	assert.Equal(t, []string{"go"}, outputRule.Languages)
	assert.Equal(t, []string{"testing"}, outputRule.Frameworks)
	assert.Equal(t, "Rule content here", outputRule.Content)
	assert.Equal(t, map[string]any{"key": "value"}, outputRule.Variables)
	assert.Equal(t, "test/rule.md", outputRule.FilePath)
	assert.Equal(t, "test-source", outputRule.Source)
}

func TestJSONWriter_WriteRulesList_MultipleRules(t *testing.T) {
	writer := NewJSONWriter()

	rules := []*domain.Rule{
		{
			ID:          "rule-1",
			Title:       "First Rule",
			Description: "First test rule",
			Tags:        []string{"tag1"},
		},
		{
			ID:          "rule-2",
			Title:       "Second Rule",
			Description: "Second test rule",
			Tags:        []string{"tag2"},
		},
	}

	metadata := ListMetadata{
		Command:       "rules list",
		TotalRules:    2,
		FilteredRules: 2,
		Timestamp:     time.Date(2025, 9, 14, 19, 0, 0, 0, time.UTC),
	}

	output := captureStdout(t, func() {
		err := writer.WriteRulesList(rules, metadata)
		require.NoError(t, err)
	})

	// Parse and verify JSON
	var result JSONRulesListOutput
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Len(t, result.Rules, 2)
	assert.Equal(t, "rule-1", result.Rules[0].ID)
	assert.Equal(t, "rule-2", result.Rules[1].ID)
}

func TestJSONWriter_WriteRulesList_ValidJSONFormat(t *testing.T) {
	writer := NewJSONWriter()

	rule := &domain.Rule{
		ID:          "format-test",
		Title:       "JSON Format Test",
		Description: "Testing JSON formatting",
		Tags:        []string{"format"},
	}

	metadata := ListMetadata{
		Command:       "rules list",
		TotalRules:    1,
		FilteredRules: 1,
		Timestamp:     time.Date(2025, 9, 14, 19, 0, 0, 0, time.UTC),
	}

	output := captureStdout(t, func() {
		err := writer.WriteRulesList([]*domain.Rule{rule}, metadata)
		require.NoError(t, err)
	})

	// Verify it's valid JSON
	var jsonData interface{}
	err := json.Unmarshal([]byte(output), &jsonData)
	require.NoError(t, err)

	// Verify it's properly formatted (indented)
	assert.Contains(t, output, "  \"command\":")
	assert.Contains(t, output, "  \"version\":")
	assert.Contains(t, output, "  \"metadata\":")
	assert.Contains(t, output, "  \"rules\":")
}

func TestJSONWriter_WriteRulesList_TimestampFormat(t *testing.T) {
	writer := NewJSONWriter()

	testTime := time.Date(2025, 9, 14, 19, 30, 45, 123456789, time.UTC)
	metadata := ListMetadata{
		Command:       "rules list",
		TotalRules:    0,
		FilteredRules: 0,
		Timestamp:     testTime,
	}

	output := captureStdout(t, func() {
		err := writer.WriteRulesList([]*domain.Rule{}, metadata)
		require.NoError(t, err)
	})

	// Parse and verify timestamp format
	var result JSONRulesListOutput
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Should preserve the exact timestamp
	assert.Equal(t, testTime, result.Metadata.Timestamp)
}

func TestNewJSONWriter(t *testing.T) {
	writer := NewJSONWriter()
	assert.NotNil(t, writer)
	assert.Implements(t, (*Writer)(nil), writer)
}
