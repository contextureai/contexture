package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewContentBuilder_BuildPreviewContent(t *testing.T) {
	// Create test styles
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	metadataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	builder := NewPreviewContentBuilder(titleStyle, labelStyle, metadataStyle, errorStyle, emptyStyle)

	t.Run("nil rule", func(t *testing.T) {
		content := builder.BuildPreviewContent(nil)
		assert.Contains(t, content, "No rule selected")
		assert.Contains(t, content, "Please select a rule to view its details.")
	})

	t.Run("basic rule with title only", func(t *testing.T) {
		rule := &domain.Rule{
			Title: "Test Rule",
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Test Rule")
	})

	t.Run("rule with description", func(t *testing.T) {
		rule := &domain.Rule{
			Title:       "Test Rule",
			Description: "This is a test description",
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Test Rule")
		assert.Contains(t, content, "Description")
		assert.Contains(t, content, "This is a test description")
	})

	t.Run("rule with tags", func(t *testing.T) {
		rule := &domain.Rule{
			Title: "Test Rule",
			Tags:  []string{"security", "authentication"},
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Tags: ")
		assert.Contains(t, content, "security, authentication")
	})

	t.Run("rule with languages", func(t *testing.T) {
		rule := &domain.Rule{
			Title:     "Test Rule",
			Languages: []string{"TypeScript", "JavaScript"},
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Languages: ")
		assert.Contains(t, content, "TypeScript, JavaScript")
	})

	t.Run("rule with frameworks", func(t *testing.T) {
		rule := &domain.Rule{
			Title:      "Test Rule",
			Frameworks: []string{"React", "Next.js"},
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Frameworks: ")
		assert.Contains(t, content, "React, Next.js")
	})

	t.Run("rule with trigger", func(t *testing.T) {
		rule := &domain.Rule{
			Title: "Test Rule",
			Trigger: &domain.RuleTrigger{
				Type:  triggerTypeGlob,
				Globs: []string{"*.ts", "*.tsx"},
			},
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Trigger: ")
		assert.Contains(t, content, "glob")
		assert.Contains(t, content, "*.ts, *.tsx")
	})

	t.Run("rule with trigger but no globs", func(t *testing.T) {
		rule := &domain.Rule{
			Title: "Test Rule",
			Trigger: &domain.RuleTrigger{
				Type: "manual",
			},
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Trigger: ")
		assert.Contains(t, content, "manual")
		assert.NotContains(t, content, "()")
	})

	t.Run("complete rule with all metadata", func(t *testing.T) {
		rule := &domain.Rule{
			Title:       "Complete Test Rule",
			Description: "A comprehensive test rule",
			Tags:        []string{"security", "validation"},
			Languages:   []string{"Go", "TypeScript"},
			Frameworks:  []string{"gin", "fastify"},
			Trigger: &domain.RuleTrigger{
				Type:  triggerTypeGlob,
				Globs: []string{"*.go", "*.ts"},
			},
		}

		content := builder.BuildPreviewContent(rule)

		// Check all components are present
		assert.Contains(t, content, "Complete Test Rule")
		assert.Contains(t, content, "A comprehensive test rule")
		assert.Contains(t, content, "Tags: security, validation")
		assert.Contains(t, content, "Languages: Go, TypeScript")
		assert.Contains(t, content, "Frameworks: gin, fastify")
		assert.Contains(t, content, "Trigger: glob (*.go, *.ts)")
	})
}

func TestPreviewContentBuilder_BuildPreviewContentWithMarkdown(t *testing.T) {
	// Create test styles
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	metadataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	builder := NewPreviewContentBuilder(titleStyle, labelStyle, metadataStyle, errorStyle, emptyStyle)

	t.Run("nil rule", func(t *testing.T) {
		mockRenderer := func(_ string) (string, error) {
			return "rendered", nil
		}

		content := builder.BuildPreviewContentWithMarkdown(nil, mockRenderer)
		assert.Contains(t, content, "No rule selected")
	})

	t.Run("rule without content", func(t *testing.T) {
		rule := &domain.Rule{
			Title: "Test Rule",
		}

		mockRenderer := func(_ string) (string, error) {
			return "rendered", nil
		}

		content := builder.BuildPreviewContentWithMarkdown(rule, mockRenderer)
		assert.Contains(t, content, "Test Rule")
		assert.NotContains(t, content, "Content")
	})

	t.Run("successful markdown rendering", func(t *testing.T) {
		rule := &domain.Rule{
			Title:   "Test Rule",
			Content: "# Markdown Content",
		}

		mockRenderer := func(content string) (string, error) {
			if content == "# Markdown Content" {
				return "<h1>Markdown Content</h1>", nil
			}
			return "", errors.New("unexpected content")
		}

		content := builder.BuildPreviewContentWithMarkdown(rule, mockRenderer)
		assert.Contains(t, content, "Test Rule")
		assert.Contains(t, content, "Content")
		assert.Contains(t, content, "<h1>Markdown Content</h1>")
	})

	t.Run("markdown rendering error with MarkdownRenderError", func(t *testing.T) {
		rule := &domain.Rule{
			Title:   "Test Rule",
			Content: "# Bad Markdown",
		}

		mockRenderer := func(_ string) (string, error) {
			return "", &MarkdownRenderError{
				Type:    "render_failed",
				Message: "parsing failed",
				Cause:   errors.New("parsing failed"),
			}
		}

		content := builder.BuildPreviewContentWithMarkdown(rule, mockRenderer)
		assert.Contains(t, content, "Test Rule")
		assert.Contains(t, content, "⚠")
		assert.Contains(t, content, "Content (Plain Text)")
		assert.Contains(t, content, "# Bad Markdown")
	})

	t.Run("markdown rendering error with generic error", func(t *testing.T) {
		rule := &domain.Rule{
			Title:   "Test Rule",
			Content: "# Bad Markdown",
		}

		mockRenderer := func(_ string) (string, error) {
			return "", errors.New("generic error")
		}

		content := builder.BuildPreviewContentWithMarkdown(rule, mockRenderer)
		assert.Contains(t, content, "Test Rule")
		assert.Contains(t, content, "⚠")
		assert.Contains(t, content, "Preview error: Unable to format the rule content")
		assert.Contains(t, content, "Content (Plain Text)")
		assert.Contains(t, content, "# Bad Markdown")
	})

	t.Run("rule with empty content", func(t *testing.T) {
		rule := &domain.Rule{
			Title:   "Test Rule",
			Content: "",
		}

		mockRenderer := func(_ string) (string, error) {
			return "should not be called", nil
		}

		content := builder.BuildPreviewContentWithMarkdown(rule, mockRenderer)
		assert.Contains(t, content, "Test Rule")
		assert.Contains(t, content, "No content available for preview.")
		assert.NotContains(t, content, "should not be called")
	})
}

func TestNewPreviewContentBuilder(t *testing.T) {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	metadataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	builder := NewPreviewContentBuilder(titleStyle, labelStyle, metadataStyle, errorStyle, emptyStyle)

	require.NotNil(t, builder)
	assert.Equal(t, titleStyle, builder.titleStyle)
	assert.Equal(t, labelStyle, builder.labelStyle)
	assert.Equal(t, metadataStyle, builder.metadataStyle)
	assert.Equal(t, errorStyle, builder.errorStyle)
	assert.Equal(t, emptyStyle, builder.emptyStyle)
}

func TestCalculatePreviewDimensions(t *testing.T) {
	tests := []struct {
		name           string
		terminalWidth  int
		terminalHeight int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "normal terminal size",
			terminalWidth:  120,
			terminalHeight: 40,
			expectedWidth:  112, // 120 - 8
			expectedHeight: 32,  // 40 - 8
		},
		{
			name:           "small terminal width",
			terminalWidth:  40,
			terminalHeight: 30,
			expectedWidth:  40, // minimum enforced
			expectedHeight: 22, // 30 - 8
		},
		{
			name:           "small terminal height",
			terminalWidth:  80,
			terminalHeight: 15,
			expectedWidth:  72, // 80 - 8
			expectedHeight: 10, // minimum enforced
		},
		{
			name:           "very small terminal",
			terminalWidth:  30,
			terminalHeight: 10,
			expectedWidth:  40, // minimum enforced
			expectedHeight: 10, // minimum enforced
		},
		{
			name:           "large terminal",
			terminalWidth:  200,
			terminalHeight: 60,
			expectedWidth:  192, // 200 - 8
			expectedHeight: 52,  // 60 - 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := CalculatePreviewDimensions(tt.terminalWidth, tt.terminalHeight)
			assert.Equal(t, tt.expectedWidth, width, "width should match expected")
			assert.Equal(t, tt.expectedHeight, height, "height should match expected")
		})
	}
}

func TestBuildPreviewContent_EdgeCases(t *testing.T) {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	metadataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	builder := NewPreviewContentBuilder(titleStyle, labelStyle, metadataStyle, errorStyle, emptyStyle)

	t.Run("empty strings in slices", func(t *testing.T) {
		rule := &domain.Rule{
			Title:      "Test Rule",
			Tags:       []string{"", "valid", ""},
			Languages:  []string{"Go", ""},
			Frameworks: []string{""},
		}

		content := builder.BuildPreviewContent(rule)
		// Should still display the fields, even with empty strings
		assert.Contains(t, content, "Tags: ")
		assert.Contains(t, content, "Languages: ")
		assert.Contains(t, content, "Frameworks: ")
	})

	t.Run("whitespace in descriptions", func(t *testing.T) {
		rule := &domain.Rule{
			Title:       "Test Rule",
			Description: "   Trimmed description   ",
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "   Trimmed description   ")
	})

	t.Run("newlines in description", func(t *testing.T) {
		rule := &domain.Rule{
			Title:       "Test Rule",
			Description: "Line 1\nLine 2\nLine 3",
		}

		content := builder.BuildPreviewContent(rule)
		assert.Contains(t, content, "Line 1\nLine 2\nLine 3")
		lines := strings.Split(content, "\n")
		assert.Greater(t, len(lines), 3, "should preserve newlines in description")
	})
}
