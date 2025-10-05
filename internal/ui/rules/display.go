// Package rules provides UI components for displaying rules
package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/ui"
)

// DisplayOptions configures rule display behavior
type DisplayOptions struct {
	ShowSource    bool
	ShowTriggers  bool
	ShowVariables bool
	ShowTags      bool
	Pattern       string // Regex pattern for filtering rules
}

// DefaultDisplayOptions returns sensible defaults for rule display
func DefaultDisplayOptions() DisplayOptions {
	return DisplayOptions{
		ShowSource:    true,
		ShowTriggers:  false,
		ShowVariables: false,
		ShowTags:      false,
	}
}

// DisplayStyles contains styling configuration for rule display using colors from existing TUI components
type DisplayStyles struct {
	header     lipgloss.Style
	rulePath   lipgloss.Style
	ruleTitle  lipgloss.Style
	ruleSource lipgloss.Style
	metadata   lipgloss.Style
	muted      lipgloss.Style
}

// createDisplayStyles creates styles consistent with existing TUI components
func createDisplayStyles() DisplayStyles {
	theme := ui.DefaultTheme()

	// Colors from TUI components
	primaryPink := lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	secondaryPurple := lipgloss.AdaptiveColor{Light: "#C084FC", Dark: "#9333EA"}
	darkGray := lipgloss.AdaptiveColor{Light: "#A0A0A0", Dark: "#585858"}

	return DisplayStyles{
		header: lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryPink),
		rulePath: lipgloss.NewStyle().
			Foreground(secondaryPurple).
			Bold(true),
		ruleTitle: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
			MarginLeft(2),
		ruleSource: lipgloss.NewStyle().
			Foreground(darkGray).
			MarginLeft(2),
		metadata: lipgloss.NewStyle().
			Foreground(theme.Muted).
			MarginLeft(2),
		muted: lipgloss.NewStyle().
			Foreground(theme.Muted),
	}
}

// filterRulesByPattern filters rules based on a regex pattern matching across multiple fields
func filterRulesByPattern(rules []*domain.Rule, pattern string) ([]*domain.Rule, error) {
	if pattern == "" {
		return rules, nil
	}

	// Compile the regex pattern
	regex, err := regexp.Compile("(?i)" + pattern) // Case-insensitive
	if err != nil {
		return nil, err
	}

	var filtered []*domain.Rule
	for _, rule := range rules {
		if ruleMatchesPattern(rule, regex) {
			filtered = append(filtered, rule)
		}
	}

	return filtered, nil
}

// ruleMatchesPattern checks if a rule matches the given regex pattern
func ruleMatchesPattern(rule *domain.Rule, regex *regexp.Regexp) bool {
	// Check ID
	if regex.MatchString(rule.ID) {
		return true
	}

	// Check Title
	if regex.MatchString(rule.Title) {
		return true
	}

	// Check Description
	if regex.MatchString(rule.Description) {
		return true
	}

	// Check Tags
	for _, tag := range rule.Tags {
		if regex.MatchString(tag) {
			return true
		}
	}

	// Check Frameworks
	for _, framework := range rule.Frameworks {
		if regex.MatchString(framework) {
			return true
		}
	}

	// Check Languages
	for _, language := range rule.Languages {
		if regex.MatchString(language) {
			return true
		}
	}

	// Check Source
	if regex.MatchString(rule.Source) {
		return true
	}

	// Check FilePath
	if regex.MatchString(rule.FilePath) {
		return true
	}

	return false
}

