package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
