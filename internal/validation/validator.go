// Package validation provides validation for the Contexture CLI.
// It consolidates all validation logic into a single, consistent approach using
// the validator/v10 library with custom tags and error formatting.
package validation

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/go-playground/validator/v10"
)

// Constants for validation
const (
	// MaxRuleIDLength is the maximum allowed length for a rule ID
	MaxRuleIDLength = 200

	// MinRuleIDLength is the minimum allowed length for a rule ID
	MinRuleIDLength = 1

	// MaxTitleLength is the maximum allowed length for a rule title
	MaxTitleLength = 80

	// MaxDescriptionLength is the maximum allowed length for a rule description
	MaxDescriptionLength = 200

	// MinTagCount is the minimum number of tags required
	MinTagCount = 1

	// MaxTagCount is the maximum number of tags allowed
	MaxTagCount = 10

	// ValidationOperation is the operation name for validation errors
	ValidationOperation = "validate"
)

// Validator provides validation for all Contexture types
type Validator interface {
	// ValidateRule validates a rule
	ValidateRule(rule *domain.Rule) *domain.ValidationResult

	// ValidateRules validates multiple rules
	ValidateRules(rules []*domain.Rule) *BatchResult

	// ValidateProject validates a project configuration
	ValidateProject(config *domain.Project) error

	// ValidateFormatConfig validates a format configuration
	ValidateFormatConfig(config *domain.FormatConfig) error

	// ValidateRuleRef validates a rule reference
	ValidateRuleRef(ref domain.RuleRef) error

	// ValidateRuleID validates a rule ID format
	ValidateRuleID(ruleID string) error

	// ValidateGitURL validates a git repository URL
	ValidateGitURL(gitURL string) error

	// ValidateWithContext validates with additional context
	ValidateWithContext(ctx context.Context, value any, name string) error
}

// BatchResult represents the result of validating multiple rules
type BatchResult struct {
	TotalRules  int
	ValidRules  int
	Results     []*Result
	AllValid    bool
	HasWarnings bool
}

// Result represents the result of validating a single rule
type Result struct {
	RuleID string
	Valid  bool
	Errors []string
}

// defaultValidator implements the Validator interface
type defaultValidator struct {
	v *validator.Validate
}

// NewValidator creates a new validator
func NewValidator() (Validator, error) {
	v := validator.New()

	uv := &defaultValidator{v: v}

	// Register custom validation tags
	customValidators := map[string]validator.Func{
		"ruleref":        uv.validateRuleRef,
		"ruleid":         uv.validateRuleIDTag,
		"formattype":     uv.validateFormatType,
		"giturl":         uv.validateGitURLTag,
		"contexturepath": uv.validateContexturePath,
	}

	for tag, fn := range customValidators {
		if err := v.RegisterValidation(tag, fn); err != nil {
			return nil, contextureerrors.WithOpf(
				"register validation",
				"failed to register %s validation: %w", tag, err,
			)
		}
	}

	// Use JSON field names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return uv, nil
}

// ValidateRule validates a single rule
func (v *defaultValidator) ValidateRule(rule *domain.Rule) *domain.ValidationResult {
	result := &domain.ValidationResult{
		Valid:    true,
		Errors:   make([]error, 0),
		Warnings: make([]domain.ValidationWarning, 0),
	}

	if rule == nil {
		result.AddError("rule", "Rule cannot be nil", "NULL_RULE")
		return result
	}

	// Use struct validation with tags
	if err := v.v.Struct(rule); err != nil {
		result.Valid = false
		v.addStructValidationErrors(err, result)
	}

	// Additional business rules
	if strings.TrimSpace(rule.Content) == "" {
		result.AddError("content", "rule content cannot be empty", "EMPTY_CONTENT")
	}

	// Check for duplicate tags
	if len(rule.Tags) > 0 {
		seen := make(map[string]bool)
		for _, tag := range rule.Tags {
			if seen[tag] {
				result.AddError("tags", "duplicate tag: "+tag, "DUPLICATE_TAG")
				break // Only report first duplicate
			}
			seen[tag] = true
		}
	}

	// Validate rule ID format if present
	if rule.ID != "" {
		if err := v.ValidateRuleID(rule.ID); err != nil {
			result.AddError("id", err.Error(), "INVALID_ID")
		}
	}

	// Additional trigger validation (avoid duplicating struct validation errors)
	if rule.Trigger != nil {
		// Only do custom validation if struct validation passed for the trigger
		if !v.hasStructErrors(result, "globs", "type") {
			if err := v.validateTrigger(rule.Trigger); err != nil {
				result.AddError("trigger", err.Error(), "INVALID_TRIGGER")
			}
		}
	}

	return result
}

