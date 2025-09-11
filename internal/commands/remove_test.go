// Package commands provides CLI command implementations
package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestNewRemoveCommand(t *testing.T) {
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewRemoveCommand(deps)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
}

func TestRemoveAction(t *testing.T) {
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	// Create a context with empty arguments to simulate CLI with no args
	ctx := context.Background()

	// Test with no arguments (should show interactive mode but fail due to no config)
	app := &cli.Command{
		Name: "test",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return RemoveAction(ctx, cmd, deps)
		},
	}

	err := app.Run(ctx, []string{"test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no project configuration found")
}

func TestRemoveCommand_Execute_NoConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	tempDir := "/tmp/test-remove"
	_ = fs.MkdirAll(tempDir, 0o755)

	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewRemoveCommand(deps)

	// Create mock CLI command
	cliCmd := &cli.Command{}

	// Test with no project configuration (should fail)
	err := cmd.Execute(context.Background(), cliCmd, []string{"[contexture:test/rule]"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no project configuration found")
}

func TestRemoveCommand_CustomSourceRules(t *testing.T) {
	// Create temporary directory using real filesystem for this test
	tempDir := t.TempDir()

	// Use real filesystem for this test since we need actual directory operations
	realFS := afero.NewOsFs()
	deps := &dependencies.Dependencies{
		FS:      realFS,
		Context: context.Background(),
	}

	// Create a test configuration with both standard and custom source rules
	configContent := `version: 1
formats:
    - type: claude
      enabled: true
rules:
    - id: '[contexture:languages/go/basics]'
      commitHash: abc123
    - id: '[contexture(git@github.com:user/custom-repo.git):test/custom-rule]'
      source: git@github.com:user/custom-repo.git
      commitHash: def456
    - id: '[contexture(https://github.com/org/rules.git):security/auth,v1.0]'
      source: https://github.com/org/rules.git
      commitHash: ghi789
`
	configPath := filepath.Join(tempDir, ".contexture.yaml")
	err := afero.WriteFile(realFS, configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Save current directory and restore after test
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	//nolint:usetesting // Need to restore original dir at end of test
	defer func() { _ = os.Chdir(originalWd) }()

	cmd := NewRemoveCommand(deps)
	cliCmd := &cli.Command{}

	// Test removing a custom source rule with SSH URL
	t.Run("removes custom SSH source rule", func(t *testing.T) {
		// Change to temp directory for this test
		t.Chdir(tempDir)

		err = cmd.Execute(context.Background(), cliCmd, []string{"[contexture(git@github.com:user/custom-repo.git):test/custom-rule]"})
		require.NoError(t, err)

		// Verify rule was removed from config
		updatedConfig, err := afero.ReadFile(realFS, configPath)
		require.NoError(t, err)
		assert.NotContains(t, string(updatedConfig), "git@github.com:user/custom-repo.git")
		assert.Contains(t, string(updatedConfig), "[contexture:languages/go/basics]") // Other rules should remain
		assert.Contains(t, string(updatedConfig), "https://github.com/org/rules.git") // Other rules should remain

		// Restore config for next test
		err = afero.WriteFile(realFS, configPath, []byte(configContent), 0o644)
		require.NoError(t, err)
	})

	// Test removing a custom source rule with HTTPS URL and branch
	t.Run("removes custom HTTPS source rule with branch", func(t *testing.T) {
		t.Chdir(tempDir)

		err = cmd.Execute(context.Background(), cliCmd, []string{"[contexture(https://github.com/org/rules.git):security/auth,v1.0]"})
		require.NoError(t, err)

		// Verify rule was removed from config
		updatedConfig, err := afero.ReadFile(realFS, configPath)
		require.NoError(t, err)
		assert.NotContains(t, string(updatedConfig), "https://github.com/org/rules.git")
		assert.Contains(t, string(updatedConfig), "[contexture:languages/go/basics]")    // Other rules should remain
		assert.Contains(t, string(updatedConfig), "git@github.com:user/custom-repo.git") // Other rules should remain

		// Restore config for next test
		err = afero.WriteFile(realFS, configPath, []byte(configContent), 0o644)
		require.NoError(t, err)
	})

	// Test removing non-existent custom source rule
	t.Run("handles non-existent custom source rule", func(t *testing.T) {
		t.Chdir(tempDir)

		err = cmd.Execute(context.Background(), cliCmd, []string{"[contexture(git@github.com:nonexistent/repo.git):missing/rule]"})
		require.NoError(t, err) // Should not error, just log warning

		// Verify no rules were removed
		updatedConfig, err := afero.ReadFile(realFS, configPath)
		require.NoError(t, err)
		assert.Contains(t, string(updatedConfig), "[contexture:languages/go/basics]")
		assert.Contains(t, string(updatedConfig), "git@github.com:user/custom-repo.git")
	})
}

func TestRemoveCommand_RuleIDMatching(t *testing.T) {
	// Create temporary directory using real filesystem for this test
	tempDir := t.TempDir()

	// Use real filesystem for this test since we need actual directory operations
	realFS := afero.NewOsFs()
	deps := &dependencies.Dependencies{
		FS:      realFS,
		Context: context.Background(),
	}

	// Create configuration with various rule formats
	configContent := `version: 1
formats:
    - type: claude
      enabled: true
rules:
    - id: '[contexture:simple/rule]'
      commitHash: abc123
    - id: '[contexture(default-repo):another/rule]'
      commitHash: def456
    - id: '[contexture(git@custom.com:user/repo.git):custom/rule]'
      source: git@custom.com:user/repo.git
      commitHash: ghi789
`
	configPath := filepath.Join(tempDir, ".contexture.yaml")
	err := afero.WriteFile(realFS, configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Save current directory and restore after test
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	//nolint:usetesting // Need to restore original dir at end of test
	defer func() { _ = os.Chdir(originalWd) }()

	removeCmd := NewRemoveCommand(deps)
	cliCmd := &cli.Command{}

	tests := []struct {
		name            string
		ruleIDToRemove  string
		shouldBeRemoved string
		shouldRemain    []string
	}{
		{
			name:            "remove standard contexture rule",
			ruleIDToRemove:  "[contexture:simple/rule]",
			shouldBeRemoved: "simple/rule",
			shouldRemain:    []string{"another/rule", "custom/rule"},
		},
		{
			name:            "remove rule with default repo reference",
			ruleIDToRemove:  "[contexture(default-repo):another/rule]",
			shouldBeRemoved: "another/rule",
			shouldRemain:    []string{"simple/rule", "custom/rule"},
		},
		{
			name:            "remove custom source rule",
			ruleIDToRemove:  "[contexture(git@custom.com:user/repo.git):custom/rule]",
			shouldBeRemoved: "custom/rule",
			shouldRemain:    []string{"simple/rule", "another/rule"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Restore config for each test
			err := afero.WriteFile(realFS, configPath, []byte(configContent), 0o644)
			require.NoError(t, err)

			// Change to temp directory for this test
			t.Chdir(tempDir)

			err = removeCmd.Execute(context.Background(), cliCmd, []string{tt.ruleIDToRemove})
			require.NoError(t, err)

			// Verify correct rule was removed and others remain
			updatedConfig, err := afero.ReadFile(realFS, configPath)
			require.NoError(t, err)
			configStr := string(updatedConfig)

			assert.NotContains(t, configStr, tt.shouldBeRemoved)
			for _, remaining := range tt.shouldRemain {
				assert.Contains(t, configStr, remaining)
			}
		})
	}
}
