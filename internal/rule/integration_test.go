package rule

import (
	"context"
	"fmt"
	"testing"

	"github.com/contextureai/contexture/internal/provider"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestRuleProcessingPipeline tests the complete rule processing workflow
func TestRuleProcessingPipeline(t *testing.T) {
	t.Parallel()
	// Set up components
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)

	// Create test rule content
	testRuleContent := `---
title: "Integration Test Rule"
description: "A rule for testing the complete processing pipeline"
tags: ["test", "integration", "pipeline"]
trigger:
  type: "always"
  description: "Always apply this rule"
languages: ["go", "javascript"]
frameworks: ["gin", "express"]
scope: "global"
global: true
variables:
  severity: "high"
  category: "testing"
---

# {{.rule.title}}

This rule tests the complete processing pipeline.

## Variables
- Severity: {{.severity}}
- Category: {{.category}}
- Environment: {{default_if_empty .environment "development"}}

## Languages
Applies to: {{join_and .rule.languages}}

## Frameworks  
Works with: {{join_and .rule.frameworks}}

## Generated Info
Generated on {{.date}} by {{.contexture.engine}}.`

	// Mock the Clone method to create test data
	mockRepo.On("Clone", mock.Anything, "https://github.com/test/repo.git", mock.AnythingOfType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			tempPath := args.Get(2).(string)
			// Create test data structure in the cloned repo
			_ = fs.MkdirAll(tempPath+"/core/integration", 0o755)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/integration/test-rule.md",
				[]byte(testRuleContent),
				0o644,
			)

			// Additional test rules for multiple rules test
			rule2Content := `---
title: "Second Test Rule"
description: "Another test rule"
tags: ["test", "second"]
---

# {{.rule.title}}

This is the second rule with variable: {{default_if_empty .testVar "default"}}.`

			rule3Content := `---
title: "Third Test Rule"
description: "Yet another test rule"
tags: ["test", "third"]
---

# {{.rule.title}}

Third rule content.`

			_ = afero.WriteFile(
				fs,
				tempPath+"/core/integration/rule2.md",
				[]byte(rule2Content),
				0o644,
			)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/integration/rule3.md",
				[]byte(rule3Content),
				0o644,
			)
		}).
		Return(nil)

	// Create components
	fetcher := NewFetcher(fs, mockRepo, FetcherConfig{
		DefaultURL: "https://github.com/test/repo.git",
	}, provider.NewRegistry())

	parser := NewParser()
	processor := NewProcessor()
	validator := NewValidator()

	ctx := context.Background()
	ruleID := "[contexture:core/integration/test-rule]"

	t.Run("complete pipeline - fetch, parse, validate, process", func(t *testing.T) {
		// Step 1: Fetch the rule (now includes parsing)
		rule, err := fetcher.FetchRule(ctx, ruleID)
		require.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, ruleID, rule.ID)
		assert.Contains(t, rule.Content, "This rule tests the complete processing pipeline")

		// Check parsed fields from frontmatter
		assert.Equal(t, "Integration Test Rule", rule.Title)
		assert.Equal(
			t,
			"A rule for testing the complete processing pipeline",
			rule.Description,
		)
		assert.Equal(t, []string{"test", "integration", "pipeline"}, rule.Tags)
		assert.Equal(t, []string{"go", "javascript"}, rule.Languages)
		assert.Equal(t, []string{"gin", "express"}, rule.Frameworks)

		// Check trigger
		assert.NotNil(t, rule.Trigger)
		assert.Equal(t, domain.TriggerAlways, rule.Trigger.Type)

		// Check variables
		assert.NotNil(t, rule.Variables)
		assert.Equal(t, "high", rule.Variables["severity"])
		assert.Equal(t, "testing", rule.Variables["category"])

		// Step 2: Validate the parsed rule
		validationResult := validator.ValidateRule(rule)
		assert.True(t, validationResult.Valid)
		assert.Empty(t, validationResult.Errors)

		// Step 4: Process the rule with context
		ruleContext := &domain.RuleContext{
			Variables: map[string]any{
				"environment": "test",
			},
			Globals: map[string]any{
				"project": "contexture",
			},
		}

		processedRule, err := processor.ProcessRule(rule, ruleContext)
		require.NoError(t, err)
		assert.NotNil(t, processedRule)

		// Verify content is raw (not processed) - format layer will handle templating
		content := processedRule.Content
		assert.Contains(t, content, "# {{.rule.title}}")
		assert.Contains(t, content, "Severity: {{.severity}}")
		assert.Contains(t, content, "Category: {{.category}}")
		assert.Contains(t, content, "Environment: {{default_if_empty .environment")

		// Verify variables are available for format processing
		assert.NotNil(t, processedRule.Variables)
		assert.Equal(t, "high", processedRule.Variables["severity"])
		assert.Equal(t, "testing", processedRule.Variables["category"])
		assert.Equal(t, "test", processedRule.Variables["environment"])

		// Verify attribution
		assert.NotEmpty(t, processedRule.Attribution)
		assert.Contains(t, processedRule.Attribution, ruleID)
	})

	t.Run("pipeline with multiple rules", func(t *testing.T) {
		// Test files are already created by the Clone mock

		ruleIDs := []string{
			"[contexture:core/integration/test-rule]",
			"[contexture:core/integration/rule2]",
			"[contexture:core/integration/rule3]",
		}

		// Fetch multiple rules (now includes parsing)
		rules, err := fetcher.FetchRules(ctx, ruleIDs)
		require.NoError(t, err)
		assert.Len(t, rules, 3)

		// Validate all rules
		batchValidationResult := validator.ValidateRules(rules)
		assert.Equal(t, 3, batchValidationResult.TotalRules)
		assert.Equal(t, 3, batchValidationResult.ValidRules)
		assert.True(t, batchValidationResult.AllValid)

		// Process all rules
		ruleContext := &domain.RuleContext{
			Variables: map[string]any{
				"testVar": "processed",
			},
		}

		processedRules, err := processor.ProcessRulesWithContext(ctx, rules, ruleContext)
		require.NoError(t, err)
		assert.Len(t, processedRules, 3)

		// Verify each processed rule
		titles := make(map[string]bool)
		for _, processed := range processedRules {
			titles[processed.Rule.Title] = true
			assert.NotEmpty(t, processed.Content)
			assert.NotEmpty(t, processed.Attribution)
		}

		assert.True(t, titles["Integration Test Rule"])
		assert.True(t, titles["Second Test Rule"])
		assert.True(t, titles["Third Test Rule"])
	})

	t.Run("pipeline error handling", func(t *testing.T) {
		// Test with non-existent rule
		_, err := fetcher.FetchRule(ctx, "[contexture:nonexistent/rule]")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rule not found")

		// Test with invalid rule content
		invalidContent := `---
title: "Invalid Rule"
# Missing required fields
---

Invalid content`

		rule := &domain.Rule{
			ID:      "[contexture:test/invalid]",
			Content: invalidContent,
		}

		metadata := Metadata{
			ID:       rule.ID,
			FilePath: "/invalid.md",
			Source:   "test",
			Ref:      "main",
		}

		_, err = parser.ParseRule(rule.Content, metadata)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation")

		// Test validation with invalid rule
		invalidRule := &domain.Rule{
			ID:          "",
			Title:       "",
			Description: "",
			Tags:        []string{},
			Content:     "",
		}

		validationResult := validator.ValidateRule(invalidRule)
		assert.False(t, validationResult.Valid)
		assert.NotEmpty(t, validationResult.Errors)

		// Test processing with invalid template syntax - processor should not error since it doesn't process templates
		ruleWithBadTemplate := &domain.Rule{
			ID:          "[contexture:test/bad-template]",
			Title:       "Bad Template",
			Description: "Rule with bad template",
			Tags:        []string{"test"},
			Content:     "{{if unclosed}} no end",
		}

		processedBadRule, err := processor.ProcessRule(ruleWithBadTemplate, &domain.RuleContext{})
		require.NoError(t, err) // No error expected - template validation happens in format layer
		assert.NotNil(t, processedBadRule)
		assert.Equal(t, ruleWithBadTemplate.Content, processedBadRule.Content) // Raw content preserved
	})
}