// ValidateRules validates multiple rules
func (v *defaultValidator) ValidateRules(rules []*domain.Rule) *BatchResult {
	result := &BatchResult{
		TotalRules: len(rules),
		Results:    make([]*Result, len(rules)),
		AllValid:   true,
	}

	for i, rule := range rules {
		validationResult := v.ValidateRule(rule)
		vr := &Result{
			RuleID: rule.ID,
			Valid:  validationResult.Valid,
			Errors: make([]string, 0, len(validationResult.Errors)),
		}

		for _, err := range validationResult.Errors {
			vr.Errors = append(vr.Errors, err.Error())
		}

		result.Results[i] = vr
		if validationResult.Valid {
			result.ValidRules++
		} else {
			result.AllValid = false
		}
		if len(validationResult.Warnings) > 0 {
			result.HasWarnings = true
		}
	}

	return result
}

// ValidateProject validates a project configuration
func (v *defaultValidator) ValidateProject(config *domain.Project) error {
	if config == nil {
		return contextureerrors.WithOpf(
			ValidationOperation+" project",
			"project configuration cannot be nil",
		)
	}

	if err := v.v.Struct(config); err != nil {
		return v.formatValidationError(err, "project")
	}

	// Business rules
	if len(config.Formats) > 0 {
		hasEnabled := false
		formatTypes := make(map[domain.FormatType]bool)

		for _, format := range config.Formats {
			if format.Enabled {
				hasEnabled = true
			}
			if formatTypes[format.Type] {
				return contextureerrors.WithOpf(
					ValidationOperation+" project",
					"duplicate format type: %s", format.Type,
				)
			}
			formatTypes[format.Type] = true
		}

		if !hasEnabled {
			return contextureerrors.WithOpf(
				ValidationOperation+" project",
				"at least one format must be enabled",
			)
		}
	}

	// Validate unique rule IDs
	ruleIDs := make(map[string]bool)
	for _, rule := range config.Rules {
		if ruleIDs[rule.ID] {
			return contextureerrors.WithOpf(
				ValidationOperation+" project",
				"duplicate rule ID: %s", rule.ID,
			)
		}
		ruleIDs[rule.ID] = true
	}

	return nil
}

// ValidateFormatConfig validates a format configuration
func (v *defaultValidator) ValidateFormatConfig(config *domain.FormatConfig) error {
	if config == nil {
		return contextureerrors.WithOpf(
			ValidationOperation+" format config",
			"format configuration cannot be nil",
		)
	}

	if err := v.v.Struct(config); err != nil {
		return v.formatValidationError(err, "format config")
	}

	return nil
}

// ValidateRuleRef validates a rule reference
func (v *defaultValidator) ValidateRuleRef(ref domain.RuleRef) error {
	if err := v.v.Struct(ref); err != nil {
		return v.formatValidationError(err, "rule reference")
	}

	if ref.ID == "" {
		return contextureerrors.WithOpf(
			ValidationOperation+" rule reference",
			"rule ID cannot be empty",
		)
	}

	return v.ValidateRuleID(ref.ID)
}

// ValidateRuleID validates a rule ID format
func (v *defaultValidator) ValidateRuleID(ruleID string) error {
	if ruleID == "" {
		return contextureerrors.ValidationErrorf("rule_id", "rule ID cannot be empty")
	}

	if len(ruleID) > MaxRuleIDLength {
		return contextureerrors.ValidationErrorf(
			"rule_id",
			"rule ID exceeds maximum length of %d characters", MaxRuleIDLength,
		)
	}

	// Check rule ID format using switch for cleaner logic
	switch {
	case strings.HasPrefix(ruleID, "[contexture"):
		// Full format [contexture:path] or [contexture(source):path,branch]{variables}
		// Use the existing regex pattern to validate the complete format
		if !domain.RuleIDPatternRegex.MatchString(ruleID) {
			if !strings.HasSuffix(ruleID, "]") && !strings.HasSuffix(ruleID, "}") {
				return contextureerrors.ValidationErrorf(
					"rule_id",
					"invalid rule ID format: missing closing bracket",
				)
			}
			if !strings.Contains(ruleID, ":") {
				return contextureerrors.ValidationErrorf(
					"rule_id",
					"invalid rule ID format: missing colon separator",
				)
			}
			return contextureerrors.ValidationErrorf("rule_id", "invalid rule ID format")
		}
	case strings.HasPrefix(ruleID, "@"):
		// Provider syntax @provider/path - valid, skip character validation
		return nil
	default:
		// For non-contexture rule IDs, check for invalid characters
		invalidChars := []string{"!", "#", "$", "%", "^", "&", "*", "(", ")", "+", "=", "{", "}", "[", "]", "|", "\\", ":", ";", "\"", "'", "<", ">", "?", ",", " "}
		for _, char := range invalidChars {
			if strings.Contains(ruleID, char) {
				return contextureerrors.ValidationErrorf(
					"rule_id",
					"invalid rule ID format: contains invalid character '%s'", char,
				)
			}
		}
	}

	return nil
}

