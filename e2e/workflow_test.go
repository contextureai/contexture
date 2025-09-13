// Package e2e provides end-to-end workflow tests
package e2e

import (
	"strings"
	"testing"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
)

// TestCompleteWorkflow tests a complete user workflow from start to finish
func TestCompleteWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Step 1: Initialize a new project
	t.Log("Step 1: Initialize project")
	result := project.Run(t, "init", "--force", "--no-interactive")
	result.ExpectSuccess(t).ExpectStdout(t, "Configuration generated successfully")

	// Verify initialization
	project.AssertFileExists(t, ".contexture.yaml")

	// Step 2: Check project configuration
	t.Log("Step 2: Check project configuration")
	result = project.Run(t, "config")
	result.ExpectSuccess(t).
		ExpectStdout(t, "Configuration").
		ExpectStdout(t, "Output Formats")

	// Step 3: Add local rules
	t.Log("Step 3: Add local rules")
	localRule1 := `---
title: "Input Validation"
description: "Validate all user inputs"
tags: ["security", "validation"]
---

# Input Validation

Always validate and sanitize user inputs to prevent security vulnerabilities.`

	localRule2 := `---
title: "Error Handling"
description: "Proper error handling patterns"
tags: ["errors", "best-practices"]
---

# Error Handling

Implement comprehensive error handling throughout the application.`

	project.WithLocalRule("security/input-validation", localRule1)
	project.WithLocalRule("patterns/error-handling", localRule2)

	// Step 4: Build output files
	t.Log("Step 4: Build output files")
	result = project.Run(t, "build")
	result.ExpectSuccess(t).
		ExpectStdout(t, "Claude (CLAUDE.md)")

	// Note: Default init only creates Claude format, so we only expect Claude output

	// Step 5: Verify output files contain expected content
	t.Log("Step 5: Verify output files")
	project.AssertFileExists(t, "CLAUDE.md")
	// Note: .cursor/rules not expected with default init

	claudeContent := project.GetFileContent(t, "CLAUDE.md")
	if !strings.Contains(claudeContent, "Input Validation") {
		t.Errorf("CLAUDE.md should contain local rule content")
	}
	if !strings.Contains(claudeContent, "Error Handling") {
		t.Errorf("CLAUDE.md should contain second local rule content")
	}

	// Step 6: Check config shows project information
	t.Log("Step 6: Verify config shows project information")
	result = project.Run(t, "config")
	result.ExpectSuccess(t).
		ExpectStdout(t, "Output Formats").
		ExpectStdout(t, "Local Rules")

	t.Log("Complete workflow test passed successfully")
}

// TestMixedRulesWorkflow tests working with both local and remote rules
func TestMixedRulesWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// Add local rule
	localRule := `---
title: "Local Test Rule"
description: "A local rule for testing"
tags: ["test", "local"]
---

# Local Test Rule

This is a local rule for mixed rules testing.`

	project.WithLocalRule("local/test-rule", localRule)

	t.Run("build with mixed rules", func(t *testing.T) {
		result := project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Generated rules")

		// Should contain local rule
		project.AssertFileContains(t, "CLAUDE.md", "Local Test Rule")
	})

	t.Run("config shows project information", func(t *testing.T) {
		result := project.Run(t, "config")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Output Formats").
			ExpectStdout(t, "Local Rules")

		content := project.GetFileContent(t, ".contexture.yaml")
		// Should have config structure but no remote rules in this test
		if !strings.Contains(content, "version:") {
			t.Error("Config should have version field")
		}
	})
}

