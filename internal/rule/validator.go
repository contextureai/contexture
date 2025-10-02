// Package rule provides rule validation using the validation system.
package rule

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/validation"
)

// Validator for rule validation operations
type Validator interface {
	ValidateRule(rule *domain.Rule) *domain.ValidationResult
	ValidateRules(rules []*domain.Rule) *validation.BatchResult
	ValidateRuleContent(content string) *domain.ValidationResult
	ValidateRuleID(ruleID string) error
	ValidateGitURL(gitURL string) error
}

// DefaultValidator implements rule validation
type DefaultValidator struct {
	v validation.Validator
}

// NewValidator creates a new rule validator
// Panics if validator creation fails, as this indicates a programming error
// that should be caught during development/testing
func NewValidator() Validator {
	v, err := validation.NewValidator()
	if err != nil {
		panic(fmt.Sprintf("failed to create validator: %v", err))
	}
	return &DefaultValidator{v: v}
}

// ValidateRule performs validation on a single rule
func (d *DefaultValidator) ValidateRule(rule *domain.Rule) *domain.ValidationResult {
	log.Debug("Validating rule", "id", rule.ID)
	return d.v.ValidateRule(rule)
}

// ValidateRules validates a batch of rules
func (d *DefaultValidator) ValidateRules(rules []*domain.Rule) *validation.BatchResult {
	log.Debug("Validating rule batch", "count", len(rules))
	return d.v.ValidateRules(rules)
}

// ValidateRuleContent validates raw rule content
func (d *DefaultValidator) ValidateRuleContent(content string) *domain.ValidationResult {
	result := &domain.ValidationResult{
		Valid:    true,
		Errors:   make([]error, 0),
		Warnings: make([]domain.ValidationWarning, 0),
	}

	if strings.TrimSpace(content) == "" {
		result.AddError("content", "Content cannot be empty", "EMPTY_CONTENT")
	}

	return result
}

// ValidateRuleID validates a rule ID format
func (d *DefaultValidator) ValidateRuleID(ruleID string) error {
	return d.v.ValidateRuleID(ruleID)
}

// ValidateGitURL validates a git repository URL
func (d *DefaultValidator) ValidateGitURL(gitURL string) error {
	return d.v.ValidateGitURL(gitURL)
}
