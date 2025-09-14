// Package commands provides CLI command implementations
package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestNewListCommand(t *testing.T) {
	deps := createTestDependencies()

	cmd := NewListCommand(deps)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
	assert.NotNil(t, cmd.ruleFetcher)
	assert.NotNil(t, cmd.registry)
}

func TestListAction(t *testing.T) {
	deps := createTestDependencies()

	app := createTestApp("test", func(ctx context.Context, cmd *cli.Command) error {
		return ListAction(ctx, cmd, deps)
	})

	err := runTestApp(app)
	assertNoProjectConfigError(t, err)
}

func TestListCommand_Execute_NoConfig(t *testing.T) {
	testCommandExecuteNoConfig(t, "list", func(ctx context.Context, cliCmd *cli.Command) error {
		deps := createTestDependencies()
		cmd := NewListCommand(deps)
		return cmd.Execute(ctx, cliCmd)
	})
}

// TestListCommand_SpecificBehavior tests list-specific functionality
func TestListCommand_SpecificBehavior(t *testing.T) {
	// Test list-specific behavior here if needed
	deps := createTestDependencies()
	cmd := NewListCommand(deps)
	assert.NotNil(t, cmd.ruleFetcher, "ListCommand should have ruleFetcher")
}

// Integration tests for list command output are covered by:
// 1. Unit tests in internal/ui/rules/display_test.go for output formatting
// 2. Existing command structure tests in this file
// 3. E2E tests in the e2e/ directory for full command workflows
