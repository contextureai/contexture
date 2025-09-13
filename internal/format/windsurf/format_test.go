package windsurf

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWindsurfOutputDir = "/output/.windsurf/rules"

func TestNewFormat(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	assert.NotNil(t, f)
}

func TestNewFormatWithMode(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	t.Run("single file mode", func(t *testing.T) {
		f := NewFormatWithMode(fs, ModeSingleFile)
		assert.Equal(t, ModeSingleFile, f.mode)
	})

	t.Run("multi file mode", func(t *testing.T) {
		f := NewFormatWithMode(fs, ModeMultiFile)
		assert.Equal(t, ModeMultiFile, f.mode)
	})
}

func TestFormat_SetMode(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	// Default mode is ModeMultiFile
	assert.Equal(t, ModeMultiFile, f.GetMode())

	f.SetMode(ModeSingleFile)
	assert.Equal(t, ModeSingleFile, f.GetMode())

	f.SetMode(ModeMultiFile)
	assert.Equal(t, ModeMultiFile, f.GetMode())
}

func TestFormat_Transform_SingleFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeSingleFile)

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
	assert.Equal(t, "rules.md", transformed.Filename)
	assert.Equal(t, ".windsurf/rules/rules.md", transformed.RelativePath)
	assert.NotZero(t, transformed.TransformedAt)

	// Check content
	assert.Contains(t, transformed.Content, "Authentication Rule")
	assert.Contains(t, transformed.Content, "A rule for secure authentication")
	assert.Contains(t, transformed.Content, "Always validate user credentials")

	// Check metadata
	assert.Equal(t, "windsurf", transformed.Metadata["format"])
	assert.Equal(t, "single", transformed.Metadata["mode"])
	assert.Equal(t, "rules.md", transformed.Metadata["filename"])
}

func TestFormat_Transform_MultiFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeMultiFile)

	rule := &domain.Rule{
		ID:          "[contexture:security/authentication]",
		Title:       "Authentication Rule",
		Description: "A rule for secure authentication",
		Tags:        []string{"security", "auth"},
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
	assert.Equal(t, "security-authentication.md", transformed.Filename)
	assert.Equal(t, ".windsurf/rules/security-authentication.md", transformed.RelativePath)

	// Check metadata
	assert.Equal(t, "windsurf", transformed.Metadata["format"])
	assert.Equal(t, "multi", transformed.Metadata["mode"])
}

func TestFormat_Validate(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	tests := []struct {
		name         string
		mode         OutputMode
		rule         *domain.Rule
		wantValid    bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name: "valid rule - single file",
			mode: ModeSingleFile,
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
			name: "valid rule - multi file",
			mode: ModeMultiFile,
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
			name: "long filename warning - multi file",
			mode: ModeMultiFile,
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
			name:         "invalid rule - missing fields",
			mode:         ModeSingleFile,
			rule:         &domain.Rule{},
			wantValid:    false,
			wantErrors:   3, // missing ID, title, content
			wantWarnings: 2, // missing description, tags
		},
		{
			name: "invalid rule - exceeds character limit",
			mode: ModeSingleFile,
			rule: &domain.Rule{
				ID:          "[contexture:test/long]",
				Title:       "Long Rule",
				Description: "A rule with content exceeding Windsurf character limit",
				Tags:        []string{"test"},
				Content:     strings.Repeat("This is very long content. ", 450), // ~12150 chars, exceeds 12000 limit
			},
			wantValid:    false,
			wantErrors:   1, // character limit exceeded
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatWithMode(fs, tt.mode)
			result, err := f.Validate(tt.rule)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantValid, result.Valid)
			assert.Len(t, result.Errors, tt.wantErrors)
			assert.Len(t, result.Warnings, tt.wantWarnings)
			assert.Equal(t, "windsurf", result.Metadata["format"])
			assert.Equal(t, string(tt.mode), result.Metadata["mode"])
		})
	}
}

func TestFormat_Write_SingleFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeSingleFile)

	rules := []*domain.TransformedRule{
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule1]",
				Title: "Rule 1",
			},
			Content:  "Content of rule 1",
			Filename: "rules.md",
		},
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule2]",
				Title: "Rule 2",
			},
			Content:  "Content of rule 2",
			Filename: "rules.md",
		},
	}

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	err := f.Write(rules, config)
	require.NoError(t, err)

	// Check that single file was created
	content, err := afero.ReadFile(fs, "/output/.windsurf/rules/rules.md")
	require.NoError(t, err)
	assert.NotEmpty(t, content)

	contentStr := string(content)

	// Check header
	assert.Contains(t, contentStr, "# Windsurf Rules")
	assert.Contains(t, contentStr, "2 contexture rules")
	assert.Contains(t, contentStr, "Mode: Single File")

	// Check rules content
	assert.Contains(t, contentStr, "Content of rule 1")
	assert.Contains(t, contentStr, "Content of rule 2")

	// Check separator
	assert.Contains(t, contentStr, "---")

	// Check footer
	assert.Contains(t, contentStr, "generated by Contexture CLI")
	assert.Contains(t, contentStr, "single-file mode")
}

