// Package e2e provides end-to-end tests for the Contexture CLI
package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

const (
	// Binary path - assumes binary is built in bin/ directory
	binaryPath = "./bin/contexture"
	// Environment variable values
	envTrue = "true"
)

// TestCLIBasics tests basic CLI functionality
func TestCLIBasics(t *testing.T) {
	// Resolve binary path like NewTestProject does
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// If we're in the e2e directory, go up one level to project root
	if filepath.Base(cwd) == "e2e" {
		cwd = filepath.Dir(cwd)
	}

	var absBinaryPath string
	if filepath.IsAbs(binaryPath) {
		absBinaryPath = binaryPath
	} else {
		absBinaryPath = filepath.Join(cwd, binaryPath)
	}

	runner := helpers.NewCLIRunner(absBinaryPath).WithWorkDir(".")

	t.Run("help command", func(t *testing.T) {
		result := runner.Run(t, "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "AI assistant rule management").
			ExpectStdout(t, "Commands:")
	})

	t.Run("version command", func(t *testing.T) {
		result := runner.Run(t, "--version")
		result.ExpectSuccess(t).
			ExpectStdout(t, "contexture version")
	})

	t.Run("invalid command", func(t *testing.T) {
		result := runner.Run(t, "invalid-command")
		result.ExpectFailure(t).
			ExpectStderr(t, "No help topic for")
	})

	t.Run("no arguments shows help", func(t *testing.T) {
		result := runner.Run(t)
		result.ExpectSuccess(t).
			ExpectStdout(t, "AI assistant rule management")
	})
}

// TestInitCommand tests the init command functionality
func TestInitCommand(t *testing.T) {
	fs := afero.NewOsFs() // Use real filesystem for e2e tests
	project := helpers.NewTestProject(t, fs, binaryPath)

	t.Run("basic init", func(t *testing.T) {
		result := project.Run(t, "init", "--force", "--no-interactive")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Configuration generated successfully")

		// Verify config file was created
		project.AssertFileExists(t, ".contexture.yaml")
		project.AssertFileContains(t, ".contexture.yaml", "version:")
	})

	t.Run("init with default format", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		result := project.Run(t, "init", "--force", "--no-interactive")
		result.ExpectSuccess(t)

		// Check that default claude format is enabled
		content := project.GetFileContent(t, ".contexture.yaml")
		require.Contains(t, content, "claude")
		require.Contains(t, content, "enabled: true")
	})

	t.Run("init in existing project fails without force", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// First init should succeed
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Second init without force should fail
		result := project.Run(t, "init", "--no-interactive")
		result.ExpectFailure(t).
			ExpectStderr(t, "already exists")
	})

	t.Run("init with force overwrites existing", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		// Create initial config
		project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

		// Force init should succeed
		result := project.Run(t, "init", "--force", "--no-interactive")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Configuration generated successfully")
	})
}

