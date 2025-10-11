package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("creates_app_with_dependencies", func(t *testing.T) {
		deps := dependencies.NewForTesting(context.Background())
		app := New(deps)

		assert.NotNil(t, app)
		assert.Equal(t, deps, app.deps)
		assert.NotNil(t, app.actions)
	})

	t.Run("creates_default_dependencies_when_nil", func(t *testing.T) {
		app := New(nil)

		assert.NotNil(t, app)
		assert.NotNil(t, app.deps)
		assert.NotNil(t, app.actions)
	})
}

func TestRun(t *testing.T) {
	t.Run("returns_zero_on_success", func(t *testing.T) {
		// Test with help flag which should succeed quickly
		exitCode := Run([]string{"contexture", "--help"})
		assert.Equal(t, 0, exitCode)
	})

	t.Run("returns_non_zero_on_error", func(t *testing.T) {
		// SKIP: The CLI framework calls os.Exit directly for invalid commands
		// This is a limitation of testing CLI applications
		t.Skip("CLI framework calls os.Exit directly for invalid commands, cannot test exit code")
	})

	t.Run("handles_empty_args", func(t *testing.T) {
		// Should show help and return 0
		exitCode := Run([]string{"contexture"})
		assert.Equal(t, 0, exitCode)
	})
}

func TestApplication_Execute(t *testing.T) {
	t.Run("executes_help_successfully", func(t *testing.T) {
		deps := dependencies.NewForTesting(context.Background())
		app := New(deps)
		ctx := context.Background()

		err := app.Execute(ctx, []string{"contexture", "--help"})
		assert.NoError(t, err)
	})

	t.Run("handles_version_flag", func(t *testing.T) {
		deps := dependencies.NewForTesting(context.Background())
		app := New(deps)
		ctx := context.Background()

		err := app.Execute(ctx, []string{"contexture", "--version"})
		assert.NoError(t, err)
	})

	t.Run("respects_context_cancellation", func(_ *testing.T) {
		deps := dependencies.NewForTesting(context.Background())
		app := New(deps)

		// Create context that cancels quickly
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// This might or might not fail depending on timing, but should not panic
		_ = app.Execute(ctx, []string{"contexture", "--help"})
	})
}

func TestApplication_buildCLIApp(t *testing.T) {
	deps := dependencies.NewForTesting(context.Background())
	app := New(deps)

	cli := app.buildCLIApp()

	t.Run("has_basic_properties", func(t *testing.T) {
		assert.Equal(t, "contexture", cli.Name)
		assert.Equal(t, "AI assistant rule management", cli.Usage)
		assert.NotEmpty(t, cli.Version)
		assert.NotNil(t, cli.Authors)
		assert.NotEmpty(t, cli.Description)
	})

	t.Run("has_all_commands", func(t *testing.T) {
		expectedCommands := []string{
			"init",
			"rules",
			"build",
			"config",
		}
		commandNames := make([]string, len(cli.Commands))

		for i, cmd := range cli.Commands {
			commandNames[i] = cmd.Name
		}

		for _, expected := range expectedCommands {
			assert.Contains(t, commandNames, expected, "should have command: %s", expected)
		}
	})

	t.Run("has_global_flags", func(t *testing.T) {
		flagNames := make([]string, len(cli.Flags))
		for i, flag := range cli.Flags {
			flagNames[i] = flag.Names()[0]
		}

		assert.Contains(t, flagNames, "verbose")
	})

	t.Run("commands_have_actions", func(t *testing.T) {
		for _, cmd := range cli.Commands {
			assert.NotNil(t, cmd.Action, "command %s should have action", cmd.Name)
		}
	})
}

func TestApplication_buildCommands(t *testing.T) {
	t.Parallel()
	deps := dependencies.NewForTesting(context.Background())
	app := New(deps)

	commands := app.buildCommands()

	t.Run("returns_expected_number_of_commands", func(t *testing.T) {
		assert.Len(t, commands, 6) // init, rules, build, query, config, providers
	})

	t.Run("all_commands_have_required_fields", func(t *testing.T) {
		for _, cmd := range commands {
			assert.NotEmpty(t, cmd.Name, "command should have name")
			assert.NotEmpty(t, cmd.Usage, "command should have usage")
			assert.NotEmpty(t, cmd.Description, "command should have description")
			assert.NotNil(t, cmd.Action, "command should have action")
		}
	})
}