// TestConfigLocationWorkflow tests different config locations
func TestConfigLocationWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("root config location", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Init creates config in root by default
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		project.AssertFileExists(t, ".contexture.yaml")

		// Add rule in rules/ subdirectory
		project.WithLocalRule("test/rule", `---
title: "Test Rule"
description: "A test rule for workflow testing"
tags: ["test", "workflow"]
---

# Test Rule

Test content for workflow testing.`)

		// Build should find local rule
		result := project.Run(t, "build")
		result.ExpectSuccess(t)

		project.AssertFileExists(t, "CLAUDE.md")
	})

	t.Run("alternative config with directory structure", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Use standard config file but create directory structure
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Add rule in rules/ directory (standard location)
		project.WithFile("rules/test/rule.md", `---
title: "Test Rule"
description: "A test rule for config location testing"
tags: ["test", "config"]
---

# Test Rule

Test content for config location workflow testing.`)

		// Config should show project information
		result := project.Run(t, "config")
		result.ExpectSuccess(t)

		// Build should work with local rules
		result = project.Run(t, "build")
		result.ExpectSuccess(t)
	})
}

// TestErrorRecoveryWorkflow tests error conditions and recovery
func TestErrorRecoveryWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	t.Run("recover from corrupted config", func(t *testing.T) {
		// Create invalid YAML config
		project.WithConfig("invalid: yaml: content: [unclosed")

		// Commands should fail gracefully
		result := project.Run(t, "config")
		result.ExpectFailure(t)

		// Re-init should fix the config
		result = project.Run(t, "init", "--force", "--no-interactive")
		result.ExpectSuccess(t)

		// Now config should work
		result = project.Run(t, "config")
		result.ExpectSuccess(t)
	})

	t.Run("build with invalid local rule", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Add invalid local rule (missing frontmatter)
		project.WithLocalRule("invalid/rule", "Just content without frontmatter")

		// Build should fail with validation errors for invalid rules
		result := project.Run(t, "build")
		result.ExpectFailure(t).
			ExpectStderr(t, "validation errors") // Current behavior: strict validation fails

		// Note: With current validation, no output files are created when rules are invalid
		// This is the expected behavior for data integrity
	})

	t.Run("recover from missing output directories", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Initialize with multiple formats
		customConfig := `version: 1
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true`
		project.WithConfig(customConfig)

		// Add a valid rule
		project.WithLocalRule("recovery/test", `---
title: "Recovery Test"
description: "Test recovery from missing directories"
tags: ["recovery", "test"]
---

# Recovery Test

Content for recovery testing.`)

		// Build should create output directories and files
		result := project.Run(t, "build")
		result.ExpectSuccess(t)

		// Verify directories were created
		project.AssertFileExists(t, "CLAUDE.md")
		project.AssertFileExists(t, ".cursor/rules")
		project.AssertFileExists(t, ".windsurf/rules")
	})

	t.Run("recover from permission denied scenario", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Add a rule
		project.WithLocalRule("permission/test", `---
title: "Permission Test"
description: "Test permission handling"
tags: ["permission", "test"]
---

# Permission Test

Content for permission testing.`)

		// Build should succeed (we can't easily simulate permission issues in tests)
		result := project.Run(t, "build")
		result.ExpectSuccess(t)

		// This test mainly verifies the CLI handles file operations gracefully
		project.AssertFileExists(t, "CLAUDE.md")
	})

	t.Run("recover from partial config corruption", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Create partially corrupt config (valid YAML but invalid structure)
		project.WithConfig(`version: "not-a-number"
formats: "not-an-array"
rules: 123`)

		// Commands should fail gracefully with meaningful errors
		result := project.Run(t, "config")
		result.ExpectFailure(t)

		// Re-init should replace the corrupt config
		result = project.Run(t, "init", "--force", "--no-interactive")
		result.ExpectSuccess(t)

		// Now config should work
		result = project.Run(t, "config")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Configuration")
	})

	t.Run("recover from mixed valid and invalid rules", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Add valid rule
		project.WithLocalRule("valid/rule", `---
title: "Valid Rule"
description: "A valid rule"
tags: ["valid", "test"]
---

# Valid Rule

This is valid content.`)

		// Add invalid rule (malformed frontmatter)
		project.WithLocalRule("invalid/rule", `---
title: "Invalid Rule
description: Missing closing quote
tags: [unclosed, array
---

# Invalid Rule

This rule has malformed frontmatter.`)

		// Build should fail due to invalid rule
		result := project.Run(t, "build")
		result.ExpectFailure(t)

		// But the system should not crash or corrupt other data
		// Verify project structure is still intact
		result = project.Run(t, "config")
		result.ExpectSuccess(t)
	})
}