// TestConfigCommand tests config command functionality
func TestConfigCommand(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project first
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("show config", func(t *testing.T) {
		result := project.Run(t, "config")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Configuration").
			ExpectStdout(t, "Output Formats")
	})

	t.Run("config help", func(t *testing.T) {
		result := project.Run(t, "config", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Show and manage project configuration")
	})

	t.Run("config in non-project directory", func(t *testing.T) {
		emptyProject := helpers.NewTestProject(t, fs, binaryPath)
		result := emptyProject.Run(t, "config")
		result.ExpectFailure(t).
			ExpectStderr(t, "no configuration file found")
	})
}

// TestConfigFormatsCommand tests config formats subcommand functionality
func TestConfigFormatsCommand(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project first
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("config formats list", func(t *testing.T) {
		result := project.Run(t, "config", "formats", "list")
		result.ExpectSuccess(t).
			ExpectStdout(t, "claude")
	})

	t.Run("config formats add", func(t *testing.T) {
		result := project.Run(t, "config", "formats", "add", "cursor")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Format added:")
	})

	t.Run("config formats enable", func(t *testing.T) {
		result := project.Run(t, "config", "formats", "enable", "cursor")
		result.ExpectSuccess(t).
			ExpectStdout(t, "already enabled")
	})

	t.Run("config formats disable", func(t *testing.T) {
		result := project.Run(t, "config", "formats", "disable", "cursor")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Format disabled:")
	})

	t.Run("config formats remove", func(t *testing.T) {
		result := project.Run(t, "config", "formats", "remove", "cursor")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Format removed:")
	})
}

// TestRulesCommand tests rules command functionality
func TestRulesCommand(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project first
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("rules help", func(t *testing.T) {
		result := project.Run(t, "rules", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Manage project rules").
			ExpectStdout(t, "Commands:")
	})

	t.Run("rules ls with no rules", func(t *testing.T) {
		// Rules ls now handles no TTY gracefully and shows a message
		result := project.Run(t, "rules", "ls")
		result.ExpectSuccess(t).
			ExpectStdout(t, "No rules found")
	})

	t.Run("rules add help", func(t *testing.T) {
		result := project.Run(t, "rules", "add", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Add rules to the project")
	})

	t.Run("rules rm help", func(t *testing.T) {
		result := project.Run(t, "rules", "rm", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Remove rules from the project")
	})

	t.Run("rules update help", func(t *testing.T) {
		result := project.Run(t, "rules", "update", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Update rules to latest versions")
	})

	t.Run("rules remove functionality", func(t *testing.T) {
		// Add a rule first
		project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "test/rule"
    source: "local"`)

		// Test removing the rule
		result := project.Run(t, "rules", "remove", "test/rule")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Rule removed successfully!")
	})

	t.Run("rules remove non-existent", func(t *testing.T) {
		result := project.Run(t, "rules", "remove", "non-existent/rule")
		// CLI treats remove as idempotent (success even if rule doesn't exist)
		result.ExpectSuccess(t)
	})

	t.Run("rules update with no rules", func(t *testing.T) {
		result := project.Run(t, "rules", "update", "--dry-run")
		result.ExpectSuccess(t).
			ExpectStdout(t, "No rules configured to update")
	})
}

// TestBuildCommand tests build command functionality
func TestBuildCommand(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project first
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("build with no rules", func(t *testing.T) {
		result := project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStderr(t, "No rules configured") // Shows info message instead of header

		// With no rules, no output files are created (expected behavior)
		// The build succeeds but doesn't generate any format files
		// Header is intentionally not shown when no rules exist
	})

	t.Run("build help", func(t *testing.T) {
		result := project.Run(t, "build", "--help")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Build output files")
	})

	t.Run("build in non-project directory", func(t *testing.T) {
		emptyProject := helpers.NewTestProject(t, fs, binaryPath)
		result := emptyProject.Run(t, "build")
		result.ExpectFailure(t).
			ExpectStderr(t, "config locate failed") // Current error message format
	})
}

// TestLocalRules tests local rules functionality
func TestLocalRules(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project and add local rule
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	localRuleContent := `---
title: "Test Local Rule"
description: "A test rule for local testing"
tags: ["test", "local"]
---

# Test Local Rule

This is a test rule content for local rules testing.`

	project.WithLocalRule("project/test-rule", localRuleContent)

	t.Run("build with local rules", func(t *testing.T) {
		result := project.Run(t, "build")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Generated rules")

		// Should include local rule in output
		project.AssertFileExists(t, "CLAUDE.md")
		project.AssertFileContains(t, "CLAUDE.md", "Test Local Rule")
	})

	t.Run("config shows project information", func(t *testing.T) {
		result := project.Run(t, "config")
		result.ExpectSuccess(t).
			ExpectStdout(t, "Output Formats").
			ExpectStdout(t, "Local Rules")
	})
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	fs := afero.NewOsFs()

	// Resolve binary path like NewTestProject does
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// If we're in the e2e directory, go up one level to project root
	if filepath.Base(cwd) == "e2e" {
		cwd = filepath.Dir(cwd)
	}

	var absBinaryPath string
	if filepath.IsAbs(binaryPath) {
		absBinaryPath = binaryPath
	} else {
		absBinaryPath = filepath.Join(cwd, binaryPath)
	}

	runner := helpers.NewCLIRunner(absBinaryPath).WithWorkDir(".")

	t.Run("commands in non-project directory", func(t *testing.T) {
		project := helpers.NewTestProject(t, fs, binaryPath)

		tests := []struct {
			name string
			args []string
		}{
			{"build", []string{"build"}},
			{"rules ls", []string{"rules", "ls"}},
			{"config", []string{"config"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := project.Run(t, tt.args...)
				result.ExpectFailure(t).
					ExpectStderr(t, "config locate failed")
			})
		}
	})

	t.Run("invalid flags", func(t *testing.T) {
		result := runner.Run(t, "init", "--invalid-flag")
		result.ExpectFailure(t)
	})

	t.Run("command timeout", func(t *testing.T) {
		// Test with very short timeout to simulate hanging command
		shortRunner := helpers.NewCLIRunner(binaryPath).WithTimeout(1)

		// This should timeout quickly due to TTY requirement
		result := shortRunner.Run(t, "rules", "ls")
		result.ExpectFailure(t)
	})
}
