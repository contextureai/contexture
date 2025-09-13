// Package e2e provides integration tests for multi-component scenarios
package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
)

// TestFullApplicationIntegration tests complete application workflows
func TestFullApplicationIntegration(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Complete workflow: init → add rules → configure → build → verify
	t.Run("complete application lifecycle", func(t *testing.T) {
		// Step 1: Initialize project
		result := project.Run(t, "init", "--force", "--no-interactive")
		result.ExpectSuccess(t).ExpectStdout(t, "Configuration generated successfully")

		// Step 2: Add multiple formats
		project.Run(t, "config", "formats", "add", "cursor").ExpectSuccess(t)
		project.Run(t, "config", "formats", "add", "windsurf").ExpectSuccess(t)

		// Step 3: Add various types of rules
		// Local rules with different categories
		categories := []string{"security", "performance", "testing", "documentation"}
		for i, category := range categories {
			ruleContent := `---
title: "` + category + ` Best Practices"
description: "Best practices for ` + category + `"
tags: ["` + category + `", "best-practices", "integration-test"]
languages: ["javascript", "typescript", "go"]
frameworks: ["react", "express", "gin"]
---

# ` + category + ` Best Practices

This rule covers ` + category + ` best practices for integration testing.

## Key Points

1. Always follow ` + category + ` guidelines
2. Test thoroughly
3. Document your approach

## Examples

` + "```javascript" + `
// Example ` + category + ` implementation
function example` + category + `() {
    console.log("` + category + ` implementation");
}
` + "```" + `

## References

- Best practices guide
- Security checklist
- Performance metrics`

			project.WithLocalRule(category+"/best-practices", ruleContent)
			t.Logf("Added %s rule (%d/%d)", category, i+1, len(categories))
		}

		// Step 4: Build with all formats
		result = project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Claude (CLAUDE.md)").
			ExpectStdout(t, "Cursor (.cursor/rules/)").
			ExpectStdout(t, "Windsurf (.windsurf/rules/)")

		// Step 5: Verify all outputs exist and contain expected content
		// CLAUDE.md is a single file
		project.AssertFileExists(t, "CLAUDE.md")
		claudeContent := project.GetFileContent(t, "CLAUDE.md")
		expectedItems := []string{"security Best Practices", "performance Best Practices", "testing Best Practices", "documentation Best Practices"}
		for _, expected := range expectedItems {
			if !strings.Contains(claudeContent, expected) {
				t.Errorf("CLAUDE.md missing expected content: %s", expected)
			}
		}

		// Cursor and Windsurf formats create directories with individual rule files
		project.AssertFileExists(t, ".cursor/rules")
		cursorContent := project.GetDirectoryContent(t, ".cursor/rules")
		for _, expected := range []string{"security", "performance", "testing", "documentation"} {
			if !strings.Contains(cursorContent, expected) {
				t.Errorf("Cursor rules missing expected content: %s", expected)
			}
		}

		project.AssertFileExists(t, ".windsurf/rules")
		windsurfContent := project.GetDirectoryContent(t, ".windsurf/rules")
		for _, expected := range []string{"security", "performance", "testing", "documentation"} {
			if !strings.Contains(windsurfContent, expected) {
				t.Errorf("Windsurf rules missing expected content: %s", expected)
			}
		}

		// Step 6: Verify configuration shows complete project state
		result = project.Run(t, "config")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Configuration").
			ExpectStdout(t, "claude").
			ExpectStdout(t, "cursor").
			ExpectStdout(t, "windsurf").
			ExpectStdout(t, "Local Rules")

		t.Log("Complete application lifecycle test passed")
	})
}

// TestCrossComponentIntegration tests interaction between different components
func TestCrossComponentIntegration(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	t.Run("config and rules integration", func(t *testing.T) {
		// Initialize project
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Test config formats integration
		project.Run(t, "config", "formats", "add", "cursor").ExpectSuccess(t)
		project.Run(t, "config", "formats", "enable", "cursor").ExpectSuccess(t)

		// Verify format is now in config
		result := project.Run(t, "config", "formats", "list")
		result.ExpectSuccess(t).ExpectStdout(t, "cursor")

		// Add rule and verify it works with the new format
		project.WithLocalRule("integration/test", `---
title: "Integration Test Rule"
description: "Testing cross-component integration"
tags: ["integration", "test"]
---

# Integration Test

Cross-component integration testing.`)

		// Build should use both formats
		result = project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Claude").
			ExpectStdout(t, "Cursor")

		// Verify both outputs exist
		project.AssertFileExists(t, "CLAUDE.md")
		project.AssertFileExists(t, ".cursor/rules")
	})

	t.Run("rules and build integration", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Test rules ls shows no rules initially
		result := project.Run(t, "rules", "ls")
		result.ExpectSuccess(t).ExpectStdout(t, "No rules found")

		// Add local rules
		project.WithLocalRule("test1/rule", `---
title: "Test Rule 1"
description: "First test rule"
tags: ["test1"]
---

# Test Rule 1

Content for test rule 1.`)

		project.WithLocalRule("test2/rule", `---
title: "Test Rule 2"
description: "Second test rule"
tags: ["test2"]
---

# Test Rule 2

Content for test rule 2.`)

		// Build should include both rules
		result = project.Run(t, "build")
		result.ExpectSuccess(t)

		// Verify both rules are in output
		content := project.GetFileContent(t, "CLAUDE.md")
		if !strings.Contains(content, "Test Rule 1") {
			t.Error("Build output missing Test Rule 1")
		}
		if !strings.Contains(content, "Test Rule 2") {
			t.Error("Build output missing Test Rule 2")
		}
	})
}

