package rule

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/cache"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
)

// GitRuleFetcher handles fetching rules from Git repositories
type GitRuleFetcher struct {
	fs       afero.Fs
	parser   Parser
	cache    *cache.SimpleCache
	repo     git.Repository
	idParser IDParser
}

// NewGitRuleFetcher creates a new Git rule fetcher
func NewGitRuleFetcher(
	fs afero.Fs,
	parser Parser,
	cache *cache.SimpleCache,
	repo git.Repository,
	idParser IDParser,
) *GitRuleFetcher {
	return &GitRuleFetcher{
		fs:       fs,
		parser:   parser,
		cache:    cache,
		repo:     repo,
		idParser: idParser,
	}
}

// FetchRule fetches a single rule from Git
func (f *GitRuleFetcher) FetchRule(ctx context.Context, ruleID string) (*domain.Rule, error) {
	log.Debug("Fetching rule from Git", "ruleID", ruleID)

	// Parse the rule ID
	parsed, err := f.idParser.ParseRuleID(ruleID)
	if err != nil {
		return nil, err
	}

	// Get repository from cache (clones if needed)
	repoDir, err := f.cache.GetRepository(ctx, parsed.Source, parsed.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Construct the full path to the rule file
	rulePath := filepath.Join(repoDir, parsed.RulePath+".md")

	// Read the rule file (EAFP - Easier to Ask Forgiveness than Permission)
	data, err := afero.ReadFile(f.fs, rulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, contextureerrors.WithOp("FetchRule", contextureerrors.ErrRuleNotFound)
		}
		return nil, fmt.Errorf("failed to read rule file: %w", err)
	}

	metadata := Metadata{
		ID:        ruleID,
		FilePath:  parsed.RulePath,
		Source:    parsed.Source,
		Ref:       parsed.Ref,
		Variables: parsed.Variables,
	}
	rule, err := f.parser.ParseRule(string(data), metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rule: %w", err)
	}

	// Add source information
	rule.ID = ruleID
	rule.Source = parsed.Source
	rule.Ref = parsed.Ref
	rule.FilePath = parsed.RulePath

	// Merge variables from parsed ID with rule variables
	if len(parsed.Variables) > 0 {
		if rule.Variables == nil {
			rule.Variables = make(map[string]any)
		}
		for key, value := range parsed.Variables {
			rule.Variables[key] = value
		}
	}

	log.Debug("Successfully fetched rule from Git", "ruleID", ruleID)
	return rule, nil
}

// FetchRuleAtCommit fetches a rule at a specific commit hash
func (f *GitRuleFetcher) FetchRuleAtCommit(ctx context.Context, ruleID, commitHash string) (*domain.Rule, error) {
	log.Debug("Fetching rule at specific commit", "ruleID", ruleID, "commitHash", commitHash)

	// Parse the rule ID
	parsed, err := f.idParser.ParseRuleID(ruleID)
	if err != nil {
		return nil, err
	}

	// Get repository from cache (clones if needed)
	repoDir, err := f.cache.GetRepository(ctx, parsed.Source, parsed.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Read the rule file at the specific commit using the injected repository implementation
	repo := f.repo
	if repo == nil {
		repo = git.NewRepository(f.fs)
	}
	ruleFilePath := parsed.RulePath + ".md"
	data, err := repo.GetFileAtCommit(repoDir, ruleFilePath, commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to read rule file at commit %s: %w", commitHash, err)
	}

	metadata := Metadata{
		ID:        ruleID,
		FilePath:  parsed.RulePath,
		Source:    parsed.Source,
		Ref:       parsed.Ref,
		Variables: parsed.Variables,
	}
	rule, err := f.parser.ParseRule(string(data), metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rule: %w", err)
	}

	// Add source information
	rule.ID = ruleID
	rule.Source = parsed.Source
	rule.Ref = parsed.Ref
	rule.FilePath = parsed.RulePath

	// Merge variables from parsed ID with rule variables
	if len(parsed.Variables) > 0 {
		if rule.Variables == nil {
			rule.Variables = make(map[string]any)
		}
		for key, value := range parsed.Variables {
			rule.Variables[key] = value
		}
	}

	log.Debug("Successfully fetched rule at commit", "ruleID", ruleID, "commitHash", commitHash)
	return rule, nil
}

// ListAvailableRules lists all available rules in a Git repository
func (f *GitRuleFetcher) ListAvailableRules(
	ctx context.Context,
	source, ref string,
) ([]string, error) {
	log.Debug("Listing available rules from Git", "source", source, "ref", ref)

	// Get repository from cache
	repoDir, err := f.cache.GetRepository(ctx, source, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Walk the repository directory to find rule files
	var ruleFiles []string
	err = afero.Walk(f.fs, repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-files and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip README.md and other non-rule files
		if strings.HasSuffix(strings.ToLower(path), "readme.md") {
			return nil
		}

		// Get relative path from repository directory
		relPath, err := filepath.Rel(repoDir, path)
		if err != nil {
			return err
		}

		// Remove .md extension to get rule ID path
		ruleID := strings.TrimSuffix(relPath, ".md")

		// Convert backslashes to forward slashes for consistency
		ruleID = strings.ReplaceAll(ruleID, "\\", "/")

		ruleFiles = append(ruleFiles, ruleID)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk repository directory: %w", err)
	}

	log.Debug("Found rules in Git repository", "count", len(ruleFiles))
	return ruleFiles, nil
}

// ListAvailableRulesWithStructure lists all available rules in a Git repository with folder structure
func (f *GitRuleFetcher) ListAvailableRulesWithStructure(
	ctx context.Context,
	source, ref string,
) (*domain.RuleNode, error) {
	log.Debug("Listing available rules with structure from Git", "source", source, "ref", ref)

	// Get the flat list of rules first
	ruleFiles, err := f.ListAvailableRules(ctx, source, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to list available rules: %w", err)
	}

	// Build the tree structure
	tree := domain.NewRuleTree(ruleFiles)

	log.Debug("Built rule tree structure", "total_rules", len(ruleFiles))
	return tree, nil
}
