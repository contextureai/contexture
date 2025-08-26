// Package rule provides rule processing functionality
package rule

import (
	"context"
	"fmt"
	"sync"

	"github.com/contextureai/contexture/internal/domain"
)

// FetchRulesParallel fetches rules in parallel with a worker pool
func FetchRulesParallel(
	ctx context.Context,
	fetcher Fetcher,
	ruleRefs []domain.RuleRef,
	maxWorkers int,
) ([]*domain.Rule, error) {
	if maxWorkers <= 0 {
		maxWorkers = domain.DefaultMaxWorkers
	}

	type result struct {
		rule *domain.Rule
		err  error
		id   string
	}

	results := make(chan result, len(ruleRefs))
	semaphore := make(chan struct{}, maxWorkers)

	var wg sync.WaitGroup

	// Start workers
	for _, ruleRef := range ruleRefs {
		wg.Add(1)
		go func(ref domain.RuleRef) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Fetch rule - use specific commit hash if available
			var rule *domain.Rule
			var err error

			if ref.CommitHash != "" {
				// Try to fetch at specific commit if the fetcher supports it
				if compositeFetcher, ok := fetcher.(*CompositeFetcher); ok {
					rule, err = compositeFetcher.FetchRuleAtCommitWithSource(ctx, ref.ID, ref.CommitHash, ref.Source)
				} else {
					// Fallback to regular fetch for other fetcher types
					rule, err = fetcher.FetchRule(ctx, ref.ID)
				}
			} else {
				// Regular fetch without commit hash, use source-aware method if available
				if compositeFetcher, ok := fetcher.(*CompositeFetcher); ok {
					rule, err = compositeFetcher.FetchRuleWithSource(ctx, ref.ID, ref.Source)
				} else {
					// Fallback to regular fetch for other fetcher types
					rule, err = fetcher.FetchRule(ctx, ref.ID)
				}
			}

			if err != nil {
				results <- result{rule: nil, err: err, id: ref.ID}
				return
			}

			// Merge variables from RuleRef with fetched rule
			// RuleRef variables take precedence over rule variables
			if len(ref.Variables) > 0 {
				if rule.Variables == nil {
					rule.Variables = make(map[string]any)
				}
				for key, value := range ref.Variables {
					rule.Variables[key] = value
				}
			}

			results <- result{rule: rule, err: nil, id: ref.ID}
		}(ruleRef)
	}

	// Close results when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var rules []*domain.Rule
	var errors []error

	for res := range results {
		if res.err != nil {
			errors = append(errors, fmt.Errorf("rule %s: %w", res.id, res.err))
			continue
		}
		rules = append(rules, res.rule)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to fetch some rules: %w", combineErrors(errors))
	}

	return rules, nil
}

// ExtractRuleIDsFromContent finds all rule IDs in the given content
func ExtractRuleIDsFromContent(content string) []string {
	re := domain.RuleIDExtractPatternRegex
	matches := re.FindAllString(content, -1)

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string

	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			unique = append(unique, match)
		}
	}

	return unique
}

// GitFetcher is an alias for backward compatibility
type GitFetcher = GitRuleFetcher

// Content represents a rule's content and metadata for bulk parsing
type Content struct {
	Content  string
	Metadata Metadata
}

// ParseResult represents the result of parsing multiple rules
type ParseResult struct {
	Rules   []*domain.Rule
	Errors  []error
	Skipped []string
}

// ParseRules parses multiple rules and returns results with error handling
func ParseRules(parser Parser, rules []Content) *ParseResult {
	result := &ParseResult{
		Rules:   make([]*domain.Rule, 0, len(rules)),
		Errors:  make([]error, 0),
		Skipped: make([]string, 0),
	}

	for _, ruleContent := range rules {
		rule, err := parser.ParseRule(ruleContent.Content, ruleContent.Metadata)
		if err != nil {
			result.Errors = append(
				result.Errors,
				fmt.Errorf("failed to parse rule %s: %w", ruleContent.Metadata.ID, err),
			)
			result.Skipped = append(result.Skipped, ruleContent.Metadata.ID)
			continue
		}
		result.Rules = append(result.Rules, rule)
	}

	return result
}
