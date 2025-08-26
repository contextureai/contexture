package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/ui"
)

// Shared color constants used across TUI components
var (
	darkGray        = lipgloss.AdaptiveColor{Light: "#A0A0A0", Dark: "#585858"}
	mutedGray       = lipgloss.AdaptiveColor{Light: "#909090", Dark: "#606060"}
	borderColor     = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	primaryPink     = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	secondaryPurple = lipgloss.AdaptiveColor{Light: "#C084FC", Dark: "#9333EA"}
	triggerTypeGlob = domain.TriggerType("glob")
)

// StyleConfig contains the styling configuration for rendering rule lines
type StyleConfig struct {
	Theme             *ui.Theme
	SelectedColor     lipgloss.Color
	SelectedDescColor lipgloss.Color
	SelectedMetaColor lipgloss.Color
	BorderColor       lipgloss.AdaptiveColor
	DimmedPath        lipgloss.Style
	DimmedDesc        lipgloss.Style
	DimmedMeta        lipgloss.Style
}

// renderStyledLinesForEmptyFilter renders styled lines when the filter is empty
// This is shared between file_browser.go and rule_selector.go to avoid duplication
func renderStyledLinesForEmptyFilter(
	lines []string,
	rulePath string,
	checkboxStyled string,
	isSelected bool,
	config StyleConfig,
) []string {
	var styledLines []string

	// Determine line indices based on whether path exists
	pathIndex := -1
	titleIndex := 0
	descIndex := 1

	if rulePath != "" {
		pathIndex = 0
		titleIndex = 1
		descIndex = 2
	}

	for i, line := range lines {
		switch i {
		case pathIndex:
			// Rule path line
			var pathStyle lipgloss.Style
			if isSelected {
				pathStyle = lipgloss.NewStyle().
					Border(lipgloss.ThickBorder(), false, false, false, true).
					BorderForeground(config.BorderColor).
					Foreground(config.SelectedMetaColor).
					Padding(0, 0, 0, 1).
					Faint(true)
			} else {
				pathStyle = config.DimmedPath
			}
			styledLines = append(styledLines, pathStyle.Render(line))
		case titleIndex:
			// Title with checkbox - use appropriate style based on selection for padding
			var titleStyle lipgloss.Style
			if isSelected {
				// Selected items need border padding even when dimmed - use darker muted style
				titleStyle = lipgloss.NewStyle().
					Border(lipgloss.ThickBorder(), false, false, false, true).
					BorderForeground(config.BorderColor).
					Foreground(lipgloss.Color("#6B7280")).
					// Use darker grey instead of theme.Muted
					Bold(true).
					Padding(0, 0, 0, 1).
					Faint(true)
				checkboxAndTitle := checkboxStyled + " " + line
				titleStyled := titleStyle.Render(checkboxAndTitle)
				styledLines = append(styledLines, titleStyled)
			} else {
				// For non-selected items, ensure checkbox is aligned with padding
				titleStyle = lipgloss.NewStyle().
					Foreground(config.Theme.Muted).
					Padding(0, 0, 0, 2). // Same padding as regular items
					Faint(true)
				checkboxAndTitle := checkboxStyled + " " + line
				styledLines = append(styledLines, titleStyle.Render(checkboxAndTitle))
			}
		case descIndex:
			// Description - use appropriate style based on selection
			var descStyle lipgloss.Style
			if isSelected {
				descStyle = lipgloss.NewStyle().
					Border(lipgloss.ThickBorder(), false, false, false, true).
					BorderForeground(config.BorderColor).
					Foreground(config.SelectedDescColor). // Medium pink for description when selected
					Padding(0, 0, 0, 1).
					Faint(true)
			} else {
				descStyle = config.DimmedDesc
			}
			styledLines = append(styledLines, descStyle.Render(line))
		default:
			// Metadata - use appropriate style based on selection
			var metadataStyle lipgloss.Style
			if isSelected {
				metadataStyle = lipgloss.NewStyle().
					Border(lipgloss.ThickBorder(), false, false, false, true).
					BorderForeground(config.BorderColor).
					Foreground(config.SelectedMetaColor). // Darker pink for metadata when selected
					Padding(0, 0, 0, 1).
					Faint(true)
			} else {
				metadataStyle = config.DimmedMeta // Use dimmed metadata style
			}
			styledLines = append(styledLines, metadataStyle.Render(line))
		}
	}

	return styledLines
}
