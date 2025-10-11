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

const (
	// queryBatchSize is the number of rules to fetch and process at once
	// This reduces memory usage by avoiding loading all rules at once
	queryBatchSize = 50
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

	// Get limit for early exit optimization
	limit := cmd.Int("limit")

	// Fetch and filter rules with streaming and early exit
	filtered, err := c.fetchAndFilterRules(ctx, cmd, queryStr, useExpr, limit)
	if err != nil {
		return err
	}

	// Output results
	return c.outputResults(filtered, queryStr, useExpr, cmd)
}

// fetchAndFilterRules fetches rules with streaming and early exit when limit is reached
func (c *QueryCommand) fetchAndFilterRules(ctx context.Context, cmd *cli.Command, queryStr string, useExpr bool, limit int) ([]*domain.Rule, error) {
	providerFilter := cmd.StringSlice("provider")

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

	// Collect matching rules with early exit
	matchedRules := make([]*domain.Rule, 0, limit)

	// Iterate through providers
	for _, provider := range providers {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return matchedRules, err
		}

		// Check if we've reached the limit
		if limit > 0 && len(matchedRules) >= limit {
			break
		}

		// List available rules from this provider
		ruleIDs, err := c.ruleFetcher.ListAvailableRules(ctx, provider.URL, provider.DefaultBranch)
		if err != nil {
			fmt.Printf("Warning: Failed to list rules from provider %s: %v\n", provider.Name, err)
			continue
		}

		// Process rules in batches
		for i := 0; i < len(ruleIDs); i += queryBatchSize {
			// Check for context cancellation
			if err := ctx.Err(); err != nil {
				return matchedRules, err
			}

			// Check if we've reached the limit
			if limit > 0 && len(matchedRules) >= limit {
				break
			}

			// Determine batch end
			end := i + queryBatchSize
			if end > len(ruleIDs) {
				end = len(ruleIDs)
			}
			batch := ruleIDs[i:end]

			// Fetch and filter batch
			for _, ruleID := range batch {
				// Check for context cancellation periodically
				if err := ctx.Err(); err != nil {
					return matchedRules, err
				}

				// Early exit if we've hit the limit
				if limit > 0 && len(matchedRules) >= limit {
					break
				}

				// Construct full rule ID with provider context
				fullRuleID := "@" + provider.Name + "/" + ruleID

				fetchedRule, err := c.ruleFetcher.FetchRule(ctx, fullRuleID)
				if err != nil {
					// Skip rules that fail to fetch
					continue
				}

				// Apply filter immediately
				var matches bool
				if useExpr {
					matches, err = c.evaluator.EvaluateExpr(fetchedRule, queryStr)
					if err != nil {
						return nil, err
					}
				} else {
					matches = c.evaluator.MatchesText(fetchedRule, queryStr)
				}

				// Add to results if it matches
				if matches {
					matchedRules = append(matchedRules, fetchedRule)
				}
			}
		}
	}

	return matchedRules, nil
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
