// Package commands provides CLI command implementations
package commands

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/tui"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/urfave/cli/v3"
)

// InitCommand implements the init command
type InitCommand struct {
	projectManager *project.Manager
	registry       *format.Registry
}

// NewInitCommand creates a new init command
func NewInitCommand(deps *dependencies.Dependencies) *InitCommand {
	return &InitCommand{
		projectManager: project.NewManager(deps.FS),
		registry:       format.GetDefaultRegistry(deps.FS),
	}
}

// Execute runs the init command
func (c *InitCommand) Execute(_ context.Context, cmd *cli.Command) error {
	noInteractive := cmd.Bool("no-interactive")
	force := cmd.Bool("force")

	return c.initProjectConfig(force, noInteractive)
}

// initProjectConfig initializes project-specific configuration
func (c *InitCommand) initProjectConfig(force, noInteractive bool) error {
	// Check if configuration already exists
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	existingConfig, _ := c.projectManager.LoadConfig(currentDir)

	if existingConfig != nil && !force {
		log.Error("Configuration already exists", "path", existingConfig.Path)
		log.Info("Use --force to overwrite existing configuration")
		return fmt.Errorf("configuration already exists")
	}

	// Show command header
	fmt.Printf("%s\n\n", ui.CommandHeader("init"))

	// Show welcome message
	theme := ui.DefaultTheme()
	welcomeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	fmt.Println(welcomeStyle.Render("Welcome to Contexture!"))
	fmt.Println("Let's set up your project configuration.")
	fmt.Println()

	// Handle non-interactive mode
	if noInteractive {
		return c.initProjectNonInteractive(currentDir)
	}

	// Interactive form for configuration
	var selectedFormats []string
	var useContextureDir bool

	form := ui.ConfigureHuhForm(huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select output formats").
				Description("Choose which formats you want to generate").
				Options(c.registry.GetUIOptions([]string{"claude"})...).
				Value(&selectedFormats).
				Validate(func(val []string) error {
					if len(val) == 0 {
						return fmt.Errorf("at least one format must be selected")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Configuration location").
				Description("Store configuration in .contexture/ directory?").
				Affirmative("Yes (.contexture/)").
				Negative("No (project root)").
				Value(&useContextureDir),
		),
	))

	err = tui.HandleFormError(form.Run())
	if err != nil {
		// Handle user cancellation gracefully
		if errors.Is(err, tui.ErrUserCancelled) {
			log.Info("Initialization cancelled")
			return nil
		}
		return err
	}

	// Convert selected formats to FormatTypes
	var formatTypes []domain.FormatType
	for _, selected := range selectedFormats {
		formatTypes = append(formatTypes, domain.FormatType(selected))
	}

	// Determine configuration location
	var location domain.ConfigLocation
	if useContextureDir {
		location = domain.ConfigLocationContexture
	} else {
		location = domain.ConfigLocationRoot
	}

	// Create the configuration
	config, err := c.projectManager.InitConfig(currentDir, formatTypes, location)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	// Success message
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	readyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	mutedStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)

	fmt.Printf("%s %s\n",
		successStyle.Render("Configuration generated successfully!"),
		mutedStyle.Render(fmt.Sprintf("[%s]", getRelativeConfigPath(currentDir, location))),
	)
	fmt.Println()
	fmt.Println(readyStyle.Render("Ready to get started?"))
	fmt.Println("  Browse available rules with: contexture rules add")
	fmt.Println("  Add a rule with: contexture rules add <rule-id>")

	log.Debug("Project initialized",
		"formats", len(config.Formats),
		"location", location,
		"path", domain.GetConfigPath(currentDir, location))

	return nil
}

// initProjectNonInteractive initializes project config without interactive prompts
func (c *InitCommand) initProjectNonInteractive(currentDir string) error {
	// Use default settings for non-interactive mode
	formatTypes := []domain.FormatType{domain.FormatClaude} // Default to Claude format
	location := domain.ConfigLocationRoot                   // Default to project root

	// Create the configuration
	config, err := c.projectManager.InitConfig(currentDir, formatTypes, location)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	// Success message
	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	readyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	mutedStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)

	fmt.Printf("%s %s\n",
		successStyle.Render("Configuration generated successfully!"),
		mutedStyle.Render(fmt.Sprintf("[%s]", getRelativeConfigPath(currentDir, location))),
	)
	fmt.Println()
	fmt.Println(readyStyle.Render("Ready to get started?"))
	fmt.Println("  Browse available rules with: contexture rules add")
	fmt.Println("  Add a rule with: contexture rules add <rule-id>")

	log.Debug("Project initialized (non-interactive)",
		"formats", len(config.Formats),
		"location", location,
		"path", domain.GetConfigPath(currentDir, location))

	return nil
}

// InitAction is the CLI action handler for the init command
func InitAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	initCmd := NewInitCommand(deps)
	return initCmd.Execute(ctx, cmd)
}

// getRelativeConfigPath returns the relative path for displaying config location
func getRelativeConfigPath(_ string, location domain.ConfigLocation) string {
	switch location {
	case domain.ConfigLocationContexture:
		return ".contexture/.contexture.yaml"
	case domain.ConfigLocationRoot:
		return ".contexture.yaml"
	default:
		return ".contexture.yaml"
	}
}
