package validation

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)
	require.NotNil(t, v)
}

func TestValidateRule(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name      string
		rule      *domain.Rule
		wantValid bool
		wantError string
	}{
		{
			name:      "nil rule",
			rule:      nil,
			wantValid: false,
			wantError: "Rule cannot be nil",
		},
		{
			name: "valid rule",
			rule: &domain.Rule{
				ID:          "[contexture:test/rule]",
				Title:       "Test Rule",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "Rule content",
			},
			wantValid: true,
		},
		{
			name: "missing required fields",
			rule: &domain.Rule{
				ID: "[contexture:test/rule]",
			},
			wantValid: false,
		},
		{
			name: "invalid rule ID",
			rule: &domain.Rule{
				ID:          "invalid[rule]id",
				Title:       "Test Rule",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "Rule content",
			},
			wantValid: false, // Contains invalid characters
			wantError: "invalid character",
		},
		{
			name: "rule with invalid trigger",
			rule: &domain.Rule{
				ID:          "[contexture:test/rule]",
				Title:       "Test Rule",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "Rule content",
				Trigger: &domain.RuleTrigger{
					Type: "invalid",
				},
			},
			wantValid: false,
			wantError: "must be one of:",
		},
		{
			name: "glob trigger without globs",
			rule: &domain.Rule{
				ID:          "[contexture:test/rule]",
				Title:       "Test Rule",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "Rule content",
				Trigger: &domain.RuleTrigger{
					Type:  "glob",
					Globs: []string{},
				},
			},
			wantValid: false,
			wantError: "glob trigger must have globs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateRule(tt.rule)
			assert.Equal(t, tt.wantValid, result.Valid)

			if tt.wantError != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Error(), tt.wantError) {
						found = true
						break
					}
				}
				assert.True(
					t,
					found,
					"Expected error containing '%s' but got errors: %v",
					tt.wantError,
					result.Errors,
				)
			}
		})
	}
}