// ValidateGitURL validates a git repository URL
func (v *defaultValidator) ValidateGitURL(gitURL string) error {
	if gitURL == "" {
		return contextureerrors.ValidationErrorf("git_url", "git URL cannot be empty")
	}

	// Accept common git URL formats
	if !strings.HasPrefix(gitURL, "https://") &&
		!strings.HasPrefix(gitURL, "http://") &&
		!strings.HasPrefix(gitURL, "git@") &&
		!strings.HasPrefix(gitURL, "ssh://") {
		return contextureerrors.ValidationErrorf(
			"git_url",
			"invalid git URL format: must start with https://, http://, git@, or ssh://",
		)
	}

	return nil
}

// ValidateWithContext validates with additional context
func (v *defaultValidator) ValidateWithContext(
	ctx context.Context,
	value any,
	name string,
) error {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := v.v.Struct(value); err != nil {
		return v.formatValidationError(err, name)
	}

	return nil
}

// Helper methods

// hasStructErrors checks if result contains validation errors for any of the specified fields
func (v *defaultValidator) hasStructErrors(result *domain.ValidationResult, fields ...string) bool {
	for _, err := range result.Errors {
		var validationErr *contextureerrors.Error
		if errors.As(err, &validationErr) {
			for _, field := range fields {
				if validationErr.Field == field {
					return true
				}
			}
		}
	}
	return false
}

func (v *defaultValidator) validateRuleRef(fl validator.FieldLevel) bool {
	ref, ok := fl.Field().Interface().(domain.RuleRef)
	if !ok {
		return false
	}
	return len(ref.ID) >= MinRuleIDLength && len(ref.ID) <= MaxRuleIDLength
}

func (v *defaultValidator) validateRuleIDTag(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	return v.ValidateRuleID(id) == nil
}

func (v *defaultValidator) validateFormatType(fl validator.FieldLevel) bool {
	ft, ok := fl.Field().Interface().(domain.FormatType)
	if !ok {
		return false
	}
	// Valid format types
	switch ft {
	case domain.FormatClaude, domain.FormatCursor, domain.FormatWindsurf:
		return true
	default:
		return false
	}
}

func (v *defaultValidator) validateGitURLTag(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	return v.ValidateGitURL(url) == nil
}

func (v *defaultValidator) validateContexturePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	// Basic path validation
	return path != "" && !strings.Contains(path, "..")
}

func (v *defaultValidator) validateTrigger(trigger *domain.RuleTrigger) error {
	if trigger.Type == "" {
		return contextureerrors.ValidationErrorf("trigger_type", "trigger type cannot be empty")
	}

	// Validate based on trigger type
	switch trigger.Type {
	case domain.TriggerGlob:
		if len(trigger.Globs) == 0 {
			return contextureerrors.ValidationErrorf("trigger_globs", "glob trigger must have globs")
		}
	case domain.TriggerAlways, domain.TriggerManual, domain.TriggerModel:
		// No additional validation needed
	default:
		return contextureerrors.ValidationErrorf(
			"trigger_type",
			"unknown trigger type: %s", trigger.Type,
		)
	}

	return nil
}

func (v *defaultValidator) addStructValidationErrors(err error, result *domain.ValidationResult) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, fieldErr := range validationErrs {
			msg := v.getErrorMessage(fieldErr)
			result.AddError(fieldErr.Field(), msg, strings.ToUpper(fieldErr.Tag()))
		}
	} else {
		result.AddError("", err.Error(), "VALIDATION_ERROR")
	}
}

func (v *defaultValidator) formatValidationError(err error, entityType string) error {
	var sb strings.Builder

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for i, fieldErr := range validationErrs {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("field '")
			sb.WriteString(fieldErr.Field())
			sb.WriteString("': ")
			sb.WriteString(v.getErrorMessage(fieldErr))
		}
	} else {
		sb.WriteString(err.Error())
	}

	return contextureerrors.WithOpf(
		ValidationOperation+" "+entityType,
		"validation failed: %s", sb.String(),
	)
}

func (v *defaultValidator) getErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "is required"
	case "min":
		return "must be at least " + fieldErr.Param()
	case "max":
		return "must be at most " + fieldErr.Param()
	case "len":
		return "must be exactly " + fieldErr.Param() + " characters"
	case "oneof":
		return "must be one of: " + fieldErr.Param()
	case "ruleid":
		return "must be a valid rule ID"
	case "ruleref":
		return "must be a valid rule reference"
	case "formattype":
		return "must be a valid format type"
	case "giturl":
		return "must be a valid git URL"
	case "contexturepath":
		return "must be a valid path"
	default:
		return "failed " + fieldErr.Tag() + " validation"
	}
}
