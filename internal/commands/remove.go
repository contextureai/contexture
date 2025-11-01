// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
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
	"github.com/contextureai/contexture/internal/output"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

// Note: Using shared types from add.go to avoid duplication

// RemoveCommand implements the remove command
type RemoveCommand struct {
	projectManager *project.Manager
	registry       *format.Registry
	ruleFetcher    rule.Fetcher
	ruleGenerator  *RuleGenerator
}

// NewRemoveCommand creates a new remove command
func NewRemoveCommand(deps *dependencies.Dependencies) *RemoveCommand {
	ruleFetcher := rule.NewFetcher(deps.FS, newOpenRepository(deps.FS), rule.FetcherConfig{}, deps.ProviderRegistry)
	registry := format.GetDefaultRegistry(deps.FS)

	return &RemoveCommand{
		projectManager: project.NewManager(deps.FS),
		registry:       registry,
		ruleFetcher:    ruleFetcher,
		ruleGenerator: NewRuleGenerator(
			ruleFetcher,
			rule.NewValidator(),
			rule.NewProcessor(),
			registry,
			deps.FS,
		),
	}
}

// Execute runs the remove command
func (c *RemoveCommand) Execute(ctx context.Context, cmd *cli.Command, ruleIDs []string) error {
	// Check if JSON output mode - if so, suppress all terminal output
	outputFormat := output.Format(cmd.String("output"))
	isJSONMode := outputFormat == output.FormatJSON

	if !isJSONMode {
		// Show header like other commands
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
		fmt.Printf("%s\n\n", headerStyle.Render("Remove Rules"))
	}

	// Check if global flag is set
	isGlobal := cmd.Bool("global")

	// Load configuration
	config, _, err := loadConfigByScope(c.projectManager, isGlobal)
	if err != nil {
		if !isGlobal {
			return contextureerrors.Wrap(err, "load project configuration").
				WithSuggestions("Run 'contexture init' to initialize a new project")
		}
		return err
	}

	var currentDir string
	if !isGlobal {
		currentDir, err = os.Getwd()
		if err != nil {
			return contextureerrors.Wrap(err, "get current directory")
		}
	}

	// Find rules to remove
	var rulesToRemove []string
	var notFound []string

	for _, ruleID := range ruleIDs {
		// Try both simple format and full format for matching
		switch {
		case c.projectManager.HasRule(config, ruleID):
			rulesToRemove = append(rulesToRemove, ruleID)
		case c.projectManager.HasRule(config, fmt.Sprintf("[contexture:%s]", ruleID)):
			// If the rule exists in full format, add it in the format it's stored
			rulesToRemove = append(rulesToRemove, fmt.Sprintf("[contexture:%s]", ruleID))
		default:
			notFound = append(notFound, ruleID)
		}
	}

	// Report not found rules
	if len(notFound) > 0 && !isJSONMode {
		log.Warn("Rules not found in configuration", "rules", notFound)
	}

	if len(rulesToRemove) == 0 {
		// Handle output format when no rules to remove
		outputFormat := output.Format(cmd.String("output"))
		outputManager, err := output.NewManager(outputFormat)
		if err != nil {
			return contextureerrors.Wrap(err, "create output manager")
		}

		// Write empty output
		metadata := output.RemoveMetadata{
			RulesRemoved: []string{},
		}

		err = outputManager.WriteRulesRemove(metadata)
		if err != nil {
			return contextureerrors.Wrap(err, "write remove output")
		}

		// For default format, also show the log message
		if !isJSONMode {
			log.Info("No rules to remove")
		}

		return nil
	}

	// Capture variables for display BEFORE removing rules from configuration
	ruleVariablesMap := make(map[string]map[string]any)
	for _, ruleID := range rulesToRemove {
		for _, configRule := range config.Rules {
			if configRule.ID == ruleID ||
				configRule.ID == fmt.Sprintf("[contexture:%s]", ruleID) ||
				strings.TrimPrefix(strings.TrimSuffix(configRule.ID, "]"), "[contexture:") == ruleID {
				ruleVariablesMap[ruleID] = configRule.Variables
				break
			}
		}
	}

	// Remove rules from configuration
	var removedRules []string
	for _, ruleID := range rulesToRemove {
		err := c.projectManager.RemoveRule(config, ruleID)
		if err != nil {
			log.Error("Failed to remove rule from configuration", "rule", ruleID, "error", err)
			continue
		}
		removedRules = append(removedRules, ruleID)
	}

	if len(removedRules) == 0 {
		return contextureerrors.ValidationErrorf("rules", "failed to remove any rules")
	}

	// Automatically clean outputs (skip for global)
	if !isGlobal {
		err = c.removeFromOutputs(ctx, config, removedRules, currentDir)
		if err != nil {
			log.Warn("Failed to clean some outputs", "error", err)
		}

		// Clean up empty directories after removing rules, similar to build command
		targetFormats := config.GetEnabledFormats()
		for _, formatConfig := range targetFormats {
			format, err := c.registry.CreateFormat(formatConfig.Type, afero.NewOsFs(), nil)
			if err != nil {
				log.Warn("Failed to create format for cleanup", "format", formatConfig.Type, "error", err)
				continue
			}
			if err := format.CleanupEmptyDirectories(&formatConfig); err != nil {
				log.Warn("Failed to cleanup empty directories", "format", formatConfig.Type, "error", err)
			}
		}
	}

	// Save updated configuration
	if isGlobal {
		err = c.projectManager.SaveGlobalConfig(config)
	} else {
		location := c.projectManager.GetConfigLocation(currentDir, false)
		err = c.projectManager.SaveConfig(config, location, currentDir)
	}
	if err != nil {
		return contextureerrors.Wrap(err, "save configuration")
	}

	// Auto-rebuild after removing global rules (similar to add command)
	if isGlobal && !isJSONMode {
		if err := c.rebuildAfterGlobalRemove(ctx); err != nil {
			log.Warn("Failed to auto-rebuild after removing global rules", "error", err)
			// Don't fail the remove - rules were removed successfully
		}
	}

	// Handle output format
	outputManager, err := output.NewManager(outputFormat)
	if err != nil {
		return contextureerrors.Wrap(err, "create output manager")
	}

	// Collect removed rule IDs for output
	var removedRuleIDs []string
	for _, ruleID := range removedRules {
		// Extract display-friendly rule ID using domain package logic
		displayRuleID := domain.ExtractRulePath(ruleID)
		if displayRuleID == "" {
			displayRuleID = ruleID
		}
		removedRuleIDs = append(removedRuleIDs, displayRuleID)
	}

	// Write output using the appropriate format
	metadata := output.RemoveMetadata{
		RulesRemoved: removedRuleIDs,
	}

	err = outputManager.WriteRulesRemove(metadata)
	if err != nil {
		return contextureerrors.Wrap(err, "write remove output")
	}

	// For default format, also display the detailed information
	if outputFormat == output.FormatDefault {
		theme := ui.DefaultTheme()
		successStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Success)

		// Show proper singular/plural success message
		var successMessage string
		if len(removedRules) == 1 {
			successMessage = "Rule removed successfully!"
		} else {
			successMessage = "Rules removed successfully!"
		}

		fmt.Println(successStyle.Render(successMessage))

		// List the removed rules like in add command
		for _, ruleID := range removedRules {
			// Extract display-friendly rule ID using domain package logic
			displayRuleID := domain.ExtractRulePath(ruleID)
			if displayRuleID == "" {
				displayRuleID = ruleID
			}

			var variables map[string]any
			var defaultVars map[string]any

			// Parse rule to extract source information and get variables
			if parsed, err := c.ruleFetcher.ParseRuleID(ruleID); err == nil {
				// Get the configured variables we captured before removal
				variables = ruleVariablesMap[ruleID]

				// Fetch the full rule to get default variables
				if fetchedRule, fetchErr := c.ruleFetcher.FetchRule(context.Background(), ruleID); fetchErr == nil {
					defaultVars = fetchedRule.DefaultVariables
				}

				// Display the short rule ID
				fmt.Printf("  %s\n", displayRuleID)

				// Show source information for custom source rules (like in ls command)
				if parsed.Source != "" && domain.IsCustomGitSource(parsed.Source) {
					darkGrayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
					sourceDisplay := domain.FormatSourceForDisplay(parsed.Source, parsed.Ref)
					fmt.Printf("    %s\n", darkGrayStyle.Render(sourceDisplay))
				}
			} else {
				fmt.Printf("  %s\n", displayRuleID)
			}

			// Show variables only if they differ from defaults
			if rule.ShouldDisplayVariables(variables, defaultVars) {
				if variablesJSON, err := json.Marshal(variables); err == nil {
					fmt.Printf("    Variables: %s\n", string(variablesJSON))
				}
			}
		}
	}

	log.Debug("Rules removed", "count", len(removedRules))

	return nil
}

