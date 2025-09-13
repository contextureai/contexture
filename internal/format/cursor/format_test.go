package cursor

import (
	"path/filepath"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCursorOutputDir = "/output/.cursor/rules"

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
		ID:          "[contexture:security/authentication]",
		Title:       "Authentication Rule",
		Description: "A rule for secure authentication",
		Tags:        []string{"security", "auth"},
		Languages:   []string{"javascript", "typescript"},
		Frameworks:  []string{"express", "fastify"},
		Content:     "Always validate user credentials",
		Source:      "https://github.com/test/repo.git",
		Ref:         "main",
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
	assert.Equal(t, rule, transformed.Rule)
	assert.NotEmpty(t, transformed.Content)
	assert.Equal(t, "security-authentication.mdc", transformed.Filename)
	assert.Equal(t, ".cursor/rules/security-authentication.mdc", transformed.RelativePath)
	assert.NotZero(t, transformed.TransformedAt)

	// Check that content contains expected elements
	assert.Contains(t, transformed.Content, "Authentication Rule")
	assert.Contains(t, transformed.Content, "A rule for secure authentication")
	assert.Contains(t, transformed.Content, "Always validate user credentials")

	// Check metadata
	assert.Equal(t, "cursor", transformed.Metadata["format"])
	assert.Equal(t, "security-authentication.mdc", transformed.Metadata["filename"])
	assert.Equal(
		t,
		".cursor/rules/security-authentication.mdc",
		transformed.Metadata["relativePath"],
	)
}

func TestFormat_generateFilename(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	tests := []struct {
		name   string
		ruleID string
		want   string
	}{
		{
			name:   "simple path",
			ruleID: "[contexture:security/auth]",
			want:   "security-auth.mdc",
		},
		{
			name:   "complex path",
			ruleID: "[contexture:javascript/react/hooks]",
			want:   "javascript-react-hooks.mdc",
		},
		{
			name:   "with source",
			ruleID: "[contexture(github.com/test/repo):security/auth]",
			want:   "security-auth.mdc",
		},
		{
			name:   "with branch",
			ruleID: "[contexture:security/auth,main]",
			want:   "security-auth.mdc",
		},
		{
			name:   "invalid format fallback",
			ruleID: "invalid-rule-id",
			want:   "invalid-rule-id.mdc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.GenerateFilename(tt.ruleID)
			assert.Equal(t, tt.want, got)
		})
	}
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
			name: "long filename warning",
			rule: &domain.Rule{
				ID:      "[contexture:very/long/path/that/will/create/a/very/long/filename/that/might/cause/issues/on/some/filesystems]",
				Title:   "Long Path Rule",
				Content: "Content",
			},
			wantValid:    true,
			wantErrors:   0,
			wantWarnings: 2, // missing description, tags
		},
		{
			name: "invalid rule - missing ID",
			rule: &domain.Rule{
				Title:   "Invalid Rule",
				Content: "Content",
			},
			wantValid:    false,
			wantErrors:   1,
			wantWarnings: 2, // missing description, tags
		},
		{
			name:         "invalid rule - all missing",
			rule:         &domain.Rule{},
			wantValid:    false,
			wantErrors:   3, // missing ID, title, content
			wantWarnings: 2, // missing description, tags
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
			assert.Equal(t, "cursor", result.Metadata["format"])
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
			Content:      "Content of rule 1",
			Filename:     "test-rule1.mdc",
			RelativePath: ".cursor/rules/test-rule1.mdc",
		},
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule2]",
				Title: "Rule 2",
			},
			Content:      "Content of rule 2",
			Filename:     "test-rule2.mdc",
			RelativePath: ".cursor/rules/test-rule2.mdc",
		},
	}

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	err := f.Write(rules, config)
	require.NoError(t, err)

	// Check that files were created
	outputDir := testCursorOutputDir

	// Check rule1 file - now uses new tracking comment format
	content1, err := afero.ReadFile(fs, filepath.Join(outputDir, "test-rule1.mdc"))
	require.NoError(t, err)
	assert.Contains(t, string(content1), "Content of rule 1")
	assert.Contains(t, string(content1), "<!-- id: [contexture:test/rule1] -->")

	// Check rule2 file - now uses new tracking comment format
	content2, err := afero.ReadFile(fs, filepath.Join(outputDir, "test-rule2.mdc"))
	require.NoError(t, err)
	assert.Contains(t, string(content2), "Content of rule 2")
	assert.Contains(t, string(content2), "<!-- id: [contexture:test/rule2] -->")

	// Note: Index file functionality has been removed as it was unused
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

	// Check that directory was not created
	exists, err := afero.DirExists(fs, testCursorOutputDir)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFormat_Remove(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	// Create test file
	outputDir := testCursorOutputDir
	err := fs.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)

	testFile := filepath.Join(outputDir, "test-rule.mdc")
	err = afero.WriteFile(fs, testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	// Test removing existing file
	err = f.Remove("[contexture:test/rule]", config)
	require.NoError(t, err)

	// Check that file was removed
	exists, err := afero.Exists(fs, testFile)
	require.NoError(t, err)
	assert.False(t, exists)

	// Test removing non-existent file (should not error)
	err = f.Remove("[contexture:nonexistent/rule]", config)
	require.NoError(t, err)
}

