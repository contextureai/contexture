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

// FailsafeValidator is a fallback validator that fails gracefully
type FailsafeValidator struct {
	err error
}

// NewValidator creates a new rule validator
func NewValidator() Validator {
	v, err := validation.NewValidator()
	if err != nil {
		// Log the error and return a validator that always fails gracefully
		log.Error("Failed to create validator, using fallback", "error", err)
		return &FailsafeValidator{err: err}
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

// FailsafeValidator methods - all return validation errors due to initialization failure

// ValidateRule returns a validation error for FailsafeValidator
func (f *FailsafeValidator) ValidateRule(_ *domain.Rule) *domain.ValidationResult {
	result := &domain.ValidationResult{
		Valid:    false,
		Errors:   []error{fmt.Errorf("validator initialization failed: %w", f.err)},
		Warnings: make([]domain.ValidationWarning, 0),
	}
	return result
}

// ValidateRules returns a validation error for FailsafeValidator
func (f *FailsafeValidator) ValidateRules(rules []*domain.Rule) *validation.BatchResult {
	return &validation.BatchResult{
		TotalRules:  len(rules),
		ValidRules:  0,
		Results:     []*validation.Result{},
		AllValid:    false,
		HasWarnings: false,
	}
}

// ValidateRuleContent returns a validation error for FailsafeValidator
func (f *FailsafeValidator) ValidateRuleContent(_ string) *domain.ValidationResult {
	result := &domain.ValidationResult{
		Valid:    false,
		Errors:   []error{fmt.Errorf("validator initialization failed: %w", f.err)},
		Warnings: make([]domain.ValidationWarning, 0),
	}
	return result
}

// ValidateRuleID returns a validation error for FailsafeValidator
func (f *FailsafeValidator) ValidateRuleID(_ string) error {
	return fmt.Errorf("validator initialization failed: %w", f.err)
}

// ValidateGitURL returns a validation error for FailsafeValidator
func (f *FailsafeValidator) ValidateGitURL(_ string) error {
	return fmt.Errorf("validator initialization failed: %w", f.err)
}