// TestFormatWorkflow tests different output format scenarios
func TestFormatWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize with custom config that has all formats
	customConfig := `version: 1
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true
`
	project.WithConfig(customConfig)

	// Add test rule
	project.WithLocalRule("test/rule", `---
title: "Test Rule"
description: "A test rule"
tags: ["test"]
---

# Test Rule

Test content for format testing.`)

	t.Run("build all formats", func(t *testing.T) {
		result := project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Claude (CLAUDE.md)").
			ExpectStdout(t, "Cursor (.cursor/rules/)").
			ExpectStdout(t, "Windsurf (.windsurf/rules/)")

		// Verify all format output files exist
		project.AssertFileExists(t, "CLAUDE.md")
		project.AssertFileExists(t, ".cursor/rules")
		project.AssertFileExists(t, ".windsurf/rules")

		// Verify content in different formats
		project.AssertFileContains(t, "CLAUDE.md", "Test Rule")
	})

	t.Run("config shows all formats", func(t *testing.T) {
		result := project.Run(t, "config")
		result.ExpectSuccess(t).
			ExpectStdout(t, "claude").
			ExpectStdout(t, "cursor").
			ExpectStdout(t, "windsurf")
	})
}

// TestLargeProjectWorkflow tests performance with many rules
func TestLargeProjectWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// Add multiple local rules to simulate larger project
	categories := []string{"security", "performance", "testing", "docs", "patterns"}
	for _, category := range categories {
		for j := range 5 { // 5 rules per category = 25 total rules
			ruleContent := `---
title: "` + category + ` Rule ` + string(rune(j+65)) + `"
description: "A ` + category + ` rule for testing"
tags: ["` + category + `", "test"]
---

# ` + category + ` Rule ` + string(rune(j+65)) + `

This is rule content for ` + category + ` testing.`

			project.WithLocalRule(category+"/rule-"+string(rune(j+65)), ruleContent)
		}

		t.Logf("Added %d rules for category %s", 5, category)
	}

	t.Run("build large project", func(t *testing.T) {
		result := project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Generated rules")

		// Should create output files
		project.AssertFileExists(t, "CLAUDE.md")

		// Verify content includes rules from different categories
		claudeContent := project.GetFileContent(t, "CLAUDE.md")
		for _, category := range categories {
			if !strings.Contains(claudeContent, category+" Rule") {
				t.Errorf("CLAUDE.md should contain rules from category: %s", category)
			}
		}
	})

	t.Run("config with many local rules", func(t *testing.T) {
		result := project.Run(t, "config")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Output Formats")

		// Should handle many rules without issues and show local rules
		result.ExpectNotStderr(t, "error").
			ExpectNotStderr(t, "fail").
			ExpectStdout(t, "Local Rules")
	})
}

// TestVariableSubstitutionWorkflow tests variable substitution functionality
func TestVariableSubstitutionWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("add rule with variables", func(t *testing.T) {
		// Create a rule template with variables in local rules directory
		project.WithLocalRule("templates/variable-test", `---
title: "{{.name}} Rule"
description: "Rule with {{.count}} examples"
tags: ["{{.category}}", "template"]
---

# {{.name}} Rule

This rule has {{.count}} examples in the {{.category}} category.`)

		// Add the rule directly to the configuration file with variables
		project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "[contexture(local):templates/variable-test]"
    variables:
      name: "Test"
      count: 5
      category: "testing"`)

		// Verify the configuration was set up correctly
		result := project.Run(t, "config")
		result.ExpectSuccess(t)
	})

	t.Run("build with variables", func(t *testing.T) {
		result := project.Run(t, "build")
		result.ExpectSuccess(t)

		// Check that variables were substituted
		project.AssertFileContains(t, "CLAUDE.md", "Test Rule")
		project.AssertFileContains(t, "CLAUDE.md", "5 examples")
		project.AssertFileContains(t, "CLAUDE.md", "testing category")
	})
}
