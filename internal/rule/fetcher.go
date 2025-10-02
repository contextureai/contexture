package rule

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/cache"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
)

const (
	defaultRulesRepo = "https://github.com/contextureai/rules.git"
	defaultBranch    = "main"
)

// CompositeFetcher implements rule fetching with separated concerns
type CompositeFetcher struct {
	gitFetcher   *GitRuleFetcher
	localFetcher Fetcher
	idParser     IDParser
	maxWorkers   int
}

// NewFetcher creates a new fetcher with separated components
func NewFetcher(fs afero.Fs, repository git.Repository, config FetcherConfig) Fetcher {
	if config.DefaultURL == "" {
		config.DefaultURL = defaultRulesRepo
	}
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = domain.DefaultMaxWorkers
	}

	parser := NewParser()
	idParser := NewRuleIDParser(config.DefaultURL)
	simpleCache := cache.NewSimpleCache(fs, repository)

	gitFetcher := NewGitRuleFetcher(fs, parser, simpleCache, repository, idParser)
	localFetcher := NewLocalFetcher(fs, ".")

	return &CompositeFetcher{
		gitFetcher:   gitFetcher,
		localFetcher: localFetcher,
		idParser:     idParser,
		maxWorkers:   config.MaxWorkers,
	}
}

// FetchRule fetches a single rule by ID
func (f *CompositeFetcher) FetchRule(ctx context.Context, ruleID string) (*domain.Rule, error) {
	// Check if it's a local path
	if isLocalPath(ruleID) {
		return f.localFetcher.FetchRule(ctx, ruleID)
	}

	// Otherwise use Git fetcher
	return f.gitFetcher.FetchRule(ctx, ruleID)
}

// FetchRuleWithSource fetches a single rule by ID with explicit source information
func (f *CompositeFetcher) FetchRuleWithSource(ctx context.Context, ruleID, source string) (*domain.Rule, error) {
	// If source is explicitly "local", use local fetcher
	if source == "local" {
		return f.localFetcher.FetchRule(ctx, ruleID)
	}

	// For any other source (including empty/default), use git fetcher
	return f.gitFetcher.FetchRule(ctx, ruleID)
}

// FetchRuleAtCommit fetches a rule at a specific commit hash
func (f *CompositeFetcher) FetchRuleAtCommit(ctx context.Context, ruleID, commitHash string) (*domain.Rule, error) {
	// Check if it's a local path (local fetcher doesn't support commit hashes)
	if isLocalPath(ruleID) {
		return f.localFetcher.FetchRule(ctx, ruleID) // Fallback to regular fetch
	}

	// Use Git fetcher for commit-specific fetch
	return f.gitFetcher.FetchRuleAtCommit(ctx, ruleID, commitHash)
}

// FetchRuleAtCommitWithSource fetches a rule at a specific commit hash with explicit source information
func (f *CompositeFetcher) FetchRuleAtCommitWithSource(ctx context.Context, ruleID, commitHash, source string) (*domain.Rule, error) {
	// If source is explicitly "local", use local fetcher (local fetcher doesn't support commit hashes)
	if source == "local" {
		return f.localFetcher.FetchRule(ctx, ruleID) // Fallback to regular fetch
	}

	// For any other source (including empty/default), use git fetcher for commit-specific fetch
	return f.gitFetcher.FetchRuleAtCommit(ctx, ruleID, commitHash)
}

// FetchRules fetches multiple rules concurrently
func (f *CompositeFetcher) FetchRules(
	ctx context.Context,
	ruleIDs []string,
) ([]*domain.Rule, error) {
	if len(ruleIDs) == 0 {
		return []*domain.Rule{}, nil
	}

	log.Debug("Fetching multiple rules", "count", len(ruleIDs))

	// Use a worker pool for concurrent fetching with bounded concurrency
	type result struct {
		rule *domain.Rule
		err  error
		id   string
	}

	maxWorkers := f.maxWorkers
	if maxWorkers <= 0 {
		maxWorkers = domain.DefaultMaxWorkers
	}

	semaphore := make(chan struct{}, maxWorkers)
	resultChan := make(chan result, len(ruleIDs))
	var wg sync.WaitGroup

	for _, ruleID := range ruleIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			rule, err := f.FetchRule(ctx, id)

			select {
			case resultChan <- result{rule: rule, err: err, id: id}:
			case <-ctx.Done():
				return
			}
		}(ruleID)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var rules []*domain.Rule
	var errors []error

	for {
		select {
		case res, ok := <-resultChan:
			if !ok {
				if len(errors) > 0 {
					return nil, fmt.Errorf("failed to fetch %d rules: %w", len(errors), combineErrors(errors))
				}
				log.Debug("Successfully fetched all rules", "count", len(rules))
				return rules, nil
			}
			if res.err != nil {
				errors = append(errors, fmt.Errorf("failed to fetch rule %s: %w", res.id, res.err))
			} else {
				rules = append(rules, res.rule)
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while fetching rules: %w", ctx.Err())
		}
	}
}

// ParseRuleID delegates to the ID parser
func (f *CompositeFetcher) ParseRuleID(ruleID string) (*domain.ParsedRuleID, error) {
	return f.idParser.ParseRuleID(ruleID)
}

// ListAvailableRules lists all available rules in a repository
func (f *CompositeFetcher) ListAvailableRules(
	ctx context.Context,
	source, branch string,
) ([]string, error) {
	// Use default source if not specified
	if source == "" {
		source = defaultRulesRepo
	}
	if branch == "" {
		branch = defaultBranch
	}

	// Use the git fetcher to clone the repository and list rules
	return f.gitFetcher.ListAvailableRules(ctx, source, branch)
}

// ListAvailableRulesWithStructure lists all available rules in a repository with folder structure
func (f *CompositeFetcher) ListAvailableRulesWithStructure(
	ctx context.Context,
	source, branch string,
) (*domain.RuleNode, error) {
	// Use default source if not specified
	if source == "" {
		source = defaultRulesRepo
	}
	if branch == "" {
		branch = defaultBranch
	}

	// Use the git fetcher to get structured list
	return f.gitFetcher.ListAvailableRulesWithStructure(ctx, source, branch)
}

// isLocalPath checks if a path is a local file path
func isLocalPath(path string) bool {
	// Check if it's an absolute path or starts with ./ or ../
	if filepath.IsAbs(path) ||
		strings.HasPrefix(path, "./") ||
		strings.HasPrefix(path, "../") {
		return true
	}

	// Check if it's NOT a contexture rule ID format (simple relative paths are local)
	// Contexture rule IDs start with [contexture: or contain special chars like :, [, ]
	if !strings.HasPrefix(path, "[contexture:") &&
		!strings.Contains(path, ":") &&
		!strings.Contains(path, "[") &&
		!strings.Contains(path, "]") &&
		!strings.Contains(path, "{") { // JSON5 variables
		return true
	}

	return false
}
