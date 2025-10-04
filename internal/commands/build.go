// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

// BuildCommand implements the build command
type BuildCommand struct {
	projectManager *project.Manager
	ruleGenerator  *RuleGenerator
	registry       *format.Registry
	fs             afero.Fs
}

// NewBuildCommand creates a new build command
func NewBuildCommand(deps *dependencies.Dependencies) *BuildCommand {
	registry := format.GetDefaultRegistry(deps.FS)
	return &BuildCommand{
		projectManager: project.NewManager(deps.FS),
		ruleGenerator: NewRuleGenerator(
			rule.NewFetcher(deps.FS, newOpenRepository(deps.FS), rule.FetcherConfig{}),
			rule.NewValidator(),
			rule.NewProcessor(),
			registry,
			deps.FS,
		),
		registry: registry,
		fs:       deps.FS,
	}
}

// Execute runs the build command
func (c *BuildCommand) Execute(ctx context.Context, cmd *cli.Command) error {
	// Load project configuration
	configLoad, err := LoadProjectConfig(c.projectManager)
	if err != nil {
		return err
	}

	config := configLoad.Config

	if len(config.Rules) == 0 {
		log.Info("No rules configured")

		// Clean up empty directories for all enabled formats even when no rules exist
		targetFormats := c.getTargetFormats(config, cmd.StringSlice("formats"))
		for _, formatConfig := range targetFormats {
			format, err := c.registry.CreateFormat(formatConfig.Type, c.fs, nil)
			if err != nil {
				log.Warn("Failed to create format for cleanup", "format", formatConfig.Type, "error", err)
				continue
			}
			c.ruleGenerator.cleanupEmptyFormatDirectory(format, &formatConfig)
		}

		return nil
	}

	// Show header like add and list commands
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	fmt.Printf("%s\n\n", headerStyle.Render("Build Rules"))

	// Get target formats (either user-specified or all enabled)
	targetFormats := c.getTargetFormats(config, cmd.StringSlice("formats"))
	if len(targetFormats) == 0 {
		return contextureerrors.ValidationErrorf("formats", "no target formats available")
	}

	log.Debug("Starting build",
		"rules", len(config.Rules),
		"formats", len(targetFormats))

	// Show which formats will be built
	if cmd.Bool("verbose") {
		fmt.Printf("Building for formats: ")
		formatNames := make([]string, len(targetFormats))
		for i, format := range targetFormats {
			if handler, exists := c.registry.GetHandler(format.Type); exists {
				formatNames[i] = handler.GetDisplayName()
			} else {
				formatNames[i] = string(format.Type)
			}
		}
		fmt.Printf("%s\n", strings.Join(formatNames, ", "))
	}

	// Use shared rule generator with consistent UI styling
	err = c.ruleGenerator.GenerateRules(ctx, config, targetFormats)
	if err != nil {
		return contextureerrors.Wrap(err, "generate rules")
	}

	log.Debug("Build completed successfully")

	return nil
}

// getTargetFormats determines which formats to generate based on user input and configuration
func (c *BuildCommand) getTargetFormats(
	config *domain.Project,
	requestedFormats []string,
) []domain.FormatConfig {
	allEnabledFormats := config.GetEnabledFormats()

	// If no specific formats requested, return all enabled formats
	if len(requestedFormats) == 0 {
		return allEnabledFormats
	}

	// Convert requested format strings to FormatType
	var requestedTypes []domain.FormatType
	for _, formatStr := range requestedFormats {
		switch strings.ToLower(formatStr) {
		case "claude":
			requestedTypes = append(requestedTypes, domain.FormatClaude)
		case "cursor":
			requestedTypes = append(requestedTypes, domain.FormatCursor)
		case "windsurf":
			requestedTypes = append(requestedTypes, domain.FormatWindsurf)
		default:
			log.Warn("Unknown format requested", "format", formatStr)
		}
	}

	// Filter enabled formats to only include requested ones
	var targetFormats []domain.FormatConfig
	for _, enabledFormat := range allEnabledFormats {
		for _, requestedType := range requestedTypes {
			if enabledFormat.Type == requestedType {
				targetFormats = append(targetFormats, enabledFormat)
				break
			}
		}
	}

	return targetFormats
}

// BuildAction is the CLI action handler for the build command
func BuildAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	buildCmd := NewBuildCommand(deps)
	return buildCmd.Execute(ctx, cmd)
}