// TestRuleServiceIntegration tests a higher-level service that combines all components
func TestRuleServiceIntegration(t *testing.T) {
	t.Parallel()
	// Create a service that combines all rule processing components
	service := NewRuleService()

	assert.NotNil(t, service)

	// Test service creation
	assert.IsType(t, &RuleService{}, service)
}

// RuleService combines all rule processing components into a unified service
type RuleService struct {
	fetcher   Fetcher
	parser    Parser
	processor Processor
	validator Validator
}

// NewRuleService creates a new rule service with all components
func NewRuleService() *RuleService {
	fs := afero.NewOsFs()
	gitRepo := git.NewRepository(fs)

	return &RuleService{
		fetcher:   NewFetcher(fs, gitRepo, FetcherConfig{}, provider.NewRegistry()),
		parser:    NewParser(),
		processor: NewProcessor(),
		validator: NewValidator(),
	}
}

// ProcessRuleByID is a convenience method that handles the full pipeline
func (s *RuleService) ProcessRuleByID(
	ctx context.Context,
	ruleID string,
	ruleContext *domain.RuleContext,
) (*domain.ProcessedRule, error) {
	// Fetch
	rule, err := s.fetcher.FetchRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rule: %w", err)
	}

	// Parse
	metadata := Metadata{
		ID:       rule.ID,
		FilePath: rule.FilePath,
		Source:   rule.Source,
		Ref:      rule.Ref,
	}

	parsedRule, err := s.parser.ParseRule(rule.Content, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rule: %w", err)
	}

	// Validate
	validationResult := s.validator.ValidateRule(parsedRule)
	if !validationResult.Valid {
		return nil, fmt.Errorf(
			"rule validation failed: %w",
			combineErrors(validationResult.Errors),
		)
	}

	// Process
	processedRule, err := s.processor.ProcessRule(parsedRule, ruleContext)
	if err != nil {
		return nil, fmt.Errorf("failed to process rule: %w", err)
	}

	return processedRule, nil
}