// TestMultiFormatIntegration tests behavior with multiple output formats
func TestMultiFormatIntegration(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	t.Run("all formats with complex rules", func(t *testing.T) {
		// Initialize with all formats
		project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true`)

		// Add complex rule with all metadata fields
		complexRule := `---
title: "Complex Integration Rule"
description: "A comprehensive rule testing all metadata fields"
tags: ["complex", "integration", "comprehensive", "metadata"]
languages: ["javascript", "typescript", "python", "go", "rust"]
frameworks: ["react", "vue", "angular", "express", "fastapi", "gin", "actix"]
trigger:
  type: "glob"
  globs: ["*.js", "*.ts", "*.py", "*.go", "*.rs"]
---

# Complex Integration Rule

This rule tests integration across all supported formats with comprehensive metadata.

## Supported Languages

The rule applies to:
- JavaScript/TypeScript for frontend development
- Python for backend and data processing
- Go for systems programming
- Rust for performance-critical applications

## Framework Integration

Works with popular frameworks:
- **Frontend**: React, Vue, Angular
- **Backend**: Express, FastAPI, Gin, Actix

## Code Examples

` + "```javascript" + `
// JavaScript example
function complexIntegration() {
    return "Integration testing with JavaScript";
}
` + "```" + `

` + "```python" + `
# Python example  
def complex_integration():
    return "Integration testing with Python"
` + "```" + `

` + "```go" + `
// Go example
func ComplexIntegration() string {
    return "Integration testing with Go"
}
` + "```" + `

## Trigger Conditions

This rule is triggered by files matching:
- *.js, *.ts (JavaScript/TypeScript)
- *.py (Python)
- *.go (Go)
- *.rs (Rust)

## Integration Testing Notes

- Tests cross-format compatibility
- Validates metadata processing
- Ensures trigger handling works
- Verifies content rendering across formats`

		project.WithLocalRule("integration/complex-rule", complexRule)

		// Build should process the complex rule for all formats
		result := project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Claude").
			ExpectStdout(t, "Cursor").
			ExpectStdout(t, "Windsurf")

		// Verify all formats contain the complex rule content
		expectedElements := []string{
			"Complex Integration Rule",
			"comprehensive rule testing",
			"JavaScript/TypeScript",
			"Integration testing with",
		}

		// Check CLAUDE.md (single file)
		project.AssertFileExists(t, "CLAUDE.md")
		claudeContent := project.GetFileContent(t, "CLAUDE.md")
		for _, element := range expectedElements {
			if !strings.Contains(claudeContent, element) {
				t.Errorf("CLAUDE.md missing expected element: %s", element)
			}
		}

		// Check .cursor/rules (directory)
		project.AssertFileExists(t, ".cursor/rules")
		cursorContent := project.GetDirectoryContent(t, ".cursor/rules")
		for _, element := range expectedElements {
			if !strings.Contains(cursorContent, element) {
				t.Errorf("Cursor rules missing expected element: %s", element)
			}
		}

		// Check .windsurf/rules (directory)
		project.AssertFileExists(t, ".windsurf/rules")
		windsurfContent := project.GetDirectoryContent(t, ".windsurf/rules")
		for _, element := range expectedElements {
			if !strings.Contains(windsurfContent, element) {
				t.Errorf("Windsurf rules missing expected element: %s", element)
			}
		}
	})
}

// TestErrorStateIntegration tests error handling across components
func TestErrorStateIntegration(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("cascading error handling", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Start with invalid config
		project.WithConfig("invalid: yaml: structure")

		// All commands should fail gracefully
		commands := [][]string{
			{"config"},
			{"build"},
			{"rules", "ls"},
			{"config", "formats", "list"},
		}

		for _, cmd := range commands {
			result := project.Run(t, cmd...)
			result.ExpectFailure(t)
			// Should not crash or hang
		}

		// Recovery should work
		result := project.Run(t, "init", "--force", "--no-interactive")
		result.ExpectSuccess(t)

		// Now all commands should work
		for _, cmd := range commands {
			// Skip rules ls as it may need TTY
			if len(cmd) >= 2 && cmd[0] == "rules" && cmd[1] == "ls" {
				continue
			}
			result := project.Run(t, cmd...)
			result.ExpectSuccess(t)
		}
	})
}

// TestPerformanceIntegration tests performance with realistic loads
func TestPerformanceIntegration(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	t.Run("large project performance", func(t *testing.T) {
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Add all three formats
		project.Run(t, "config", "formats", "add", "cursor").ExpectSuccess(t)
		project.Run(t, "config", "formats", "add", "windsurf").ExpectSuccess(t)

		// Generate many rules (50 total across 5 categories)
		categories := []string{"security", "performance", "testing", "documentation", "architecture"}

		start := time.Now()
		for _, category := range categories {
			for i := range 10 { // 10 rules per category
				ruleContent := `---
title: "` + category + ` Rule ` + string(rune(i+'A')) + `"
description: "Rule ` + string(rune(i+'A')) + ` for ` + category + ` category"
tags: ["` + category + `", "rule-` + string(rune(i+'A')) + `", "performance-test"]
languages: ["javascript", "typescript", "go"]
---

# ` + category + ` Rule ` + string(rune(i+'A')) + `

This is performance testing rule ` + string(rune(i+'A')) + ` in the ` + category + ` category.

## Implementation

Content for rule ` + string(rune(i+'A')) + ` with sufficient detail for testing.

` + "```javascript" + `
function performanceTest` + string(rune(i+'A')) + `() {
    return "` + category + ` rule ` + string(rune(i+'A')) + ` implementation";
}
` + "```" + `

## Notes

This rule is part of the performance testing suite for large projects.`

				project.WithLocalRule(category+"/rule-"+string(rune(i+'A')), ruleContent)
			}
		}

		ruleCreationTime := time.Since(start)
		t.Logf("Created 50 rules in %v", ruleCreationTime)

		// Build all formats - should complete within reasonable time
		buildStart := time.Now()
		result := project.Run(t, "build")
		buildTime := time.Since(buildStart)

		result.ExpectSuccess(t)
		t.Logf("Built all formats in %v", buildTime)

		// Verify performance is acceptable (less than 30 seconds for 50 rules)
		if buildTime > 30*time.Second {
			t.Errorf("Build took too long: %v (expected < 30s)", buildTime)
		}

		// Verify all outputs exist and have reasonable sizes
		// Check CLAUDE.md (single file)
		project.AssertFileExists(t, "CLAUDE.md")
		claudeContent := project.GetFileContent(t, "CLAUDE.md")
		for _, category := range categories {
			if !strings.Contains(claudeContent, category+" Rule") {
				t.Errorf("CLAUDE.md missing content from category %s", category)
			}
		}
		if len(claudeContent) < 20*1024 { // 20KB minimum for 50 rules
			t.Errorf("CLAUDE.md seems too small: %d bytes", len(claudeContent))
		}

		// Check .cursor/rules (directory)
		project.AssertFileExists(t, ".cursor/rules")
		cursorContent := project.GetDirectoryContent(t, ".cursor/rules")
		for _, category := range categories {
			if !strings.Contains(cursorContent, category+" Rule") {
				t.Errorf("Cursor rules missing content from category %s", category)
			}
		}
		if len(cursorContent) < 20*1024 { // 20KB minimum for 50 rules
			t.Errorf("Cursor rules seem too small: %d bytes", len(cursorContent))
		}

		// Check .windsurf/rules (directory)
		project.AssertFileExists(t, ".windsurf/rules")
		windsurfContent := project.GetDirectoryContent(t, ".windsurf/rules")
		for _, category := range categories {
			if !strings.Contains(windsurfContent, category+" Rule") {
				t.Errorf("Windsurf rules missing content from category %s", category)
			}
		}
		if len(windsurfContent) < 20*1024 { // 20KB minimum for 50 rules
			t.Errorf("Windsurf rules seem too small: %d bytes", len(windsurfContent))
		}
	})
}

// TestConfigurationIntegration tests various configuration scenarios
func TestConfigurationIntegration(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	t.Run("format configuration persistence", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Add and enable formats
		formats := []string{"cursor", "windsurf"}
		for _, format := range formats {
			project.Run(t, "config", "formats", "add", format).ExpectSuccess(t)
			project.Run(t, "config", "formats", "enable", format).ExpectSuccess(t)
		}

		// Verify formats are persisted
		result := project.Run(t, "config", "formats", "list")
		result.ExpectSuccess(t)
		for _, format := range formats {
			result.ExpectStdout(t, format)
		}

		// Build should use all formats
		project.WithLocalRule("config/test", `---
title: "Config Test Rule"
description: "Testing configuration persistence"
tags: ["config", "test"]
---

# Config Test

Testing configuration persistence.`)

		result = project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Claude").
			ExpectStdout(t, "Cursor").
			ExpectStdout(t, "Windsurf")
	})
}