// DisplayRuleList displays a list of rules in a compact format
func DisplayRuleList(rules []*domain.Rule, options DisplayOptions) error {
	if len(rules) == 0 {
		fmt.Println("No rules found.")
		return nil
	}

	// Apply pattern filtering if provided
	filteredRules, err := filterRulesByPattern(rules, options.Pattern)
	if err != nil {
		return contextureerrors.ValidationError("pattern", err)
	}

	if len(filteredRules) == 0 {
		if options.Pattern != "" {
			fmt.Printf("No rules found matching pattern: %s\n", options.Pattern)
		} else {
			fmt.Println("No rules found.")
		}
		return nil
	}

	styles := createDisplayStyles()

	// Header with pattern info if applicable
	headerText := "Installed Rules"
	if options.Pattern != "" {
		headerText = fmt.Sprintf("Installed Rules (pattern: %s)", options.Pattern)
	}
	fmt.Printf("%s\n\n", styles.header.Render(headerText))

	// Sort rules by path for consistent output
	sortedRules := make([]*domain.Rule, len(filteredRules))
	copy(sortedRules, filteredRules)
	sort.Slice(sortedRules, func(i, j int) bool {
		pathI := extractRulePath(sortedRules[i].ID)
		pathJ := extractRulePath(sortedRules[j].ID)
		return pathI < pathJ
	})

	// Display each rule in compact format
	for i, rule := range sortedRules {
		if i > 0 {
			fmt.Println() // Empty line between rules
		}

		// 1. Rule path (no emojis) - extract just the path for custom rules
		rulePath := extractSimpleRulePath(rule.ID)
		if rulePath == "" {
			rulePath = rule.ID
		}
		fmt.Println(styles.rulePath.Render(rulePath))

		// 2. Title on next line with indentation
		fmt.Println(styles.ruleTitle.Render(rule.Title))

		// 3. Source on third line (only if non-default)
		if options.ShowSource {
			source := formatSourceForDisplay(rule)
			if source != "" {
				fmt.Println(styles.ruleSource.Render(source))
			}
		}

		// Optional: Additional metadata on following lines
		var metadataLines []string

		if options.ShowTags && len(rule.Tags) > 0 {
			metadataLines = append(metadataLines,
				fmt.Sprintf("Tags: %s", strings.Join(rule.Tags, ", ")))
		}

		if options.ShowTriggers && rule.Trigger != nil {
			trigger := formatTrigger(rule.Trigger)
			if trigger != "" {
				metadataLines = append(metadataLines, fmt.Sprintf("Trigger: %s", trigger))
			}
		}

		if options.ShowVariables && shouldDisplayVariables(rule) {
			variables := formatVariables(rule.Variables)
			if variables != "" {
				metadataLines = append(metadataLines, fmt.Sprintf("Variables: %s", variables))
			}
		}

		// Display metadata lines
		for _, line := range metadataLines {
			fmt.Println(styles.metadata.Render(line))
		}
	}

	return nil
}

// extractRulePath extracts the display path from a rule ID
func extractRulePath(ruleID string) string {
	return domain.ExtractRuleDisplayPath(ruleID)
}

// extractSimpleRulePath extracts just the rule path without source prefix for custom rules
func extractSimpleRulePath(ruleID string) string {
	return domain.ExtractRulePath(ruleID)
}

// formatSourceForDisplay formats the source information for display
func formatSourceForDisplay(rule *domain.Rule) string {
	// Don't show source for default repository
	if rule.Source == "" || rule.Source == domain.DefaultRepository {
		return ""
	}

	// Don't show source for provider syntax rules (@provider/path)
	// These rules have a user-friendly ID already
	if strings.HasPrefix(rule.ID, "@") {
		return ""
	}

	// For local rules
	if rule.Source == "local" {
		return "local"
	}

	// For custom Git sources
	if domain.IsCustomGitSource(rule.Source) {
		return domain.FormatSourceForDisplay(rule.Source, rule.Ref)
	}

	// For other custom sources
	if rule.Ref != "" && rule.Ref != domain.DefaultBranch {
		return fmt.Sprintf("%s (%s)", rule.Source, rule.Ref)
	}

	return rule.Source
}

// formatTrigger formats trigger information for display
func formatTrigger(trigger *domain.RuleTrigger) string {
	if trigger == nil {
		return ""
	}

	switch trigger.Type {
	case domain.TriggerGlob:
		if len(trigger.Globs) > 0 {
			return fmt.Sprintf("%s (%s)", trigger.Type, strings.Join(trigger.Globs, ", "))
		}
		return string(trigger.Type)
	case domain.TriggerAlways, domain.TriggerManual, domain.TriggerModel:
		return string(trigger.Type)
	default:
		return string(trigger.Type)
	}
}

// shouldDisplayVariables checks if variables should be displayed
func shouldDisplayVariables(rule *domain.Rule) bool {
	if len(rule.Variables) == 0 {
		return false
	}

	// If there are default variables, only show if they differ
	if rule.DefaultVariables != nil {
		return !mapsEqual(rule.Variables, rule.DefaultVariables)
	}

	return true
}

// formatVariables formats variables for display
func formatVariables(variables map[string]any) string {
	if len(variables) == 0 {
		return ""
	}

	// Simple key=value format for basic types, JSON for complex types
	var parts []string
	for key, value := range variables {
		switch v := value.(type) {
		case string:
			parts = append(parts, fmt.Sprintf("%s=%s", key, v))
		case bool:
			parts = append(parts, fmt.Sprintf("%s=%t", key, v))
		case int, int32, int64:
			parts = append(parts, fmt.Sprintf("%s=%d", key, v))
		case float32, float64:
			parts = append(parts, fmt.Sprintf("%s=%f", key, v))
		default:
			parts = append(parts, fmt.Sprintf("%s=%v", key, v))
		}
	}

	return strings.Join(parts, ", ")
}

// mapsEqual compares two maps for equality
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists || valueA != valueB {
			return false
		}
	}

	return true
}