func TestApplication_buildGlobalFlags(t *testing.T) {
	t.Parallel()
	deps := dependencies.NewForTesting(context.Background())
	app := New(deps)

	flags := app.buildGlobalFlags()

	t.Run("has_verbose_flag", func(t *testing.T) {
		assert.Len(t, flags, 1)
		assert.Equal(t, "verbose", flags[0].Names()[0])
	})
}

func TestApplication_setupGlobalFlags(t *testing.T) {
	deps := dependencies.NewForTesting(context.Background())
	app := New(deps)

	// Create a mock command with verbose flag
	cli := app.buildCLIApp()

	t.Run("handles_verbose_flag", func(t *testing.T) {
		// This is somewhat difficult to test directly since setupGlobalFlags
		// is called by the CLI framework, but we can test the structure
		assert.NotNil(t, cli.Before)
	})
}

// Test individual command builders
func TestApplication_CommandBuilders(t *testing.T) {
	t.Parallel()
	deps := dependencies.NewForTesting(context.Background())
	app := New(deps)

	tests := []struct {
		name        string
		builder     func() *cli.Command
		commandName string
	}{
		{"init", func() *cli.Command { return app.buildInitCommand() }, "init"},
		{"rules", func() *cli.Command { return app.buildRulesCommand() }, "rules"},
		{"build", func() *cli.Command { return app.buildBuildCommand() }, "build"},
		{"config", func() *cli.Command { return app.buildConfigCommand() }, "config"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.builder()

			assert.Equal(t, tt.commandName, cmd.Name)
			assert.NotEmpty(t, cmd.Usage)
			assert.NotEmpty(t, cmd.Description)

			// Commands should have either an Action or subcommands
			if len(cmd.Commands) > 0 {
				// Parent command with subcommands - action is optional
				assert.NotEmpty(t, cmd.Commands, "parent commands should have subcommands")
			} else {
				// Leaf command - must have action
				assert.NotNil(t, cmd.Action, "leaf commands should have action")
			}

			// Commands should have flags slice (might be empty) - no assertion needed
		})
	}
}

func TestApplication_Integration(t *testing.T) {
	t.Run("help_commands_execute", func(t *testing.T) {
		deps := dependencies.NewForTesting(context.Background())
		app := New(deps)
		ctx := context.Background()

		commands := []string{"init", "rules", "build", "config"}
		subcommands := []string{"rules add", "rules remove", "rules list", "rules update"}

		for _, cmdName := range commands {
			t.Run("help_"+cmdName, func(t *testing.T) {
				err := app.Execute(ctx, []string{"contexture", cmdName, "--help"})
				assert.NoError(t, err, "help for %s command should work", cmdName)
			})
		}

		for _, subCmd := range subcommands {
			t.Run("help_"+strings.ReplaceAll(subCmd, " ", "_"), func(t *testing.T) {
				args := strings.Split("contexture "+subCmd+" --help", " ")
				err := app.Execute(ctx, args)
				assert.NoError(t, err, "help for %s command should work", subCmd)
			})
		}
	})

	t.Run("application_lifecycle", func(t *testing.T) {
		// Test that creating and using the app multiple times works
		for i := range 3 {
			deps := dependencies.NewForTesting(context.Background())
			app := New(deps)
			ctx := context.Background()

			err := app.Execute(ctx, []string{"contexture", "--help"})
			assert.NoError(t, err, "iteration %d should work", i)
		}
	})
}

// Benchmark application creation and basic operations
func BenchmarkNew(b *testing.B) {
	deps := dependencies.NewForTesting(context.Background())

	b.ResetTimer()
	for range b.N {
		_ = New(deps)
	}
}

func BenchmarkRun(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		_ = Run([]string{"contexture", "--help"})
	}
}

func BenchmarkApplication_Execute(b *testing.B) {
	deps := dependencies.NewForTesting(context.Background())
	app := New(deps)
	ctx := context.Background()
	args := []string{"contexture", "--help"}

	b.ResetTimer()
	for range b.N {
		_ = app.Execute(ctx, args)
	}
}
