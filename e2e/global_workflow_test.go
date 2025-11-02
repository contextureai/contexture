// Package e2e provides end-to-end workflow tests
// This file contains E2E tests for global rules functionality
package e2e

import (
	"strings"
	"testing"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
)

const testGoContextRule = "@contexture/languages/go/context"

// TestGlobalRuleWorkflow tests the complete global rules workflow
func TestGlobalRuleWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()

	// Create test project with custom home directory
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Step 1: Initialize project
	t.Log("Step 1: Initialize project")
	result := project.Run(t, "init", "--force", "--no-interactive")
	result.ExpectSuccess(t).ExpectStdout(t, "Configuration generated successfully")

	// Step 2: Add global rule using a real rule from contexture repository
	t.Log("Step 2: Add global rule")
	ruleID := testGoContextRule

	result = project.Run(t, "rules", "add", "-g", ruleID)
	result.ExpectSuccess(t)

	// Step 3: List rules - should show global rule
	t.Log("Step 3: List rules - verify global rule appears")
	result = project.Run(t, "rules", "list")
	result.ExpectSuccess(t).
		ExpectStdout(t, ruleID)

	// Step 4: Add project rule with same ID (override)
	t.Log("Step 4: Add project rule to override global")
	result = project.Run(t, "rules", "add", ruleID)
	result.ExpectSuccess(t)

	// Step 5: List again - should still show the rule
	t.Log("Step 5: List rules - verify rule still appears")
	result = project.Run(t, "rules", "list")
	result.ExpectSuccess(t).
		ExpectStdout(t, ruleID)

	// Step 6: Build should use project version
	t.Log("Step 6: Build with project override")
	result = project.Run(t, "build")
	result.ExpectSuccess(t)

	// Verify file exists
	project.AssertFileExists(t, "CLAUDE.md")

	// Step 7: Remove project rule
	t.Log("Step 7: Remove project rule")
	result = project.Run(t, "rules", "remove", ruleID)
	result.ExpectSuccess(t)

	// Step 8: List should show global rule again
	t.Log("Step 8: List rules - verify global rule still appears")
	result = project.Run(t, "rules", "list")
	result.ExpectSuccess(t).
		ExpectStdout(t, ruleID)
}

// TestGlobalConfigCommands tests all global config-related commands
func TestGlobalConfigCommands(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	ruleID := testGoContextRule

	t.Run("rules add -g creates global config", func(t *testing.T) {
		result := project.Run(t, "rules", "add", "-g", ruleID)
		result.ExpectSuccess(t)
	})

	t.Run("rules list shows global rules", func(t *testing.T) {
		result := project.Run(t, "rules", "list")
		result.ExpectSuccess(t).
			ExpectStdout(t, ruleID)
	})

	t.Run("rules remove -g removes from global config", func(t *testing.T) {
		result := project.Run(t, "rules", "remove", "-g", ruleID)
		result.ExpectSuccess(t)

		// List should not show the rule anymore
		result = project.Run(t, "rules", "list")
		result.ExpectSuccess(t)

		// Should not contain the removed rule
		if strings.Contains(result.Stdout, ruleID) {
			t.Errorf("Removed global rule should not appear in list")
		}
	})

	t.Run("config -g shows global configuration", func(t *testing.T) {
		// First add a rule back so we have something in global config
		project.Run(t, "rules", "add", "-g", ruleID).ExpectSuccess(t)

		result := project.Run(t, "config", "-g")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Global Configuration").
			ExpectStdout(t, "Rules")
	})
}

// TestGlobalAndProjectRules tests interaction between global and project rules
func TestGlobalAndProjectRules(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	globalRule := testGoContextRule

	// Add global rule
	t.Log("Add rule to global config")
	project.Run(t, "rules", "add", "-g", globalRule).ExpectSuccess(t)

	// List should show global rule
	t.Log("List should show global rule")
	result := project.Run(t, "rules", "list")
	result.ExpectSuccess(t).
		ExpectStdout(t, globalRule)

	// Build should succeed with global rule
	// Global rules go to ~/.claude/CLAUDE.md, not project CLAUDE.md
	t.Log("Build should succeed with global rule")
	result = project.Run(t, "build")
	result.ExpectSuccess(t)
}

// TestGlobalProvidersWorkflow tests adding providers to global config
func TestGlobalProvidersWorkflow(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("providers add -g adds to global config", func(t *testing.T) {
		result := project.Run(t, "providers", "add", "-g", "testprovider", "https://github.com/test/rules.git")
		result.ExpectSuccess(t)
	})

	t.Run("providers list shows global providers", func(t *testing.T) {
		result := project.Run(t, "providers", "list")
		result.ExpectSuccess(t).
			ExpectStdout(t, "testprovider")
	})

	t.Run("providers remove -g removes from global config", func(t *testing.T) {
		result := project.Run(t, "providers", "remove", "-g", "testprovider")
		result.ExpectSuccess(t)

		// Verify it's removed
		result = project.Run(t, "providers", "list")
		result.ExpectSuccess(t)

		// Should not contain removed provider
		if strings.Contains(result.Stdout, "testprovider") && !strings.Contains(result.Stdout, "No providers") {
			t.Errorf("Removed provider should not appear in list")
		}
	})
}

// TestGlobalConfigLazyInitialization tests that global config is created on first use
func TestGlobalConfigLazyInitialization(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// First global operation should create config
	t.Log("First -g operation should create global config")
	result := project.Run(t, "rules", "add", "-g", "@contexture/languages/go/context")
	result.ExpectSuccess(t)

	// config -g should now work
	result = project.Run(t, "config", "-g")
	result.ExpectSuccess(t).
		ExpectStdout(t, "Global Configuration")
}

// TestGlobalConfigWithVariables tests rules with variables in global config
func TestGlobalConfigWithVariables(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// Add to global with variable
	t.Log("Add rule with variable to global config")
	result := project.Run(t, "rules", "add", "-g", "@contexture/languages/go/context", "--var", "testvar=global-value")
	result.ExpectSuccess(t)

	// Build should succeed with global variable
	// Global rules go to ~/.claude/CLAUDE.md, not project CLAUDE.md
	result = project.Run(t, "build")
	result.ExpectSuccess(t)

	// Override in project with different variable
	t.Log("Override with project rule using different variable")
	result = project.Run(t, "rules", "add", "@contexture/languages/go/context", "--var", "testvar=project-value")
	result.ExpectSuccess(t)

	// Build should succeed with project variable override
	// Now CLAUDE.md should exist because we added a project rule
	result = project.Run(t, "build")
	result.ExpectSuccess(t)
	project.AssertFileExists(t, "CLAUDE.md")
}

// TestBuildCommandMergesGlobalAndProject tests that build handles global rules correctly
func TestBuildCommandMergesGlobalAndProject(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// Add global rule
	project.Run(t, "rules", "add", "-g", "@contexture/languages/go/context").ExpectSuccess(t)

	// Build should succeed with global rule
	result := project.Run(t, "build")
	result.ExpectSuccess(t)

	// Global rules should NOT appear in project CLAUDE.md
	// They should only go to ~/.claude/CLAUDE.md
	// Note: We can't easily test ~/.claude/CLAUDE.md in this test environment
	// so we just verify the build succeeded without error
}
