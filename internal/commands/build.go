// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
			rule.NewFetcher(deps.FS, newOpenRepository(deps.FS), rule.FetcherConfig{}, deps.ProviderRegistry),
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
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	// Load merged configuration (global + project + local rules)
	merged, err := c.projectManager.LoadConfigMergedWithLocalRules(currentDir)
	if err != nil {
		return contextureerrors.Wrap(err, "load configuration").
			WithSuggestions("Run 'contexture init' to create a project configuration")
	}

	// Separate user rules from project rules
	var projectRules, userRules []domain.RuleRef
	for _, rws := range merged.MergedRules {
		if rws.Source == domain.RuleSourceGlobal {
			userRules = append(userRules, rws.RuleRef)
		} else {
			projectRules = append(projectRules, rws.RuleRef)
		}
	}

	// Create project config for generation
	config := &domain.Project{}
	*config = *merged.Project

	if len(projectRules) == 0 && len(userRules) == 0 {
		fmt.Fprintln(os.Stderr, "No rules configured")

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
		"project_rules", len(projectRules),
		"user_rules", len(userRules),
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

	// Generate rules per format based on user rules mode
	err = c.generateWithUserRulesHandling(ctx, config, targetFormats, projectRules, userRules)
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

// generateWithUserRulesHandling generates rules for each format based on user rules mode
func (c *BuildCommand) generateWithUserRulesHandling(
	ctx context.Context,
	baseConfig *domain.Project,
	targetFormats []domain.FormatConfig,
	projectRules, userRules []domain.RuleRef,
) error {
	for _, formatConfig := range targetFormats {
		caps, _ := c.registry.GetCapabilities(formatConfig.Type)
		mode := formatConfig.GetEffectiveUserRulesMode()

		// Determine which rules to use based on mode
		var rulesToGenerate []domain.RuleRef

		switch mode {
		case domain.UserRulesNative:
			// For native mode, only generate project rules to project
			// User rules will be generated to native location separately
			rulesToGenerate = projectRules
		case domain.UserRulesProject:
			// For project mode, include both
			rulesToGenerate = append(append([]domain.RuleRef{}, projectRules...), userRules...)
		case domain.UserRulesDisabled:
			// For disabled mode, only project rules
			rulesToGenerate = projectRules
		}

		// Generate project rules if any
		if len(rulesToGenerate) > 0 {
			config := &domain.Project{}
			*config = *baseConfig
			config.Rules = rulesToGenerate

			if err := c.ruleGenerator.GenerateRules(ctx, config, []domain.FormatConfig{formatConfig}); err != nil {
				return err
			}
		}

		// For native mode, generate user rules to native location separately
		if mode == domain.UserRulesNative && caps.SupportsUserRules && len(userRules) > 0 {
			userConfig := &domain.Project{}
			*userConfig = *baseConfig
			userConfig.Rules = userRules

			// Create format config for user rules
			userFormatConfig := formatConfig
			userDir := filepath.Dir(caps.UserRulesPath)
			userFormatConfig.BaseDir = userDir
			userFormatConfig.IsUserRules = true // Mark as user rules generation

			if err := c.ruleGenerator.GenerateRules(ctx, userConfig, []domain.FormatConfig{userFormatConfig}); err != nil {
				log.Warn("Failed to generate user rules to native location",
					"format", formatConfig.Type,
					"path", caps.UserRulesPath,
					"error", err)
			}
		}
	}

	return nil
}

// BuildAction is the CLI action handler for the build command
func BuildAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	buildCmd := NewBuildCommand(deps)
	return buildCmd.Execute(ctx, cmd)
}
