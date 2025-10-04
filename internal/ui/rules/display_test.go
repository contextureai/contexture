package rules

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureOutput captures stdout during test execution
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	// Save original stdout
	originalStdout := os.Stdout
	defer func() {
		os.Stdout = originalStdout
	}()

	// Create a pipe to capture output
	r, w, err := os.Pipe()
	require.NoError(t, err)

	// Replace stdout with our pipe writer
	os.Stdout = w

	// Channel to collect output
	outputChan := make(chan string, 1)

	// Start goroutine to read from pipe
	go func() {
		defer func() { _ = r.Close() }()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputChan <- buf.String()
	}()

	// Execute the function
	fn()

	// Close writer to signal EOF to reader
	_ = w.Close()

	// Get the captured output
	return <-outputChan
}

func TestDisplayRuleList_EmptyRules(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	output := captureOutput(t, func() {
		err := DisplayRuleList([]*domain.Rule{}, DefaultDisplayOptions())
		assert.NoError(t, err)
	})

	assert.Equal(t, "No rules found.\n", output)
}

func TestDisplayRuleList_SingleDefaultRule(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:          "[contexture:languages/go/testing]",
			Title:       "Go Testing Best Practices",
			Description: "Comprehensive guide for Go testing",
			Tags:        []string{"go", "testing"},
			Source:      "",
		},
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, DefaultDisplayOptions())
		assert.NoError(t, err)
	})

	// Should contain header without count
	assert.Contains(t, output, "Installed Rules")
	assert.NotContains(t, output, "(1 total)")

	// Should contain rule path and title
	assert.Contains(t, output, "languages/go/testing")
	assert.Contains(t, output, "Go Testing Best Practices")

	// Should not contain source line for default rules
	assert.NotContains(t, output, "Source:")
	assert.NotContains(t, output, "contexture")
}

func TestDisplayRuleList_CustomGitSourceRule(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:          "[contexture(git@github.com:user/repo.git):test/example,feature]",
			Title:       "Custom Rule Example",
			Description: "A rule from a custom repository",
			Source:      "git@github.com:user/repo.git",
			Ref:         "feature",
		},
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, DefaultDisplayOptions())
		assert.NoError(t, err)
	})

	// Should contain simple rule path (not full ID)
	assert.Contains(t, output, "test/example")
	assert.NotContains(t, output, "git@github.com:user/repo.git")

	// Should contain title
	assert.Contains(t, output, "Custom Rule Example")

	// Should contain source line without "Source:" prefix
	assert.Contains(t, output, "git@github.com:user/repo (feature)")
	assert.NotContains(t, output, "Source:")
}

func TestDisplayRuleList_LocalRule(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:     "local/custom-rule",
			Title:  "Local Custom Rule",
			Source: "local",
		},
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, DefaultDisplayOptions())
		assert.NoError(t, err)
	})

	// Should contain rule path and title
	assert.Contains(t, output, "local/custom-rule")
	assert.Contains(t, output, "Local Custom Rule")

	// Should contain source line for local rules
	assert.Contains(t, output, "local")
	assert.NotContains(t, output, "Source:")
}

func TestDisplayRuleList_MultipleMixedRules(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:     "[contexture:security/validation]",
			Title:  "Input Validation",
			Source: "",
		},
		{
			ID:     "[contexture(git@github.com:user/repo.git):custom/rule]",
			Title:  "Custom Security Rule",
			Source: "git@github.com:user/repo.git",
			Ref:    "main",
		},
		{
			ID:     "local/project-rule",
			Title:  "Project Specific Rule",
			Source: "local",
		},
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, DefaultDisplayOptions())
		assert.NoError(t, err)
	})

	// Should contain all rule paths
	assert.Contains(t, output, "security/validation")
	assert.Contains(t, output, "custom/rule")
	assert.Contains(t, output, "local/project-rule")

	// Should contain all titles
	assert.Contains(t, output, "Input Validation")
	assert.Contains(t, output, "Custom Security Rule")
	assert.Contains(t, output, "Project Specific Rule")

	// Should have appropriate source lines
	assert.Contains(t, output, "git@github.com:user/repo")
	assert.Contains(t, output, "local")

	// Default rule should not have source
	lines := strings.Split(output, "\n")
	validationRuleIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "security/validation") {
			validationRuleIndex = i
			break
		}
	}
	assert.GreaterOrEqual(t, validationRuleIndex, 0, "Should find validation rule")

	// Next line should be title, line after that should be empty or next rule
	if validationRuleIndex+2 < len(lines) {
		nextLine := strings.TrimSpace(lines[validationRuleIndex+2])
		// Should not contain source info
		assert.NotContains(t, nextLine, "github.com")
		assert.NotContains(t, nextLine, "git@")
	}
}

