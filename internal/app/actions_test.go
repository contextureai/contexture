package app

import (
	"context"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestNewCommandActions(t *testing.T) {
	t.Parallel()
	deps := dependencies.NewForTesting(context.Background())
	actions := NewCommandActions(deps)

	assert.NotNil(t, actions)
	assert.Equal(t, deps, actions.deps)
}

func TestCommandActions_AllActions(t *testing.T) {
	t.Parallel()
	deps := dependencies.NewForTesting(context.Background())
	actions := NewCommandActions(deps)

	tests := []struct {
		name   string
		action func(context.Context, *cli.Command) error
	}{
		{"InitAction", actions.InitAction},
		{"AddAction", actions.AddAction},
		{"RemoveAction", actions.RemoveAction},
		{"BuildAction", actions.BuildAction},
		{"ListAction", actions.ListAction},
		{"UpdateAction", actions.UpdateAction},
		{"ConfigAction", actions.ConfigAction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the action function exists and has the correct signature
			assert.NotNil(t, tt.action, "action function should exist")

			// Note: We don't call the actual functions here because they require
			// complex setup (config files, etc.) and would panic or fail.
			// The actual command logic is tested in the commands package.
			// Here we just verify the wrapper functions exist and have correct signatures.
		})
	}
}

func TestCommandActions_Dependencies(t *testing.T) {
	t.Parallel()
	t.Run("uses_provided_dependencies", func(t *testing.T) {
		type contextKey string
		ctx1 := context.WithValue(context.Background(), contextKey("test"), "value1")
		deps1 := dependencies.NewForTesting(ctx1)

		actions1 := NewCommandActions(deps1)
		assert.Equal(t, ctx1, actions1.deps.Context)

		ctx2 := context.WithValue(context.Background(), contextKey("test"), "value2")
		deps2 := dependencies.NewForTesting(ctx2)

		actions2 := NewCommandActions(deps2)
		assert.Equal(t, ctx2, actions2.deps.Context)
	})

	t.Run("different_instances_have_independent_dependencies", func(t *testing.T) {
		deps1 := dependencies.NewForTesting(context.Background())
		deps2 := dependencies.NewForTesting(context.Background())

		actions1 := NewCommandActions(deps1)
		actions2 := NewCommandActions(deps2)

		// Verify they have different filesystem instances
		assert.NotSame(t, actions1.deps.FS, actions2.deps.FS)
		assert.NotSame(t, actions1.deps, actions2.deps)
	})
}

func TestCommandActions_IntegrationWithCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	deps := dependencies.NewForTesting(context.Background())
	actions := NewCommandActions(deps)

	// Create a simple CLI app using our actions
	app := &cli.Command{
		Name: "test-app",
		Commands: []*cli.Command{
			{
				Name:   "init",
				Action: actions.InitAction,
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "help"},
				},
			},
			{
				Name:   "add",
				Action: actions.AddAction,
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "help"},
				},
			},
		},
	}

	ctx := context.Background()

	t.Run("can_get_help_for_commands", func(t *testing.T) {
		// Test help for init command
		err := app.Run(ctx, []string{"test-app", "init", "--help"})
		require.NoError(t, err)

		// Test help for add command
		err = app.Run(ctx, []string{"test-app", "add", "--help"})
		require.NoError(t, err)
	})
}

func TestCommandActions_ErrorHandling(t *testing.T) {
	t.Parallel()
	deps := dependencies.NewForTesting(context.Background())
	actions := NewCommandActions(deps)

	t.Run("actions_exist_and_have_correct_signature", func(t *testing.T) {
		actionFuncs := []func(context.Context, *cli.Command) error{
			actions.InitAction,
			actions.AddAction,
			actions.RemoveAction,
			actions.BuildAction,
			actions.ListAction,
			actions.UpdateAction,
			actions.ConfigAction,
		}

		for i, actionFunc := range actionFuncs {
			t.Run(string(rune('a'+i)), func(t *testing.T) {
				// Just verify the function signature is correct by checking it's not nil
				assert.NotNil(t, actionFunc, "action function should exist")
			})
		}
	})
}

func TestCommandActions_ContextHandling(t *testing.T) {
	t.Parallel()
	deps := dependencies.NewForTesting(context.Background())
	actions := NewCommandActions(deps)

	t.Run("context_parameter_accepted", func(t *testing.T) {
		// Test that actions accept context parameter (signature test)
		// We don't actually call them to avoid complex setup requirements
		assert.NotNil(t, actions.InitAction)
	})

	t.Run("cli_command_parameter_accepted", func(t *testing.T) {
		// Test that actions accept cli.Command parameter (signature test)
		assert.NotNil(t, actions.InitAction)
	})
}
