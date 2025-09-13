package claude

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormat(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	assert.NotNil(t, f)
}

func TestFormat_Transform(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	rule := &domain.Rule{
		ID:          "[contexture:test/rule]",
		Title:       "Test Rule",
		Description: "A test rule for validation",
		Tags:        []string{"test", "validation"},
		Content:     "This is the rule content",
		Source:      "https://github.com/test/repo.git",
		Ref:         "main",
	}

	processedRule := &domain.ProcessedRule{
		Rule:    rule,
		Content: rule.Content, // Use the raw content from rule
		Context: &domain.RuleContext{},
		Variables: map[string]any{
			"testVar": "processed", // Example variable for template processing
		},
	}
	transformed, err := f.Transform(processedRule)

	require.NoError(t, err)
	assert.NotNil(t, transformed)
	assert.Equal(t, rule, transformed.Rule)
	assert.NotEmpty(t, transformed.Content)
	assert.Equal(t, "CLAUDE.md", transformed.Filename)
	assert.Equal(t, "CLAUDE.md", transformed.RelativePath)
	assert.NotZero(t, transformed.TransformedAt)

	// Check that content contains expected elements
	assert.Contains(t, transformed.Content, "Test Rule")
	assert.Contains(t, transformed.Content, "A test rule for validation")
	assert.Contains(t, transformed.Content, "test and validation")
	assert.Contains(t, transformed.Content, "This is the rule content")
	// Transform output should NOT contain tracking comments (they're added during Write)
	// Note: Claude template doesn't include source and branch in the output

	// Check metadata
	assert.Equal(t, "claude", transformed.Metadata["format"])
	assert.Equal(t, "CLAUDE.md", transformed.Metadata["filename"])
}

func TestFormat_Transform_MinimalRule(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	rule := &domain.Rule{
		ID:      "[contexture:test/minimal]",
		Title:   "Minimal Rule",
		Content: "Minimal content",
	}

	processedRule := &domain.ProcessedRule{
		Rule:      rule,
		Content:   rule.Content, // Use the raw content from rule
		Context:   &domain.RuleContext{},
		Variables: map[string]any{},
	}
	transformed, err := f.Transform(processedRule)

	require.NoError(t, err)
	assert.NotNil(t, transformed)
	assert.Contains(t, transformed.Content, "Minimal Rule")
	assert.Contains(t, transformed.Content, "Minimal content")
}

func TestFormat_Validate(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	tests := []struct {
		name         string
		rule         *domain.Rule
		wantValid    bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name: "valid rule",
			rule: &domain.Rule{
				ID:          "[contexture:test/valid]",
				Title:       "Valid Rule",
				Description: "A valid rule",
				Tags:        []string{"test"},
				Content:     "Valid content",
			},
			wantValid:    true,
			wantErrors:   0,
			wantWarnings: 0,
		},
		{
			name: "rule with warnings",
			rule: &domain.Rule{
				ID:      "[contexture:test/warnings]",
				Title:   "Rule with Warnings",
				Content: "Content without description or tags",
			},
			wantValid:    true,
			wantErrors:   0,
			wantWarnings: 2, // missing description and tags
		},
		{
			name: "invalid rule - missing ID",
			rule: &domain.Rule{
				Title:   "Invalid Rule",
				Content: "Content",
			},
			wantValid:    false,
			wantErrors:   1,
			wantWarnings: 2, // missing description and tags
		},
		{
			name: "invalid rule - missing title",
			rule: &domain.Rule{
				ID:      "[contexture:test/invalid]",
				Content: "Content",
			},
			wantValid:    false,
			wantErrors:   1,
			wantWarnings: 2, // missing description and tags
		},
		{
			name: "invalid rule - missing content",
			rule: &domain.Rule{
				ID:    "[contexture:test/invalid]",
				Title: "Invalid Rule",
			},
			wantValid:    false,
			wantErrors:   1,
			wantWarnings: 2, // missing description and tags
		},
		{
			name:         "invalid rule - all missing",
			rule:         &domain.Rule{},
			wantValid:    false,
			wantErrors:   3, // missing ID, title, content
			wantWarnings: 2, // missing description and tags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Validate(tt.rule)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantValid, result.Valid)
			assert.Len(t, result.Errors, tt.wantErrors)
			assert.Len(t, result.Warnings, tt.wantWarnings)
			assert.Equal(t, "claude", result.Metadata["format"])
		})
	}
}

func TestFormat_Write(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	// Create test rules
	rules := []*domain.TransformedRule{
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule1]",
				Title: "Rule 1",
			},
			Content:  "Content of rule 1",
			Filename: "CLAUDE.md",
		},
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule2]",
				Title: "Rule 2",
			},
			Content:  "Content of rule 2",
			Filename: "CLAUDE.md",
		},
	}

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	err := f.Write(rules, config)
	require.NoError(t, err)

	// Check that file was created
	content, err := afero.ReadFile(fs, "/output/CLAUDE.md")
	require.NoError(t, err)
	assert.NotEmpty(t, content)

	contentStr := string(content)

	// Check header
	assert.Contains(t, contentStr, "# claude.md")

	// Check rules content
	assert.Contains(t, contentStr, "Content of rule 1")
	assert.Contains(t, contentStr, "Content of rule 2")

	// Check separator
	assert.Contains(t, contentStr, "---")

	// Check footer
	assert.Contains(t, contentStr, "Generated by Contexture CLI")
}

