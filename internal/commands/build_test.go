// Package commands provides CLI command implementations
package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestNewBuildCommand(t *testing.T) {
	deps := createTestDependencies()

	cmd := NewBuildCommand(deps)

	// Build command should have all required components
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
	assert.NotNil(t, cmd.ruleGenerator, "Build command requires rule generator")
	assert.NotNil(t, cmd.registry)
}

func TestBuildAction(t *testing.T) {
	deps := createTestDependencies()

	app := createTestApp("test", func(ctx context.Context, cmd *cli.Command) error {
		return BuildAction(ctx, cmd, deps)
	})

	err := runTestApp(app)
	assertNoProjectConfigError(t, err)
}

func TestBuildCommand_Execute_NoConfig(t *testing.T) {
	testCommandExecuteNoConfig(t, "build", func(ctx context.Context, cliCmd *cli.Command) error {
		deps := createTestDependencies()
		cmd := NewBuildCommand(deps)
		return cmd.Execute(ctx, cliCmd)
	})
}

// TestBuildCommand_SpecificBehavior tests build-specific functionality
func TestBuildCommand_SpecificBehavior(t *testing.T) {
	// Test build-specific behavior here if needed
	deps := createTestDependencies()
	cmd := NewBuildCommand(deps)
	assert.NotNil(t, cmd.ruleGenerator, "BuildCommand should have ruleGenerator")
}
