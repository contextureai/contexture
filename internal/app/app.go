// Package app provides the main application structure for contexture
package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/log"
	helpCLI "github.com/contextureai/contexture/internal/cli"
	"github.com/contextureai/contexture/internal/dependencies"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/version"
	"github.com/urfave/cli/v3"
)

// Application represents the main contexture CLI application
type Application struct {
	deps    *dependencies.Dependencies
	actions *CommandActions
}

// New creates a new Application instance with proper dependency injection
func New(deps *dependencies.Dependencies) *Application {
	if deps == nil {
		deps = dependencies.New(context.Background())
	}

	return &Application{
		deps:    deps,
		actions: NewCommandActions(deps),
	}
}

// Run executes the application with proper error handling and returns an exit code
func Run(args []string) int {
	// Create dependencies
	ctx := context.Background()
	deps := dependencies.New(ctx)
	app := New(deps)

	if err := app.Execute(ctx, args); err != nil {
		// Display the error
		contextureerrors.Display(err)

		// Get exit code
		var e *contextureerrors.Error
		if errors.As(err, &e) {
			return e.ExitCode()
		}

		// Default exit code
		return 1
	}

	return 0
}

// Execute runs the CLI application with the given context and arguments
func (a *Application) Execute(ctx context.Context, args []string) error {
	// Save and restore original help printer to avoid global state mutation
	originalHelpPrinter := cli.HelpPrinter
	defer func() {
		cli.HelpPrinter = originalHelpPrinter
	}()

	app := a.buildCLIApp()
	return app.Run(ctx, args)
}

// buildCLIApp constructs the CLI application structure
func (a *Application) buildCLIApp() *cli.Command {
	// Set up custom help printer for this execution
	helpPrinter := helpCLI.NewHelpPrinter()
	cli.HelpPrinter = func(w io.Writer, templ string, data any) {
		if err := helpPrinter.Print(w, templ, data); err != nil {
			// For backward compatibility, we write the error but don't panic
			_, _ = fmt.Fprintf(w, "Error rendering help: %v\n", err)
		}
	}

	app := &cli.Command{
		Name:    "contexture",
		Usage:   "AI assistant rule management",
		Version: version.Get().Version,
		Authors: []any{
			"Contexture Contributors",
		},
		Description:        "Contexture helps you manage AI assistant rules across multiple formats (Claude, Cursor, Windsurf).",
		CustomHelpTemplate: helpCLI.AppHelpTemplate,
		Commands:           a.buildCommands(),
		Flags:              a.buildGlobalFlags(),
		Before:             a.setupGlobalFlags,
	}

	return app
}

// buildCommands creates all CLI commands
func (a *Application) buildCommands() []*cli.Command {
	return []*cli.Command{
		a.buildInitCommand(),
		a.buildRulesCommand(),
		a.buildBuildCommand(),
		a.buildConfigCommand(),
	}
}

// buildGlobalFlags creates global application flags
func (a *Application) buildGlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "Enable verbose logging",
		},
	}
}

// setupGlobalFlags handles global flag setup before command execution
func (a *Application) setupGlobalFlags(
	ctx context.Context,
	cmd *cli.Command,
) (context.Context, error) {
	if cmd.Bool("verbose") {
		// Enable debug logging
		log.SetLevel(log.DebugLevel)
	}
	return ctx, nil
}

// Command builders - extracted for better testability and organization

func (a *Application) buildInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new project configuration",
		Description: `Initialize a new Contexture project in the current directory.
This will create a configuration file and set up output formats.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Overwrite existing configuration",
			},
			&cli.BoolFlag{
				Name:  "no-interactive",
				Usage: "Skip interactive prompts (for CI/CD usage)",
			},
		},
		Action: a.actions.InitAction,
	}
}

func (a *Application) buildRulesCommand() *cli.Command {
	return &cli.Command{
		Name:  "rules",
		Usage: "Manage project rules",
		Description: `Manage rules for your Contexture project.
Rules define the AI assistant instructions and configurations.

Use subcommands to manage specific aspects of your rules.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Action:             a.actions.RulesAction,
		Commands: []*cli.Command{
			a.buildRulesAddCommand(),
			a.buildRulesRemoveCommand(),
			a.buildRulesListCommand(),
			a.buildRulesUpdateCommand(),
		},
	}
}

func (a *Application) buildRulesAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add rules to the project",
		ArgsUsage: "[rule-id...]",
		Description: `Add one or more rules to the current project.

Rule IDs can be specified in multiple ways:
• Short format: path/to/rule
• Full format: [contexture:path/to/rule]  
• Custom sources: [contexture(source):path/to/rule,branch]
• With flags: path/to/rule --source https://github.com/user/repo --ref branch

Examples:
  contexture rules add languages/go/testing
  contexture rules add [contexture:languages/go/testing]
  contexture rules add [contexture(git@github.com:user/rules):custom/rule,main]
  contexture rules add my/rule --source https://github.com/user/repo --ref v1.0`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "data",
				Usage: "Additional rule data or variables (JSON format)",
			},
			&cli.StringSliceFlag{
				Name:  "var",
				Usage: "Set a variable (can be used multiple times): --var key=value or --var key='{\"complex\": \"json\"}'",
			},
			&cli.StringFlag{
				Name:    "source",
				Aliases: []string{"src"},
				Usage:   "Custom source repository to pull from",
			},
			&cli.StringFlag{
				Name:  "ref",
				Usage: "Git branch or tag reference",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "default",
				Usage:   "Output format (default, json)",
			},
		},
		Action: a.actions.AddAction,
	}
}

