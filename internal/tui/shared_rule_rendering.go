package tui

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/ui"
)

// ruleItemStyles defines the styling for rule items across all TUI components
type ruleItemStyles struct {
	normalPath     lipgloss.Style
	normalTitle    lipgloss.Style
	normalDesc     lipgloss.Style
	normalMeta     lipgloss.Style
	selectedPath   lipgloss.Style
	selectedTitle  lipgloss.Style
	selectedDesc   lipgloss.Style
	selectedMeta   lipgloss.Style
	dimmedPath     lipgloss.Style
	dimmedTitle    lipgloss.Style
	dimmedDesc     lipgloss.Style
	dimmedMeta     lipgloss.Style
	matchHighlight lipgloss.Style
}

// createRuleItemStyles creates the standard rule item styles used across all components
func createRuleItemStyles() ruleItemStyles {
	theme := ui.DefaultTheme()

	return ruleItemStyles{
		normalPath: lipgloss.NewStyle().
			Foreground(darkGray).
			Padding(0, 0, 0, 2),
		normalTitle: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
			Padding(0, 0, 0, 2),
		normalDesc: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 0, 0, 2),
		normalMeta: lipgloss.NewStyle().
			Foreground(darkGray).
			Padding(0, 0, 0, 2),
		selectedPath: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(borderColor).
			Foreground(mutedGray).
			Padding(0, 0, 0, 1),
		selectedTitle: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(borderColor).
			Foreground(primaryPink).
			Bold(true).
			Padding(0, 0, 0, 1),
		selectedDesc: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(borderColor).
			Foreground(secondaryPurple).
			Padding(0, 0, 0, 1),
		selectedMeta: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(borderColor).
			Foreground(mutedGray).
			Padding(0, 0, 0, 1),
		dimmedPath: lipgloss.NewStyle().
			Foreground(darkGray).
			Padding(0, 0, 0, 2).
			Faint(true),
		dimmedTitle: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 0, 0, 2).
			Faint(true),
		dimmedDesc: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 0, 0, 2).
			Faint(true),
		dimmedMeta: lipgloss.NewStyle().
			Foreground(darkGray).
			Padding(0, 0, 0, 2).
			Faint(true),
		matchHighlight: lipgloss.NewStyle().
			Foreground(primaryPink),
	}
}

// extractRulePath extracts the rule path from a contexture rule ID
// Handles formats: [contexture:path/rule], [contexture(source):path/rule], [contexture:path/rule,branch]{variables}
func extractRulePath(ruleID string) string {
	if ruleID == "" {
		return ""
	}

	// Remove contexture wrapper: [contexture:...] or [contexture(source):...]
	pathPart := strings.TrimPrefix(ruleID, "[contexture:")
	if strings.HasPrefix(ruleID, "[contexture(") {
		// Handle format: [contexture(source):path/rule]
		parts := strings.SplitN(pathPart, "):", 2)
		if len(parts) == 2 {
			pathPart = parts[1]
		}
	}
	pathPart = strings.TrimSuffix(pathPart, "]")

	// Remove variables part if present (path/rule,branch]{variables} or path/rule]{variables})
	if bracketIdx := strings.Index(pathPart, "]{"); bracketIdx != -1 {
		pathPart = pathPart[:bracketIdx]
	}

	// Remove branch suffix if present (path/rule,branch)
	if commaIdx := strings.Index(pathPart, ","); commaIdx != -1 {
		pathPart = pathPart[:commaIdx]
	}
	return pathPart
}

// extractRuleVariables extracts variables from a contexture rule ID
// Returns the variables as a JSON string, or empty string if no variables
func extractRuleVariables(ruleID string) string {
	if ruleID == "" {
		return ""
	}

	// Look for variables part: {variables}
	if startIdx := strings.Index(ruleID, "]{"); startIdx != -1 {
		// Found variables after the bracket
		variablesPart := ruleID[startIdx+2:] // Skip "]{"
		if endIdx := strings.LastIndex(variablesPart, "}"); endIdx != -1 {
			return "{" + variablesPart[:endIdx+1]
		}
	} else if startIdx := strings.Index(ruleID, "}{"); startIdx != -1 {
		// Handle case where there's no closing bracket before variables
		variablesPart := ruleID[startIdx+2:] // Skip "}{"
		if endIdx := strings.LastIndex(variablesPart, "}"); endIdx != -1 {
			return "{" + variablesPart[:endIdx+1]
		}
	}

	return ""
}