func TestDisplayRuleList_WithTags(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:    "[contexture:languages/go/testing]",
			Title: "Go Testing",
			Tags:  []string{"go", "testing", "best-practices"},
		},
	}

	options := DisplayOptions{
		ShowSource: false,
		ShowTags:   true,
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, options)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Tags: go, testing, best-practices")
}

func TestDisplayRuleList_WithTriggers(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:    "[contexture:languages/go/testing]",
			Title: "Go Testing",
			Trigger: &domain.RuleTrigger{
				Type:  domain.TriggerGlob,
				Globs: []string{"**/*_test.go", "**/test_*.go"},
			},
		},
	}

	options := DisplayOptions{
		ShowSource:   false,
		ShowTriggers: true,
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, options)
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Trigger: glob (**/*_test.go, **/test_*.go)")
}

func TestDisplayRuleList_WithVariables(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:    "[contexture:security/validation]",
			Title: "Input Validation",
			Variables: map[string]any{
				"strict_mode": true,
				"max_length":  255,
				"encoding":    "utf-8",
			},
			DefaultVariables: map[string]any{
				"strict_mode": false,
				"max_length":  100,
			},
		},
	}

	options := DisplayOptions{
		ShowSource:    false,
		ShowVariables: true,
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, options)
		assert.NoError(t, err)
	})

	// Should show variables since they differ from defaults
	assert.Contains(t, output, "Variables:")
	assert.Contains(t, output, "strict_mode=true")
	assert.Contains(t, output, "max_length=255")
	assert.Contains(t, output, "encoding=utf-8")
}

func TestDisplayRuleList_NoVariablesWhenSameAsDefaults(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	rules := []*domain.Rule{
		{
			ID:    "[contexture:security/validation]",
			Title: "Input Validation",
			Variables: map[string]any{
				"strict_mode": false,
				"max_length":  100,
			},
			DefaultVariables: map[string]any{
				"strict_mode": false,
				"max_length":  100,
			},
		},
	}

	options := DisplayOptions{
		ShowSource:    false,
		ShowVariables: true,
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, options)
		assert.NoError(t, err)
	})

	// Should not show variables when they match defaults
	assert.NotContains(t, output, "Variables:")
}

