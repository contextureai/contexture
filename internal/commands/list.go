// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/output"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/provider"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/urfave/cli/v3"
)

// ListCommand implements the list command
type ListCommand struct {
	projectManager   *project.Manager
	ruleFetcher      rule.Fetcher
	registry         *format.Registry
	providerRegistry *provider.Registry
}

// NewListCommand creates a new list command
func NewListCommand(deps *dependencies.Dependencies) *ListCommand {
	return &ListCommand{
		projectManager:   project.NewManager(deps.FS),
		ruleFetcher:      rule.NewFetcher(deps.FS, newOpenRepository(deps.FS), rule.FetcherConfig{}, deps.ProviderRegistry),
		registry:         format.GetDefaultRegistry(deps.FS),
		providerRegistry: deps.ProviderRegistry,
	}
}

// Execute runs the list command
func (c *ListCommand) Execute(ctx context.Context, cmd *cli.Command) error {
	return c.listInstalledRules(ctx, cmd)
}

// listInstalledRules lists rules configured in the current project
func (c *ListCommand) listInstalledRules(ctx context.Context, cmd *cli.Command) error {
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	configResult, err := c.projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		return contextureerrors.Wrap(err, "load project configuration").
			WithSuggestions("Run 'contexture init' to initialize a new project")
	}

	config := configResult.Config

	// Load providers from project config into registry
	if err := c.providerRegistry.LoadFromProject(config); err != nil {
		return contextureerrors.Wrap(err, "load providers")
	}

	// Fetch the actual rules from the rule references
	rules, err := c.fetchRulesFromReferences(ctx, config.Rules)
	if err != nil {
		return contextureerrors.Wrap(err, "fetch rules")
	}

	// Use simple rule list display
	return c.showRuleList(rules, cmd)
}

// fetchRulesFromReferences fetches the actual rule content from rule references
func (c *ListCommand) fetchRulesFromReferences(
	ctx context.Context,
	ruleRefs []domain.RuleRef,
) ([]*domain.Rule, error) {
	if len(ruleRefs) == 0 {
		return []*domain.Rule{}, nil
	}

	rules := make([]*domain.Rule, 0, len(ruleRefs))
	var lastError error

	for _, ruleRef := range ruleRefs {
		// Convert RuleRef to rule ID format expected by the fetcher
		ruleID := ruleRef.ID
		if ruleID == "" {
			// If no ID, skip this rule
			continue
		}

		// Fetch the rule content
		rule, err := c.ruleFetcher.FetchRule(ctx, ruleID)
		if err != nil {
			lastError = err
			// Log the error but continue with other rules
			fmt.Printf("Warning: Failed to fetch rule %s: %v\n", ruleID, err)
			continue
		}

		// Merge configured variables with the fetched rule
		// The fetched rule has DefaultVariables, but we need to set Variables to the configured ones
		if ruleRef.Variables != nil {
			rule.Variables = ruleRef.Variables
		}

		rules = append(rules, rule)
	}

	// If we failed to fetch any rules and had errors, return the last error
	if len(rules) == 0 && len(ruleRefs) > 0 && lastError != nil {
		return nil, contextureerrors.Wrap(lastError, "fetch rules")
	}

	return rules, nil
}

// showRuleList displays rules using the configured output format
func (c *ListCommand) showRuleList(ruleList []*domain.Rule, cmd *cli.Command) error {
	// Determine output format
	outputFormat := output.Format(cmd.String("output"))

	// Create output manager
	outputMgr, err := output.NewManager(outputFormat)
	if err != nil {
		return err
	}

	totalRules := len(ruleList)
	pattern := cmd.String("pattern")

	// Prepare metadata - for now, both counts are the same since filtering
	// happens inside the writers. This will be refined when we extract
	// the filtering logic for better JSON metadata
	metadata := output.ListMetadata{
		Pattern:       pattern,
		TotalRules:    totalRules,
		FilteredRules: totalRules, // This will be corrected by the writers

	}

	// Write output in requested format
	return outputMgr.WriteRulesList(ruleList, metadata)
}

// ListAction is the CLI action handler for the list command
func ListAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	listCmd := NewListCommand(deps)
	return listCmd.Execute(ctx, cmd)
}