func (a *Application) buildRulesRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Remove rules from the project",
		ArgsUsage: "[rule-id...]",
		Description: `Remove one or more rules from the current project.
This will update the configuration and clean generated files.
Rule IDs are required as arguments.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "default",
				Usage:   "Output format (default, json)",
			},
		},
		Action: a.actions.RemoveAction,
	}
}

func (a *Application) buildBuildCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build output files for all configured formats",
		Description: `Build output files based on the configured rules and formats.
This will fetch all rules, process templates, and write format-specific files.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed progress information",
			},
			&cli.StringSliceFlag{
				Name:  "formats",
				Usage: "Build for specific formats only (claude, cursor, windsurf)",
			},
		},
		Action: a.actions.BuildAction,
	}
}

func (a *Application) buildRulesListCommand() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List rules",
		Description: `List rules configured in the current project.
To add rules, use 'contexture rules add' with rule IDs.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "pattern",
				Aliases: []string{"p"},
				Usage:   "Filter rules by regex pattern (matches ID, title, description, tags, etc.)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format (default, json)",
				Value:   "default",
			},
		},
		Action: a.actions.ListAction,
	}
}

func (a *Application) buildRulesUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update rules to latest versions",
		Description: `Update configured rules to their latest versions.
This will check for updates and optionally apply them.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Check for updates without applying them",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Skip confirmation prompts",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "default",
				Usage:   "Output format (default, json)",
			},
		},
		Action: a.actions.UpdateAction,
	}
}

func (a *Application) buildConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Show and manage project configuration",
		Description: `Display current project configuration and manage output formats.

Use subcommands to manage specific aspects of your configuration.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags:              []cli.Flag{},
		Action:             a.actions.ConfigAction,
		Commands: []*cli.Command{
			a.buildConfigShowCommand(),
			a.buildConfigFormatsCommand(),
		},
	}
}

func (a *Application) buildConfigFormatsCommand() *cli.Command {
	return &cli.Command{
		Name:    "formats",
		Aliases: []string{"fmt"},
		Usage:   "Manage output formats",
		Description: `Manage output formats for your Contexture project.

Use subcommands to perform specific format operations.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags:              []cli.Flag{},
		Action:             a.actions.ConfigFormatsAction,
		Commands: []*cli.Command{
			a.buildConfigFormatsListCommand(),
			a.buildConfigFormatsAddCommand(),
			a.buildConfigFormatsRemoveCommand(),
			a.buildConfigFormatsEnableCommand(),
			a.buildConfigFormatsDisableCommand(),
		},
	}
}

func (a *Application) buildConfigFormatsListCommand() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List all configured formats",
		Description: `Display all configured output formats with their current status.

Shows which formats are enabled or disabled for the current project.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags:              []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return a.actions.ConfigFormatsListAction(ctx, cmd, a.deps)
		},
	}
}

func (a *Application) buildConfigFormatsAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add one or more formats",
		ArgsUsage: "[format-type...] (if no args provided, shows interactive selection)",
		Description: `Add output formats to your project configuration.

Available formats: claude, cursor, windsurf

When run without arguments, shows an interactive selection menu.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags:              []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return a.actions.ConfigFormatsAddAction(ctx, cmd, a.deps)
		},
	}
}

func (a *Application) buildConfigFormatsRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Remove one or more formats",
		ArgsUsage: "[format-type...] (if no args provided, shows interactive selection)",
		Description: `Remove output formats from your project configuration.

When run without arguments, shows an interactive selection menu of configured formats.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags:              []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return a.actions.ConfigFormatsRemoveAction(ctx, cmd, a.deps)
		},
	}
}

func (a *Application) buildConfigFormatsEnableCommand() *cli.Command {
	return &cli.Command{
		Name:      "enable",
		Usage:     "Enable a specific format",
		ArgsUsage: "[format-type] (if no args provided, shows interactive selection)",
		Description: `Enable an output format that was previously disabled.

The format must already be added to the project configuration.

When run without arguments, shows an interactive selection menu.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags:              []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return a.actions.ConfigFormatsEnableAction(ctx, cmd, a.deps)
		},
	}
}

func (a *Application) buildConfigFormatsDisableCommand() *cli.Command {
	return &cli.Command{
		Name:      "disable",
		Usage:     "Disable a specific format",
		ArgsUsage: "[format-type] (if no args provided, shows interactive selection)",
		Description: `Disable an output format without removing it from the configuration.

At least one format must remain enabled in the project.

When run without arguments, shows an interactive selection menu.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Flags:              []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return a.actions.ConfigFormatsDisableAction(ctx, cmd, a.deps)
		},
	}
}

func (a *Application) buildConfigShowCommand() *cli.Command {
	return &cli.Command{
		Name:  "show",
		Usage: "Show current project configuration",
		Description: `Display the current project configuration including enabled formats and rules.

This is the default action when running 'contexture config' without subcommands.`,
		CustomHelpTemplate: helpCLI.CommandHelpTemplate,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return a.actions.ConfigAction(ctx, cmd)
		},
	}
}