// ProcessMultipleRules processes multiple rules with comprehensive error handling
func (s *RuleService) ProcessMultipleRules(
	ctx context.Context,
	ruleIDs []string,
	ruleContext *domain.RuleContext,
) (*RuleProcessingResult, error) {
	result := &RuleProcessingResult{
		Processed: make([]*domain.ProcessedRule, 0),
		Failed:    make([]RuleProcessingError, 0),
		Summary:   ProcessingSummary{},
	}

	// Fetch all rules
	rules, err := s.fetcher.FetchRules(ctx, ruleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules: %w", err)
	}

	result.Summary.TotalRequested = len(ruleIDs)
	result.Summary.SuccessfullyFetched = len(rules)

	// Process each rule through the pipeline
	for _, rule := range rules {
		// Parse
		metadata := Metadata{
			ID:       rule.ID,
			FilePath: rule.FilePath,
			Source:   rule.Source,
			Ref:      rule.Ref,
		}

		parsedRule, err := s.parser.ParseRule(rule.Content, metadata)
		if err != nil {
			result.Failed = append(result.Failed, RuleProcessingError{
				RuleID: rule.ID,
				Stage:  "parsing",
				Error:  err,
			})
			continue
		}

		// Validate
		validationResult := s.validator.ValidateRule(parsedRule)
		if !validationResult.Valid {
			result.Failed = append(result.Failed, RuleProcessingError{
				RuleID: rule.ID,
				Stage:  "validation",
				Error: fmt.Errorf(
					"validation errors: %w",
					combineErrors(validationResult.Errors),
				),
			})
			continue
		}

		// Process
		processedRule, err := s.processor.ProcessRule(parsedRule, ruleContext)
		if err != nil {
			result.Failed = append(result.Failed, RuleProcessingError{
				RuleID: rule.ID,
				Stage:  "processing",
				Error:  err,
			})
			continue
		}

		result.Processed = append(result.Processed, processedRule)
	}

	result.Summary.SuccessfullyProcessed = len(result.Processed)
	result.Summary.FailedProcessing = len(result.Failed)

	return result, nil
}

// RuleProcessingResult represents the result of processing multiple rules
type RuleProcessingResult struct {
	Processed []*domain.ProcessedRule
	Failed    []RuleProcessingError
	Summary   ProcessingSummary
}

// RuleProcessingError represents an error that occurred during rule processing
type RuleProcessingError struct {
	RuleID string
	Stage  string // "fetching", "parsing", "validation", "processing"
	Error  error
}

// ProcessingSummary provides a summary of the processing operation
type ProcessingSummary struct {
	TotalRequested        int
	SuccessfullyFetched   int
	SuccessfullyProcessed int
	FailedProcessing      int
}

// SuccessRate returns the success rate as a percentage
func (r *RuleProcessingResult) SuccessRate() float64 {
	if r.Summary.TotalRequested == 0 {
		return 100.0
	}
	return float64(r.Summary.SuccessfullyProcessed) / float64(r.Summary.TotalRequested) * 100.0
}

// HasErrors returns true if any errors occurred during processing
func (r *RuleProcessingResult) HasErrors() bool {
	return len(r.Failed) > 0
}
