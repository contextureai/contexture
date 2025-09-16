// Package rule provides rule processing functionality
package rule

import (
	"context"

	"github.com/contextureai/contexture/internal/domain"
)

// Fetcher interface for rule fetching operations
type Fetcher interface {
	FetchRule(ctx context.Context, ruleID string) (*domain.Rule, error)
	FetchRules(ctx context.Context, ruleIDs []string) ([]*domain.Rule, error)
	ParseRuleID(ruleID string) (*domain.ParsedRuleID, error)
	ListAvailableRules(ctx context.Context, source, ref string) ([]string, error)
	ListAvailableRulesWithStructure(ctx context.Context, source, ref string) (*domain.RuleNode, error)
}

// SourceAwareFetcher can fetch a rule while honoring an explicit source hint.
type SourceAwareFetcher interface {
	FetchRuleWithSource(ctx context.Context, ruleID, source string) (*domain.Rule, error)
}

// CommitAwareFetcher can fetch a rule pinned to a specific commit with source awareness.
type CommitAwareFetcher interface {
	FetchRuleAtCommitWithSource(ctx context.Context, ruleID, commitHash, source string) (*domain.Rule, error)
}

// CommitFetcher can fetch a rule pinned to a specific commit without source hints.
type CommitFetcher interface {
	FetchRuleAtCommit(ctx context.Context, ruleID, commitHash string) (*domain.Rule, error)
}

// Parser interface for rule parsing operations
type Parser interface {
	ParseRule(content string, metadata Metadata) (*domain.Rule, error)
	ParseContent(content string) (frontmatter map[string]any, body string, err error)
	ValidateRule(rule *domain.Rule) error
}

// Processor interface for rule processing operations
type Processor interface {
	ProcessRule(rule *domain.Rule, context *domain.RuleContext) (*domain.ProcessedRule, error)
	ProcessRules(rules []*domain.Rule, context *domain.RuleContext) ([]*domain.ProcessedRule, error)
	ProcessRulesWithContext(
		ctx context.Context,
		rules []*domain.Rule,
		context *domain.RuleContext,
	) ([]*domain.ProcessedRule, error)
	ProcessTemplate(content string, variables map[string]any) (string, error)
	GenerateAttribution(rule *domain.Rule) string
}

// FetcherConfig configures the rule fetcher
type FetcherConfig struct {
	DefaultURL string
	MaxWorkers int
}

// Metadata contains metadata about a rule file
type Metadata struct {
	ID        string
	FilePath  string
	Source    string
	Ref       string
	Variables map[string]any // Variables from parsed rule ID
}
