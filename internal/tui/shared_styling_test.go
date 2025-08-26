package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/stretchr/testify/assert"
)

// Test constants to avoid linter warnings
const (
	testRulePath = "typescript/strict-config"
	testTitle    = "TypeScript Strict Configuration"
	testDesc     = "Enforces strict TypeScript compiler options"
	testMeta     = "Languages: TypeScript"
	uncheckedBox = "☐"
	checkedBox   = "☑"
)

func TestSharedColorConstants(t *testing.T) {
	// Test that shared color constants are properly defined
	assert.NotNil(t, darkGray)
	assert.NotNil(t, mutedGray)
	assert.NotNil(t, borderColor)
	assert.NotNil(t, primaryPink)
	assert.NotNil(t, secondaryPurple)
	assert.NotNil(t, triggerTypeGlob)

	// Test specific color values
	expectedDarkGray := lipgloss.AdaptiveColor{Light: "#A0A0A0", Dark: "#585858"}
	assert.Equal(t, expectedDarkGray, darkGray)

	expectedMutedGray := lipgloss.AdaptiveColor{Light: "#909090", Dark: "#606060"}
	assert.Equal(t, expectedMutedGray, mutedGray)

	expectedBorderColor := lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	assert.Equal(t, expectedBorderColor, borderColor)

	expectedPrimaryPink := lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	assert.Equal(t, expectedPrimaryPink, primaryPink)

	expectedSecondaryPurple := lipgloss.AdaptiveColor{Light: "#C084FC", Dark: "#9333EA"}
	assert.Equal(t, expectedSecondaryPurple, secondaryPurple)

	assert.Equal(t, "glob", string(triggerTypeGlob))
}

func TestStyleConfig(t *testing.T) {
	// Test StyleConfig struct can be created
	theme := ui.DefaultTheme()
	config := StyleConfig{
		Theme:             &theme,
		SelectedColor:     lipgloss.Color("#FF0000"),
		SelectedDescColor: lipgloss.Color("#00FF00"),
		SelectedMetaColor: lipgloss.Color("#0000FF"),
		BorderColor:       borderColor,
		DimmedPath:        lipgloss.NewStyle().Faint(true),
		DimmedDesc:        lipgloss.NewStyle().Faint(true),
		DimmedMeta:        lipgloss.NewStyle().Faint(true),
	}

	assert.NotNil(t, config.Theme)
	assert.Equal(t, lipgloss.Color("#FF0000"), config.SelectedColor)
	assert.Equal(t, lipgloss.Color("#00FF00"), config.SelectedDescColor)
	assert.Equal(t, lipgloss.Color("#0000FF"), config.SelectedMetaColor)
	assert.Equal(t, borderColor, config.BorderColor)
	assert.NotNil(t, config.DimmedPath)
	assert.NotNil(t, config.DimmedDesc)
	assert.NotNil(t, config.DimmedMeta)
}

func getStyleConfig() StyleConfig {
	theme := ui.DefaultTheme()
	config := StyleConfig{
		Theme:             &theme,
		SelectedColor:     lipgloss.Color("#FF0000"),
		SelectedDescColor: lipgloss.Color("#00FF00"),
		SelectedMetaColor: lipgloss.Color("#0000FF"),
		BorderColor:       borderColor,
		DimmedPath: lipgloss.NewStyle().
			Foreground(darkGray).
			Padding(0, 0, 0, 2).
			Faint(true),
		DimmedDesc: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 0, 0, 2).
			Faint(true),
		DimmedMeta: lipgloss.NewStyle().
			Foreground(darkGray).
			Padding(0, 0, 0, 2).
			Faint(true),
	}

	return config
}

func TestRenderStyledLinesForEmptyFilter_WithPath(t *testing.T) {
	config := getStyleConfig()

	lines := []string{
		testRulePath,
		testTitle,
		testDesc,
		testMeta,
	}

	rulePath := testRulePath
	checkboxStyled := uncheckedBox
	isSelected := false

	result := renderStyledLinesForEmptyFilter(lines, rulePath, checkboxStyled, isSelected, config)

	// Should have 4 styled lines (path, title, desc, metadata)
	assert.Len(t, result, 4)

	// All results should be non-empty strings
	for i, line := range result {
		assert.NotEmpty(t, line, "Line %d should not be empty", i)
	}

	// Title line should contain checkbox
	assert.Contains(t, result[1], checkboxStyled)
}

