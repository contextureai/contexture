package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFormatType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		formatType FormatType
		expected   string
	}{
		{
			name:       "claude format",
			formatType: FormatClaude,
			expected:   "claude",
		},
		{
			name:       "cursor format",
			formatType: FormatCursor,
			expected:   "cursor",
		},
		{
			name:       "windsurf format",
			formatType: FormatWindsurf,
			expected:   "windsurf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.formatType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationResult_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		result   *ValidationResult
		expected string
	}{
		{
			name:     "valid result",
			result:   &ValidationResult{Valid: true},
			expected: "",
		},
		{
			name:     "invalid with no errors",
			result:   &ValidationResult{Valid: false, Errors: []error{}},
			expected: "validation failed",
		},
		{
			name: "invalid with errors",
			result: &ValidationResult{
				Valid: false,
				Errors: []error{
					errors.New("title is required"),
					errors.New("tags must not be empty"),
				},
			},
			expected: "title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationResult_AddError(t *testing.T) {
	t.Parallel()
	result := &ValidationResult{Valid: true}

	result.AddError("title", "title is required", "required")

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Error(), "title")
	assert.Contains(t, result.Errors[0].Error(), "title is required")
	assert.Contains(t, result.Errors[0].Error(), "required")
}

func TestValidationResult_AddWarning(t *testing.T) {
	t.Parallel()
	result := &ValidationResult{Valid: true}

	result.AddWarning("tags", "consider adding more tags", "suggestion")

	assert.True(t, result.Valid) // Warnings don't affect validity
	assert.Len(t, result.Warnings, 1)
	assert.Equal(t, "tags", result.Warnings[0].Field)
	assert.Equal(t, "consider adding more tags", result.Warnings[0].Message)
	assert.Equal(t, "suggestion", result.Warnings[0].Code)
}

func TestTransformedRule_GetContentHash(t *testing.T) {
	t.Parallel()
	rule := &Rule{ID: "[contexture:test/rule]"}
	transformed := &TransformedRule{
		Rule:    rule,
		Content: "test content",
	}

	hash := transformed.GetContentHash()
	// SHA256 hash of "test content"
	expected := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"
	assert.Equal(t, expected, hash)

	// Same content should produce same hash
	transformed2 := &TransformedRule{
		Rule:    rule,
		Content: "test content",
	}
	assert.Equal(t, hash, transformed2.GetContentHash())
}

func TestTransformedRule_GetAbsolutePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		rule     *TransformedRule
		baseDir  string
		expected string
	}{
		{
			name: "with filename only",
			rule: &TransformedRule{
				Filename:     "test.md",
				RelativePath: "",
			},
			baseDir:  "/base",
			expected: "/base/test.md",
		},
		{
			name: "with relative path",
			rule: &TransformedRule{
				Filename:     "test.md",
				RelativePath: "subdir/test.md",
			},
			baseDir:  "/base",
			expected: "/base/subdir/test.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rule.GetAbsolutePath(tt.baseDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatConfig_GetEffectiveUserRulesMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		config   FormatConfig
		expected UserRulesOutputMode
	}{
		{
			name: "windsurf with unset mode returns native",
			config: FormatConfig{
				Type:          FormatWindsurf,
				UserRulesMode: "",
			},
			expected: UserRulesNative,
		},
		{
			name: "windsurf with explicit native mode",
			config: FormatConfig{
				Type:          FormatWindsurf,
				UserRulesMode: UserRulesNative,
			},
			expected: UserRulesNative,
		},
		{
			name: "windsurf with project mode",
			config: FormatConfig{
				Type:          FormatWindsurf,
				UserRulesMode: UserRulesProject,
			},
			expected: UserRulesProject,
		},
		{
			name: "windsurf with disabled mode",
			config: FormatConfig{
				Type:          FormatWindsurf,
				UserRulesMode: UserRulesDisabled,
			},
			expected: UserRulesDisabled,
		},
		{
			name: "claude with unset mode returns native",
			config: FormatConfig{
				Type:          FormatClaude,
				UserRulesMode: "",
			},
			expected: UserRulesNative,
		},
		{
			name: "claude with explicit native mode",
			config: FormatConfig{
				Type:          FormatClaude,
				UserRulesMode: UserRulesNative,
			},
			expected: UserRulesNative,
		},
		{
			name: "claude with project mode",
			config: FormatConfig{
				Type:          FormatClaude,
				UserRulesMode: UserRulesProject,
			},
			expected: UserRulesProject,
		},
		{
			name: "cursor with unset mode returns project",
			config: FormatConfig{
				Type:          FormatCursor,
				UserRulesMode: "",
			},
			expected: UserRulesProject,
		},
		{
			name: "cursor with explicit project mode",
			config: FormatConfig{
				Type:          FormatCursor,
				UserRulesMode: UserRulesProject,
			},
			expected: UserRulesProject,
		},
		{
			name: "cursor with disabled mode",
			config: FormatConfig{
				Type:          FormatCursor,
				UserRulesMode: UserRulesDisabled,
			},
			expected: UserRulesDisabled,
		},
		{
			name: "unknown format defaults to project",
			config: FormatConfig{
				Type:          "unknown",
				UserRulesMode: "",
			},
			expected: UserRulesProject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetEffectiveUserRulesMode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatConfig_ShouldOmitUserRulesMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		config   FormatConfig
		expected bool
	}{
		{
			name: "windsurf with empty mode should omit",
			config: FormatConfig{
				Type:          FormatWindsurf,
				UserRulesMode: "",
			},
			expected: true,
		},
		{
			name: "windsurf with native mode should omit (default)",
			config: FormatConfig{
				Type:          FormatWindsurf,
				UserRulesMode: UserRulesNative,
			},
			expected: true,
		},
		{
			name: "windsurf with project mode should NOT omit",
			config: FormatConfig{
				Type:          FormatWindsurf,
				UserRulesMode: UserRulesProject,
			},
			expected: false,
		},
		{
			name: "claude with empty mode should omit",
			config: FormatConfig{
				Type:          FormatClaude,
				UserRulesMode: "",
			},
			expected: true,
		},
		{
			name: "claude with native mode should omit (default)",
			config: FormatConfig{
				Type:          FormatClaude,
				UserRulesMode: UserRulesNative,
			},
			expected: true,
		},
		{
			name: "cursor with empty mode should omit",
			config: FormatConfig{
				Type:          FormatCursor,
				UserRulesMode: "",
			},
			expected: true,
		},
		{
			name: "cursor with project mode should omit (default)",
			config: FormatConfig{
				Type:          FormatCursor,
				UserRulesMode: UserRulesProject,
			},
			expected: true,
		},
		{
			name: "cursor with disabled mode should NOT omit",
			config: FormatConfig{
				Type:          FormatCursor,
				UserRulesMode: UserRulesDisabled,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldOmitUserRulesMode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatConfig_MarshalYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		config   FormatConfig
		expected string
	}{
		{
			name: "windsurf with default native mode omits userRulesMode",
			config: FormatConfig{
				Type:          FormatWindsurf,
				Enabled:       true,
				UserRulesMode: UserRulesNative,
			},
			expected: "type: windsurf\nenabled: true\n",
		},
		{
			name: "windsurf with empty mode omits userRulesMode",
			config: FormatConfig{
				Type:          FormatWindsurf,
				Enabled:       true,
				UserRulesMode: "",
			},
			expected: "type: windsurf\nenabled: true\n",
		},
		{
			name: "windsurf with non-default project mode includes userRulesMode",
			config: FormatConfig{
				Type:          FormatWindsurf,
				Enabled:       true,
				UserRulesMode: UserRulesProject,
			},
			expected: "type: windsurf\nenabled: true\nuserRulesMode: project\n",
		},
		{
			name: "cursor with default project mode omits userRulesMode",
			config: FormatConfig{
				Type:          FormatCursor,
				Enabled:       true,
				UserRulesMode: UserRulesProject,
			},
			expected: "type: cursor\nenabled: true\n",
		},
		{
			name: "cursor with non-default disabled mode includes userRulesMode",
			config: FormatConfig{
				Type:          FormatCursor,
				Enabled:       true,
				UserRulesMode: UserRulesDisabled,
			},
			expected: "type: cursor\nenabled: true\nuserRulesMode: disabled\n",
		},
		{
			name: "claude with template and default mode",
			config: FormatConfig{
				Type:          FormatClaude,
				Enabled:       true,
				Template:      "CLAUDE.template.md",
				UserRulesMode: UserRulesNative,
			},
			expected: "type: claude\nenabled: true\ntemplate: CLAUDE.template.md\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(tt.config)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}