// extractRulePathWithLocalIndicator extracts the rule path and adds a local indicator for local rules
func extractRulePathWithLocalIndicator(rule *domain.Rule) string {
	rulePath := extractRulePath(rule.ID)

	// If the rule source is "local", add a local indicator
	if rule.Source == "local" {
		if rulePath == "" {
			return "[local] " + rule.ID
		}
		return "[local] " + rulePath
	}

	return rulePath
}

// buildRuleMetadata builds the metadata lines for a rule (tags, languages, frameworks, trigger, variables)
func buildRuleMetadata(rule *domain.Rule) (string, string, string) {
	var basicMetadataParts []string
	var basicMetadataLine, triggerLine, variablesLine string

	// Add Languages and Frameworks
	if len(rule.Languages) > 0 {
		basicMetadataParts = append(
			basicMetadataParts,
			"Languages: "+strings.Join(rule.Languages, ", "),
		)
	}
	if len(rule.Frameworks) > 0 {
		basicMetadataParts = append(
			basicMetadataParts,
			"Frameworks: "+strings.Join(rule.Frameworks, ", "),
		)
	}

	// Add Tags
	if len(rule.Tags) > 0 {
		basicMetadataParts = append(basicMetadataParts, "Tags: "+strings.Join(rule.Tags, ", "))
	}

	// Combine basic metadata into single line
	basicMetadataLine = strings.Join(basicMetadataParts, " • ")

	// Build trigger line separately
	if rule.Trigger != nil {
		triggerLine = "Trigger: " + string(rule.Trigger.Type)
		if rule.Trigger.Type == triggerTypeGlob && len(rule.Trigger.Globs) > 0 {
			triggerLine += " (" + strings.Join(rule.Trigger.Globs, ", ") + ")"
		}
	}

	// Build variables line if variables exist
	if len(rule.Variables) > 0 {
		if variablesJSON, err := json.Marshal(rule.Variables); err == nil {
			variablesLine = "Variables: " + string(variablesJSON)
		}
	}

	return basicMetadataLine, triggerLine, variablesLine
}

