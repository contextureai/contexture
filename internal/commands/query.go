// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/output"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/provider"
	"github.com/contextureai/contexture/internal/query"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/urfave/cli/v3"
)

// QueryCommand implements the query command
type QueryCommand struct {
	projectManager   *project.Manager
	ruleFetcher      rule.Fetcher
	providerRegistry *provider.Registry
	evaluator        query.Evaluator
}

// NewQueryCommand creates a new query command
func NewQueryCommand(deps *dependencies.Dependencies) *QueryCommand {
	return &QueryCommand{
		projectManager:   project.NewManager(deps.FS),
		ruleFetcher:      rule.NewFetcher(deps.FS, newOpenRepository(deps.FS), rule.FetcherConfig{}, deps.ProviderRegistry),
		providerRegistry: deps.ProviderRegistry,
		evaluator:        query.NewEvaluator(),
	}
}

// Execute runs the query command
func (c *QueryCommand) Execute(ctx context.Context, cmd *cli.Command) error {
	// Get query string
	queryStr := cmd.Args().First()
	if queryStr == "" {
		return contextureerrors.ValidationErrorf("query", "query string is required")
	}

	// Determine query mode
	useExpr := cmd.Bool("expr")

	// Get current directory and load configuration (for providers)
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	// Load project config to get custom providers
	configResult, err := c.projectManager.LoadConfig(currentDir)
	if err == nil {
		// Load providers from project config into registry
		if err := c.providerRegistry.LoadFromProject(configResult.Config); err != nil {
			return contextureerrors.Wrap(err, "load providers")
		}
	}
	// If config doesn't exist, that's okay - we'll just use default providers

	// Fetch rules
	rules, err := c.fetchRules(ctx, cmd)
	if err != nil {
		return err
	}

	// Apply appropriate filter
	var filtered []*domain.Rule
	if useExpr {
		filtered, err = c.filterWithExpr(rules, queryStr)
		if err != nil {
			return err
		}
	} else {
		filtered = c.filterWithText(rules, queryStr)
	}

	// Apply limit if specified
	limit := cmd.Int("limit")
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Output results
	return c.outputResults(filtered, queryStr, useExpr, cmd)
}

// fetchRules fetches rules based on command flags
func (c *QueryCommand) fetchRules(ctx context.Context, cmd *cli.Command) ([]*domain.Rule, error) {
	providerFilter := cmd.StringSlice("provider")

	// Fetch rules from all providers or specific providers
	return c.fetchProviderRules(ctx, providerFilter)
}

// fetchProviderRules fetches rules from all configured providers or specific ones
func (c *QueryCommand) fetchProviderRules(ctx context.Context, providerFilter []string) ([]*domain.Rule, error) {
	// Get list of providers to query
	providers := c.providerRegistry.ListProviders()

	// Filter providers if specified
	if len(providerFilter) > 0 {
		filtered := make([]*domain.Provider, 0)
		for _, p := range providers {
			for _, filter := range providerFilter {
				if p.Name == filter {
					filtered = append(filtered, p)
					break
				}
			}
		}
		providers = filtered
	}

	if len(providers) == 0 {
		return nil, contextureerrors.ValidationErrorf("provider", "no providers found")
	}

	// Fetch rules from each provider
	allRules := make([]*domain.Rule, 0)
	for _, provider := range providers {
		// List available rules from this provider
		ruleIDs, err := c.ruleFetcher.ListAvailableRules(ctx, provider.URL, provider.DefaultBranch)
		if err != nil {
			fmt.Printf("Warning: Failed to list rules from provider %s: %v\n", provider.Name, err)
			continue
		}

		// Fetch each rule
		for _, ruleID := range ruleIDs {
			// Construct full rule ID with provider context
			// Always include the provider prefix to ensure proper routing
			fullRuleID := "@" + provider.Name + "/" + ruleID

			fetchedRule, err := c.ruleFetcher.FetchRule(ctx, fullRuleID)
			if err != nil {
				// Skip rules that fail to fetch
				continue
			}

			allRules = append(allRules, fetchedRule)
		}
	}

	return allRules, nil
}

// filterWithText filters rules using simple text matching
func (c *QueryCommand) filterWithText(rules []*domain.Rule, queryStr string) []*domain.Rule {
	filtered := make([]*domain.Rule, 0)
	for _, r := range rules {
		if c.evaluator.MatchesText(r, queryStr) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// filterWithExpr filters rules using expr expression evaluation
func (c *QueryCommand) filterWithExpr(rules []*domain.Rule, exprStr string) ([]*domain.Rule, error) {
	filtered := make([]*domain.Rule, 0)
	for _, r := range rules {
		matches, err := c.evaluator.EvaluateExpr(r, exprStr)
		if err != nil {
			return nil, err
		}
		if matches {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// outputResults outputs the filtered rules
func (c *QueryCommand) outputResults(rules []*domain.Rule, queryStr string, useExpr bool, cmd *cli.Command) error {
	// Determine output format
	outputFormat := output.Format(cmd.String("output"))

	// Create output manager
	outputMgr, err := output.NewManager(outputFormat)
	if err != nil {
		return err
	}

	// Prepare query type string
	queryType := "text"
	if useExpr {
		queryType = "expr"
	}

	// Prepare metadata
	metadata := output.QueryMetadata{
		Query:        queryStr,
		QueryType:    queryType,
		TotalResults: len(rules),
	}

	// Write output in requested format
	return outputMgr.WriteQueryResults(rules, metadata)
}

// QueryAction is the CLI action handler for the query command
func QueryAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	queryCmd := NewQueryCommand(deps)
	return queryCmd.Execute(ctx, cmd)
}
