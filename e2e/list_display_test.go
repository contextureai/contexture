package e2e

import (
	"strings"
	"testing"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestListCommand_EmptyProject(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Run list command
	result := project.Run(t, "rules", "list").ExpectSuccess(t)

	// Verify empty output
	result.ExpectStdout(t, "No rules found.")
	result.ExpectNotStdout(t, "Installed Rules")
}

func TestListCommand_WithRules(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Run list command
	result := project.Run(t, "rules", "list").ExpectSuccess(t)

	// Verify output format
	lines := strings.Split(result.Stdout, "\n")

	// Should have header without count
	headerFound := false
	for _, line := range lines {
		if strings.Contains(line, "Installed Rules") {
			headerFound = true
			assert.NotContains(t, line, "total)", "Header should not show count")
			break
		}
	}
	assert.True(t, headerFound, "Should contain header 'Installed Rules'")

	// Should contain rule path
	result.ExpectStdout(t, "languages/go/testing")

	// Should not contain "Source:" prefix for default rules
	result.ExpectNotStdout(t, "Source:")
}

func TestListCommand_OutputFormat(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Run list command
	result := project.Run(t, "rules", "list").ExpectSuccess(t)

	lines := strings.Split(result.Stdout, "\n")

	// Find the rule path line and verify formatting
	rulePathLineIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "languages/go/testing" {
			rulePathLineIndex = i
			break
		}
	}

	assert.GreaterOrEqual(t, rulePathLineIndex, 0, "Should find rule path line")

	if rulePathLineIndex >= 0 {
		// Rule path should not be indented
		ruleLine := lines[rulePathLineIndex]
		assert.False(t, strings.HasPrefix(ruleLine, " "), "Rule path should not be indented")

		// Next line should be the title and should be indented
		if rulePathLineIndex+1 < len(lines) {
			titleLine := lines[rulePathLineIndex+1]
			assert.True(t, strings.HasPrefix(titleLine, "  "), "Title should be indented with 2 spaces")
		}
	}
}

func TestListCommand_NoProject(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Run list command without initializing project
	result := project.Run(t, "rules", "list").ExpectFailure(t)

	// Should provide helpful error message
	result.ExpectStderr(t, "no project configuration found")
}
