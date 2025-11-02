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
	"github.com/contextureai/contexture/internal/ui"
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
		if rws.Source == domain.RuleSourceUser {
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

		// List files that will be deleted
		targetFormats := c.getTargetFormats(config, cmd.StringSlice("formats"))
		var filesToDelete []string
		for _, formatConfig := range targetFormats {
			format, err := c.registry.CreateFormat(formatConfig.Type, c.fs, nil)
			if err != nil {
				continue
			}

			outputPath := format.GetOutputPath(&formatConfig)
			if outputPath == "" {
				continue
			}

			// Check if file or directory exists using afero.Fs
			metadata := format.GetMetadata()
			stat, err := c.fs.Stat(outputPath)
			if err != nil {
				// Path doesn't exist, skip
				continue
			}

			if metadata.IsDirectory {
				// Check if directory has files
				if stat.IsDir() {
					files, err := afero.ReadDir(c.fs, outputPath)
					if err == nil && len(files) > 0 {
						// Add the directory itself
						filesToDelete = append(filesToDelete, outputPath+"/")
					}
				}
			} else {
				// File exists
				if !stat.IsDir() {
					filesToDelete = append(filesToDelete, outputPath)
				}
			}
		}

		// If there are files to delete, ask for consent
		if len(filesToDelete) > 0 {
			fmt.Println("\nThe following output files will be deleted:")
			for _, file := range filesToDelete {
				fmt.Printf("  - %s\n", file)
			}
			fmt.Println()

			// Skip prompt if --force flag is set or if running in non-interactive mode
			if !cmd.Bool("force") {
				fmt.Print("Do you want to continue? (y/N): ")
				var response string
				_, _ = fmt.Scanln(&response) // Ignore error - empty input is valid
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Aborted. No files were deleted.")
					return nil
				}
			}

			// Delete files by calling Write with empty rules
			// This triggers the new deletion logic in format handlers
			fmt.Println()
			for _, formatConfig := range targetFormats {
				format, err := c.registry.CreateFormat(formatConfig.Type, c.fs, nil)
				if err != nil {
					continue
				}

				// Call Write with empty rules to trigger deletion
				if err := format.Write([]*domain.TransformedRule{}, &formatConfig); err != nil {
					log.Warn("Failed to delete output", "format", formatConfig.Type, "error", err)
				}
			}

			fmt.Println("Output files deleted successfully.")
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

	// Clean up orphaned rules before generation
	c.cleanupOrphanedRules(ctx, targetFormats, projectRules, userRules)

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
	// Group formats by what rules they need to generate
	var projectFormats []domain.FormatConfig
	var userFormats []domain.FormatConfig // Formats for native user rules generation
	theme := ui.DefaultTheme()
	mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	for _, formatConfig := range targetFormats {
		caps, _ := c.registry.GetCapabilities(formatConfig.Type)
		mode := formatConfig.GetEffectiveUserRulesMode()

		// Determine which rules to use based on mode
		switch mode {
		case domain.UserRulesNative:
			// For native mode, generate project rules to project location
			if len(projectRules) > 0 {
				projectFormats = append(projectFormats, formatConfig)
			}

			// Generate user rules to native location if supported
			if len(userRules) > 0 {
				if caps.SupportsUserRules {
					// Create format config for user rules
					userFormatConfig := formatConfig
					userDir := filepath.Dir(caps.UserRulesPath)
					userFormatConfig.BaseDir = userDir
					userFormatConfig.IsUserRules = true
					userFormats = append(userFormats, userFormatConfig)
				} else {
					// Warn that this format doesn't support global rules
					handler, _ := c.registry.GetHandler(formatConfig.Type)
					displayName := string(formatConfig.Type)
					if handler != nil {
						displayName = handler.GetDisplayName()
					}
					fmt.Printf("  %s %s does not support global rules\n",
						mutedStyle.Render("⚠"),
						displayName)
				}
			}

		case domain.UserRulesProject:
			// For project mode, include both project and user rules in project location
			if len(projectRules) > 0 || len(userRules) > 0 {
				projectFormats = append(projectFormats, formatConfig)
			}

		case domain.UserRulesDisabled:
			// For disabled mode, only project rules
			if len(projectRules) > 0 {
				projectFormats = append(projectFormats, formatConfig)
			}
		}
	}

	// Generate project rules - all formats in a single operation for clean grouping
	if len(projectFormats) > 0 {
		config := &domain.Project{}
		*config = *baseConfig
		// Use merged rules (project + user) for all formats to enable single-pass generation
		config.Rules = append(append([]domain.RuleRef{}, projectRules...), userRules...)

		// Pass information about whether global rules are present
		hasGlobalRules := len(userRules) > 0
		if err := c.ruleGenerator.GenerateRulesWithScopeAndWarning(ctx, config, projectFormats, "project", hasGlobalRules); err != nil {
			return err
		}
	}

	// Generate user rules for all native formats that support it - SINGLE fetch
	if len(userFormats) > 0 && len(userRules) > 0 {
		userConfig := &domain.Project{}
		*userConfig = *baseConfig
		userConfig.Rules = userRules

		// Generate once for ALL user formats with [global] tag
		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, userConfig, userFormats, "global"); err != nil {
			log.Warn("Failed to generate user rules to native location", "error", err)
		}
	}

	return nil
}

// cleanupOrphanedRules removes rule files that exist in outputs but not in config
func (c *BuildCommand) cleanupOrphanedRules(
	_ context.Context,
	targetFormats []domain.FormatConfig,
	projectRules, userRules []domain.RuleRef,
) {
	// Build a set of expected rule IDs for quick lookup
	expectedRules := make(map[string]bool)
	for _, rule := range projectRules {
		expectedRules[extractRulePath(rule.ID)] = true
	}
	for _, rule := range userRules {
		expectedRules[extractRulePath(rule.ID)] = true
	}

	// For each format, list installed rules and remove orphaned ones
	for _, formatConfig := range targetFormats {
		format, err := c.registry.CreateFormat(formatConfig.Type, c.fs, nil)
		if err != nil {
			log.Warn("Failed to create format for cleanup", "format", formatConfig.Type, "error", err)
			continue
		}

		// Get currently installed rules
		installedRules, err := format.List(&formatConfig)
		if err != nil {
			log.Debug("Failed to list installed rules for cleanup", "format", formatConfig.Type, "error", err)
			// This might fail if no rules are installed yet, which is fine
			continue
		}

		// Remove orphaned rules
		for _, installed := range installedRules {
			rulePath := extractRulePath(installed.Rule.ID)
			if !expectedRules[rulePath] {
				log.Debug("Removing orphaned rule", "rule", installed.Rule.ID, "format", formatConfig.Type)
				if err := format.Remove(installed.Rule.ID, &formatConfig); err != nil {
					log.Warn("Failed to remove orphaned rule", "rule", installed.Rule.ID, "error", err)
				}
			}
		}

		// Also handle UserRulesMode changes for formats that merge global rules
		mode := formatConfig.GetEffectiveUserRulesMode()
		if mode == domain.UserRulesDisabled {
			// Remove any global rules that might be in project files
			for _, userRule := range userRules {
				// Try to remove - ignore errors as rule might not exist
				_ = format.Remove(userRule.ID, &formatConfig)
			}
		}
	}
}

// extractRulePath extracts the base rule path for comparison
func extractRulePath(ruleID string) string {
	// Use domain package's ExtractRulePath for consistent behavior
	return domain.ExtractRulePath(ruleID)
}

// BuildAction is the CLI action handler for the build command
func BuildAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	buildCmd := NewBuildCommand(deps)
	return buildCmd.Execute(ctx, cmd)
}
