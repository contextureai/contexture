// Package rule provides rule processing functionality
package rule

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
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
				if commitSourceFetcher, ok := fetcher.(CommitAwareFetcher); ok {
					rule, err = commitSourceFetcher.FetchRuleAtCommitWithSource(ctx, ref.ID, ref.CommitHash, ref.Source)
				} else if commitFetcher, ok := fetcher.(CommitFetcher); ok {
					rule, err = commitFetcher.FetchRuleAtCommit(ctx, ref.ID, ref.CommitHash)
				} else {
					// Fallback to regular fetch for other fetcher types
					rule, err = fetcher.FetchRule(ctx, ref.ID)
				}
			} else {
				// Regular fetch without commit hash, use source-aware method if available
				if sourceFetcher, ok := fetcher.(SourceAwareFetcher); ok {
					rule, err = sourceFetcher.FetchRuleWithSource(ctx, ref.ID, ref.Source)
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
			errors = append(errors, contextureerrors.Wrap(res.err, "rule "+res.id))
			continue
		}
		rules = append(rules, res.rule)
	}

	if len(errors) > 0 {
		return nil, contextureerrors.Wrap(combineErrors(errors), "failed to fetch some rules")
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
				contextureerrors.Wrap(err, "failed to parse rule "+ruleContent.Metadata.ID),
			)
			result.Skipped = append(result.Skipped, ruleContent.Metadata.ID)
			continue
		}
		result.Rules = append(result.Rules, rule)
	}

	return result
}

// ShouldDisplayVariables determines if variables should be displayed in UI
// by checking if any variable differs from its default value
func ShouldDisplayVariables(variables, defaults map[string]any) bool {
	if len(variables) == 0 {
		return false
	}

	// If no defaults are provided, show variables if they exist
	if len(defaults) == 0 {
		return true
	}

	// Check if any variable differs from its default
	for key, value := range variables {
		defaultValue, hasDefault := defaults[key]
		if !hasDefault {
			// Variable exists but has no default - should display
			return true
		}
		if !reflect.DeepEqual(value, defaultValue) {
			// Variable differs from default - should display
			return true
		}
	}

	// All variables match their defaults - no need to display
	return false
}

// FilterNonDefaultVariables returns only variables that differ from their defaults
func FilterNonDefaultVariables(variables, defaults map[string]any) map[string]any {
	if len(variables) == 0 {
		return nil
	}

	filtered := make(map[string]any)

	for key, value := range variables {
		defaultValue, hasDefault := defaults[key]
		if !hasDefault || !reflect.DeepEqual(value, defaultValue) {
			filtered[key] = value
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}

// SortRulesDeterministically sorts rules by their normalized ID for consistent output
// This ensures that generated files have the same order every time, preventing
// unnecessary git diffs when rules are added/removed.
func SortRulesDeterministically(rules []*domain.Rule, parser IDParser) []*domain.Rule {
	if len(rules) == 0 {
		return rules
	}

	// Create a sorted copy
	sorted := make([]*domain.Rule, len(rules))
	copy(sorted, rules)

	// Sort by normalized ID (case-insensitive, alphabetical)
	// Use stable sort to preserve order for rules with same normalized ID
	sort.SliceStable(sorted, func(i, j int) bool {
		idI := normalizeRuleIDForSort(sorted[i].ID, parser)
		idJ := normalizeRuleIDForSort(sorted[j].ID, parser)
		return idI < idJ
	})

	return sorted
}

// normalizeRuleIDForSort extracts the path from a rule ID and normalizes it for sorting
func normalizeRuleIDForSort(ruleID string, parser IDParser) string {
	if parser == nil {
		// Fallback: just use the ID lowercased
		return strings.ToLower(ruleID)
	}

	// Parse the rule ID to extract the path
	parsed, err := parser.ParseRuleID(ruleID)
	if err != nil {
		// If parsing fails, use the ID as-is (lowercased)
		return strings.ToLower(ruleID)
	}

	// Use the rule path (normalized to lowercase for case-insensitive sorting)
	return strings.ToLower(parsed.RulePath)
}
