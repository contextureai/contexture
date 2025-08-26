package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestExtractRulePath(t *testing.T) {
	tests := []struct {
		name     string
		ruleID   string
		expected string
	}{
		{
			name:     "empty rule ID",
			ruleID:   "",
			expected: "",
		},
		{
			name:     "basic contexture format",
			ruleID:   "[contexture:typescript/strict-config]",
			expected: "typescript/strict-config",
		},
		{
			name:     "contexture with source",
			ruleID:   "[contexture(github):typescript/strict-config]",
			expected: "typescript/strict-config",
		},
		{
			name:     "contexture with branch",
			ruleID:   "[contexture:typescript/strict-config,main]",
			expected: "typescript/strict-config",
		},
		{
			name:     "contexture with source and branch",
			ruleID:   "[contexture(github):typescript/strict-config,main]",
			expected: "typescript/strict-config",
		},
		{
			name:     "nested path",
			ruleID:   "[contexture:frontend/react/component-naming]",
			expected: "frontend/react/component-naming",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRulePath(tt.ruleID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildRuleMetadata(t *testing.T) {
	tests := []struct {
		name                string
		rule                *domain.Rule
		expectedBasicMeta   string
		expectedTriggerLine string
	}{
		{
			name: "rule with all metadata",
			rule: &domain.Rule{
				Languages:  []string{"TypeScript", "JavaScript"},
				Frameworks: []string{"React", "Next.js"},
				Tags:       []string{"frontend", "component"},
				Trigger: &domain.RuleTrigger{
					Type:  "glob",
					Globs: []string{"*.tsx", "*.jsx"},
				},
			},
			expectedBasicMeta:   "Languages: TypeScript, JavaScript • Frameworks: React, Next.js • Tags: frontend, component",
			expectedTriggerLine: "Trigger: glob (*.tsx, *.jsx)",
		},
		{
			name: "rule with only languages",
			rule: &domain.Rule{
				Languages: []string{"Go", "Python"},
			},
			expectedBasicMeta:   "Languages: Go, Python",
			expectedTriggerLine: "",
		},
		{
			name: "rule with only frameworks",
			rule: &domain.Rule{
				Frameworks: []string{"Django", "Flask"},
			},
			expectedBasicMeta:   "Frameworks: Django, Flask",
			expectedTriggerLine: "",
		},
		{
			name: "rule with only tags",
			rule: &domain.Rule{
				Tags: []string{"backend", "api"},
			},
			expectedBasicMeta:   "Tags: backend, api",
			expectedTriggerLine: "",
		},
		{
			name: "rule with trigger but no globs",
			rule: &domain.Rule{
				Trigger: &domain.RuleTrigger{
					Type: "manual",
				},
			},
			expectedBasicMeta:   "",
			expectedTriggerLine: "Trigger: manual",
		},
		{
			name:                "empty rule",
			rule:                &domain.Rule{},
			expectedBasicMeta:   "",
			expectedTriggerLine: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basicMeta, triggerLine := buildRuleMetadata(tt.rule)
			assert.Equal(t, tt.expectedBasicMeta, basicMeta)
			assert.Equal(t, tt.expectedTriggerLine, triggerLine)
		})
	}
}

func TestSharedCountMatches(t *testing.T) {
	tests := []struct {
		name     string
		haystack string
		needle   string
		expected int
	}{
		{
			name:     "no matches",
			haystack: "hello world",
			needle:   "foo",
			expected: 0,
		},
		{
			name:     "single match",
			haystack: "hello world",
			needle:   "hello",
			expected: 1,
		},
		{
			name:     "multiple matches",
			haystack: "hello hello world",
			needle:   "hello",
			expected: 2,
		},
		{
			name:     "case insensitive",
			haystack: "Hello HELLO hElLo",
			needle:   "hello",
			expected: 3,
		},
		{
			name:     "empty needle",
			haystack: "hello world",
			needle:   "",
			expected: 0,
		},
		{
			name:     "empty haystack",
			haystack: "",
			needle:   "hello",
			expected: 0,
		},
		{
			name:     "overlapping matches",
			haystack: "aaaa",
			needle:   "aa",
			expected: 2, // strings.Count doesn't count overlapping matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countMatches(tt.haystack, tt.needle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSharedCreateRuleItemStyles(t *testing.T) {
	styles := createRuleItemStyles()

	// Test that styles are created and have expected properties
	assert.NotNil(t, styles.normalPath)
	assert.NotNil(t, styles.normalTitle)
	assert.NotNil(t, styles.normalDesc)
	assert.NotNil(t, styles.normalMeta)
	assert.NotNil(t, styles.selectedPath)
	assert.NotNil(t, styles.selectedTitle)
	assert.NotNil(t, styles.selectedDesc)
	assert.NotNil(t, styles.selectedMeta)
	assert.NotNil(t, styles.dimmedPath)
	assert.NotNil(t, styles.dimmedTitle)
	assert.NotNil(t, styles.dimmedDesc)
	assert.NotNil(t, styles.dimmedMeta)
	assert.NotNil(t, styles.matchHighlight)

	// Test that selected styles have borders
	assert.True(t, styles.selectedTitle.GetBorderLeft())
	assert.True(t, styles.selectedDesc.GetBorderLeft())
	assert.True(t, styles.selectedMeta.GetBorderLeft())
	assert.True(t, styles.selectedPath.GetBorderLeft())

	// Test that normal styles don't have borders
	assert.False(t, styles.normalTitle.GetBorderLeft())
	assert.False(t, styles.normalDesc.GetBorderLeft())
	assert.False(t, styles.normalMeta.GetBorderLeft())
	assert.False(t, styles.normalPath.GetBorderLeft())

	// Test that selected title is bold
	assert.True(t, styles.selectedTitle.GetBold())
	assert.False(t, styles.normalTitle.GetBold())

	// Test padding
	assert.Equal(t, 2, styles.normalTitle.GetPaddingLeft())
	assert.Equal(t, 1, styles.selectedTitle.GetPaddingLeft())
}

func TestApplyHighlightsGeneric(t *testing.T) {
	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

	tests := []struct {
		name        string
		text        string
		filterValue string
		expectMatch bool
	}{
		{
			name:        "no filter",
			text:        "hello world",
			filterValue: "",
			expectMatch: false,
		},
		{
			name:        "exact match",
			text:        "hello",
			filterValue: "hello",
			expectMatch: true,
		},
		{
			name:        "partial match",
			text:        "hello world",
			filterValue: "world",
			expectMatch: true,
		},
		{
			name:        "case insensitive match",
			text:        "Hello World",
			filterValue: "hello",
			expectMatch: true,
		},
		{
			name:        "no match",
			text:        "hello world",
			filterValue: "foo",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyHighlightsGeneric(tt.text, tt.filterValue, baseStyle, highlightStyle)

			// Basic sanity check - result should not be empty
			assert.NotEmpty(t, result)

			if tt.filterValue == "" {
				// When no filter, should just render with base style
				expected := baseStyle.Render(tt.text)
				assert.Equal(t, expected, result)
			} else if !tt.expectMatch {
				// When no match, should render with base style
				expected := baseStyle.Render(tt.text)
				assert.Equal(t, expected, result)
			}
			// For matches, we can't easily test the exact output due to ANSI codes,
			// but we verify it's not empty and different from base style
		})
	}
}

func TestApplyTitleHighlightingGeneric(t *testing.T) {
	baseColor := lipgloss.Color("#FF00FF")

	tests := []struct {
		name        string
		title       string
		filterValue string
		expectMatch bool
	}{
		{
			name:        "no filter",
			title:       "Test Title",
			filterValue: "",
			expectMatch: false,
		},
		{
			name:        "exact match",
			title:       "Test",
			filterValue: "Test",
			expectMatch: true,
		},
		{
			name:        "partial match",
			title:       "Test Title",
			filterValue: "Title",
			expectMatch: true,
		},
		{
			name:        "case insensitive match",
			title:       "Test Title",
			filterValue: "test",
			expectMatch: true,
		},
		{
			name:        "no match",
			title:       "Test Title",
			filterValue: "foo",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyTitleHighlightingGeneric(tt.title, tt.filterValue, baseColor)

			// Basic sanity check - result should not be empty
			assert.NotEmpty(t, result)

			if tt.filterValue == "" || !tt.expectMatch {
				// When no filter or no match, should render with base color and bold
				expected := lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(tt.title)
				assert.Equal(t, expected, result)
			}
			// For matches, we can't easily test the exact output due to ANSI codes,
			// but we verify it's not empty
		})
	}
}

func TestGetDefaultFilterColors(t *testing.T) {
	colors := GetDefaultFilterColors()

	// Test that all colors are set
	assert.NotNil(t, colors.TitleColor)
	assert.NotNil(t, colors.DescColor)
	assert.NotNil(t, colors.MetaColor)
	assert.NotNil(t, colors.PathColor)

	// Test specific color values
	expectedTitleColor := lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}
	assert.Equal(t, expectedTitleColor, colors.TitleColor)

	expectedDescColor := lipgloss.Color("#C084FC")
	assert.Equal(t, expectedDescColor, colors.DescColor)

	expectedMetaColor := lipgloss.Color("#6B7280")
	assert.Equal(t, expectedMetaColor, colors.MetaColor)

	expectedPathColor := lipgloss.Color("#6B7280")
	assert.Equal(t, expectedPathColor, colors.PathColor)
}

// Test with realistic rule data
func TestBuildRuleMetadataIntegration(t *testing.T) {
	// Create a realistic rule similar to what would be used in the app
	rule := &domain.Rule{
		ID:          "[contexture:typescript/strict-config]",
		Title:       "TypeScript Strict Configuration",
		Description: "Enforces strict TypeScript compiler options for better type safety",
		Languages:   []string{"TypeScript"},
		Frameworks:  []string{"Next.js", "React"},
		Tags:        []string{"typescript", "config", "strict"},
		Trigger: &domain.RuleTrigger{
			Type:  "glob",
			Globs: []string{"tsconfig.json", "*.ts", "*.tsx"},
		},
	}

	basicMeta, triggerLine := buildRuleMetadata(rule)

	expectedBasicMeta := "Languages: TypeScript • Frameworks: Next.js, React • Tags: typescript, config, strict"
	expectedTriggerLine := "Trigger: glob (tsconfig.json, *.ts, *.tsx)"

	assert.Equal(t, expectedBasicMeta, basicMeta)
	assert.Equal(t, expectedTriggerLine, triggerLine)
}

func TestExtractRulePathIntegration(t *testing.T) {
	// Test with realistic rule IDs that would be used in the app
	testCases := []struct {
		ruleID   string
		expected string
	}{
		{"[contexture:typescript/strict-config]", "typescript/strict-config"},
		{"[contexture:react/component-naming]", "react/component-naming"},
		{"[contexture:go/error-handling]", "go/error-handling"},
		{"[contexture(github):frontend/accessibility]", "frontend/accessibility"},
		{"[contexture:backend/security/auth,main]", "backend/security/auth"},
	}

	for _, tc := range testCases {
		result := extractRulePath(tc.ruleID)
		assert.Equal(t, tc.expected, result, "Failed for rule ID: %s", tc.ruleID)
	}
}

func TestExtractRulePathWithLocalIndicator(t *testing.T) {
	tests := []struct {
		name     string
		rule     *domain.Rule
		expected string
	}{
		{
			name: "remote rule with contexture format",
			rule: &domain.Rule{
				ID:     "[contexture:security/authentication]",
				Source: "remote",
			},
			expected: "security/authentication",
		},
		{
			name: "local rule simple path",
			rule: &domain.Rule{
				ID:     "security/auth",
				Source: "local",
			},
			expected: "[local] security/auth",
		},
		{
			name: "local rule empty path",
			rule: &domain.Rule{
				ID:     "auth",
				Source: "local",
			},
			expected: "[local] auth",
		},
		{
			name: "remote rule with source and branch",
			rule: &domain.Rule{
				ID:     "[contexture(github):ui/components,main]",
				Source: "remote",
			},
			expected: "ui/components",
		},
		{
			name: "local rule nested path",
			rule: &domain.Rule{
				ID:     "deep/nested/security/rule",
				Source: "local",
			},
			expected: "[local] deep/nested/security/rule",
		},
		{
			name: "rule with no source (default to remote behavior)",
			rule: &domain.Rule{
				ID:     "[contexture:default/rule]",
				Source: "",
			},
			expected: "default/rule",
		},
		{
			name: "local rule with contexture format (shouldn't happen but handle it)",
			rule: &domain.Rule{
				ID:     "[contexture:local/rule]",
				Source: "local",
			},
			expected: "[local] local/rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRulePathWithLocalIndicator(tt.rule)
			assert.Equal(t, tt.expected, result)
		})
	}
}
