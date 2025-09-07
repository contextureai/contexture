package base

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseFormat(t *testing.T) {
	fs := afero.NewMemMapFs()
	base := NewBaseFormat(fs, domain.FormatClaude)

	assert.NotNil(t, base)
	assert.Equal(t, domain.FormatClaude, base.formatType)
	assert.NotNil(t, base.fs)
	assert.NotNil(t, base.templateEngine)
}

func TestBaseFormat_ValidateRule(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	tests := []struct {
		name           string
		rule           *domain.Rule
		expectValid    bool
		expectErrors   int
		expectWarnings int
	}{
		{
			name: "valid rule",
			rule: &domain.Rule{
				ID:          "[contexture:test/rule]",
				Title:       "Test Rule",
				Description: "A test rule",
				Tags:        []string{"test"},
				Content:     "Rule content",
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "missing required fields",
			rule: &domain.Rule{
				Description: "A test rule",
				Tags:        []string{"test"},
			},
			expectValid:    false,
			expectErrors:   3, // ID, Title, Content
			expectWarnings: 0,
		},
		{
			name: "missing recommended fields",
			rule: &domain.Rule{
				ID:      "[contexture:test/rule]",
				Title:   "Test Rule",
				Content: "Rule content",
			},
			expectValid:    true,
			expectErrors:   0,
			expectWarnings: 2, // Description, Tags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.ValidateRule(tt.rule)

			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Len(t, result.Errors, tt.expectErrors)
			assert.Len(t, result.Warnings, tt.expectWarnings)
			assert.Equal(t, "claude", result.Metadata["format"])
		})
	}
}

func TestBaseFormat_ProcessTemplate(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	rule := &domain.Rule{
		ID:          "[contexture:test/rule]",
		Title:       "Test Rule",
		Description: "A test rule",
		Tags:        []string{"test", "example"},
		Content:     "Rule content",
		Source:      "https://github.com/test/repo.git",
		Ref:         "main",
	}

	template := "# {{.title}}\n\n{{.description}}\n\nTags: {{range .tags}}{{.}} {{end}}\n\n{{.content}}"

	result, err := base.ProcessTemplate(rule, template)

	require.NoError(t, err)
	assert.Contains(t, result, "# Test Rule")
	assert.Contains(t, result, "A test rule")
	assert.Contains(t, result, "test example")
	assert.Contains(t, result, "Rule content")
}

func TestBaseFormat_ProcessTemplateWithVars(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	rule := &domain.Rule{
		ID:      "[contexture:test/rule]",
		Title:   "Test Rule",
		Content: "Rule content",
	}

	template := "# {{.title}}\n\nCustom: {{.custom}}\n\n{{.content}}"
	additionalVars := map[string]any{
		"custom": "custom value",
	}

	result, err := base.ProcessTemplateWithVars(rule, template, additionalVars)

	require.NoError(t, err)
	assert.Contains(t, result, "# Test Rule")
	assert.Contains(t, result, "Custom: custom value")
	assert.Contains(t, result, "Rule content")
}

func TestBaseFormat_CreateTransformedRule(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	rule := &domain.Rule{
		ID:      "[contexture:test/rule]",
		Title:   "Test Rule",
		Content: "Rule content",
	}

	content := "transformed content"
	filename := "test.md"
	relativePath := "subdir/test.md"
	metadata := map[string]any{"test": "value"}

	transformed := base.CreateTransformedRule(rule, content, filename, relativePath, metadata)

	assert.Equal(t, rule, transformed.Rule)
	assert.Equal(t, content, transformed.Content)
	assert.Equal(t, filename, transformed.Filename)
	assert.Equal(t, relativePath, transformed.RelativePath)
	assert.Equal(t, "claude", transformed.Metadata["format"])
	assert.Equal(t, filename, transformed.Metadata["filename"])
	assert.Equal(t, relativePath, transformed.Metadata["relativePath"])
	assert.Equal(t, "value", transformed.Metadata["test"])
	assert.NotZero(t, transformed.TransformedAt)
}