func TestFormat_Write_MultiFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeMultiFile)

	rules := []*domain.TransformedRule{
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule1]",
				Title: "Rule 1",
			},
			Content:      "Content of rule 1",
			Filename:     "test-rule1.md",
			RelativePath: ".windsurf/rules/test-rule1.md",
		},
		{
			Rule: &domain.Rule{
				ID:    "[contexture:test/rule2]",
				Title: "Rule 2",
			},
			Content:      "Content of rule 2",
			Filename:     "test-rule2.md",
			RelativePath: ".windsurf/rules/test-rule2.md",
		},
	}

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	err := f.Write(rules, config)
	require.NoError(t, err)

	// Check that individual files were created
	outputDir := testWindsurfOutputDir

	// Check rule1 file - now uses new tracking comment format
	content1, err := afero.ReadFile(fs, filepath.Join(outputDir, "test-rule1.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content1), "Content of rule 1")
	assert.Contains(t, string(content1), "<!-- id: [contexture:test/rule1] -->")

	// Check rule2 file
	content2, err := afero.ReadFile(fs, filepath.Join(outputDir, "test-rule2.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content2), "Content of rule 2")
}

func TestFormat_Remove_SingleFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeSingleFile)

	// Create test file
	outputDir := testWindsurfOutputDir
	err := fs.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)

	testFile := filepath.Join(outputDir, "rules.md")
	err = afero.WriteFile(fs, testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	// Test removing from single file (should not error but may not actually remove)
	err = f.Remove("[contexture:test/rule]", config)
	require.NoError(t, err)
}

func TestFormat_Remove_MultiFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeMultiFile)

	// Create test file
	outputDir := testWindsurfOutputDir
	err := fs.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)

	testFile := filepath.Join(outputDir, "test-rule.md")
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
}

func TestFormat_List_SingleFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeSingleFile)

	t.Run("file does not exist", func(t *testing.T) {
		config := &domain.FormatConfig{
			BaseDir: "/output",
		}

		rules, err := f.List(config)
		require.NoError(t, err)
		assert.Empty(t, rules)
	})

	t.Run("file exists", func(t *testing.T) {
		// Create test file
		outputDir := testWindsurfOutputDir
		err := fs.MkdirAll(outputDir, 0o755)
		require.NoError(t, err)

		content := `# Windsurf Rules

This file contains rules for Windsurf.

---

# Test Windsurf File

Some content here

<!-- id: [contexture:test/rule] -->`
		err = afero.WriteFile(fs, filepath.Join(outputDir, "rules.md"), []byte(content), 0o644)
		require.NoError(t, err)

		config := &domain.FormatConfig{
			BaseDir: "/output",
		}

		rules, err := f.List(config)
		require.NoError(t, err)
		assert.Len(t, rules, 1)

		rule := rules[0]
		assert.Equal(t, "[contexture:test/rule]", rule.ID())
		assert.Equal(t, "Test Windsurf File", rule.Title())
		assert.Equal(t, "rules.md", rule.Filename)
		assert.Equal(t, ".windsurf/rules/rules.md", rule.RelativePath)
	})
}

func TestFormat_List_MultiFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeMultiFile)

	// Create test directory and files
	outputDir := testWindsurfOutputDir
	err := fs.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)

	// Create test rule files
	rule1Content := `<!-- Contexture Rule: [contexture:test/rule1] -->
# Test Rule 1

This is a test rule.`

	rule2Content := `<!-- Contexture Rule: [contexture:test/rule2] -->
# Test Rule 2

Another test rule.`

	err = afero.WriteFile(fs, filepath.Join(outputDir, "test-rule1.md"), []byte(rule1Content), 0o644)
	require.NoError(t, err)

	err = afero.WriteFile(fs, filepath.Join(outputDir, "test-rule2.md"), []byte(rule2Content), 0o644)
	require.NoError(t, err)

	// Create index file (should be ignored)
	err = afero.WriteFile(fs, filepath.Join(outputDir, "index.md"), []byte("index content"), 0o644)
	require.NoError(t, err)

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	rules, err := f.List(config)
	require.NoError(t, err)
	assert.Len(t, rules, 2)

	// Check first rule
	rule1 := findRuleByFilename(rules, "test-rule1.md")
	assert.NotNil(t, rule1)
	assert.Equal(t, "[contexture:test/rule1]", rule1.ID())
	assert.Equal(t, "Test Rule 1", rule1.Title())
	assert.Equal(t, "test-rule1.md", rule1.Filename)

	// Check second rule
	rule2 := findRuleByFilename(rules, "test-rule2.md")
	assert.NotNil(t, rule2)
	assert.Equal(t, "[contexture:test/rule2]", rule2.ID())
	assert.Equal(t, "Test Rule 2", rule2.Title())
}

