package rule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	assert.NotNil(t, validator)
	assert.Implements(t, (*Validator)(nil), validator)
}

func TestValidateRule(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	tests := []struct {
		name     string
		rule     *domain.Rule
		wantErr  bool
		errCount int
	}{
		{
			name: "valid rule",
			rule: &domain.Rule{
				ID:          "[contexture:test/valid]",
				Title:       "Valid Rule",
				Description: "A valid test rule",
				Tags:        []string{"test", "valid"},
				Content:     "This is valid content",
			},
			wantErr:  false,
			errCount: 0,
		},
		{
			name: "missing ID",
			rule: &domain.Rule{
				Title:       "Rule without ID",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "This rule has no ID",
			},
			wantErr:  true,
			errCount: 1, // Only missing ID
		},
		{
			name: "missing title",
			rule: &domain.Rule{
				ID:          "[contexture:test/no-title]",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "This rule has no title",
			},
			wantErr:  true,
			errCount: 1, // Only missing title
		},
		{
			name: "missing content",
			rule: &domain.Rule{
				ID:          "[contexture:test/no-content]",
				Title:       "Rule without content",
				Description: "A test rule",
				Tags:        []string{"test"},
			},
			wantErr:  true,
			errCount: 2, // Missing content field + empty content validation
		},
		{
			name: "invalid rule ID format",
			rule: &domain.Rule{
				ID:          "invalid!@#$%format",
				Title:       "Rule with invalid ID",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "Content here",
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "empty title (whitespace only)",
			rule: &domain.Rule{
				ID:          "[contexture:test/empty-title]",
				Title:       "   ",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "Content here",
			},
			wantErr:  false, // Whitespace title is actually valid for required tag
			errCount: 0,
		},
		{
			name: "empty content (whitespace only)",
			rule: &domain.Rule{
				ID:          "[contexture:test/empty-content]",
				Title:       "Valid Title",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "   ",
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "multiple validation errors",
			rule: &domain.Rule{
				ID:      "",
				Title:   "",
				Content: "",
			},
			wantErr:  true,
			errCount: 6, // ID, title, description, tags, content (struct) + content (business rule)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateRule(tt.rule)
			require.NotNil(t, result)

			if tt.wantErr {
				assert.False(t, result.Valid)
				assert.Len(t, result.Errors, tt.errCount)
			} else {
				assert.True(t, result.Valid)
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestValidateRules(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	rules := []*domain.Rule{
		{
			ID:          "[contexture:test/valid1]",
			Title:       "Valid Rule 1",
			Description: "A valid test rule",
			Tags:        []string{"test"},
			Content:     "Valid content 1",
		},
		{
			ID:          "[contexture:test/valid2]",
			Title:       "Valid Rule 2",
			Description: "Another valid test rule",
			Tags:        []string{"test"},
			Content:     "Valid content 2",
		},
		{
			ID:    "[contexture:test/invalid]",
			Title: "Invalid Rule",
			// Missing description, tags, and content
		},
	}

	result := validator.ValidateRules(rules)
	require.NotNil(t, result)

	assert.Equal(t, 3, result.TotalRules)
	assert.Equal(t, 2, result.ValidRules) // Two valid rules
	assert.Len(t, result.Results, 3)

	// Check individual results
	assert.True(t, result.Results[0].Valid)
	assert.True(t, result.Results[1].Valid)
	assert.False(t, result.Results[2].Valid)
	assert.NotEmpty(t, result.Results[2].Errors) // Should have validation errors
}

func TestValidateRuleContent(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid content",
			content: "This is valid content",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "whitespace only content",
			content: "   \n\t  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateRuleContent(tt.content)
			require.NotNil(t, result)

			if tt.wantErr {
				assert.False(t, result.Valid)
				assert.NotEmpty(t, result.Errors)
			} else {
				assert.True(t, result.Valid)
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestValidateRuleID(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	tests := []struct {
		name    string
		ruleID  string
		wantErr bool
	}{
		{
			name:    "valid rule ID",
			ruleID:  "[contexture:test/valid]",
			wantErr: false,
		},
		{
			name:    "valid rule ID with source",
			ruleID:  "[contexture(custom):test/valid]",
			wantErr: false,
		},
		{
			name:    "valid rule ID with branch",
			ruleID:  "[contexture:test/valid,main]",
			wantErr: false,
		},
		{
			name:    "valid rule ID with source and branch",
			ruleID:  "[contexture(custom):test/valid,main]",
			wantErr: false,
		},
		{
			name:    "empty rule ID",
			ruleID:  "",
			wantErr: true,
		},
		{
			name:    "invalid format - no brackets",
			ruleID:  "contexture:test/invalid",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong prefix",
			ruleID:  "[invalid:test/rule]",
			wantErr: true,
		},
		{
			name:    "invalid format - missing colon (space is invalid char)",
			ruleID:  "[contexture test/rule]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRuleID(tt.ruleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGitURL(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	tests := []struct {
		name    string
		gitURL  string
		wantErr bool
	}{
		{
			name:    "valid HTTPS URL",
			gitURL:  "https://github.com/user/repo.git",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			gitURL:  "http://gitlab.com/user/repo.git",
			wantErr: false,
		},
		{
			name:    "valid SSH URL",
			gitURL:  "git@github.com:user/repo.git",
			wantErr: false,
		},
		{
			name:    "empty URL",
			gitURL:  "",
			wantErr: true,
		},
		{
			name:    "invalid format - no protocol",
			gitURL:  "github.com/user/repo.git",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong protocol",
			gitURL:  "ftp://github.com/user/repo.git",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateGitURL(tt.gitURL)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationBatchResult(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	// Create a mix of valid and invalid rules
	rules := []*domain.Rule{
		{
			ID:          "[contexture:test/valid1]",
			Title:       "Valid Rule 1",
			Description: "A valid test rule",
			Tags:        []string{"test"},
			Content:     "Content 1",
		},
		{
			ID:    "[contexture:test/invalid1]",
			Title: "Invalid Rule 1",
			// Missing description, tags, and content - will cause multiple errors
		},
		{
			ID:          "[contexture:test/valid2]",
			Title:       "Valid Rule 2",
			Description: "Another valid test rule",
			Tags:        []string{"test"},
			Content:     "Content 2",
		},
		{
			// Missing ID, description, tags, and content - will cause multiple errors
			Title: "Invalid Rule 2",
		},
	}

	result := validator.ValidateRules(rules)
	require.NotNil(t, result)

	assert.Equal(t, 4, result.TotalRules)
	assert.Equal(t, 2, result.ValidRules)
	// Test individual results
	assert.Equal(t, "[contexture:test/valid1]", result.Results[0].RuleID)
	assert.True(t, result.Results[0].Valid)

	assert.Equal(t, "[contexture:test/invalid1]", result.Results[1].RuleID)
	assert.False(t, result.Results[1].Valid)
	assert.NotEmpty(t, result.Results[1].Errors)

	assert.Equal(t, "[contexture:test/valid2]", result.Results[2].RuleID)
	assert.True(t, result.Results[2].Valid)

	assert.Empty(t, result.Results[3].RuleID) // Empty ID
	assert.False(t, result.Results[3].Valid)
	assert.NotEmpty(t, result.Results[3].Errors)
}

// Test edge cases and error scenarios
func TestValidatorEdgeCases(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	t.Run("nil rule", func(t *testing.T) {
		// This should not panic, but will have validation errors
		result := validator.ValidateRule(&domain.Rule{})
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("empty rules slice", func(t *testing.T) {
		result := validator.ValidateRules([]*domain.Rule{})
		assert.Equal(t, 0, result.TotalRules)
		assert.Equal(t, 0, result.ValidRules)
		// No errors expected for empty slice
		assert.Empty(t, result.Results)
	})

	t.Run("nil rules slice", func(t *testing.T) {
		result := validator.ValidateRules(nil)
		assert.Equal(t, 0, result.TotalRules)
		assert.Equal(t, 0, result.ValidRules)
		// No errors expected for empty slice
		assert.Empty(t, result.Results)
	})

	t.Run("very long rule ID", func(t *testing.T) {
		longPath := strings.Repeat("a", 1000)
		ruleID := fmt.Sprintf("[contexture:test/%s]", longPath)

		// Should fail due to length limit
		err := validator.ValidateRuleID(ruleID)
		require.Error(t, err) // Length exceeds maximum
		assert.Contains(t, err.Error(), "exceeds maximum length")
	})
}

// Test that the minimal validator doesn't have the old complex validation
func TestMinimalValidatorBehavior(t *testing.T) {
	t.Parallel()
	validator := NewValidator()

	// These should all pass with the minimal validator
	// (they would have failed with the old enhanced validator)

	t.Run("large content allowed", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "[contexture:test/large-content]",
			Title:       "Large Content Rule",
			Description: "A rule with large content",
			Tags:        []string{"test"},
			Content:     strings.Repeat("a", 10000), // Very large content
		}

		result := validator.ValidateRule(rule)
		assert.True(t, result.Valid) // Should pass - no size limits
		assert.Empty(t, result.Errors)
	})

	t.Run("no forbidden pattern checking", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "[contexture:test/patterns]",
			Title:       "Rule with patterns",
			Description: "A rule with potentially dangerous patterns",
			Tags:        []string{"test"},
			Content:     "eval() and document.write() and dangerous stuff",
		}

		result := validator.ValidateRule(rule)
		assert.True(t, result.Valid) // Should pass - no pattern checking
		assert.Empty(t, result.Errors)
	})

	t.Run("no performance limits", func(t *testing.T) {
		// Create rule with fields that would violate validation limits
		manyTags := make([]string, 100)
		for i := range manyTags {
			manyTags[i] = fmt.Sprintf("tag%d", i)
		}
		rule := &domain.Rule{
			ID:          "[contexture:test/limits]",
			Title:       strings.Repeat("Very Long Title ", 10), // Very long title (will fail)
			Description: "A rule testing limits",
			Tags:        manyTags, // Too many tags (will fail)
			Content:     "Content here",
		}

		result := validator.ValidateRule(rule)
		assert.False(t, result.Valid) // Will fail due to validation limits
		// Should have validation errors for title length and tag count
		assert.NotEmpty(t, result.Errors)
	})
}
