// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/git"
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
}

// NewRemoveCommand creates a new remove command
func NewRemoveCommand(deps *dependencies.Dependencies) *RemoveCommand {
	return &RemoveCommand{
		projectManager: project.NewManager(deps.FS),
		registry:       format.GetDefaultRegistry(deps.FS),
		ruleFetcher:    rule.NewFetcher(deps.FS, git.NewRepository(deps.FS), rule.FetcherConfig{}),
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
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := c.projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	// Find rules to remove
	var rulesToRemove []string
	var notFound []string

	for _, ruleID := range ruleIDs {
		// Try both simple format and full format for matching
		switch {
		case c.projectManager.HasRule(configResult.Config, ruleID):
			rulesToRemove = append(rulesToRemove, ruleID)
		case c.projectManager.HasRule(configResult.Config, fmt.Sprintf("[contexture:%s]", ruleID)):
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
			return fmt.Errorf("failed to create output manager: %w", err)
		}

		// Write empty output
		metadata := output.RemoveMetadata{
			RulesRemoved: []string{},
		}

		err = outputManager.WriteRulesRemove(metadata)
		if err != nil {
			return fmt.Errorf("failed to write remove output: %w", err)
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
		for _, configRule := range configResult.Config.Rules {
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
		err := c.projectManager.RemoveRule(configResult.Config, ruleID)
		if err != nil {
			log.Error("Failed to remove rule from configuration", "rule", ruleID, "error", err)
			continue
		}
		removedRules = append(removedRules, ruleID)
	}

	if len(removedRules) == 0 {
		return fmt.Errorf("failed to remove any rules")
	}

	// Automatically clean outputs
	err = c.removeFromOutputs(ctx, configResult.Config, removedRules, currentDir)
	if err != nil {
		log.Warn("Failed to clean some outputs", "error", err)
	}

	// Clean up empty directories after removing rules, similar to build command
	targetFormats := configResult.Config.GetEnabledFormats()
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

	// Save updated configuration
	err = c.projectManager.SaveConfig(configResult.Config, configResult.Location, currentDir)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Handle output format
	outputManager, err := output.NewManager(outputFormat)
	if err != nil {
		return fmt.Errorf("failed to create output manager: %w", err)
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
		return fmt.Errorf("failed to write remove output: %w", err)
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

	log.Debug("Rules removed",
		"count", len(removedRules),
		"config_path", configResult.Path)

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
			errors = append(errors, fmt.Sprintf("failed to create format %s: %v",
				formatConfig.Type, err))
			continue
		}

		for _, ruleID := range ruleIDs {
			err := format.Remove(ruleID, &formatConfig)
			if err != nil {
				errors = append(errors, fmt.Sprintf("failed to remove %s from %s: %v",
					ruleID, formatConfig.Type, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// RemoveAction is the CLI action handler for the remove command
func RemoveAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	ruleIDs := cmd.Args().Slice()
	removeCmd := NewRemoveCommand(deps)

	// If no rule IDs provided, show helpful error message
	if len(ruleIDs) == 0 {
		return fmt.Errorf("no rule IDs provided\n\nUsage:\n  contexture rules remove [rule-id...]\n\nExamples:\n  # Remove specific rules (simple format)\n  contexture rules remove languages/go/code-organization testing/unit-tests\n  \n  # Remove rules (full format)\n  contexture rules remove \"[contexture:languages/go/advanced-patterns]\" \"[contexture:security/input-validation]\"\n  \n  # Remove from custom source\n  contexture rules remove my/custom-rule\n\nTo see installed rules:\n  Use 'contexture rules list' to see currently installed rules\n  \nRun 'contexture rules remove --help' for more options")
	}

	return removeCmd.Execute(ctx, cmd, ruleIDs)
}