func TestFormat_generateMultiFileFilename(t *testing.T) {
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
			want:   "security-auth.md",
		},
		{
			name:   "complex path",
			ruleID: "[contexture:javascript/react/hooks]",
			want:   "javascript-react-hooks.md",
		},
		{
			name:   "with source",
			ruleID: "[contexture(github.com/test/repo):security/auth]",
			want:   "security-auth.md",
		},
		{
			name:   "invalid format fallback",
			ruleID: "invalid-rule-id",
			want:   "invalid-rule-id.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the BaseFormat.GenerateFilename method instead
			got := f.GenerateFilename(tt.ruleID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_getSingleFileFilename(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	filename := f.GetSingleFileFilename()
	assert.Equal(t, "rules.md", filename)
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
			want:   ".windsurf/rules",
		},
		{
			name:   "empty base dir",
			config: &domain.FormatConfig{},
			want:   ".windsurf/rules",
		},
		{
			name: "with base dir",
			config: &domain.FormatConfig{
				BaseDir: "/output",
			},
			want: testWindsurfOutputDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.getOutputDir(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
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
			name:    "no h1 header",
			content: "## H2 Header\n\nContent",
			want:    "",
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

func TestFormat_getDefaultTemplate(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	template := f.GetDefaultTemplate()
	assert.NotEmpty(t, template)
	assert.Contains(t, template, "{{.title}}")
	assert.Contains(t, template, "{{.content}}")
	// Check for YAML frontmatter
	assert.Contains(t, template, "---")
	assert.Contains(t, template, "trigger:")
}

func TestFormat_Headers(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormat(fs)

	t.Run("single file header", func(t *testing.T) {
		header := f.getSingleFileHeader(5)
		assert.Contains(t, header, "# Windsurf Rules")
		assert.Contains(t, header, "5 contexture rules")
		assert.Contains(t, header, "Mode: Single File")
	})

	t.Run("single file footer", func(t *testing.T) {
		footer := f.getSingleFileFooter()
		assert.Contains(t, footer, "single-file mode")
		assert.Contains(t, footer, "Do not edit manually")
	})
}

func TestFormat_Write_CharacterLimitValidation(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeSingleFile) // Use single-file mode

	// Create rules that should succeed since no total limit exists (only per-file limit of 12000 chars)
	rules := []*domain.TransformedRule{
		{
			Rule: &domain.Rule{
				ID:      "[contexture:test/rule1]",
				Title:   "Rule 1",
				Content: strings.Repeat("Very long content. ", 200), // ~3800 chars
			},
			Content:  strings.Repeat("Very long content. ", 200), // ~3800 chars
			Filename: "rule1.md",
		},
		{
			Rule: &domain.Rule{
				ID:      "[contexture:test/rule2]",
				Title:   "Rule 2",
				Content: strings.Repeat("Very long content. ", 200), // ~3800 chars
			},
			Content:  strings.Repeat("Very long content. ", 200), // ~3800 chars
			Filename: "rule2.md",
		},
		{
			Rule: &domain.Rule{
				ID:      "[contexture:test/rule3]",
				Title:   "Rule 3",
				Content: strings.Repeat("Very long content. ", 250), // ~4750 chars
			},
			Content:  strings.Repeat("Very long content. ", 250), // ~4750 chars
			Filename: "rule3.md",
		},
		// Total: ~12350 chars, but no total limit so should succeed
	}

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	// Should succeed since no total character limit exists (only per-file limit)
	err := f.Write(rules, config)
	require.NoError(t, err)

	// Check that file was created
	exists, err := afero.Exists(fs, "/output/.windsurf/rules/rules.md")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFormat_Write_CharacterLimitValidation_MultiFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	f := NewFormatWithMode(fs, ModeMultiFile) // Use multi-file mode to test per-file limit

	// Create rules where one exceeds the per-file character limit (12000 chars)
	rules := []*domain.TransformedRule{
		{
			Rule: &domain.Rule{
				ID:      "[contexture:test/rule1]",
				Title:   "Rule 1",
				Content: strings.Repeat("Short content. ", 50), // ~750 chars
			},
			Content:  strings.Repeat("Short content. ", 50), // ~750 chars
			Filename: "rule1.md",
		},
		{
			Rule: &domain.Rule{
				ID:      "[contexture:test/rule2]",
				Title:   "Rule 2",
				Content: strings.Repeat("Very long content. ", 650), // ~12350 chars, exceeds 12000 limit
			},
			Content:  strings.Repeat("Very long content. ", 650), // ~12350 chars
			Filename: "rule2.md",
		},
	}

	config := &domain.FormatConfig{
		BaseDir: "/output",
	}

	// Should fail due to individual rule character limit
	err := f.Write(rules, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds Windsurf per-file limit")
	assert.Contains(t, err.Error(), "12000")
	assert.Contains(t, err.Error(), "[contexture:test/rule2]")
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