func TestFormat_Write_EmptyRules(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	err := f.Write([]*domain.TransformedRule{}, config)
	require.NoError(t, err)

	// Check that no file was created
	exists, err := afero.Exists(fs, "/output/CLAUDE.md")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFormat_Write_DirectoryCreation(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	rules := []*domain.TransformedRule{
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule]",
				Title: "Test Rule",
			},
			Content: "Test content",
		},
	}

	config := &domain.FormatConfig{
		BaseDir: "/deep/nested/path",
	}

	err := f.Write(rules, config)
	require.NoError(t, err)

	// Check that directories were created
	exists, err := afero.DirExists(fs, "/deep/nested/path")
	require.NoError(t, err)
	assert.True(t, exists)

	// Check that file exists
	exists, err = afero.Exists(fs, "/deep/nested/path/CLAUDE.md")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFormat_List(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	t.Run("file does not exist", func(t *testing.T) {
		config := &domain.FormatConfig{
			BaseDir: "/output",
		}

		rules, err := f.List(config)
		require.NoError(t, err)
		assert.Empty(t, rules)
	})

	t.Run("file exists", func(t *testing.T) {
		// Create test file with new tracking comment format
		content := `# Contexture Rules

This file contains 1 contexture rules for Claude AI assistant.

Generated at: 2024-01-01 12:00:00

---

# Test Rule

This is a test rule for validation.

**Tags:** test, validation

Some rule content here.

<!-- id: [contexture:test/rule] -->`
		err := fs.MkdirAll("/output", 0o755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, "/output/CLAUDE.md", []byte(content), 0o644)
		require.NoError(t, err)

		config := &domain.FormatConfig{
			BaseDir: "/output",
		}

		rules, err := f.List(config)
		require.NoError(t, err)
		assert.Len(t, rules, 1)

		rule := rules[0]
		assert.Equal(t, "[contexture:test/rule]", rule.ID())
		assert.Equal(t, "Test Rule", rule.Title())
		assert.Equal(t, "local", rule.Source())
		assert.Equal(t, "CLAUDE.md", rule.Filename)
		assert.Equal(t, "CLAUDE.md", rule.RelativePath)
		assert.Equal(t, int64(len(content)), rule.Size)
		assert.NotEmpty(t, rule.ContentHash)
		assert.NotZero(t, rule.InstalledAt)
	})
}

func TestFormat_getOutputPath(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	tests := []struct {
		name   string
		config *domain.FormatConfig
		want   string
	}{
		{
			name:   "nil config",
			config: nil,
			want:   "CLAUDE.md",
		},
		{
			name:   "empty config",
			config: &domain.FormatConfig{},
			want:   "CLAUDE.md",
		},
		{
			name: "with base dir",
			config: &domain.FormatConfig{
				BaseDir: "/output",
			},
			want: "/output/CLAUDE.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.getOutputPath(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_getDefaultTemplate(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	template := f.getDefaultTemplate()
	assert.NotEmpty(t, template)
	assert.Contains(t, template, "{{.title}}")
	assert.Contains(t, template, "{{.content}}")
	assert.Contains(t, template, "{{.description}}")
	assert.Contains(t, template, "{{join_and .tags}}")
	assert.Contains(t, template, "**Applies:**")
	assert.Contains(t, template, "{{join_and .frameworks}}")
	assert.Contains(t, template, "{{if .trigger}}")
}

func TestFormat_getFileHeader(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	header := f.getFileHeader(5)
	assert.Contains(t, header, "# claude.md")
}

func TestFormat_getFileFooter(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	footer := f.getFileFooter()
	assert.Contains(t, footer, "Generated by Contexture CLI")
	// Note: The footer doesn't contain "Do not edit manually" message
}

func TestFormat_getOutputFilename(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	filename := f.getOutputFilename()
	assert.Equal(t, domain.ClaudeOutputFile, filename)
}

func TestFormat_extractBasePath(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	tests := []struct {
		name     string
		ruleID   string
		expected string
	}{
		{
			name:     "simple rule ID",
			ruleID:   "test/rule1",
			expected: "test/rule1",
		},
		{
			name:     "full format without variables",
			ruleID:   "[contexture:test/rule1]",
			expected: "test/rule1",
		},
		{
			name:     "full format with variables",
			ruleID:   "[contexture:test/rule1]{\"extended\": true}",
			expected: "test/rule1",
		},
		{
			name:     "complex path with variables",
			ruleID:   "[contexture:languages/go/advanced-patterns]{\"strict\": false, \"target\": \"es2022\"}",
			expected: "languages/go/advanced-patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.extractBasePath(tt.ruleID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