// applyHighlightsGeneric applies highlighting by finding matches within text
func applyHighlightsGeneric(
	text, filterValue string,
	baseStyle, highlightStyle lipgloss.Style,
) string {
	if filterValue == "" {
		return baseStyle.Render(text)
	}

	// Find all occurrences of the filter in this text (case-insensitive)
	textLower := strings.ToLower(text)
	filterLower := strings.ToLower(filterValue)

	if !strings.Contains(textLower, filterLower) {
		return baseStyle.Render(text)
	}

	// Build highlighted text by splitting and reassembling
	result := strings.Builder{}
	remaining := text
	remainingLower := textLower

	// Get base color from the baseStyle for non-highlighted text
	baseColor := baseStyle.GetForeground()

	for len(remainingLower) > 0 {
		pos := strings.Index(remainingLower, filterLower)
		if pos == -1 {
			// No more matches, add the rest with base color
			if remaining != "" {
				normalTextStyle := lipgloss.NewStyle().Foreground(baseColor)
				result.WriteString(normalTextStyle.Render(remaining))
			}
			break
		}

		// Add text before the match with base color
		if pos > 0 {
			normalTextStyle := lipgloss.NewStyle().Foreground(baseColor)
			result.WriteString(normalTextStyle.Render(remaining[:pos]))
		}

		// Add the matched text with highlight color
		matchedText := remaining[pos : pos+len(filterLower)]
		result.WriteString(highlightStyle.Render(matchedText))

		// Continue with the rest
		remaining = remaining[pos+len(filterLower):]
		remainingLower = remainingLower[pos+len(filterLower):]
	}

	// Apply only the structural styles (border, padding) without color to preserve our color handling
	structuralStyle := lipgloss.NewStyle()

	// Copy border style if it exists
	if baseStyle.GetBorderTop() || baseStyle.GetBorderRight() || baseStyle.GetBorderBottom() ||
		baseStyle.GetBorderLeft() {
		structuralStyle = structuralStyle.
			Border(lipgloss.ThickBorder(),
				baseStyle.GetBorderTop(),
				baseStyle.GetBorderRight(),
				baseStyle.GetBorderBottom(),
				baseStyle.GetBorderLeft())
		// Try to preserve border color - use a default if we can't extract it
		borderColor := lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
		structuralStyle = structuralStyle.BorderForeground(borderColor)
	}

	// Copy padding
	if baseStyle.GetPaddingLeft() > 0 || baseStyle.GetPaddingRight() > 0 ||
		baseStyle.GetPaddingTop() > 0 || baseStyle.GetPaddingBottom() > 0 {
		structuralStyle = structuralStyle.Padding(
			baseStyle.GetPaddingTop(),
			baseStyle.GetPaddingRight(),
			baseStyle.GetPaddingBottom(),
			baseStyle.GetPaddingLeft(),
		)
	}

	// Copy bold styling
	if baseStyle.GetBold() {
		structuralStyle = structuralStyle.Bold(true)
	}

	return structuralStyle.Render(result.String())
}

// applyTitleHighlightingGeneric applies pink base with white highlights for selected titles
func applyTitleHighlightingGeneric(
	title, filterValue string,
	baseColor lipgloss.TerminalColor,
) string {
	if filterValue == "" {
		return lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(title)
	}

	// Find all occurrences of the filter in the title (case-insensitive)
	titleLower := strings.ToLower(title)
	filterLower := strings.ToLower(filterValue)

	if !strings.Contains(titleLower, filterLower) {
		return lipgloss.NewStyle().Foreground(baseColor).Bold(true).Render(title)
	}

	// Colors for highlighting
	pinkStyle := lipgloss.NewStyle().Foreground(baseColor).Bold(true)
	whiteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true)

	// Build the styled title
	result := ""
	remaining := title
	remainingLower := titleLower

	for len(remainingLower) > 0 {
		pos := strings.Index(remainingLower, filterLower)
		if pos == -1 {
			// No more matches, add the rest with pink
			if remaining != "" {
				result += pinkStyle.Render(remaining)
			}
			break
		}

		// Add text before the match with pink
		if pos > 0 {
			result += pinkStyle.Render(remaining[:pos])
		}

		// Add the matched text with white
		matchedText := remaining[pos : pos+len(filterLower)]
		result += whiteStyle.Render(matchedText)

		// Continue with the rest
		remaining = remaining[pos+len(filterLower):]
		remainingLower = remainingLower[pos+len(filterLower):]
	}

	return result
}

// FilterColors contains the color scheme for filtered text rendering
type FilterColors struct {
	TitleColor lipgloss.AdaptiveColor
	DescColor  lipgloss.Color
	MetaColor  lipgloss.Color
	PathColor  lipgloss.Color
}

// GetDefaultFilterColors returns the default filter color scheme
func GetDefaultFilterColors() FilterColors {
	return FilterColors{
		TitleColor: lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}, // White/dark for titles
		DescColor:  lipgloss.Color("#C084FC"),                                 // Purple for descriptions
		MetaColor:  lipgloss.Color("#6B7280"),                                 // Dark grey for metadata
		PathColor:  lipgloss.Color("#6B7280"),                                 // Same as metadata
	}
}

// countMatches counts how many times needle appears in haystack
func countMatches(haystack, needle string) int {
	if needle == "" {
		return 0
	}
	return strings.Count(strings.ToLower(haystack), strings.ToLower(needle))
}
