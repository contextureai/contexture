// Package commands provides CLI command implementations
package commands

import (
	"context"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// createTestDependencies creates test dependencies with a memory filesystem
func createTestDependencies() *dependencies.Dependencies {
	fs := afero.NewMemMapFs()
	return &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}
}

// createTestApp creates a test CLI app that executes the given action
func createTestApp(name string, action func(context.Context, *cli.Command) error) *cli.Command {
	return &cli.Command{
		Name:   name,
		Action: action,
	}
}

// runTestApp runs a test app and returns the error
func runTestApp(app *cli.Command) error {
	ctx := context.Background()
	return app.Run(ctx, []string{app.Name})
}

// assertNoProjectConfigError asserts that the error indicates no project configuration
func assertNoProjectConfigError(t *testing.T, err error) {
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no project configuration found")
}

// testCommandExecuteNoConfig tests a command's Execute method with no project configuration
func testCommandExecuteNoConfig(t *testing.T, commandName string, executeFunc func(context.Context, *cli.Command) error) {
	deps := createTestDependencies()
	tempDir := "/tmp/test-" + commandName
	_ = deps.FS.MkdirAll(tempDir, 0o755)

	// Create mock CLI command
	cliCmd := &cli.Command{}

	// Test with no project configuration (should fail)
	err := executeFunc(context.Background(), cliCmd)
	assertNoProjectConfigError(t, err)
}