func TestRenderStyledLinesForEmptyFilter_WithoutPath(t *testing.T) {
	config := getStyleConfig()

	lines := []string{
		"TypeScript Strict Configuration",
		"Enforces strict TypeScript compiler options",
		"Languages: TypeScript",
	}

	rulePath := "" // No path
	checkboxStyled := uncheckedBox
	isSelected := false

	result := renderStyledLinesForEmptyFilter(lines, rulePath, checkboxStyled, isSelected, config)

	// Should have 3 styled lines (title, desc, metadata)
	assert.Len(t, result, 3)

	// All results should be non-empty strings
	for i, line := range result {
		assert.NotEmpty(t, line, "Line %d should not be empty", i)
	}

	// Title line (first) should contain checkbox
	assert.Contains(t, result[0], checkboxStyled)
}

func TestRenderStyledLinesForEmptyFilter_Selected(t *testing.T) {
	config := getStyleConfig()

	lines := []string{
		testRulePath,
		testTitle,
		testDesc,
		testMeta,
	}

	rulePath := testRulePath
	checkboxStyled := checkedBox
	isSelected := true // Selected item

	result := renderStyledLinesForEmptyFilter(lines, rulePath, checkboxStyled, isSelected, config)

	// Should have 4 styled lines
	assert.Len(t, result, 4)

	// All results should be non-empty strings
	for i, line := range result {
		assert.NotEmpty(t, line, "Line %d should not be empty", i)
	}

	// Title line should contain checked checkbox
	assert.Contains(t, result[1], checkboxStyled)
}

func TestRenderStyledLinesForEmptyFilter_EmptyLines(t *testing.T) {
	config := getStyleConfig()

	lines := []string{}
	rulePath := ""
	checkboxStyled := uncheckedBox
	isSelected := false

	result := renderStyledLinesForEmptyFilter(lines, rulePath, checkboxStyled, isSelected, config)

	// Should return empty slice for empty input
	assert.Empty(t, result)
}

func TestRenderStyledLinesForEmptyFilter_SingleLine(t *testing.T) {
	config := getStyleConfig()

	lines := []string{"Single Title"}
	rulePath := ""
	checkboxStyled := uncheckedBox
	isSelected := false

	result := renderStyledLinesForEmptyFilter(lines, rulePath, checkboxStyled, isSelected, config)

	// Should have 1 styled line
	assert.Len(t, result, 1)
	assert.NotEmpty(t, result[0])
	assert.Contains(t, result[0], checkboxStyled)
	assert.Contains(t, result[0], "Single Title")
}

// Integration test with realistic data
func TestRenderStyledLinesForEmptyFilter_Integration(t *testing.T) {
	config := getStyleConfig()

	// Test with realistic rule data
	lines := []string{
		testRulePath,
		"TypeScript Strict Configuration",
		"Enforces strict TypeScript compiler options for better type safety",
		"Languages: TypeScript • Frameworks: Next.js, React • Tags: typescript, config, strict",
		"Trigger: glob (tsconfig.json, *.ts, *.tsx)",
	}

	rulePath := testRulePath
	checkboxStyled := checkedBox
	isSelected := true

	result := renderStyledLinesForEmptyFilter(lines, rulePath, checkboxStyled, isSelected, config)

	// Should have all 5 lines styled
	assert.Len(t, result, 5)

	// Verify each line is properly styled and not empty
	for i, line := range result {
		assert.NotEmpty(t, line, "Line %d should not be empty", i)
		// Lines should contain the original content somewhere in the styled output
		// (This is a basic check - exact matching is complex due to ANSI codes)
	}

	// Title line should contain checkbox and title
	assert.Contains(t, result[1], checkboxStyled)
}
