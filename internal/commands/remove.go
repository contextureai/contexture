// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/git"
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
	// Show command header
	fmt.Println(ui.CommandHeader("remove"))
	fmt.Println()
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
	if len(notFound) > 0 {
		log.Warn("Rules not found in configuration", "rules", notFound)
	}

	if len(rulesToRemove) == 0 {
		log.Info("No rules to remove")
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

	// Automatically clean outputs unless --keep-outputs is specified
	if !cmd.Bool("keep-outputs") {
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
	}

	// Save updated configuration
	err = c.projectManager.SaveConfig(configResult.Config, configResult.Location, currentDir)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Success message
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
		// Extract simple rule ID for display (remove [contexture:] wrapper if present)
		displayRuleID := ruleID
		var variables map[string]any
		var defaultVars map[string]any

		if strings.HasPrefix(ruleID, "[contexture:") {
			// Parse to extract just the path component
			parsed, err := c.ruleFetcher.ParseRuleID(ruleID)
			if err == nil && parsed.RulePath != "" {
				displayRuleID = parsed.RulePath

				// Get the configured variables we captured before removal
				variables = ruleVariablesMap[ruleID]

				// Fetch the full rule to get default variables
				if fetchedRule, fetchErr := c.ruleFetcher.FetchRule(context.Background(), ruleID); fetchErr == nil {
					defaultVars = fetchedRule.DefaultVariables
				}
			}
		}
		fmt.Printf("  %s\n", displayRuleID)

		// Show variables only if they differ from defaults
		if rule.ShouldDisplayVariables(variables, defaultVars) {
			if variablesJSON, err := json.Marshal(variables); err == nil {
				fmt.Printf("    Variables: %s\n", string(variablesJSON))
			}
		}
	}

	log.Debug("Rules removed",
		"count", len(removedRules),
		"config_path", configResult.Path)

	return nil
}

// ShowInstalledRules lists installed rules from project configuration for removal
func (c *RemoveCommand) ShowInstalledRules(ctx context.Context, cmd *cli.Command) error {
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := c.projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	if len(configResult.Config.Rules) == 0 {
		log.Info("No rules are currently installed in this project")
		return nil
	}

	// Extract rule IDs from configuration, but only include removable rules
	var installedRuleIDs []string
	var localRuleCount int
	for _, ruleRef := range configResult.Config.Rules {
		// Skip local rules - they can't be removed through the configuration
		if ruleRef.Source == "local" {
			localRuleCount++
			continue
		}

		// Use the full rule ID as stored in configuration
		// Don't attempt to convert to "simple format" as it breaks custom source rules
		installedRuleIDs = append(installedRuleIDs, ruleRef.ID)
	}

	// Inform user about local rules if any exist
	if localRuleCount > 0 {
		log.Debug("Local rules found", "count", localRuleCount, "note", "Local rules are files on disk and cannot be removed through configuration")
	}

	// No filters available since --search flag was removed

	if len(installedRuleIDs) == 0 {
		if localRuleCount > 0 {
			log.Info("No remote rules found", "local_rules", localRuleCount)
			fmt.Printf("No remote rules can be removed from this project.\n")
			fmt.Printf("Found %d local rules (files on disk) - to remove local rules, delete the .md files from the rules directory.\n", localRuleCount)
		} else {
			log.Info("No installed rules found")
		}
		return nil
	}

	// Sort rules for consistent output
	sort.Strings(installedRuleIDs)

	// Always show interactive mode since no flags are available for non-interactive filtering
	return c.showInteractiveRulesForRemoving(ctx, cmd, installedRuleIDs, configResult)
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

// showInteractiveRulesForRemoving shows an interactive searchable list of rules for removal
func (c *RemoveCommand) showInteractiveRulesForRemoving(
	ctx context.Context,
	cmd *cli.Command,
	ruleIDs []string,
	configResult *domain.ConfigResult,
) error {
	// Fetch detailed rule information with spinner
	detailSpinner := ui.NewBubblesSpinner("Loading rule details")
	fmt.Print(detailSpinner.View())

	var detailedRules []*domain.Rule
	for _, ruleID := range ruleIDs {
		// Use source-aware fetching to ensure we fetch from remote repository
		type sourceAwareFetcher interface {
			FetchRuleWithSource(ctx context.Context, ruleID, source string) (*domain.Rule, error)
		}

		var rule *domain.Rule
		var fetchErr error

		if compositeFetcher, ok := c.ruleFetcher.(sourceAwareFetcher); ok {
			// Use the source-aware method to force remote fetching (empty source = default/remote)
			rule, fetchErr = compositeFetcher.FetchRuleWithSource(ctx, ruleID, "")
		} else {
			// Fallback to regular fetch
			rule, fetchErr = c.ruleFetcher.FetchRule(ctx, ruleID)
		}

		if fetchErr != nil {
			log.Warn("Failed to fetch rule details", "rule", ruleID, "error", fetchErr)
			// Create a minimal rule object for display if fetch fails
			minimalRule := &domain.Rule{
				ID:          ruleID, // Use the actual rule ID as stored in config
				Title:       ruleID,
				Description: "Failed to load rule details",
				Tags:        []string{"unknown"},
			}
			detailedRules = append(detailedRules, minimalRule)
			continue
		}

		// Find the configured variables for this rule from the project configuration
		var configuredVariables map[string]any
		for _, configRule := range configResult.Config.Rules {
			// Use the centralized matching logic from project manager
			if configRule.ID == ruleID {
				configuredVariables = configRule.Variables
				break
			}
		}

		// Merge configured variables with the fetched rule
		// The fetched rule already has DefaultVariables populated, we just need to set Variables
		if configuredVariables != nil {
			rule.Variables = configuredVariables
		}

		detailedRules = append(detailedRules, rule)
	}
	detailSpinner.Stop("") // Stop without message

	if len(detailedRules) == 0 {
		fmt.Println("No rule details could be loaded.")
		return nil
	}

	// Show interactive list for rule selection
	selectedRules, err := c.showRuleListSelectionForRemoval(detailedRules)
	if err != nil {
		return err
	}

	if len(selectedRules) == 0 {
		log.Info("No rules selected for removal")
		return nil
	}

	// Process selected rules using existing remove logic
	return c.Execute(ctx, cmd, selectedRules)
}

// showRuleListSelectionForRemoval shows an interactive bubbles list for rule removal selection
func (c *RemoveCommand) showRuleListSelectionForRemoval(rules []*domain.Rule) ([]string, error) {
	return showInteractiveRuleSelection(rules, "Select Rules to Remove")
}

// RemoveAction is the CLI action handler for the remove command
func RemoveAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	ruleIDs := cmd.Args().Slice()
	removeCmd := NewRemoveCommand(deps)

	// If no rule IDs provided, show installed rules for removal
	if len(ruleIDs) == 0 {
		return removeCmd.ShowInstalledRules(ctx, cmd)
	}

	return removeCmd.Execute(ctx, cmd, ruleIDs)
}
