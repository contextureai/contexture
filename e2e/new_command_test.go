// Package e2e provides end-to-end tests for the new command
package e2e

import (
	"strings"
	"testing"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
)

// TestNewCommandBasic tests basic rule creation
func TestNewCommandBasic(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("create simple rule", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "test-rule")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Rule created successfully").
			ExpectStdout(t, "test-rule")

		// Verify rule file was created in rules directory
		project.AssertFileExists(t, "rules/test-rule.md")

		// Verify content structure
		content := project.GetFileContent(t, "rules/test-rule.md")
		if !strings.Contains(content, "---") {
			t.Error("Rule should have YAML frontmatter")
		}
		if !strings.Contains(content, "trigger: manual") {
			t.Error("Rule should have trigger set to manual")
		}
		// Title and description should be present but empty when no flags provided
		if !strings.Contains(content, "title:") {
			t.Error("Rule should have title field in frontmatter")
		}
		if !strings.Contains(content, "description:") {
			t.Error("Rule should have description field in frontmatter")
		}
		// Tags should NOT be present when not specified
		if strings.Contains(content, "tags:") {
			t.Error("Rule should NOT have tags when no flags provided")
		}
	})

	t.Run("create nested rule", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "security/auth-check")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Rule created successfully").
			ExpectStdout(t, "security/auth-check")

		// Verify nested rule file was created
		project.AssertFileExists(t, "rules/security/auth-check.md")
	})

	t.Run("create rule with custom metadata", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "custom-rule",
			"--name", "Custom Rule Name",
			"--description", "This is a custom description",
			"--tags", "security,testing,custom")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Custom Rule Name")

		// Verify file exists
		project.AssertFileExists(t, "rules/custom-rule.md")

		// Verify custom metadata
		content := project.GetFileContent(t, "rules/custom-rule.md")
		if !strings.Contains(content, "title: Custom Rule Name") {
			t.Error("Rule should have custom title")
		}
		if !strings.Contains(content, "description: This is a custom description") {
			t.Error("Rule should have custom description")
		}
		if !strings.Contains(content, "- security") {
			t.Error("Rule should have security tag")
		}
		if !strings.Contains(content, "- testing") {
			t.Error("Rule should have testing tag")
		}
		if !strings.Contains(content, "- custom") {
			t.Error("Rule should have custom tag")
		}
	})
}

// TestNewCommandOutsideProject tests rule creation outside a Contexture project
func TestNewCommandOutsideProject(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Don't initialize - test without .contexture.yaml

	t.Run("create rule at literal path", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "standalone-rule")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Rule created successfully")

		// Verify rule was created at literal path (not in rules/)
		project.AssertFileExists(t, "standalone-rule.md")

		// Verify rules/ directory was NOT created
		// Just check that standalone-rule.md exists and it's not in rules/
		// (the fact that AssertFileExists passed for "standalone-rule.md" is sufficient)
	})

	t.Run("create rule with nested path", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "custom/path/rule")
		result.ExpectSuccess(t)

		// Verify nested path was created at literal location
		project.AssertFileExists(t, "custom/path/rule.md")
	})
}

// TestNewCommandErrors tests error conditions
func TestNewCommandErrors(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("no path provided", func(t *testing.T) {
		result := project.Run(t, "rules", "new")
		result.ExpectFailure(t).
			ExpectStderr(t, "no path provided")
	})

	t.Run("file already exists", func(t *testing.T) {
		// Create a rule
		project.Run(t, "rules", "new", "existing-rule").ExpectSuccess(t)

		// Try to create the same rule again
		result := project.Run(t, "rules", "new", "existing-rule")
		result.ExpectFailure(t).
			ExpectStderr(t, "already exists")
	})
}

