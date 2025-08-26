// Package tui provides formatting utilities for TUI components
package tui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
)

// PreviewContentBuilder builds preview content for rules
type PreviewContentBuilder struct {
	titleStyle    lipgloss.Style
	labelStyle    lipgloss.Style
	metadataStyle lipgloss.Style
	errorStyle    lipgloss.Style
	emptyStyle    lipgloss.Style
}

// NewPreviewContentBuilder creates a new preview content builder with styles
func NewPreviewContentBuilder(
	titleStyle lipgloss.Style,
	labelStyle lipgloss.Style,
	metadataStyle lipgloss.Style,
	errorStyle lipgloss.Style,
	emptyStyle lipgloss.Style,
) *PreviewContentBuilder {
	return &PreviewContentBuilder{
		titleStyle:    titleStyle,
		labelStyle:    labelStyle,
		metadataStyle: metadataStyle,
		errorStyle:    errorStyle,
		emptyStyle:    emptyStyle,
	}
}

// BuildPreviewContent creates the content for the preview (pure function)
func (b *PreviewContentBuilder) BuildPreviewContent(rule *domain.Rule) string {
	var content strings.Builder

	// Handle nil rule
	if rule == nil {
		content.WriteString(b.titleStyle.Render("No rule selected"))
		content.WriteString("\n\n")
		content.WriteString(b.metadataStyle.Render("Please select a rule to view its details."))
		return content.String()
	}

	// Title
	content.WriteString(b.titleStyle.Render(rule.Title))
	content.WriteString("\n\n")

	// Description
	if rule.Description != "" {
		content.WriteString(b.labelStyle.Render("Description"))
		content.WriteString("\n")
		content.WriteString(rule.Description)
		content.WriteString("\n\n")
	}

	// Metadata
	metadataAdded := false

	if len(rule.Tags) > 0 {
		content.WriteString(b.metadataStyle.Render("Tags: "))
		content.WriteString(strings.Join(rule.Tags, ", "))
		content.WriteString("\n")
		metadataAdded = true
	}

	if len(rule.Languages) > 0 {
		content.WriteString(b.metadataStyle.Render("Languages: "))
		content.WriteString(strings.Join(rule.Languages, ", "))
		content.WriteString("\n")
		metadataAdded = true
	}

	if len(rule.Frameworks) > 0 {
		content.WriteString(b.metadataStyle.Render("Frameworks: "))
		content.WriteString(strings.Join(rule.Frameworks, ", "))
		content.WriteString("\n")
		metadataAdded = true
	}

	if rule.Trigger != nil {
		content.WriteString(b.metadataStyle.Render("Trigger: "))
		content.WriteString(string(rule.Trigger.Type))
		if rule.Trigger.Type == triggerTypeGlob && len(rule.Trigger.Globs) > 0 {
			content.WriteString(" (")
			content.WriteString(strings.Join(rule.Trigger.Globs, ", "))
			content.WriteString(")")
		}
		content.WriteString("\n")
		metadataAdded = true
	}

	if metadataAdded {
		content.WriteString("\n")
	}

	return content.String()
}

// BuildPreviewContentWithMarkdown builds preview content with markdown rendering
func (b *PreviewContentBuilder) BuildPreviewContentWithMarkdown(rule *domain.Rule, markdownRenderer func(string) (string, error)) string {
	baseContent := b.BuildPreviewContent(rule)

	if rule == nil {
		return baseContent
	}

	var content strings.Builder
	content.WriteString(baseContent)

	// Rule Content with markdown rendering
	if rule.Content != "" {
		renderedContent, err := markdownRenderer(rule.Content)
		if err != nil {
			// Use improved error handling with user-friendly messages
			var userMessage string
			var markdownErr *MarkdownRenderError
			if errors.As(err, &markdownErr) {
				userMessage = markdownErr.GetUserFriendlyMessage()
			} else {
				userMessage = "Preview error: Unable to format the rule content. Showing content as plain text."
			}

			content.WriteString(b.errorStyle.Render("⚠ " + userMessage))
			content.WriteString("\n\n")
			content.WriteString(b.labelStyle.Render("Content (Plain Text)"))
			content.WriteString("\n")
			content.WriteString(rule.Content)
		} else {
			content.WriteString(b.labelStyle.Render("Content"))
			content.WriteString("\n")
			content.WriteString(renderedContent)
		}
	} else {
		content.WriteString(b.emptyStyle.Render("No content available for preview."))
	}

	return content.String()
}

// CalculatePreviewDimensions calculates appropriate preview dimensions based on terminal size
func CalculatePreviewDimensions(terminalWidth, terminalHeight int) (int, int) {
	// Make preview take up most of the screen with some margin
	previewWidth := terminalWidth - 8   // 4 margin on each side
	previewHeight := terminalHeight - 8 // 4 margin top/bottom

	if previewWidth < 40 {
		previewWidth = 40
	}
	if previewHeight < 10 {
		previewHeight = 10
	}

	return previewWidth, previewHeight
}