func TestBaseFormat_GenerateFilename(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	tests := []struct {
		name     string
		ruleID   string
		expected string
	}{
		{
			name:     "standard rule ID",
			ruleID:   "[contexture:security/input-validation]",
			expected: "security-input-validation.md",
		},
		{
			name:     "rule ID with source",
			ruleID:   "[contexture(custom):typescript/strict-config]",
			expected: "typescript-strict-config.md",
		},
		{
			name:     "rule ID with branch",
			ruleID:   "[contexture:react/hooks,main]",
			expected: "react-hooks.md",
		},
		{
			name:     "invalid rule ID",
			ruleID:   "invalid-rule-id",
			expected: "invalid-rule-id.md",
		},
		{
			name:     "rule ID with special characters",
			ruleID:   "[contexture:test/rule@#$%]",
			expected: "test-rule____.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.GenerateFilename(tt.ruleID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseFormat_ExtractRuleIDFromFilename(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "standard filename",
			filename: "security-input-validation.md",
			expected: "[contexture:security/input/validation]",
		},
		{
			name:     "simple filename",
			filename: "test.md",
			expected: "[contexture:test]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.ExtractRuleIDFromFilename(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseFormat_ExtractTitleFromContent(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "content with title",
			content:  "# Test Title\n\nSome content here",
			expected: "Test Title",
		},
		{
			name:     "content with multiple headers",
			content:  "Some text\n# First Title\n## Second Header\n# Another Title",
			expected: "First Title",
		},
		{
			name:     "content without title",
			content:  "Some content without title",
			expected: "",
		},
		{
			name:     "content with title having extra spaces",
			content:  "#   Spaced Title   \n\nContent",
			expected: "Spaced Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.ExtractTitleFromContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseFormat_FileOperations(t *testing.T) {
	fs := afero.NewMemMapFs()
	base := NewBaseFormat(fs, domain.FormatClaude)

	t.Run("write and read file", func(t *testing.T) {
		content := []byte("test content")
		path := "/test/file.md"

		err := base.WriteFile(path, content)
		require.NoError(t, err)

		readContent, err := base.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, content, readContent)
	})

	t.Run("file exists", func(t *testing.T) {
		path := "/test/existing.md"
		err := base.WriteFile(path, []byte("content"))
		require.NoError(t, err)

		exists, err := base.FileExists(path)
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = base.FileExists("/test/nonexistent.md")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("directory operations", func(t *testing.T) {
		dir := "/test/dir"

		err := base.EnsureDirectory(dir)
		require.NoError(t, err)

		exists, err := base.DirExists(dir)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestBaseFormat_ParseRuleFromContent(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	content := `<!-- Contexture Rule: [contexture:test/rule] -->
<!-- Format: claude -->
<!-- Generated: 2024-01-01 12:00:00 -->

# Test Rule Title

Some content here`

	ruleID, title := base.ParseRuleFromContent(content)

	assert.Equal(t, "[contexture:test/rule]", ruleID)
	assert.Equal(t, "Test Rule Title", title)
}

func TestBaseFormat_CalculateContentHash(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	content := []byte("test content")
	hash := base.CalculateContentHash(content)

	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 hash length

	// Same content should generate same hash
	hash2 := base.CalculateContentHash(content)
	assert.Equal(t, hash, hash2)

	// Different content should generate different hash
	hash3 := base.CalculateContentHash([]byte("different content"))
	assert.NotEqual(t, hash, hash3)
}

func TestBaseFormat_ProcessTemplateWithVars_BooleanConversion(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	tests := []struct {
		name       string
		template   string
		rule       *domain.Rule
		expected   string
		shouldFail bool
	}{
		{
			name:     "string true converts to boolean true",
			template: "{{if .extended}}Extended content{{else}}Basic content{{end}}",
			rule: &domain.Rule{
				ID:      "[contexture:test/boolean]",
				Title:   "Boolean Test",
				Content: "Test content",
				Variables: map[string]any{
					"extended": "true",
				},
			},
			expected: "Extended content",
		},
		{
			name:     "string false converts to boolean false",
			template: "{{if .extended}}Extended content{{else}}Basic content{{end}}",
			rule: &domain.Rule{
				ID:      "[contexture:test/boolean]",
				Title:   "Boolean Test",
				Content: "Test content",
				Variables: map[string]any{
					"extended": "false",
				},
			},
			expected: "Basic content",
		},
		{
			name:     "boolean true stays boolean true",
			template: "{{if .extended}}Extended content{{else}}Basic content{{end}}",
			rule: &domain.Rule{
				ID:      "[contexture:test/boolean]",
				Title:   "Boolean Test",
				Content: "Test content",
				Variables: map[string]any{
					"extended": true,
				},
			},
			expected: "Extended content",
		},
		{
			name:     "boolean false stays boolean false",
			template: "{{if .extended}}Extended content{{else}}Basic content{{end}}",
			rule: &domain.Rule{
				ID:      "[contexture:test/boolean]",
				Title:   "Boolean Test",
				Content: "Test content",
				Variables: map[string]any{
					"extended": false,
				},
			},
			expected: "Basic content",
		},
		{
			name:     "non-boolean string stays as string",
			template: "{{.value}}",
			rule: &domain.Rule{
				ID:      "[contexture:test/string]",
				Title:   "String Test",
				Content: "Test content",
				Variables: map[string]any{
					"value": "hello",
				},
			},
			expected: "hello",
		},
		{
			name:     "empty string stays as empty string",
			template: "{{if .value}}Has value{{else}}No value{{end}}",
			rule: &domain.Rule{
				ID:      "[contexture:test/empty]",
				Title:   "Empty Test",
				Content: "Test content",
				Variables: map[string]any{
					"value": "",
				},
			},
			expected: "No value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base.ProcessTemplateWithVars(tt.rule, tt.template, map[string]any{})

			if tt.shouldFail {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseFormat_ProcessTemplateWithVars_RuleIDVariables(t *testing.T) {
	base := NewBaseFormat(afero.NewMemMapFs(), domain.FormatClaude)

	tests := []struct {
		name     string
		ruleID   string
		template string
		expected string
	}{
		{
			name:     "rule ID with boolean false variable",
			ruleID:   `[contexture:languages/go/testing]{"extended":"false"}`,
			template: "{{if .extended}}Extended template{{else}}Basic template{{end}}",
			expected: "Basic template",
		},
		{
			name:     "rule ID with boolean true variable",
			ruleID:   `[contexture:languages/go/testing]{"extended":"true"}`,
			template: "{{if .extended}}Extended template{{else}}Basic template{{end}}",
			expected: "Extended template",
		},
		{
			name:     "rule ID with string variable",
			ruleID:   `[contexture:test/template]{"name":"TestName"}`,
			template: "Hello {{.name}}!",
			expected: "Hello TestName!",
		},
		{
			name:     "rule ID with multiple variables",
			ruleID:   `[contexture:test/multi]{"extended":"false","name":"Test","count":"5"}`,
			template: "{{.name}}: {{if .extended}}Extended ({{.count}}){{else}}Basic{{end}}",
			expected: "Test: Basic",
		},
		{
			name:     "rule ID with no variables",
			ruleID:   "[contexture:test/simple]",
			template: "{{if .extended}}Extended{{else}}Basic{{end}}",
			expected: "Basic", // .extended should be false/empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the rule ID parser to parse variables from the rule ID
			parser := rule.NewRuleIDParser("https://github.com/contexture-org/rules.git")
			parsedID, err := parser.ParseRuleID(tt.ruleID)
			require.NoError(t, err)

			rule := &domain.Rule{
				ID:        tt.ruleID,
				Title:     "Test Rule",
				Content:   "Test content",
				Variables: parsedID.Variables,
			}

			result, err := base.ProcessTemplateWithVars(rule, tt.template, map[string]any{})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateTrackingComment(t *testing.T) {
	fs := afero.NewMemMapFs()
	base := NewBaseFormat(fs, domain.FormatClaude)

	tests := []struct {
		name      string
		ruleID    string
		variables map[string]any
		expected  string
	}{
		{
			name:      "simple rule ID without variables",
			ruleID:    "[contexture:languages/go/testing]",
			variables: nil,
			expected:  "<!-- id: [contexture:languages/go/testing] -->",
		},
		{
			name:      "rule ID with simple variables",
			ruleID:    "[contexture:languages/go/testing]",
			variables: map[string]any{"extended": true},
			expected:  "<!-- id: [contexture:languages/go/testing]{\"extended\":true} -->",
		},
		{
			name:   "rule ID with complex variables",
			ruleID: "[contexture:templates/readme]",
			variables: map[string]any{
				"project_name": "MyApp",
				"features":     []string{"auth", "logging"},
				"config":       map[string]any{"debug": true, "level": "info"},
			},
			expected: "<!-- id: [contexture:templates/readme]{\"config\":{\"debug\":true,\"level\":\"info\"},\"features\":[\"auth\",\"logging\"],\"project_name\":\"MyApp\"} -->",
		},
		{
			name:      "rule ID already containing variables (no duplication)",
			ruleID:    "[contexture:languages/go/testing]{\"extended\": true}",
			variables: nil,
			expected:  "<!-- id: [contexture:languages/go/testing]{\"extended\": true} -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.CreateTrackingComment(tt.ruleID, tt.variables)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateTrackingCommentFromParsed(t *testing.T) {
	fs := afero.NewMemMapFs()
	base := NewBaseFormat(fs, domain.FormatClaude)

	tests := []struct {
		name     string
		parsed   *domain.ParsedRuleID
		expected string
	}{
		{
			name: "parsed rule without variables",
			parsed: &domain.ParsedRuleID{
				Source:   "https://github.com/contextureai/rules.git",
				RulePath: "languages/go/testing",
				Ref:      "main",
			},
			expected: "<!-- id: [contexture:languages/go/testing] -->",
		},
		{
			name: "parsed rule with variables",
			parsed: &domain.ParsedRuleID{
				Source:   "https://github.com/contextureai/rules.git",
				RulePath: "languages/go/testing",
				Ref:      "main",
				Variables: map[string]any{
					"extended": true,
					"strict":   false,
				},
			},
			expected: "<!-- id: [contexture:languages/go/testing]{\"extended\":true,\"strict\":false} -->",
		},
		{
			name: "parsed rule with custom source",
			parsed: &domain.ParsedRuleID{
				Source:   "https://github.com/custom/rules.git",
				RulePath: "custom/rule",
				Ref:      "main",
				Variables: map[string]any{
					"config": map[string]any{"enabled": true},
				},
			},
			expected: "<!-- id: [contexture(https://github.com/custom/rules.git):custom/rule]{\"config\":{\"enabled\":true}} -->",
		},
		{
			name: "parsed rule with branch and variables",
			parsed: &domain.ParsedRuleID{
				Source:   "https://github.com/contextureai/rules.git",
				RulePath: "typescript/strict",
				Ref:      "v2.0.0",
				Variables: map[string]any{
					"target": "es2022",
					"strict": true,
				},
			},
			expected: "<!-- id: [contexture:typescript/strict,v2.0.0]{\"strict\":true,\"target\":\"es2022\"} -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.CreateTrackingCommentFromParsed(tt.parsed)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingCommentNoDuplication(t *testing.T) {
	fs := afero.NewMemMapFs()
	base := NewBaseFormat(fs, domain.FormatClaude)

	// This test specifically verifies that variables aren't duplicated
	// when using CreateTrackingCommentFromParsed
	parsed := &domain.ParsedRuleID{
		Source:   "https://github.com/contextureai/rules.git",
		RulePath: "languages/go/testing",
		Ref:      "main",
		Variables: map[string]any{
			"extended": true,
		},
	}

	result := base.CreateTrackingCommentFromParsed(parsed)

	// Should only contain variables once, not duplicated
	expected := "<!-- id: [contexture:languages/go/testing]{\"extended\":true} -->"
	assert.Equal(t, expected, result)

	// Verify no duplication by checking that the result doesn't contain
	// the variables JSON twice
	assert.NotContains(t, result, "{\"extended\":true}{\"extended\":true}")
}