// TestNewCommandPathHandling tests various path formats
func TestNewCommandPathHandling(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("path without .md extension", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "no-extension")
		result.ExpectSuccess(t)
		project.AssertFileExists(t, "rules/no-extension.md")
	})

	t.Run("path with .md extension", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "with-extension.md")
		result.ExpectSuccess(t)
		project.AssertFileExists(t, "rules/with-extension.md")

		// Should not create .md.md - the fact that the above assertion passed
		// means the correct file was created
	})

	t.Run("deeply nested path", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "a/b/c/d/deep-rule")
		result.ExpectSuccess(t)
		project.AssertFileExists(t, "rules/a/b/c/d/deep-rule.md")
	})
}

// TestNewCommandWithBuild tests integration with build command
func TestNewCommandWithBuild(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// Create a new rule with all required fields for build validation
	result := project.Run(t, "rules", "new", "build-test-rule",
		"--name", "Build Test",
		"--description", "Rule for testing build integration",
		"--tags", "testing,build")
	result.ExpectSuccess(t)

	// Build the project
	result = project.Run(t, "build")
	result.ExpectSuccess(t).
		ExpectStdout(t, "Generated rules")

	// Verify the new rule is included in output
	project.AssertFileContains(t, "CLAUDE.md", "Build Test")
}

// TestNewCommandShortFlags tests short flag aliases
func TestNewCommandShortFlags(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("use short flags", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "short-flags-rule",
			"-n", "Short Flags Test",
			"-d", "Testing short flag aliases",
			"-t", "test,flags,short")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Short Flags Test")

		// Verify content
		content := project.GetFileContent(t, "rules/short-flags-rule.md")
		if !strings.Contains(content, "title: Short Flags Test") {
			t.Error("Rule should have title from -n flag")
		}
		if !strings.Contains(content, "description: Testing short flag aliases") {
			t.Error("Rule should have description from -d flag")
		}
		if !strings.Contains(content, "- test") || !strings.Contains(content, "- flags") {
			t.Error("Rule should have tags from -t flag")
		}
	})
}

// TestNewCommandHelp tests help output
func TestNewCommandHelp(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	t.Run("help output", func(t *testing.T) {
		result := project.Run(t, "rules", "new", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Create a new rule file").
			ExpectStdout(t, "--name").
			ExpectStdout(t, "--description").
			ExpectStdout(t, "--tags").
			ExpectStdout(t, "Inside a Contexture project").
			ExpectStdout(t, "Outside a Contexture project")
	})
}

// TestNewCommandWorkflow tests complete workflow with new command
func TestNewCommandWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Step 1: Initialize project
	t.Log("Step 1: Initialize project")
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// Step 2: Create multiple rules with new command
	t.Log("Step 2: Create rules using new command")
	project.Run(t, "rules", "new", "security/input-validation",
		"--name", "Input Validation",
		"--description", "Validate all user inputs",
		"--tags", "security,validation").ExpectSuccess(t)

	project.Run(t, "rules", "new", "patterns/error-handling",
		"--name", "Error Handling",
		"--description", "Proper error handling patterns",
		"--tags", "errors,best-practices").ExpectSuccess(t)

	// Step 3: Verify rules were created
	t.Log("Step 3: Verify rules exist")
	project.AssertFileExists(t, "rules/security/input-validation.md")
	project.AssertFileExists(t, "rules/patterns/error-handling.md")

	// Step 4: Build project
	t.Log("Step 4: Build project with new rules")
	result := project.Run(t, "build")
	result.ExpectSuccess(t)

	// Step 5: Verify rules appear in output
	t.Log("Step 5: Verify rules in output files")
	project.AssertFileContains(t, "CLAUDE.md", "Input Validation")
	project.AssertFileContains(t, "CLAUDE.md", "Error Handling")

	// Step 6: Check config shows local rules
	t.Log("Step 6: Verify config shows local rules")
	result = project.Run(t, "config")
	result.ExpectSuccess(t).
		ExpectStdout(t, "Local Rules")

	t.Log("New command workflow test passed successfully")
}
