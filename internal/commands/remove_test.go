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