func TestValidateProject(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		config  *domain.Project
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "cannot be nil",
		},
		{
			name: "valid config",
			config: &domain.Project{
				Version: 1,
				Formats: []domain.FormatConfig{
					{Type: domain.FormatClaude, Enabled: true},
				},
				Rules: []domain.RuleRef{
					{ID: "[contexture:test/rule]"},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate format types",
			config: &domain.Project{
				Version: 1,
				Formats: []domain.FormatConfig{
					{Type: domain.FormatClaude, Enabled: true},
					{Type: domain.FormatClaude, Enabled: false},
				},
			},
			wantErr: true,
			errMsg:  "duplicate format type",
		},
		{
			name: "no enabled formats",
			config: &domain.Project{
				Version: 1,
				Formats: []domain.FormatConfig{
					{Type: domain.FormatClaude, Enabled: false},
					{Type: domain.FormatCursor, Enabled: false},
				},
			},
			wantErr: true,
			errMsg:  "at least one format must be enabled",
		},
		{
			name: "duplicate rule IDs",
			config: &domain.Project{
				Version: 1,
				Rules: []domain.RuleRef{
					{ID: "[contexture:test/rule]"},
					{ID: "[contexture:test/rule]"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate rule ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateProject(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRuleID(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		ruleID  string
		wantErr bool
	}{
		{
			name:    "empty rule ID",
			ruleID:  "",
			wantErr: true,
		},
		{
			name:    "valid full format",
			ruleID:  "[contexture:test/rule]",
			wantErr: false,
		},
		{
			name:    "valid full format with source",
			ruleID:  "[contexture(https://github.com/user/repo):test/rule]",
			wantErr: false,
		},
		{
			name:    "valid full format with branch",
			ruleID:  "[contexture:test/rule,main]",
			wantErr: false,
		},
		{
			name:    "missing closing bracket",
			ruleID:  "[contexture:test/rule",
			wantErr: true,
		},
		{
			name:    "rule ID too long",
			ruleID:  "[contexture:" + string(make([]byte, 300)) + "]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRuleID(tt.ruleID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGitURL(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "valid HTTPS URL",
			url:     "https://github.com/user/repo.git",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://github.com/user/repo.git",
			wantErr: false,
		},
		{
			name:    "valid SSH URL",
			url:     "git@github.com:user/repo.git",
			wantErr: false,
		},
		{
			name:    "valid SSH protocol URL",
			url:     "ssh://git@github.com/user/repo.git",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateGitURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWithContext(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := v.ValidateWithContext(ctx, &domain.Project{Version: 1}, "project")
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("valid with context", func(t *testing.T) {
		ctx := context.Background()
		project := &domain.Project{
			Version: 1,
		}

		err := v.ValidateWithContext(ctx, project, "project")
		assert.NoError(t, err)
	})
}

func TestValidateRules(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name             string
		rules            []*domain.Rule
		expectedValid    int
		expectedTotal    int
		expectedAllValid bool
	}{
		{
			name:             "empty rules",
			rules:            []*domain.Rule{},
			expectedValid:    0,
			expectedTotal:    0,
			expectedAllValid: true,
		},
		{
			name: "all valid rules",
			rules: []*domain.Rule{
				{
					ID:          "[contexture:test/rule1]",
					Title:       "Test Rule 1",
					Description: "A test rule",
					Tags:        []string{"test"},
					Content:     "Rule content",
				},
				{
					ID:          "[contexture:test/rule2]",
					Title:       "Test Rule 2",
					Description: "Another test rule",
					Tags:        []string{"test"},
					Content:     "Rule content",
				},
			},
			expectedValid:    2,
			expectedTotal:    2,
			expectedAllValid: true,
		},
		{
			name: "mixed valid and invalid rules",
			rules: []*domain.Rule{
				{
					ID:          "[contexture:test/rule1]",
					Title:       "Test Rule 1",
					Description: "A test rule",
					Tags:        []string{"test"},
					Content:     "Rule content",
				},
				{
					ID:      "[contexture:test/rule2]",
					Content: "Rule content",
					// Missing required fields
				},
			},
			expectedValid:    1,
			expectedTotal:    2,
			expectedAllValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateRules(tt.rules)

			assert.Equal(t, tt.expectedTotal, result.TotalRules)
			assert.Equal(t, tt.expectedValid, result.ValidRules)
			assert.Equal(t, tt.expectedAllValid, result.AllValid)
			assert.Len(t, result.Results, tt.expectedTotal)

			// Check individual results
			for i, res := range result.Results {
				assert.Equal(t, tt.rules[i].ID, res.RuleID)
			}
		})
	}
}

func TestValidateFormatConfig(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		config  *domain.FormatConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "cannot be nil",
		},
		{
			name: "valid config",
			config: &domain.FormatConfig{
				Type:    domain.FormatClaude,
				Enabled: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateFormatConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRuleRef(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		ref     domain.RuleRef
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid rule ref",
			ref: domain.RuleRef{
				ID:     "[contexture:test/rule]",
				Source: "https://github.com/user/repo.git",
			},
			wantErr: false,
		},
		{
			name: "empty rule ID",
			ref: domain.RuleRef{
				ID: "",
			},
			wantErr: true,
			errMsg:  "rule ID cannot be empty",
		},
		{
			name: "invalid rule ID format",
			ref: domain.RuleRef{
				ID: "invalid[rule]id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRuleRef(tt.ref)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatorHelperMethods(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	// Test validateRuleRef via struct validation
	t.Run("validateRuleRef via struct validation", func(t *testing.T) {
		type TestStruct struct {
			Ref domain.RuleRef `validate:"ruleref"`
		}

		validStruct := TestStruct{
			Ref: domain.RuleRef{ID: "valid-rule-id"},
		}
		invalidStruct := TestStruct{
			Ref: domain.RuleRef{ID: string(make([]byte, MaxRuleIDLength+1))}, // Too long
		}

		dv := v.(*defaultValidator)
		require.NoError(t, dv.v.Struct(validStruct))
		assert.Error(t, dv.v.Struct(invalidStruct))
	})

	// Test validateRuleIDTag via struct validation
	t.Run("validateRuleIDTag via struct validation", func(t *testing.T) {
		type TestStruct struct {
			ID string `validate:"ruleid"`
		}

		validStruct := TestStruct{ID: "[contexture:test/rule]"}
		invalidStruct := TestStruct{ID: "invalid[rule]id"}

		dv := v.(*defaultValidator)
		require.NoError(t, dv.v.Struct(validStruct))
		assert.Error(t, dv.v.Struct(invalidStruct))
	})

	// Test validateFormatType via struct validation
	t.Run("validateFormatType via struct validation", func(t *testing.T) {
		type TestStruct struct {
			Type domain.FormatType `validate:"formattype"`
		}

		validStruct := TestStruct{Type: domain.FormatClaude}
		invalidStruct := TestStruct{Type: domain.FormatType("invalid")}

		dv := v.(*defaultValidator)
		require.NoError(t, dv.v.Struct(validStruct))
		assert.Error(t, dv.v.Struct(invalidStruct))
	})

	// Test validateGitURLTag via struct validation
	t.Run("validateGitURLTag via struct validation", func(t *testing.T) {
		type TestStruct struct {
			URL string `validate:"giturl"`
		}

		validStruct := TestStruct{URL: "https://github.com/user/repo.git"}
		invalidStruct := TestStruct{URL: "not-a-url"}

		dv := v.(*defaultValidator)
		require.NoError(t, dv.v.Struct(validStruct))
		assert.Error(t, dv.v.Struct(invalidStruct))
	})

	// Test validateContexturePath via struct validation
	t.Run("validateContexturePath via struct validation", func(t *testing.T) {
		type TestStruct struct {
			Path string `validate:"contexturepath"`
		}

		validStruct := TestStruct{Path: "valid/path"}
		invalidStruct := TestStruct{Path: "../invalid/path"}

		dv := v.(*defaultValidator)
		require.NoError(t, dv.v.Struct(validStruct))
		assert.Error(t, dv.v.Struct(invalidStruct))
	})
}

func TestGetErrorMessage(t *testing.T) {
	t.Parallel()
	v, err := NewValidator()
	require.NoError(t, err)

	dv := v.(*defaultValidator)

	// Test some basic validation tags by creating validation errors
	tests := []struct {
		name        string
		structField any
		tag         string
		expectedMsg string
	}{
		{
			name: "required validation",
			structField: struct {
				Value string `validate:"required"`
			}{Value: ""},
			tag:         "required",
			expectedMsg: "is required",
		},
		{
			name: "min validation",
			structField: struct {
				Value string `validate:"min=5"`
			}{Value: "hi"},
			tag:         "min",
			expectedMsg: "must be at least",
		},
		{
			name: "max validation",
			structField: struct {
				Value string `validate:"max=3"`
			}{Value: "toolong"},
			tag:         "max",
			expectedMsg: "must be at most",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dv.v.Struct(tt.structField)
			require.Error(t, err)

			var validationErrs validator.ValidationErrors
			if errors.As(err, &validationErrs) {
				for _, fieldErr := range validationErrs {
					if fieldErr.Tag() == tt.tag {
						actualMsg := dv.getErrorMessage(fieldErr)
						assert.Contains(t, actualMsg, tt.expectedMsg)
						return
					}
				}
			}
			t.Errorf("Expected validation error with tag %s not found", tt.tag)
		})
	}
}
