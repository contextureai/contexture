// Package commands provides CLI command implementations
package commands

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// globalCliMutex protects access to global cli variables during tests
var globalCliMutex sync.Mutex

// createTestDependencies creates test dependencies with a memory filesystem
func createTestDependencies() *dependencies.Dependencies {
	return dependencies.NewForTesting(context.Background())
}

// createTestApp creates a test CLI app that executes the given action
func createTestApp(action func(context.Context, *cli.Command) error) *cli.Command {
	return &cli.Command{
		Name:   "test",
		Action: action,
	}
}

// runTestApp runs a test app and returns the error
func runTestApp(app *cli.Command) error {
	ctx := context.Background()

	// Capture error from ExitErrHandler
	var capturedErr error

	// Protect access to global cli variables with mutex
	globalCliMutex.Lock()

	// Save original handlers
	originalOsExiter := cli.OsExiter
	originalExitErrHandler := app.ExitErrHandler

	// Prevent os.Exit in tests
	cli.OsExiter = func(_ int) {
		// Don't exit, just track that exit was called
	}

	// Capture errors before exit
	app.ExitErrHandler = func(_ context.Context, _ *cli.Command, err error) {
		capturedErr = err
	}

	// Restore handlers after test
	defer func() {
		cli.OsExiter = originalOsExiter
		app.ExitErrHandler = originalExitErrHandler
		globalCliMutex.Unlock()
	}()

	// Run the app
	err := app.Run(ctx, []string{app.Name})
	// Return error from action or captured error from exit handler
	if err != nil {
		return err
	}
	return capturedErr
}

// assertNoProjectConfigError asserts that the error indicates no project configuration
func assertNoProjectConfigError(t *testing.T, err error) {
	require.Error(t, err)
	// Error message format can be either:
	// "load project configuration: ..." (old format)
	// "load configuration: load project config: ..." (new format with merged config)
	assert.True(t,
		strings.Contains(err.Error(), "load project configuration") ||
			strings.Contains(err.Error(), "load configuration"),
		"Error should mention configuration loading: %s", err.Error())
	assert.Contains(t, err.Error(), "no configuration file found")
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