func TestExtractSimpleRulePath(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	tests := []struct {
		name     string
		ruleID   string
		expected string
	}{
		{
			name:     "default repository rule",
			ruleID:   "[contexture:languages/go/testing]",
			expected: "languages/go/testing",
		},
		{
			name:     "custom repository rule",
			ruleID:   "[contexture(git@github.com:user/repo.git):test/example]",
			expected: "test/example",
		},
		{
			name:     "custom repository with branch",
			ruleID:   "[contexture(git@github.com:user/repo.git):test/example,feature]",
			expected: "test/example",
		},
		{
			name:     "local rule",
			ruleID:   "local/custom-rule",
			expected: "local/custom-rule",
		},
		{
			name:     "simple path",
			ruleID:   "simple/path",
			expected: "simple/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSimpleRulePath(tt.ruleID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatSourceForDisplay(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	tests := []struct {
		name     string
		rule     *domain.Rule
		expected string
	}{
		{
			name: "default repository",
			rule: &domain.Rule{
				Source: "",
			},
			expected: "",
		},
		{
			name: "default repository explicit",
			rule: &domain.Rule{
				Source: domain.DefaultRepository,
			},
			expected: "",
		},
		{
			name: "local rule",
			rule: &domain.Rule{
				Source: "local",
			},
			expected: "local",
		},
		{
			name: "custom git source with default branch",
			rule: &domain.Rule{
				Source: "git@github.com:user/repo.git",
				Ref:    "main",
			},
			expected: "git@github.com:user/repo",
		},
		{
			name: "custom git source with feature branch",
			rule: &domain.Rule{
				Source: "git@github.com:user/repo.git",
				Ref:    "feature-branch",
			},
			expected: "git@github.com:user/repo (feature-branch)",
		},
		{
			name: "custom source with branch",
			rule: &domain.Rule{
				Source: "custom-source",
				Ref:    "develop",
			},
			expected: "custom-source (develop)",
		},
		{
			name: "custom source without branch",
			rule: &domain.Rule{
				Source: "custom-source",
			},
			expected: "custom-source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSourceForDisplay(tt.rule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTrigger(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	tests := []struct {
		name     string
		trigger  *domain.RuleTrigger
		expected string
	}{
		{
			name:     "nil trigger",
			trigger:  nil,
			expected: "",
		},
		{
			name: "glob trigger with patterns",
			trigger: &domain.RuleTrigger{
				Type:  domain.TriggerGlob,
				Globs: []string{"**/*.go", "**/*.ts"},
			},
			expected: "glob (**/*.go, **/*.ts)",
		},
		{
			name: "glob trigger without patterns",
			trigger: &domain.RuleTrigger{
				Type:  domain.TriggerGlob,
				Globs: []string{},
			},
			expected: "glob",
		},
		{
			name: "always trigger",
			trigger: &domain.RuleTrigger{
				Type: domain.TriggerAlways,
			},
			expected: "always",
		},
		{
			name: "manual trigger",
			trigger: &domain.RuleTrigger{
				Type: domain.TriggerManual,
			},
			expected: "manual",
		},
		{
			name: "model trigger",
			trigger: &domain.RuleTrigger{
				Type: domain.TriggerModel,
			},
			expected: "model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTrigger(tt.trigger)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatVariables(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	tests := []struct {
		name      string
		variables map[string]any
		expected  string
	}{
		{
			name:      "nil variables",
			variables: nil,
			expected:  "",
		},
		{
			name:      "empty variables",
			variables: map[string]any{},
			expected:  "",
		},
		{
			name: "string variable",
			variables: map[string]any{
				"mode": "strict",
			},
			expected: "mode=strict",
		},
		{
			name: "boolean variable",
			variables: map[string]any{
				"enabled": true,
			},
			expected: "enabled=true",
		},
		{
			name: "integer variable",
			variables: map[string]any{
				"count": 42,
			},
			expected: "count=42",
		},
		{
			name: "multiple variables",
			variables: map[string]any{
				"mode":    "strict",
				"enabled": true,
				"count":   42,
			},
			// Note: map iteration order is not guaranteed, so we check components
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVariables(tt.variables)

			if tt.name == "multiple variables" {
				// For multiple variables, check that all components are present
				assert.Contains(t, result, "mode=strict")
				assert.Contains(t, result, "enabled=true")
				assert.Contains(t, result, "count=42")
				assert.Contains(t, result, ", ")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDefaultDisplayOptions(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	options := DefaultDisplayOptions()

	// Should show source by default
	assert.True(t, options.ShowSource)

	// Should not show optional metadata by default
	assert.False(t, options.ShowTriggers)
	assert.False(t, options.ShowVariables)
	assert.False(t, options.ShowTags)
}

func TestDisplayRuleList_SortingOrder(t *testing.T) {
	// t.Parallel() // Removed due to stdout capture

	// Rules in deliberately unsorted order
	rules := []*domain.Rule{
		{
			ID:    "[contexture:security/validation]",
			Title: "Validation Rules",
		},
		{
			ID:    "[contexture:languages/go/testing]",
			Title: "Go Testing",
		},
		{
			ID:    "[contexture:languages/python/style]",
			Title: "Python Style",
		},
		{
			ID:    "[contexture:deployment/docker]",
			Title: "Docker Best Practices",
		},
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, DisplayOptions{ShowSource: false})
		assert.NoError(t, err)
	})

	lines := strings.Split(output, "\n")

	// Find rule path lines (they don't start with spaces)
	var rulePaths []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(line, " ") && !strings.Contains(trimmed, "Installed Rules") {
			rulePaths = append(rulePaths, trimmed)
		}
	}

	// Should be sorted alphabetically
	expected := []string{
		"deployment/docker",
		"languages/go/testing",
		"languages/python/style",
		"security/validation",
	}

	assert.Equal(t, expected, rulePaths)
}

func TestFilterRulesByPattern_EmptyPattern(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "test-rule",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"testing"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, rules, filtered)
}

func TestFilterRulesByPattern_InvalidRegex(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "test-rule",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"testing"},
		},
	}

	_, err := filterRulesByPattern(rules, "[invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing regexp")
}

func TestFilterRulesByPattern_MatchesID(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "contexture:test-rule",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"testing"},
		},
		{
			ID:          "contexture:other-rule",
			Title:       "Other Rule",
			Description: "Another rule",
			Tags:        []string{"other"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "test-rule")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "contexture:test-rule", filtered[0].ID)
}

func TestFilterRulesByPattern_MatchesTitle(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Go Testing Rules",
			Description: "Rules for testing Go code",
			Tags:        []string{"testing"},
		},
		{
			ID:          "rule2",
			Title:       "Python Style Guide",
			Description: "Style rules for Python",
			Tags:        []string{"style"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "Go Testing")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Go Testing Rules", filtered[0].Title)
}

func TestFilterRulesByPattern_MatchesDescription(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Test Rule",
			Description: "Comprehensive testing guidelines for Go applications",
			Tags:        []string{"testing"},
		},
		{
			ID:          "rule2",
			Title:       "Style Rule",
			Description: "Code formatting rules",
			Tags:        []string{"style"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "testing guidelines")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Test Rule", filtered[0].Title)
}

