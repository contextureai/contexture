// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestDefaultTheme(t *testing.T) {
	t.Parallel()
	theme := DefaultTheme()
	assert.NotNil(t, theme)
	// Test CharmTheme-based colors
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}, theme.Primary)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#235", Dark: "#252"}, theme.Secondary)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}, theme.Success)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"}, theme.Warning)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}, theme.Error)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}, theme.Info)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#6B7280"}, theme.Muted)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#000000"}, theme.Background)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#235", Dark: "#252"}, theme.Foreground)
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"}, theme.Border)
}

func TestStyles(t *testing.T) {
	t.Parallel()
	theme := DefaultTheme()
	styles := NewStyles(theme)
	assert.NotNil(t, styles)

	t.Run("Header", func(t *testing.T) {
		result := styles.Header("Test")
		assert.Contains(t, result, "▶ Test")
	})

	t.Run("Success", func(t *testing.T) {
		result := styles.Success("Test")
		assert.Contains(t, result, "✓ Test")
	})

	t.Run("Warning", func(t *testing.T) {
		result := styles.Warning("Test")
		assert.Contains(t, result, "⚠ Test")
	})

	t.Run("Error", func(t *testing.T) {
		result := styles.Error("Test")
		assert.Contains(t, result, "✗ Test")
	})

	t.Run("Info", func(t *testing.T) {
		result := styles.Info("Test")
		assert.Contains(t, result, "ⓘ Test")
	})

	t.Run("Muted", func(t *testing.T) {
		result := styles.Muted("Test")
		assert.Contains(t, result, "Test")
	})
}

func TestStyleMethods(t *testing.T) {
	t.Parallel()
	theme := DefaultTheme()
	styles := NewStyles(theme)

	tests := []struct {
		name      string
		styleFunc func() lipgloss.Style
	}{
		{"PrimaryStyle", func() lipgloss.Style { return styles.PrimaryStyle() }},
		{"SecondaryStyle", func() lipgloss.Style { return styles.SecondaryStyle() }},
		{"SuccessStyle", func() lipgloss.Style { return styles.SuccessStyle() }},
		{"ErrorStyle", func() lipgloss.Style { return styles.ErrorStyle() }},
		{"BorderStyle", func() lipgloss.Style { return styles.BorderStyle() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tt.styleFunc()
			assert.NotNil(t, style)

			rendered := style.Render("test")
			assert.Contains(t, rendered, "test")
		})
	}
}

func TestSpinnerChars(t *testing.T) {
	t.Parallel()
	// Test that spinner characters are defined
	assert.NotEmpty(t, SpinnerChars)
	assert.Len(t, SpinnerChars, 10)

	// Check that all are different characters
	charMap := make(map[string]bool)
	for _, char := range SpinnerChars {
		assert.False(t, charMap[char], "Duplicate spinner character: %s", char)
		charMap[char] = true
	}
}

func TestHuhIntegration(t *testing.T) {
	t.Parallel()
	// Test that we can create a keymap with ESC enabled inline
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit.SetKeys("ctrl+c", "esc")
	assert.NotNil(t, keymap)
	assert.Contains(t, keymap.Quit.Keys(), "esc")
}

func TestAllComponentsWithTheme(t *testing.T) {
	t.Parallel()
	theme := DefaultTheme()

	t.Run("all_components_render", func(t *testing.T) {
		// Test that all components work with a theme
		banner := NewBanner("Test").WithTheme(theme).Render()
		assert.Contains(t, banner, "Test")

		status := NewStatusIndicator(StatusSuccess, "OK").WithTheme(theme).Render()
		assert.Contains(t, status, "OK")
		assert.Contains(t, status, "✓")

		card := NewCard("Card").WithTheme(theme).Render()
		assert.Contains(t, card, "Card")

		notification := NewNotification(
			NotificationInfo,
			"Info",
			"Message",
		).WithTheme(theme).
			Render()
		assert.Contains(t, notification, "Info")
		assert.Contains(t, notification, "ⓘ")

		divider := NewDivider().WithText("Section").WithTheme(theme).Render()
		assert.Contains(t, divider, "Section")
	})
}