// removeFromOutputs removes rules from generated format outputs
func (c *RemoveCommand) removeFromOutputs(
	_ context.Context,
	config *domain.Project,
	ruleIDs []string,
	_ string,
) error {
	var errors []string

	for _, formatConfig := range config.GetEnabledFormats() {
		format, err := c.registry.CreateFormat(formatConfig.Type, afero.NewOsFs(), nil)
		if err != nil {
			errors = append(errors, contextureerrors.Wrap(err, "create format").Error())
			continue
		}

		for _, ruleID := range ruleIDs {
			err := format.Remove(ruleID, &formatConfig)
			if err != nil {
				errors = append(errors, contextureerrors.Wrap(err, "remove from format").Error())
			}
		}
	}

	if len(errors) > 0 {
		return contextureerrors.ValidationErrorf("cleanup", "cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// rebuildAfterGlobalRemove rebuilds outputs after removing global rules
// This is similar to rebuildAfterGlobalAdd in add.go
func (c *RemoveCommand) rebuildAfterGlobalRemove(ctx context.Context) error {
	// Load global config to get the remaining global rules
	globalConfigResult, err := c.projectManager.LoadGlobalConfig()
	if err != nil {
		return contextureerrors.Wrap(err, "load global config")
	}

	globalConfig := globalConfigResult.Config

	// Get formats from global config that have native user rules support
	var userFormats []domain.FormatConfig
	for _, formatConfig := range globalConfig.Formats {
		if !formatConfig.Enabled {
			continue
		}

		caps, exists := c.registry.GetCapabilities(formatConfig.Type)
		if !exists {
			continue
		}

		// Only include formats that support native user rules
		if caps.SupportsUserRules && caps.UserRulesPath != "" {
			// Create format config for user rules generation
			userFormatConfig := formatConfig
			userDir := filepath.Dir(caps.UserRulesPath)
			userFormatConfig.BaseDir = userDir
			userFormatConfig.IsUserRules = true
			userFormats = append(userFormats, userFormatConfig)
		}
	}

	// Regenerate to native user rules locations with remaining rules
	if len(userFormats) > 0 {
		userConfig := &domain.Project{}
		*userConfig = *globalConfig
		userConfig.Rules = globalConfig.Rules

		log.Debug("Regenerating global rules to user locations after removal", "formats", len(userFormats))
		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, userConfig, userFormats, "global"); err != nil {
			log.Warn("Failed to regenerate global rules to user locations", "error", err)
		}
	}

	// Check if we're in a project context
	currentDir, err := os.Getwd()
	if err != nil {
		// Can't determine current directory, skip project rebuild
		log.Debug("Cannot determine current directory, skipping project rebuild", "error", err)
		return nil
	}

	// Try to load project config with merged rules
	merged, projectErr := c.projectManager.LoadConfigMergedWithLocalRules(currentDir)
	if projectErr != nil {
		// Not in a project, this is normal - global rules were removed successfully
		//nolint:nilerr // Intentionally returning nil - not being in a project is not an error
		return nil
	}

	// We're in a project - rebuild project files based on UserRulesMode per format
	log.Debug("Detected project context, rebuilding based on UserRulesMode")

	// Separate user (global) rules from project rules
	var projectRules, userRules []domain.RuleRef
	for _, rws := range merged.MergedRules {
		if rws.Source == domain.RuleSourceUser {
			userRules = append(userRules, rws.RuleRef)
		} else {
			projectRules = append(projectRules, rws.RuleRef)
		}
	}

	// Get enabled formats from project config
	targetFormats := merged.Project.GetEnabledFormats()
	if len(targetFormats) == 0 {
		return nil
	}

	// Group formats by which rules they need based on UserRulesMode
	var nativeOnlyFormats []domain.FormatConfig // UserRulesNative - project rules only
	var mergedFormats []domain.FormatConfig     // UserRulesProject - project + user rules
	var disabledFormats []domain.FormatConfig   // UserRulesDisabled - project rules only

	for _, formatConfig := range targetFormats {
		mode := formatConfig.GetEffectiveUserRulesMode()
		switch mode {
		case domain.UserRulesNative:
			nativeOnlyFormats = append(nativeOnlyFormats, formatConfig)
		case domain.UserRulesProject:
			mergedFormats = append(mergedFormats, formatConfig)
		case domain.UserRulesDisabled:
			disabledFormats = append(disabledFormats, formatConfig)
		}
	}

	// Generate for formats with UserRulesNative or UserRulesDisabled (project rules only)
	projectOnlyFormats := make([]domain.FormatConfig, 0, len(nativeOnlyFormats)+len(disabledFormats))
	projectOnlyFormats = append(projectOnlyFormats, nativeOnlyFormats...)
	projectOnlyFormats = append(projectOnlyFormats, disabledFormats...)
	if len(projectOnlyFormats) > 0 {
		config := &domain.Project{}
		*config = *merged.Project
		config.Rules = projectRules

		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, config, projectOnlyFormats, "project"); err != nil {
			return err
		}
	}

	// Generate for formats with UserRulesProject (project + user rules)
	if len(mergedFormats) > 0 {
		config := &domain.Project{}
		*config = *merged.Project
		config.Rules = append(append([]domain.RuleRef{}, projectRules...), userRules...)

		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, config, mergedFormats, "project"); err != nil {
			return err
		}
	}

	return nil
}

// RemoveAction is the CLI action handler for the remove command
func RemoveAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	ruleIDs := cmd.Args().Slice()
	removeCmd := NewRemoveCommand(deps)

	// If no rule IDs provided, show helpful error message
	if len(ruleIDs) == 0 {
		return contextureerrors.ValidationErrorf("rule-ids",
			"no rule IDs provided\n\nUsage:\n  contexture rules remove [rule-id...]\n\nExamples:\n  # Remove specific rules (simple format)\n  contexture rules remove languages/go/code-organization testing/unit-tests\n  \n  # Remove rules (full format)\n  contexture rules remove \"[contexture:languages/go/advanced-patterns]\" \"[contexture:security/input-validation]\"\n  \n  # Remove from custom source\n  contexture rules remove my/custom-rule\n\nTo see installed rules:\n  Use 'contexture rules list' to see currently installed rules\n  \nRun 'contexture rules remove --help' for more options")
	}

	return removeCmd.Execute(ctx, cmd, ruleIDs)
}