func TestFilterRulesByPattern_MatchesTags(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"testing", "validation"},
		},
		{
			ID:          "rule2",
			Title:       "Style Rule",
			Description: "A style rule",
			Tags:        []string{"style", "formatting"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "validation")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Test Rule", filtered[0].Title)
}

func TestFilterRulesByPattern_MatchesFrameworks(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "React Rule",
			Description: "Rules for React",
			Tags:        []string{"frontend"},
			Frameworks:  []string{"react", "nextjs"},
		},
		{
			ID:          "rule2",
			Title:       "Vue Rule",
			Description: "Rules for Vue",
			Tags:        []string{"frontend"},
			Frameworks:  []string{"vue", "nuxt"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "nextjs")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "React Rule", filtered[0].Title)
}

func TestFilterRulesByPattern_MatchesLanguages(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Go Rule",
			Description: "Rules for Go",
			Tags:        []string{"backend"},
			Languages:   []string{"go", "golang"},
		},
		{
			ID:          "rule2",
			Title:       "Python Rule",
			Description: "Rules for Python",
			Tags:        []string{"backend"},
			Languages:   []string{"python"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "golang")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Go Rule", filtered[0].Title)
}

func TestFilterRulesByPattern_MatchesSource(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Default Rule",
			Description: "Rule from default source",
			Tags:        []string{"default"},
			Source:      "contexture:default",
		},
		{
			ID:          "rule2",
			Title:       "Custom Rule",
			Description: "Rule from custom source",
			Tags:        []string{"custom"},
			Source:      "myorg:custom-rules",
		},
	}

	filtered, err := filterRulesByPattern(rules, "myorg")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Custom Rule", filtered[0].Title)
}

func TestFilterRulesByPattern_CaseInsensitive(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Test Rule",
			Description: "Testing Guidelines",
			Tags:        []string{"TESTING"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "testing")
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Test Rule", filtered[0].Title)
}

func TestFilterRulesByPattern_RegexPattern(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Test Rule 1",
			Description: "First test rule",
			Tags:        []string{"testing"},
		},
		{
			ID:          "rule2",
			Title:       "Test Rule 2",
			Description: "Second test rule",
			Tags:        []string{"testing"},
		},
		{
			ID:          "rule3",
			Title:       "Style Rule",
			Description: "Style rule",
			Tags:        []string{"style"},
		},
	}

	// Match rules ending with digit
	filtered, err := filterRulesByPattern(rules, "Rule \\d$")
	require.NoError(t, err)
	assert.Len(t, filtered, 2)
	assert.Equal(t, "Test Rule 1", filtered[0].Title)
	assert.Equal(t, "Test Rule 2", filtered[1].Title)
}

func TestFilterRulesByPattern_NoMatches(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"testing"},
		},
	}

	filtered, err := filterRulesByPattern(rules, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, filtered)
}

func TestDisplayRuleList_WithPattern(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "contexture:languages/go/testing",
			Title:       "Go Testing Standards",
			Description: "Comprehensive testing guidelines for Go applications",
			Tags:        []string{"testing", "go", "standards"},
			Source:      "contexture:default",
		},
		{
			ID:          "contexture:languages/python/style",
			Title:       "Python Style Guide",
			Description: "PEP 8 style guidelines for Python code",
			Tags:        []string{"style", "python", "pep8"},
			Source:      "contexture:default",
		},
	}

	options := DisplayOptions{
		ShowSource: true,
		Pattern:    "go",
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, options)
		require.NoError(t, err)
	})

	// Should show only the Go rule
	assert.Contains(t, output, "Go Testing Standards")
	assert.NotContains(t, output, "Python Style Guide")
	assert.Contains(t, output, "pattern: go") // Should show pattern in header
}

func TestDisplayRuleList_WithPatternNoMatches(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "contexture:languages/go/testing",
			Title:       "Go Testing Standards",
			Description: "Comprehensive testing guidelines for Go applications",
			Tags:        []string{"testing", "go", "standards"},
			Source:      "contexture:default",
		},
	}

	options := DisplayOptions{
		ShowSource: true,
		Pattern:    "nonexistent",
	}

	output := captureOutput(t, func() {
		err := DisplayRuleList(rules, options)
		require.NoError(t, err)
	})

	// Should show no matches message with pattern
	assert.Contains(t, output, "No rules found matching pattern: nonexistent")
	assert.NotContains(t, output, "Go Testing Standards")
}

func TestDisplayRuleList_WithInvalidPattern(t *testing.T) {
	rules := []*domain.Rule{
		{
			ID:          "rule1",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"testing"},
		},
	}

	options := DisplayOptions{
		Pattern: "[invalid",
	}

	err := DisplayRuleList(rules, options)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed for pattern")
}
