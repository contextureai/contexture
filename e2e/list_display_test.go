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

func TestListCommand_WithPattern(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule that we know exists
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test filtering by language - should match
	result := project.Run(t, "rules", "list", "--pattern", "go").ExpectSuccess(t)
	result.ExpectStdout(t, "pattern: go")          // Should show pattern in header
	result.ExpectStdout(t, "languages/go/testing") // Should include go rule

	// Test filtering with pattern that won't match
	result = project.Run(t, "rules", "list", "--pattern", "python").ExpectSuccess(t)
	result.ExpectStdout(t, "No rules found matching pattern: python")
	result.ExpectNotStdout(t, "go/testing")
}

func TestListCommand_WithPatternShortFlag(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test using short flag -p
	result := project.Run(t, "rules", "list", "-p", "go").ExpectSuccess(t)
	result.ExpectStdout(t, "pattern: go")
	result.ExpectStdout(t, "languages/go/testing")
}

func TestListCommand_WithPatternNoMatches(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test pattern that doesn't match anything
	result := project.Run(t, "rules", "list", "--pattern", "nonexistent").ExpectSuccess(t)
	result.ExpectStdout(t, "No rules found matching pattern: nonexistent")
	result.ExpectNotStdout(t, "go/testing")
}

func TestListCommand_WithInvalidPattern(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test invalid regex pattern
	result := project.Run(t, "rules", "list", "--pattern", "[invalid").ExpectFailure(t)
	result.ExpectStderr(t, "invalid pattern")
}

func TestListCommand_JSONOutput(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test JSON output
	result := project.Run(t, "rules", "list", "--output", "json").ExpectSuccess(t)

	// Should contain JSON structure
	result.ExpectStdout(t, `"command": "rules list"`)
	result.ExpectStdout(t, `"version": "1.0"`)
	result.ExpectStdout(t, `"metadata":`)
	result.ExpectStdout(t, `"rules":`)
	result.ExpectStdout(t, `"totalRules": 1`)
	result.ExpectStdout(t, `"filteredRules": 1`)

	// Should contain rule data
	result.ExpectStdout(t, `"id": "[contexture:languages/go/testing]"`)
	result.ExpectStdout(t, `"title": "Go Testing Best Practices"`)
}

func TestListCommand_JSONOutputShortFlag(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test JSON output with short flag
	result := project.Run(t, "rules", "list", "-o", "json").ExpectSuccess(t)

	// Should contain JSON structure
	result.ExpectStdout(t, `"command": "rules list"`)
	result.ExpectStdout(t, `"version": "1.0"`)
}

func TestListCommand_JSONOutputWithPattern(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test JSON output with pattern
	result := project.Run(t, "rules", "list", "-p", "testing", "-o", "json").ExpectSuccess(t)

	// Should contain pattern in metadata
	result.ExpectStdout(t, `"pattern": "testing"`)
	result.ExpectStdout(t, `"command": "rules list"`)
}

func TestListCommand_JSONOutputEmptyRules(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project (no rules added)
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Test JSON output with no rules
	result := project.Run(t, "rules", "list", "-o", "json").ExpectSuccess(t)

	// Should contain empty rules array
	result.ExpectStdout(t, `"rules": []`)
	result.ExpectStdout(t, `"totalRules": 0`)
	result.ExpectStdout(t, `"filteredRules": 0`)
}

func TestListCommand_InvalidOutputFormat(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Test invalid output format
	result := project.Run(t, "rules", "list", "--output", "yaml").ExpectFailure(t)
	result.ExpectStderr(t, "unsupported output format: yaml")
	result.ExpectStderr(t, "supported formats: default, json")
}

func TestListCommand_DefaultOutputFormat(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--no-interactive", "--force").ExpectSuccess(t)

	// Add a rule
	project.Run(t, "rules", "add", "languages/go/testing", "--force").ExpectSuccess(t)

	// Test explicit default output format
	result := project.Run(t, "rules", "list", "--output", "default").ExpectSuccess(t)

	// Should contain terminal format (not JSON)
	result.ExpectStdout(t, "Installed Rules")
	result.ExpectStdout(t, "languages/go/testing")
	result.ExpectStdout(t, "Go Testing Best Practices")
	result.ExpectNotStdout(t, `"command":`)
	result.ExpectNotStdout(t, `"version":`)
}