func TestFormat_List(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	t.Run("directory does not exist", func(t *testing.T) {
		config := &domain.FormatConfig{
			BaseDir: "/output",
		}

		rules, err := f.List(config)
		require.NoError(t, err)
		assert.Empty(t, rules)
	})

	t.Run("directory exists with files", func(t *testing.T) {
		// Create test directory and files
		outputDir := testCursorOutputDir
		err := fs.MkdirAll(outputDir, 0o755)
		require.NoError(t, err)

		// Create test rule files
		rule1Content := `<!-- Contexture Rule: [contexture:test/rule1] -->
# Test Rule 1

This is a test rule.`

		rule2Content := `<!-- Contexture Rule: [contexture:test/rule2] -->
# Test Rule 2

Another test rule.`

		err = afero.WriteFile(
			fs,
			filepath.Join(outputDir, "test-rule1.mdc"),
			[]byte(rule1Content),
			0o644,
		)
		require.NoError(t, err)

		err = afero.WriteFile(
			fs,
			filepath.Join(outputDir, "test-rule2.mdc"),
			[]byte(rule2Content),
			0o644,
		)
		require.NoError(t, err)

		// Create index file (should be ignored)
		err = afero.WriteFile(
			fs,
			filepath.Join(outputDir, "index.md"),
			[]byte("index content"),
			0o644,
		)
		require.NoError(t, err)

		// Create non-md file (should be ignored)
		err = afero.WriteFile(
			fs,
			filepath.Join(outputDir, "other.txt"),
			[]byte("other content"),
			0o644,
		)
		require.NoError(t, err)

		config := &domain.FormatConfig{
			BaseDir: "/output",
		}

		rules, err := f.List(config)
		require.NoError(t, err)
		assert.Len(t, rules, 2)

		// Check first rule
		rule1 := findRuleByFilename(rules, "test-rule1.mdc")
		assert.NotNil(t, rule1)
		assert.Equal(t, "[contexture:test/rule1]", rule1.ID())
		assert.Equal(t, "Test Rule 1", rule1.Title())
		assert.Equal(t, "test-rule1.mdc", rule1.Filename)
		assert.Equal(t, ".cursor/rules/test-rule1.mdc", rule1.RelativePath)
		assert.NotEmpty(t, rule1.ContentHash)

		// Check second rule
		rule2 := findRuleByFilename(rules, "test-rule2.mdc")
		assert.NotNil(t, rule2)
		assert.Equal(t, "[contexture:test/rule2]", rule2.ID())
		assert.Equal(t, "Test Rule 2", rule2.Title())
		assert.Equal(t, "test-rule2.mdc", rule2.Filename)
	})
}

func TestFormat_extractTitleFromContent(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "h1 header at start",
			content: "# Test Title\n\nContent here",
			want:    "Test Title",
		},
		{
			name:    "h1 header with spaces",
			content: "   # Spaced Title   \n\nContent",
			want:    "Spaced Title",
		},
		{
			name:    "no h1 header",
			content: "## H2 Header\n\nContent",
			want:    "",
		},
		{
			name:    "h1 header later in content",
			content: "Some text\n# Later Title\nMore content",
			want:    "Later Title",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.ExtractTitleFromContent(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_extractRuleIDFromFilename(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "simple filename",
			filename: "security-auth.mdc",
			want:     "[contexture:security/auth]",
		},
		{
			name:     "complex filename",
			filename: "javascript-react-hooks.mdc",
			want:     "[contexture:javascript/react/hooks]",
		},
		{
			name:     "single level",
			filename: "test.mdc",
			want:     "[contexture:test]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.ExtractRuleIDFromFilename(tt.filename)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_getOutputDir(t *testing.T) {
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
			want:   ".cursor/rules",
		},
		{
			name:   "empty base dir",
			config: &domain.FormatConfig{},
			want:   ".cursor/rules",
		},
		{
			name: "with base dir",
			config: &domain.FormatConfig{
				BaseDir: "/output",
			},
			want: testCursorOutputDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.getOutputDir(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_getDefaultTemplate(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	template := f.GetDefaultTemplate()
	assert.NotEmpty(t, template)
	assert.Contains(t, template, "{{.title}}")
	assert.Contains(t, template, "{{.content}}")
	assert.Contains(t, template, "{{.description}}")
	// Check for YAML frontmatter
	assert.Contains(t, template, "---")
	assert.Contains(t, template, "alwaysApply:")
}

// Helper function to find a rule by filename
func findRuleByFilename(rules []*domain.InstalledRule, filename string) *domain.InstalledRule {
	for _, rule := range rules {
		if rule.Filename == filename {
			return rule
		}
	}
	return nil
}
